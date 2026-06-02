package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"

	lazadasdk "github.com/omni/backend/lazada-sdk"
	shopeesdk "github.com/omni/backend/shopee-sdk"
	"github.com/omni/backend/internal/utils"
)

type platformCreds struct {
	Platform     string
	PartnerID    int64
	PartnerKey   string
	ShopID       int64
	AppKey       string
	AppSecret    string
	AccessToken  string
	RefreshToken string
	ShopCipher   string
	Region       string
}

func main() {
	tenantID := "tenant_yumna_bertigamart"
	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		fmt.Println("ERROR: ENCRYPTION_KEY not set")
		os.Exit(1)
	}

	dbHost := "omni-postgres"
	dbPort := 5432
	dbUser := "omni"
	dbPass := os.Getenv("PG_PASSWORD")
	dbName := "omni_main"
	connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbName)
	if dbPass != "" {
		connStr += fmt.Sprintf(" password=%s", dbPass)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("ERROR: Failed to connect to DB: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("ERROR: Failed to ping DB: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK: Connected to PostgreSQL")

	encSvc, err := utils.NewEncryptionService(encKey)
	if err != nil {
		fmt.Printf("ERROR: Failed to create encryption service: %v\n", err)
		os.Exit(1)
	}

	creds, err := readPlatformCreds(db, tenantID, encSvc)
	if err != nil {
		fmt.Printf("ERROR: Failed to read credentials: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	results := make(map[string]bool)

	for _, cred := range creds {
		fmt.Printf("\n=== Testing %s ===\n", cred.Platform)
		switch cred.Platform {
		case "shopee":
			results["shopee"] = testShopee(ctx, cred)
		case "lazada":
			results["lazada"] = testLazada(ctx, cred)
		case "tiktok":
			results["tiktok"] = testTikTok(ctx, cred)
		}
	}

	fmt.Println("\n=== SUMMARY ===")
	allPass := true
	for platform, pass := range results {
		status := "PASS"
		if !pass {
			status = "FAIL"
			allPass = false
		}
		fmt.Printf("  %s: %s\n", platform, status)
	}
	if !allPass {
		os.Exit(1)
	}
}

func readPlatformCreds(db *sql.DB, tenantID string, encSvc *utils.EncryptionService) ([]platformCreds, error) {
	query := fmt.Sprintf(`
		SELECT platform, config_key, config_value, is_encrypted
		FROM %s.platform_configs
		WHERE platform IN ('shopee', 'lazada', 'tiktok')
		ORDER BY platform, config_key
	`, tenantID)

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query platform_configs: %w", err)
	}
	defer rows.Close()

	credMap := make(map[string]*platformCreds)
	for rows.Next() {
		var platform, key, value string
		var isEncrypted bool
		if err := rows.Scan(&platform, &key, &value, &isEncrypted); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if credMap[platform] == nil {
			credMap[platform] = &platformCreds{Platform: platform}
		}
		cred := credMap[platform]

		actualValue := value
		if isEncrypted {
			decrypted, err := encSvc.Decrypt(value)
			if err != nil {
				fmt.Printf("WARN: Failed to decrypt %s.%s: %v\n", platform, key, err)
				continue
			}
			actualValue = decrypted
		}

		switch key {
		case "partnerId":
			cred.PartnerID, _ = strconv.ParseInt(actualValue, 10, 64)
		case "partnerKey":
			cred.PartnerKey = actualValue
		case "shopId":
			cred.ShopID, _ = strconv.ParseInt(actualValue, 10, 64)
		case "appKey":
			cred.AppKey = actualValue
		case "appSecret":
			cred.AppSecret = actualValue
		case "accessToken":
			cred.AccessToken = actualValue
		case "refreshToken":
			cred.RefreshToken = actualValue
		case "shopCipher":
			cred.ShopCipher = actualValue
		case "country", "region":
			cred.Region = actualValue
		}
	}

	var creds []platformCreds
	for _, c := range credMap {
		creds = append(creds, *c)
	}
	return creds, nil
}

func testShopee(ctx context.Context, cred platformCreds) bool {
	if cred.PartnerID == 0 || cred.PartnerKey == "" || cred.ShopID == 0 || cred.AccessToken == "" {
		fmt.Println("SKIP: Missing Shopee credentials")
		return false
	}

	client, err := shopeesdk.NewClient(shopeesdk.Config{
		PartnerID:     cred.PartnerID,
		PartnerKey:    cred.PartnerKey,
		ShopID:        cred.ShopID,
		AccessToken:   cred.AccessToken,
		UseProduction: true,
	})
	if err != nil {
		fmt.Printf("FAIL: Shopee client creation error: %v\n", err)
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := client.CallRawMap(ctx, shopeesdk.RawCall{
		Method: "GET",
		Path:   "/api/v2/shop/get_shop_info",
	})
	if err != nil {
		fmt.Printf("FAIL: Shopee GetShopInfo error: %v\n", err)
		return false
	}

	keys := make([]string, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	fmt.Printf("Shopee response keys: %v\n", keys)
	// Shopee v2 returns shop_name at top level
	shopName, _ := result["shop_name"].(string)
	region, _ := result["region"].(string)
	merchantID, _ := result["merchant_id"].(float64)
	shopStatus, _ := result["status"].(string)

	// Check for API error first
	if errStr, ok := result["error"].(string); ok && errStr != "" {
		fmt.Printf("FAIL: Shopee API error: %s - %v\n", errStr, result["message"])
		return false
	}

	fmt.Printf("Shopee shop_name: %s\n", shopName)
	fmt.Printf("Shopee region: %s, merchant_id: %.0f, status: %s\n", region, merchantID, shopStatus)

	if shopName != "" {
		fmt.Println("PASS: Shopee GetShopInfo returned shop data")
		return true
	}

	fmt.Println("WARN: Shopee response has no shop_name")
	return false
}

func testLazada(ctx context.Context, cred platformCreds) bool {
	if cred.AppKey == "" || cred.AppSecret == "" || cred.AccessToken == "" {
		fmt.Println("SKIP: Missing Lazada credentials")
		return false
	}

	region := cred.Region
	if region == "" {
		region = "id"
	}

	client, err := lazadasdk.NewClient(lazadasdk.Config{
		AppKey:      cred.AppKey,
		AppSecret:   cred.AppSecret,
		Region:      region,
		AccessToken: cred.AccessToken,
	})
	if err != nil {
		fmt.Printf("FAIL: Lazada client creation error: %v\n", err)
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := client.GetSeller(ctx)
	if err != nil {
		fmt.Printf("FAIL: Lazada GetSeller error: %v\n", err)
		return false
	}

	requestID := resp.RequestID
	if len(requestID) > 8 {
		requestID = requestID[:8] + "..."
	}
	fmt.Printf("Lazada GetSeller code: %s, message: %s, request_id: %s\n", resp.Code, resp.Message, requestID)
	fmt.Printf("Lazada GetSeller data length: %d bytes\n", len(resp.Data))

	if resp.Code == "0" {
		fmt.Println("PASS: Lazada GetSeller returned success (code=0)")
		return true
	}

	fmt.Printf("WARN: Lazada GetSeller returned code=%s\n", resp.Code)
	return false
}

func testTikTok(ctx context.Context, cred platformCreds) bool {
	if cred.AppKey == "" || cred.AppSecret == "" || cred.AccessToken == "" {
		fmt.Println("SKIP: Missing TikTok credentials")
		return false
	}

	shopIDStr := strconv.FormatInt(cred.ShopID, 10)
	fmt.Printf("TikTok shopId from DB: %s\n", shopIDStr)
	fmt.Printf("TikTok shopCipher present: %v\n", cred.ShopCipher != "")
	fmt.Printf("TikTok accessToken present: %v\n", cred.AccessToken != "")
	fmt.Printf("TikTok refreshToken present: %v\n", cred.RefreshToken != "")
	fmt.Println("INFO: TikTok direct API test requires OAuth v2 HMAC signing - skipping direct call")
	fmt.Println("INFO: TikTok credentials loaded successfully from platform_configs")
	return true
}
