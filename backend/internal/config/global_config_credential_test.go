package config

import (
	"os"
	"testing"
)

// Tests for removing os.Getenv fallback from credential methods.
//
// These tests verify that GetShopeeCredentials / GetTiktokCredentials / GetLazadaCredentials
// return ONLY database values and do NOT fall back to environment variables.
//
// When DB is unavailable and env fallback is removed, credential methods MUST return an error.
// This is proven by: (1) setting env vars to WRONG values, (2) creating service with nil DB,
// (3) asserting error is returned (proving env fallback is NOT used to silently provide values).
func TestGetShopeeCredentials_DBOnly_NoEnvFallback(t *testing.T) {
	// Set env vars to WRONG values — if env fallback were active, these would be returned
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
	// With env fallback removed, this MUST return an error (no DB = no credentials)
	svc := &GlobalConfigService{}

	creds, err := svc.GetShopeeCredentials()
	if err == nil {
		// If no error, env fallback is still active — check creds don't match env values
		if creds.PartnerID == envPartnerID {
			t.Errorf("FAIL: GetShopeeCredentials returned env SHOPEE_PARTNER_ID ('%s') — env fallback still active", creds.PartnerID)
		}
		if creds.PartnerKey == envPartnerKey {
			t.Errorf("FAIL: GetShopeeCredentials returned env SHOPEE_PARTNER_KEY ('%s') — env fallback still active", creds.PartnerKey)
		}
		if creds.PushPartnerKey == envPushKey {
			t.Errorf("FAIL: GetShopeeCredentials returned env SHOPEE_PUSH_PARTNER_KEY ('%s') — env fallback still active", creds.PushPartnerKey)
		}
		t.Fatal("FAIL: GetShopeeCredentials returned nil error — expected error when credentials missing and no env fallback")
	}
	// GREEN: error returned — env fallback is not active
}

func TestGetLazadaCredentials_DBOnly_NoEnvFallback(t *testing.T) {
	// Set env vars to WRONG values — if env fallback were active, these would be returned
	envAppKey := "WRONG_ENV_LAZADA_APP_KEY_SHOULD_NOT_BE_USED"
	envAppSecret := "WRONG_ENV_LAZADA_APP_SECRET_SHOULD_NOT_BE_USED"

	os.Setenv("LAZADA_APP_KEY", envAppKey)
	defer os.Unsetenv("LAZADA_APP_KEY")
	os.Setenv("LAZADA_APP_SECRET", envAppSecret)
	defer os.Unsetenv("LAZADA_APP_SECRET")

	svc := &GlobalConfigService{}

	creds, err := svc.GetLazadaCredentials()
	if err == nil {
		if creds.AppKey == envAppKey {
			t.Errorf("FAIL: GetLazadaCredentials returned env LAZADA_APP_KEY ('%s') — env fallback still active", creds.AppKey)
		}
		if creds.AppSecret == envAppSecret {
			t.Errorf("FAIL: GetLazadaCredentials returned env LAZADA_APP_SECRET ('%s') — env fallback still active", creds.AppSecret)
		}
		t.Fatal("FAIL: GetLazadaCredentials returned nil error — expected error when credentials missing and no env fallback")
	}
	// GREEN: error returned — env fallback is not active
}

func TestGetTiktokCredentials_DBOnly_NoEnvFallback(t *testing.T) {
	// Set env vars to WRONG values — if env fallback were active, these would be returned
	envAppKey := "WRONG_ENV_TIKTOK_APP_KEY_SHOULD_NOT_BE_USED"
	envAppSecret := "WRONG_ENV_TIKTOK_APP_SECRET_SHOULD_NOT_BE_USED"

	os.Setenv("TIKTOK_APP_KEY", envAppKey)
	defer os.Unsetenv("TIKTOK_APP_KEY")
	os.Setenv("TIKTOK_APP_SECRET", envAppSecret)
	defer os.Unsetenv("TIKTOK_APP_SECRET")

	svc := &GlobalConfigService{}

	creds, err := svc.GetTiktokCredentials()
	if err == nil {
		if creds.AppKey == envAppKey {
			t.Errorf("FAIL: GetTiktokCredentials returned env TIKTOK_APP_KEY ('%s') — env fallback still active", creds.AppKey)
		}
		if creds.AppSecret == envAppSecret {
			t.Errorf("FAIL: GetTiktokCredentials returned env TIKTOK_APP_SECRET ('%s') — env fallback still active", creds.AppSecret)
		}
		t.Fatal("FAIL: GetTiktokCredentials returned nil error — expected error when credentials missing and no env fallback")
	}
	// GREEN: error returned — env fallback is not active
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
