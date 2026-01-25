package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Migrator handles the migration process
type Migrator struct {
	config   MigrationConfig
	sqliteDB *sql.DB
	pgDB     *sql.DB
	report   *MigrationReport
}

// NewMigrator creates a new Migrator instance
func NewMigrator(config MigrationConfig) (*Migrator, error) {
	return &Migrator{
		config: config,
		report: &MigrationReport{
			StartTime: time.Now(),
			TenantID:  config.TenantID,
			Stats:     []MigrationStats{},
		},
	}, nil
}

// Connect establishes connections to both databases
func (m *Migrator) Connect() error {
	var err error

	m.sqliteDB, err = sql.Open("sqlite3", m.config.SQLitePath)
	if err != nil {
		return fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	m.pgDB, err = sql.Open("postgres", m.config.PostgresDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := m.sqliteDB.Ping(); err != nil {
		return fmt.Errorf("SQLite ping failed: %w", err)
	}
	if err := m.pgDB.Ping(); err != nil {
		return fmt.Errorf("PostgreSQL ping failed: %w", err)
	}

	log.Println("✅ Connected to both databases")
	return nil
}

// Close closes database connections
func (m *Migrator) Close() {
	if m.sqliteDB != nil {
		m.sqliteDB.Close()
	}
	if m.pgDB != nil {
		m.pgDB.Close()
	}
}

// SetPostgresSchema sets the PostgreSQL search path
func (m *Migrator) SetPostgresSchema(schema string) error {
	_, err := m.pgDB.Exec(fmt.Sprintf("SET search_path TO %s, public", schema))
	if err != nil {
		return fmt.Errorf("failed to set schema: %w", err)
	}
	log.Printf("📂 Using PostgreSQL schema: %s", schema)
	return nil
}

// MigrateTable migrates a single table
func (m *Migrator) MigrateTable(mapping TableMapping) (*MigrationStats, error) {
	startTime := time.Now()
	stats := &MigrationStats{TableName: mapping.SQLiteTable, Errors: []string{}}

	count, err := m.getSourceRowCount(mapping.SQLiteTable)
	if err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("Table not found: %v", err))
		return stats, nil
	}
	stats.SourceRows = count

	if count == 0 {
		log.Printf("⏭️  %s: No rows to migrate", mapping.SQLiteTable)
		return stats, nil
	}

	log.Printf("📤 Migrating %s: %d rows", mapping.SQLiteTable, count)

	if err := m.migrateRows(mapping, stats); err != nil {
		return stats, err
	}

	stats.Duration = time.Since(startTime)
	log.Printf("✅ %s: Migrated %d rows (%.2fs)", mapping.SQLiteTable, stats.MigratedRows, stats.Duration.Seconds())
	return stats, nil
}

func (m *Migrator) getSourceRowCount(tableName string) (int64, error) {
	var count int64
	err := m.sqliteDB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	return count, err
}

func (m *Migrator) migrateRows(mapping TableMapping, stats *MigrationStats) error {
	sqliteCols, pgCols := m.buildColumnLists(mapping)

	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(sqliteCols, ", "), mapping.SQLiteTable)
	rows, err := m.sqliteDB.Query(query)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	insertSQL := m.buildInsertSQL(mapping.PostgresTable, pgCols)

	tx, err := m.pgDB.Begin()
	if err != nil {
		return fmt.Errorf("transaction begin failed: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("prepare statement failed: %w", err)
	}
	defer stmt.Close()

	batch := 0
	for rows.Next() {
		values := make([]interface{}, len(sqliteCols))
		valuePtrs := make([]interface{}, len(sqliteCols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			stats.ErrorRows++
			stats.Errors = append(stats.Errors, fmt.Sprintf("Scan error: %v", err))
			if !m.config.ContinueOnError {
				return err
			}
			continue
		}

		m.applyTransforms(mapping.Columns, values)

		if !m.config.DryRun {
			_, err := stmt.Exec(values...)
			if err != nil {
				stats.ErrorRows++
				if len(stats.Errors) < 10 {
					stats.Errors = append(stats.Errors, fmt.Sprintf("Insert error: %v", err))
				}
				if !m.config.ContinueOnError {
					return fmt.Errorf("insert error: %w", err)
				}
				continue
			}
		}

		stats.MigratedRows++
		batch++

		if batch%1000 == 0 {
			log.Printf("   Progress: %d/%d rows", batch, stats.SourceRows)
		}
	}

	if !m.config.DryRun {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit failed: %w", err)
		}
	}

	return nil
}

