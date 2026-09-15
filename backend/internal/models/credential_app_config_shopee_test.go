package models

import (
	"encoding/json"
	"testing"
	"time"
)

// Phase 8 — Shopee-specific expectations on the masked response.
// Kept in a separate file (per handlers AGENTS.md file-naming convention:
// `{domain}_{platform}.go`) so credential_app_config.go stays focused on the
// data model.

// TestToMaskedResponse_ShopeeFullShape ensures every new Phase 8 field lands
// in the JSON response and every secret stays hidden.
func TestToMaskedResponse_ShopeeFullShape(t *testing.T) {
	expiry := time.Date(2026, 11, 25, 22, 59, 0, 0, time.UTC)
	testExpiry := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	cfg := &CredentialAppConfig{
		ID:                      "cfg-1",
		Platform:                "shopee",
		Region:                  "id",
		PartnerID:               2011782,
		PartnerKey:              "SUPER_SECRET_LIVE",
		TestPartnerID:           1187586,
		TestPartnerKey:          "SUPER_SECRET_TEST",
		PartnerKeyExpiresAt:     &expiry,
		TestPartnerKeyExpiresAt: &testExpiry,
		AppStatus:               "online",
		ActivePartnerEnv:        "live",
		Configured:              true,
	}
	got := cfg.ToMaskedResponse()

	if got.PartnerID != 2011782 {
		t.Errorf("PartnerID = %d, want 2011782 (ID is not a secret)", got.PartnerID)
	}
	if got.TestPartnerID != 1187586 {
		t.Errorf("TestPartnerID = %d, want 1187586", got.TestPartnerID)
	}
	if !got.TestConfigured {
		t.Errorf("TestConfigured = false, want true (both id + key present)")
	}
	if got.AppStatus != "online" {
		t.Errorf("AppStatus = %q, want online", got.AppStatus)
	}
	if got.ActivePartnerEnv != "live" {
		t.Errorf("ActivePartnerEnv = %q, want live", got.ActivePartnerEnv)
	}
	if got.PartnerKeyExpiresAt == nil || !got.PartnerKeyExpiresAt.Equal(expiry) {
		t.Errorf("PartnerKeyExpiresAt = %v, want %v", got.PartnerKeyExpiresAt, expiry)
	}

	// Secrets must never appear in the JSON serialisation of the masked
	// response OR of the raw model (json:"-" tags).
	body, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal masked: %v", err)
	}
	str := string(body)
	for _, secret := range []string{"SUPER_SECRET_LIVE", "SUPER_SECRET_TEST"} {
		if contains(str, secret) {
			t.Errorf("masked response leaked secret %q; body=%s", secret, str)
		}
	}
	rawBody, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal raw: %v", err)
	}
	rawStr := string(rawBody)
	for _, secret := range []string{"SUPER_SECRET_LIVE", "SUPER_SECRET_TEST"} {
		if contains(rawStr, secret) {
			t.Errorf("raw model leaked secret %q via JSON (json:\"-\" tag missing); body=%s", secret, rawStr)
		}
	}
}

// TestToMaskedResponse_TestConfigured_RequiresBothFields — partial sandbox
// setup (id without key or vice versa) must NOT flip TestConfigured to true.
func TestToMaskedResponse_TestConfigured_RequiresBothFields(t *testing.T) {
	cases := []struct {
		name string
		id   int64
		key  string
		want bool
	}{
		{"both_set", 100, "k", true},
		{"only_id", 100, "", false},
		{"only_key", 0, "k", false},
		{"neither", 0, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &CredentialAppConfig{Platform: "shopee", TestPartnerID: tc.id, TestPartnerKey: tc.key}
			if got := cfg.ToMaskedResponse().TestConfigured; got != tc.want {
				t.Errorf("TestConfigured = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestToMaskedResponse_DefaultsForLegacyRows — legacy rows migrated with empty
// AppStatus / ActivePartnerEnv should fall back to their DB defaults in the
// response, so the UI never renders `null` badges.
func TestToMaskedResponse_DefaultsForLegacyRows(t *testing.T) {
	cfg := &CredentialAppConfig{Platform: "shopee"}
	got := cfg.ToMaskedResponse()
	if got.AppStatus != "online" {
		t.Errorf("legacy AppStatus fallback = %q, want online", got.AppStatus)
	}
	if got.ActivePartnerEnv != "live" {
		t.Errorf("legacy ActivePartnerEnv fallback = %q, want live", got.ActivePartnerEnv)
	}
}

// contains is a tiny helper — avoiding strings.Contains import churn in a
// leaf test file.
func contains(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
