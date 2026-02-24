package analytics_test

// TikTok Finance: Statement API (batch) vs DB Comparison
//
// Auto-loads & decrypts credentials from DB using ENCRYPTION_KEY env var.
//
// Run from backend/:
//   ENCRYPTION_KEY="gvhe61ysa+PqTfD+pljdqBOmprAt9t8ELF8+JCjgN8Q=" \
//   go test ./internal/services/analytics/ -run TestFinanceStatementCompare -v \
//     -month=11 -year=2025
//
// Output: tiktok_compare_11_2025.csv in working directory

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/omni/backend/internal/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	fMonth  = flag.Int("month", 11, "Month (1-12)")
	fYear   = flag.Int("year", 2025, "Year")
	fTenant = flag.String("tenant", "yumna_bertigamart", "Tenant ID")
	fDSN    = flag.String("dsn", "postgres://omni:omni@localhost:5432/omni_main?sslmode=disable", "Postgres DSN")
)

// ─── API types ────────────────────────────────────────────────────────────────

type stmtListResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Statements    []struct{ StatementID string `json:"statement_id"` } `json:"statements"`
		NextPageToken string                                               `json:"next_page_token"`
	} `json:"data"`
}

type stmtTxResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Transactions []struct {
			OrderID          string `json:"order_id"`
			Type             string `json:"type"`
			SettlementAmount string `json:"settlement_amount"`
			RevenueAmount    string `json:"revenue_amount"`
		} `json:"transactions"`
		NextPageToken string `json:"next_page_token"`
	} `json:"data"`
}

// ─── API client ───────────────────────────────────────────────────────────────

const tiktokBaseURL = "https://open-api.tiktokglobalshop.com"

type apiCreds struct {
	appKey, appSecret, token, cipher string
}

