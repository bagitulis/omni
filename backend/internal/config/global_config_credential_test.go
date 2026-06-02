package config

import (
	"os"
	"testing"
)

// RED PHASE tests for removing os.Getenv fallback from credential methods.
//
// These tests verify that GetShopeeCredentials / GetTiktokCredentials / GetLazadaCredentials
// return ONLY database values and do NOT fall back to environment variables.
//
// Current implementation (RED): env fallback is still active -> these tests FAIL.
// Expected implementation (GREEN): env fallback removed -> these tests PASS.

func TestGetShopeeCredentials_DBOnly_NoEnvFallback(t *testing.T) {
	// Set env vars to WRONG values that should NOT be returned
	envPartnerID := "WRONG_ENV_PARTNER_ID_SHOULD_NOT_BE_USED"
	envPartnerKey := "WRONG_ENV_PARTNER_KEY_SHOULD_NOT_BE_USED"
	envPushKey := "WRONG_ENV_PUSH_KEY_SHOULD_NOT_BE_USED"

	os.Setenv("SHOPEE_PARTNER_ID", envPartnerID)
	defer os.Unsetenv("SHOPEE_PARTNER_ID")
	os.Setenv("SHOPEE_PARTNER_KEY", envPartnerKey)
	defer os.Unsetenv("SHOPEE_PARTNER_KEY")
	os.Setenv("SHOPEE_PUSH_PARTNER_KEY", envPushKey)
	defer os.Unsetenv("SHOPEE_PUSH_PARTNER_KEY")

	// Create service with nil DB — no database available
	// GetConfig will return "" because getDB() fails (no PostgreSQL)
	svc := &GlobalConfigService{}

	creds, err := svc.GetShopeeCredentials()
	if err != nil {
		t.Fatalf("GetShopeeCredentials returned error: %v", err)
	}

	// RED: env fallback is still active — returns env values instead of DB-only
	// These assertions FAIL because env fallback pollutes the result
	if creds.PartnerID == envPartnerID {
		t.Errorf("RED (must fix): GetShopeeCredentials returned env SHOPEE_PARTNER_ID ('%s') instead of DB-only value", creds.PartnerID)
	}
	if creds.PartnerKey == envPartnerKey {
		t.Errorf("RED (must fix): GetShopeeCredentials returned env SHOPEE_PARTNER_KEY ('%s') instead of DB-only value", creds.PartnerKey)
	}
	if creds.PushPartnerKey == envPushKey {
		t.Errorf("RED (must fix): GetShopeeCredentials returned env SHOPEE_PUSH_PARTNER_KEY ('%s') instead of DB-only value", creds.PushPartnerKey)
	}
}

func TestGetLazadaCredentials_DBOnly_NoEnvFallback(t *testing.T) {
	// Set env vars to WRONG values that should NOT be returned
	envAppKey := "WRONG_ENV_LAZADA_APP_KEY_SHOULD_NOT_BE_USED"
	envAppSecret := "WRONG_ENV_LAZADA_APP_SECRET_SHOULD_NOT_BE_USED"

	os.Setenv("LAZADA_APP_KEY", envAppKey)
	defer os.Unsetenv("LAZADA_APP_KEY")
	os.Setenv("LAZADA_APP_SECRET", envAppSecret)
	defer os.Unsetenv("LAZADA_APP_SECRET")

	svc := &GlobalConfigService{}

	creds, err := svc.GetLazadaCredentials()
	if err != nil {
		t.Fatalf("GetLazadaCredentials returned error: %v", err)
	}

	// RED: env fallback is still active
	if creds.AppKey == envAppKey {
		t.Errorf("RED (must fix): GetLazadaCredentials returned env LAZADA_APP_KEY ('%s') instead of DB-only value", creds.AppKey)
	}
	if creds.AppSecret == envAppSecret {
		t.Errorf("RED (must fix): GetLazadaCredentials returned env LAZADA_APP_SECRET ('%s') instead of DB-only value", creds.AppSecret)
	}
}

func TestGetTiktokCredentials_DBOnly_NoEnvFallback(t *testing.T) {
	// Set env vars to WRONG values that should NOT be returned
	envAppKey := "WRONG_ENV_TIKTOK_APP_KEY_SHOULD_NOT_BE_USED"
	envAppSecret := "WRONG_ENV_TIKTOK_APP_SECRET_SHOULD_NOT_BE_USED"

	os.Setenv("TIKTOK_APP_KEY", envAppKey)
	defer os.Unsetenv("TIKTOK_APP_KEY")
	os.Setenv("TIKTOK_APP_SECRET", envAppSecret)
	defer os.Unsetenv("TIKTOK_APP_SECRET")

	svc := &GlobalConfigService{}

	creds, err := svc.GetTiktokCredentials()
	if err != nil {
		t.Fatalf("GetTiktokCredentials returned error: %v", err)
	}

	// RED: env fallback is still active
	if creds.AppKey == envAppKey {
		t.Errorf("RED (must fix): GetTiktokCredentials returned env TIKTOK_APP_KEY ('%s') instead of DB-only value", creds.AppKey)
	}
	if creds.AppSecret == envAppSecret {
		t.Errorf("RED (must fix): GetTiktokCredentials returned env TIKTOK_APP_SECRET ('%s') instead of DB-only value", creds.AppSecret)
	}
}

func TestGetCredentials_MissingInDB_ReturnsError(t *testing.T) {
	// When credentials are absent from DB AND env fallback is removed,
	// the credential methods should return an error.
	//
	// Current implementation: always returns (empty creds, nil) — no error.
	// This test asserts err != nil, which FAILS (RED) because the
	// method never validates whether credentials are actually present.

	svc := &GlobalConfigService{}

	// Test all three platforms — none should silently return empty credentials
	t.Run("Shopee", func(t *testing.T) {
		_, err := svc.GetShopeeCredentials()
		if err == nil {
			t.Error("RED (must fix): GetShopeeCredentials should return error when credentials are missing (no env fallback)")
		}
	})

	t.Run("Lazada", func(t *testing.T) {
		_, err := svc.GetLazadaCredentials()
		if err == nil {
			t.Error("RED (must fix): GetLazadaCredentials should return error when credentials are missing (no env fallback)")
		}
	})

	t.Run("Tiktok", func(t *testing.T) {
		_, err := svc.GetTiktokCredentials()
		if err == nil {
			t.Error("RED (must fix): GetTiktokCredentials should return error when credentials are missing (no env fallback)")
		}
	})
}
