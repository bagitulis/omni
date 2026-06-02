package services_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	lazada "github.com/omni/backend/pkg/lazada"
	shopee "github.com/omni/backend/pkg/shopee"
)

// TestShopeeGetShopInfo calls Shopee GetShopInfo API and verifies response.
// Credentials are read from environment variables to avoid hardcoding secrets.
func TestShopeeGetShopInfo(t *testing.T) {
	partnerID := getEnvInt64(t, "TEST_SHOPEE_PARTNER_ID")
	partnerKey := getEnvOrSkip(t, "TEST_SHOPEE_PARTNER_KEY")
	shopID := getEnvInt64(t, "TEST_SHOPEE_SHOP_ID")
	accessToken := getEnvOrSkip(t, "TEST_SHOPEE_ACCESS_TOKEN")

	client := shopee.NewClient(partnerID, partnerKey, true) // production
	client.SetShopCredentials(shopID, accessToken)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call GetShopInfo via RawGet (Shopee v2 endpoint)
	var result map[string]interface{}
	err := client.RawGet(ctx, "/api/v2/shop/get_shop_info", nil, &result)
	if err != nil {
		t.Fatalf("Shopee GetShopInfo failed: %v", err)
	}

	// Verify response structure
	t.Logf("Shopee GetShopInfo response keys: %v", getKeys(result))

	if resp, ok := result["response"].(map[string]interface{}); ok {
		shopName, _ := resp["shop_name"].(string)
		t.Logf("Shopee shop_name: %s", shopName)
		if shopName == "" {
			t.Error("Shopee shop_name is empty")
		}
	} else {
		t.Logf("Full response: %v", result)
		// Check if there's an error in the response
		if errStr, ok := result["error"].(string); ok && errStr != "" {
			t.Errorf("Shopee API returned error: %s - %v", errStr, result["message"])
		}
	}
}

// TestLazadaGetSeller calls Lazada GetSeller API and verifies response.
func TestLazadaGetSeller(t *testing.T) {
	appKey := getEnvOrSkip(t, "TEST_LAZADA_APP_KEY")
	appSecret := getEnvOrSkip(t, "TEST_LAZADA_APP_SECRET")
	accessToken := getEnvOrSkip(t, "TEST_LAZADA_ACCESS_TOKEN")

	client := lazada.NewClient(appKey, appSecret, "id") // Indonesia
	client.SetAccessToken(accessToken)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call GetSeller endpoint — RawRequest decodes into nil (response logged only)
	err := client.RawRequest(ctx, "GET", "/seller/get", nil, nil)
	if err != nil {
		t.Fatalf("Lazada GetSeller failed: %v", err)
	}

	t.Log("Lazada GetSeller call succeeded")
}

// TestTikTokGetShopInfo calls TikTok GetShopInfo API and verifies response.
// TikTok uses a different auth mechanism (app_key + app_secret + access_token).
func TestTikTokGetShopInfo(t *testing.T) {
	appKey := getEnvOrSkip(t, "TEST_TIKTOK_APP_KEY")
	appSecret := getEnvOrSkip(t, "TEST_TIKTOK_APP_SECRET")
	accessToken := getEnvOrSkip(t, "TEST_TIKTOK_ACCESS_TOKEN")
	shopCipher := getEnvOrSkip(t, "TEST_TIKTOK_SHOP_CIPHER")

	// TikTok API call using direct HTTP (SDK is complex generated code)
	client := &tiktokHTTPClient{
		appKey:      appKey,
		appSecret:   appSecret,
		accessToken: accessToken,
		shopCipher:  shopCipher,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := client.callAPI(ctx, "POST", "/authorization/202309/shops", nil)
	if err != nil {
		t.Fatalf("TikTok GetShopInfo failed: %v", err)
	}

	t.Logf("TikTok GetShopInfo response keys: %v", getKeys(result))

	if code, ok := result["code"].(float64); ok {
		t.Logf("TikTok response code: %.0f", code)
		if code != 0 {
			t.Errorf("TikTok API returned error code: %.0f - %v", code, result["message"])
		}
	}
	if data, ok := result["data"].(map[string]interface{}); ok {
		shopName, _ := data["shop_name"].(string)
		t.Logf("TikTok shop_name: %s", shopName)
	}
}

// tiktokHTTPClient is a minimal HTTP client for TikTok API calls.
type tiktokHTTPClient struct {
	appKey      string
	appSecret   string
	accessToken string
	shopCipher  string
}

func (c *tiktokHTTPClient) callAPI(ctx context.Context, method, path string, body interface{}) (map[string]interface{}, error) {
	// TikTok API uses HMAC-SHA256 signing similar to Shopee
	// For now, return an error indicating this needs proper implementation
	return nil, fmt.Errorf("tiktok API client not yet implemented - use backend service instead")
}

// Helper functions

func getEnvOrSkip(t *testing.T, key string) string {
	t.Helper()
	val := os.Getenv(key)
	if val == "" {
		t.Skipf("Environment variable %s not set, skipping test", key)
	}
	return val
}

func getEnvInt64(t *testing.T, key string) int64 {
	t.Helper()
	val := getEnvOrSkip(t, key)
	var result int64
	_, err := fmt.Sscanf(val, "%d", &result)
	if err != nil {
		t.Fatalf("Invalid int64 for %s: %v", key, err)
	}
	return result
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
