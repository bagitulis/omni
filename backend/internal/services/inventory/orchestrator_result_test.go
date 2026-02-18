package inventory

import "testing"

func TestSummarizeStockResults(t *testing.T) {
	t.Run("returns success when one platform succeeds", func(t *testing.T) {
		success, errors := summarizeStockResults(map[string]*PlatformStockResult{
			"shopee": {Success: true},
			"lazada": {Success: false, Error: "lazada API error [code=500]: upstream timeout"},
		})

		if !success {
			t.Fatalf("expected success to be true when at least one platform succeeds")
		}
		if len(errors) != 1 || errors[0] != "lazada API error [code=500]: upstream timeout" {
			t.Fatalf("unexpected errors: %#v", errors)
		}
	})

	t.Run("returns failure when all platforms are not found", func(t *testing.T) {
		success, errors := summarizeStockResults(map[string]*PlatformStockResult{
			"shopee": {Success: false, Error: "SKU not found in Shopee"},
			"lazada": {Success: false, Error: "SKU not found in Lazada"},
		})

		if success {
			t.Fatalf("expected success to be false when no platform succeeded")
		}
		if len(errors) != 0 {
			t.Fatalf("expected no aggregated errors for not-found only results, got %#v", errors)
		}
	})
}

func TestSummarizePriceResults(t *testing.T) {
	t.Run("returns success when one platform succeeds", func(t *testing.T) {
		success, errors := summarizePriceResults(map[string]*PlatformPriceResult{
			"tiktok": {Success: true},
			"shopee": {Success: false, Error: "shopee API error [code=E3001]: invalid price"},
		})

		if !success {
			t.Fatalf("expected success to be true when at least one platform succeeds")
		}
		if len(errors) != 1 || errors[0] != "shopee API error [code=E3001]: invalid price" {
			t.Fatalf("unexpected errors: %#v", errors)
		}
	})

	t.Run("returns failure for unsupported platform error", func(t *testing.T) {
		success, errors := summarizePriceResults(map[string]*PlatformPriceResult{
			"custom": {Success: false, Error: "unsupported platform: custom"},
		})

		if success {
			t.Fatalf("expected success to be false when all platform updates fail")
		}
		if len(errors) != 1 || errors[0] != "unsupported platform: custom" {
			t.Fatalf("unexpected errors: %#v", errors)
		}
	})
}

func TestRawPlatformErrorFormatting(t *testing.T) {
	t.Run("shopee keeps raw code and message", func(t *testing.T) {
		errMsg := rawShopeeAPIError("E123", "invalid token")
		if errMsg != "E123: invalid token" {
			t.Fatalf("unexpected shopee raw error: %s", errMsg)
		}
	})

	t.Run("lazada keeps raw code and message", func(t *testing.T) {
		errMsg := rawLazadaAPIError("1000", "Invalid seller sku")
		if errMsg != "code=1000: Invalid seller sku" {
			t.Fatalf("unexpected lazada raw error: %s", errMsg)
		}
	})

	t.Run("tiktok keeps raw code and message", func(t *testing.T) {
		errMsg := rawTiktokAPIError(50001, "invalid parameter")
		if errMsg != "code=50001: invalid parameter" {
			t.Fatalf("unexpected tiktok raw error: %s", errMsg)
		}
	})
}