func (c *apiCreds) sign(path string, params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && k != "access_token" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(params[k])
	}
	raw := c.appSecret + path + sb.String() + c.appSecret
	h := hmac.New(sha256.New, []byte(c.appSecret))
	h.Write([]byte(raw))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *apiCreds) get(path string, params map[string]string, out interface{}) error {
	params["app_key"] = c.appKey
	params["timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	params["access_token"] = c.token
	if c.cipher != "" {
		params["shop_cipher"] = c.cipher
	}
	params["sign"] = c.sign(path, params)

	u, _ := url.Parse(tiktokBaseURL + path)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("x-tts-access-token", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return json.Unmarshal(body, out)
}

// ─── DB helpers ───────────────────────────────────────────────────────────────

type confRow struct {
	ConfigKey   string `gorm:"column:config_key"`
	ConfigValue string `gorm:"column:config_value"`
	IsEncrypted bool   `gorm:"column:is_encrypted"`
}

type dbOrderRow struct {
	OrderID               string  `gorm:"column:order_id"`
	TotalSettlementAmount float64 `gorm:"column:total_settlement_amount"`
}

func openDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

func maybeDecrypt(enc *utils.EncryptionService, value string, encrypted bool) string {
	if !encrypted || enc == nil || value == "" {
		return value
	}
	plain, err := enc.Decrypt(value)
	if err != nil {
		return value // fallback to raw
	}
	return plain
}

// ─── Test ─────────────────────────────────────────────────────────────────────

func TestFinanceStatementCompare(t *testing.T) {
	// This test requires a live PostgreSQL connection and real TikTok API credentials.
	// It is intended for manual debugging only. Set FINANCE_COMPARE_TEST=1 to run.
	// JIRA: N/A — manual integration test, not suitable for CI
	if os.Getenv("FINANCE_COMPARE_TEST") != "1" {
		t.Skip("Skipping manual integration test: set FINANCE_COMPARE_TEST=1 to run")
	}
	month, year, tenant := *fMonth, *fYear, *fTenant

	t.Logf("=== TikTok Finance Statement API vs DB ===")
	t.Logf("Period: %02d/%d  Tenant: %s", month, year, tenant)

	// Encryption service (from ENCRYPTION_KEY env)
	encKey := os.Getenv("ENCRYPTION_KEY")
	var encSvc *utils.EncryptionService
	if encKey != "" {
		var err error
		encSvc, err = utils.NewEncryptionService(encKey)
		if err != nil {
			t.Fatalf("NewEncryptionService: %v", err)
		}
		t.Log("Encryption service ready ✅")
	} else {
		t.Log("ENCRYPTION_KEY not set — encrypted values will not be decrypted")
	}

	db, err := openDB(*fDSN)
	if err != nil {
		t.Fatalf("DB connect: %v", err)
	}

	// 1. Load app credentials (appKey, appSecret) from system schema
	creds := &apiCreds{}
	var globalCfgs []confRow
	if err := db.Raw(`SELECT config_key, config_value, is_encrypted FROM system.global_config WHERE platform = 'tiktok'`).Scan(&globalCfgs).Error; err != nil {
		t.Fatalf("global_config: %v", err)
	}
	for _, c := range globalCfgs {
		val := maybeDecrypt(encSvc, c.ConfigValue, c.IsEncrypted)
		switch c.ConfigKey {
		case "appKey":
			creds.appKey = val
		case "appSecret":
			creds.appSecret = val
		}
	}
	if creds.appKey == "" || creds.appSecret == "" {
		t.Fatal("appKey/appSecret not found in system.global_config for platform=tiktok")
	}
	t.Logf("AppKey: %s", creds.appKey)

	// 2. Load tenant credentials (accessToken, shopCipher) from tenant schema
	var tenantCfgs []confRow
	if err := db.Raw(
		fmt.Sprintf(`SELECT config_key, config_value, is_encrypted FROM tenant_%s.platform_configs WHERE platform = 'tiktok'`, tenant),
	).Scan(&tenantCfgs).Error; err != nil {
		t.Fatalf("platform_configs: %v", err)
	}
	for _, c := range tenantCfgs {
		val := maybeDecrypt(encSvc, c.ConfigValue, c.IsEncrypted)
		switch c.ConfigKey {
		case "accessToken":
			creds.token = val
		case "shopCipher":
			creds.cipher = val
		}
	}
	if creds.token == "" {
		t.Fatal("accessToken not found or could not be decrypted — ensure ENCRYPTION_KEY is correct")
	}
	t.Logf("Token loaded (len=%d), Cipher loaded (len=%d)", len(creds.token), len(creds.cipher))

	// 3. Load DB orders for the month
	var dbRows []dbOrderRow
	if err := db.Raw(
		fmt.Sprintf(`SELECT order_id, total_settlement_amount FROM tenant_%s.tiktok_escrow_orders WHERE tenant_id = ? AND month = ? AND year = ?`, tenant),
		tenant, month, year,
	).Scan(&dbRows).Error; err != nil {
		t.Fatalf("load DB orders: %v", err)
	}
	dbSettle := make(map[string]float64, len(dbRows))
	for _, r := range dbRows {
		dbSettle[r.OrderID] = r.TotalSettlementAmount
	}
	t.Logf("DB: %d orders", len(dbSettle))

	// 4. Fetch statement IDs for the period
	startTime := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endTime := startTime.AddDate(0, 1, 0)
	t.Logf("Statements: %s → %s", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))

	var stmtIDs []string
	for page := ""; ; {
		params := map[string]string{
			"sort_field":        "statement_time",
			"sort_order":        "ASC",
			"statement_time_ge": strconv.FormatInt(startTime.Unix(), 10),
			"statement_time_lt": strconv.FormatInt(endTime.Unix(), 10),
			"page_size":         "100",
		}
		if page != "" {
			params["page_token"] = page
		}
		var resp stmtListResp
		if err := creds.get("/finance/202309/statements", params, &resp); err != nil {
			t.Fatalf("statements API: %v", err)
		}
		if resp.Code != 0 {
			t.Fatalf("statements API code=%d: %s", resp.Code, resp.Message)
		}
		for _, s := range resp.Data.Statements {
			stmtIDs = append(stmtIDs, s.StatementID)
		}
		t.Logf("  batch %d → total %d statements", len(resp.Data.Statements), len(stmtIDs))
		if resp.Data.NextPageToken == "" {
			break
		}
		page = resp.Data.NextPageToken
		time.Sleep(200 * time.Millisecond)
	}
	t.Logf("Total statements: %d", len(stmtIDs))

	// 5. Fetch transactions per statement via v202501 (100/page)
	t.Log("Fetching transactions via Statement API v202501...")
	apiSettle := make(map[string]float64)
	apiRevenue := make(map[string]float64)

	for i, id := range stmtIDs {
		for page := ""; ; {
			params := map[string]string{
				"sort_field": "order_create_time",
				"sort_order": "ASC",
				"page_size":  "100",
			}
			if page != "" {
				params["page_token"] = page
			}
			var resp stmtTxResp
			if err := creds.get(fmt.Sprintf("/finance/202501/statements/%s/statement_transactions", id), params, &resp); err != nil {
				t.Logf("[WARN] stmt %s: %v", id, err)
				break
			}
			if resp.Code != 0 {
				t.Logf("[WARN] stmt %s code=%d: %s", id, resp.Code, resp.Message)
				break
			}
			for _, tx := range resp.Data.Transactions {
				if tx.OrderID == "" || tx.Type != "ORDER" {
					continue
				}
				amt, _ := strconv.ParseFloat(tx.SettlementAmount, 64)
				rev, _ := strconv.ParseFloat(tx.RevenueAmount, 64)
				apiSettle[tx.OrderID] += amt
				apiRevenue[tx.OrderID] += rev
			}
			if resp.Data.NextPageToken == "" {
				break
			}
			page = resp.Data.NextPageToken
			time.Sleep(100 * time.Millisecond)
		}
		t.Logf("[%d/%d] stmt %s", i+1, len(stmtIDs), id)
		time.Sleep(150 * time.Millisecond)
	}
	t.Logf("API: settlement for %d orders", len(apiSettle))

	// 6. Write CSV & compare
	outFile := fmt.Sprintf("tiktok_compare_%02d_%d.csv", month, year)
	f, err := os.Create(outFile)
	if err != nil {
		t.Fatalf("create CSV: %v", err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"order_id", "db_settlement", "api_settlement", "api_revenue", "diff", "status"})

	matched, mismatched, dbOnly, apiOnly := 0, 0, 0, 0

	for orderID, dbAmt := range dbSettle {
		apiAmt := apiSettle[orderID]
		apiRev := apiRevenue[orderID]
		_, found := apiSettle[orderID]

		var status string
		switch {
		case !found:
			dbOnly++
			status = "DB_ONLY"
		case math.Abs(dbAmt-apiAmt) < 1.0:
			matched++
			status = "MATCH"
		default:
			mismatched++
			status = fmt.Sprintf("MISMATCH_%.2f", math.Abs(dbAmt-apiAmt))
		}
		_ = w.Write([]string{orderID, fmt.Sprintf("%.2f", dbAmt), fmt.Sprintf("%.2f", apiAmt), fmt.Sprintf("%.2f", apiRev), fmt.Sprintf("%.2f", apiAmt-dbAmt), status})
	}

	for orderID, apiAmt := range apiSettle {
		if _, inDB := dbSettle[orderID]; !inDB {
			apiOnly++
			_ = w.Write([]string{orderID, "0.00", fmt.Sprintf("%.2f", apiAmt), fmt.Sprintf("%.2f", apiRevenue[orderID]), fmt.Sprintf("%.2f", apiAmt), "API_ONLY"})
		}
	}
	w.Flush()

	// 7. Summary
	t.Logf("CSV saved: %s", outFile)
	t.Logf("─────────────────────────────────────────")
	t.Logf("DB orders:   %d", len(dbSettle))
	t.Logf("API orders:  %d", len(apiSettle))
	t.Logf("✅ MATCH:     %d", matched)
	t.Logf("❌ MISMATCH:  %d", mismatched)
	t.Logf("DB only:     %d  (belum settled atau refund)", dbOnly)
	t.Logf("API only:    %d  (settled tapi tidak ada di DB sync)", apiOnly)

	if mismatched > 0 {
		t.Errorf("%d mismatches ditemukan — cek %s", mismatched, outFile)
	} else {
		t.Log("✅ Statement API CONSISTENT — data akurat!")
	}
}
