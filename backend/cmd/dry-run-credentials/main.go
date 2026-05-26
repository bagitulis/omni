package main

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "omni_secure_2026")
	dbname := getEnv("DB_NAME", "omni_main")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening connection: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to %s:%s/%s: %v\n", host, port, dbname, err)
		os.Exit(1)
	}

	tenants, err := db.Query(`
		SELECT tenant_id, COALESCE(db_schema, 'tenant_' || tenant_id) AS schema_name
		FROM system.tenants
		WHERE is_active = true
		ORDER BY tenant_id
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying system.tenants: %v\n", err)
		os.Exit(1)
	}
	defer tenants.Close()

	type rowInfo struct {
		TenantID string
		Platform string
		RowShape string
		Count    int
	}
	var results []rowInfo

	for tenants.Next() {
		var tenantID, schema string
		if err := tenants.Scan(&tenantID, &schema); err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning tenant row: %v\n", err)
			continue
		}

		hasConfigKey, hasTenantID, err := detectColumns(db, schema)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tenant=%s: skipping — %v\n", tenantID, err)
			continue
		}

		shape := classifyShape(hasConfigKey, hasTenantID)
		if shape == "" {
			fmt.Fprintf(os.Stderr, "tenant=%s: unrecognized platform_configs structure\n", tenantID)
			continue
		}

		rows, err := db.Query(fmt.Sprintf(
			`SELECT COALESCE(platform, ''), COUNT(*) FROM %s.platform_configs GROUP BY platform ORDER BY platform`,
			quoteIdent(schema),
		))
		if err != nil {
			fmt.Fprintf(os.Stderr, "tenant=%s: error querying platform_configs: %v\n", tenantID, err)
			continue
		}

		for rows.Next() {
			var platform string
			var count int
			if err := rows.Scan(&platform, &count); err != nil {
				fmt.Fprintf(os.Stderr, "tenant=%s: error scanning row: %v\n", tenantID, err)
				continue
			}
			results = append(results, rowInfo{
				TenantID: tenantID,
				Platform: platform,
				RowShape: shape,
				Count:    count,
			})
		}
		rows.Close()
	}

	sort.Slice(results, func(i, j int) bool {
		a, b := results[i], results[j]
		if a.TenantID != b.TenantID {
			return a.TenantID < b.TenantID
		}
		if a.Platform != b.Platform {
			return a.Platform < b.Platform
		}
		return a.RowShape < b.RowShape
	})

	fmt.Println("tenant_id,platform,row_shape,count")
	for _, r := range results {
		fmt.Printf("%s,%s,%s,%d\n", r.TenantID, r.Platform, r.RowShape, r.Count)
	}
}

// detectColumns checks which columns exist in the tenant's platform_configs table.
func detectColumns(db *sql.DB, schema string) (hasConfigKey, hasTenantID bool, err error) {
	rows, err := db.Query(`
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = 'platform_configs'
	`, schema)
	if err != nil {
		return false, false, fmt.Errorf("cannot inspect platform_configs columns: %w", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			continue
		}
		found = true
		switch col {
		case "config_key":
			hasConfigKey = true
		case "tenant_id":
			hasTenantID = true
		}
	}
	if !found {
		return false, false, fmt.Errorf("platform_configs table not found or empty in schema %s", schema)
	}
	return hasConfigKey, hasTenantID, nil
}

// classifyShape returns "key-value" if config_key exists, "structured" if tenant_id exists, else "".
func classifyShape(hasConfigKey, hasTenantID bool) string {
	if hasConfigKey {
		return "key-value"
	}
	if hasTenantID {
		return "structured"
	}
	return ""
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