func (m *Migrator) buildColumnLists(mapping TableMapping) ([]string, []string) {
	var sqliteCols, pgCols []string
	for _, col := range mapping.Columns {
		sqliteCols = append(sqliteCols, col.SQLiteCol)
		pgCols = append(pgCols, col.PostgresCol)
	}
	return sqliteCols, pgCols
}

func (m *Migrator) buildInsertSQL(tableName string, cols []string) string {
	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT DO NOTHING",
		tableName, strings.Join(cols, ", "), strings.Join(placeholders, ", "),
	)
}

func (m *Migrator) applyTransforms(columns []ColumnMapping, values []interface{}) {
	for i, col := range columns {
		if col.Transform != nil {
			values[i] = col.Transform(values[i])
		}
	}
}

// MigrateAllTables migrates all tables defined in mappings
func (m *Migrator) MigrateAllTables() error {
	mappings := getTableMappings()
	m.report.TotalTables = len(mappings)

	for _, mapping := range mappings {
		stats, err := m.MigrateTable(mapping)
		if err != nil {
			m.report.FailedTables++
			stats.Errors = append(stats.Errors, err.Error())
		} else {
			m.report.SuccessTables++
		}
		m.report.Stats = append(m.report.Stats, *stats)
	}

	m.report.EndTime = time.Now()
	return nil
}

// VerifyMigration verifies the migration by comparing row counts
func (m *Migrator) VerifyMigration() error {
	log.Println("🔍 Verifying migration...")

	mappings := getTableMappings()
	var errors []string

	for _, mapping := range mappings {
		var sqliteCount int64
		err := m.sqliteDB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", mapping.SQLiteTable)).Scan(&sqliteCount)
		if err != nil {
			continue
		}

		var pgCount int64
		err = m.pgDB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", mapping.PostgresTable)).Scan(&pgCount)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: PostgreSQL query failed: %v", mapping.PostgresTable, err))
			continue
		}

		if sqliteCount != pgCount {
			errors = append(errors, fmt.Sprintf("%s: Row count mismatch (SQLite: %d, PostgreSQL: %d)",
				mapping.SQLiteTable, sqliteCount, pgCount))
		} else {
			log.Printf("✅ %s: %d rows verified", mapping.SQLiteTable, sqliteCount)
		}
	}

	if len(errors) > 0 {
		log.Println("❌ Verification errors:")
		for _, err := range errors {
			log.Printf("   - %s", err)
		}
		return fmt.Errorf("verification failed with %d errors", len(errors))
	}

	log.Println("✅ All tables verified successfully!")
	return nil
}

// PrintReport prints the migration report
func (m *Migrator) PrintReport() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 MIGRATION REPORT")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Tenant:       %s\n", m.report.TenantID)
	fmt.Printf("Duration:     %s\n", m.report.EndTime.Sub(m.report.StartTime))
	fmt.Printf("Tables:       %d total, %d success, %d failed\n",
		m.report.TotalTables, m.report.SuccessTables, m.report.FailedTables)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%-25s %10s %10s %10s\n", "TABLE", "SOURCE", "MIGRATED", "ERRORS")
	fmt.Println(strings.Repeat("-", 60))

	var totalSource, totalMigrated, totalErrors int64
	for _, stat := range m.report.Stats {
		fmt.Printf("%-25s %10d %10d %10d\n",
			stat.TableName, stat.SourceRows, stat.MigratedRows, stat.ErrorRows)
		totalSource += stat.SourceRows
		totalMigrated += stat.MigratedRows
		totalErrors += stat.ErrorRows
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%-25s %10d %10d %10d\n", "TOTAL", totalSource, totalMigrated, totalErrors)
	fmt.Println(strings.Repeat("=", 60))
}
