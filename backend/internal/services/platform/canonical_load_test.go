package platform

import (
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
)

// Phase 9 / Bug C — pure translation from canonical credential tables into
// the fields ShopeeConfigManager needs. Kept as pure functions so the DB
// call and the mapping can be tested independently. The DB integration test
// itself lives under testcontainers-gated repositories tests.

// TestCanonicalToShopeeFields_MinimalConnection maps store_identifier →
// ShopID and pulls the access token.
func TestCanonicalToShopeeFields_MinimalConnection(t *testing.T) {
	app := &models.CredentialAppConfig{
		PartnerID:  2011782,
		PartnerKey: "live-key",
	}
	conn := &models.CredentialConnection{
		Platform:        "shopee",
		StoreIdentifier: "530635055",
		AccessToken:     "at-live",
		RefreshToken:    "rt-live",
		TokenExpiry:     time.Now().Add(time.Hour).UnixMilli(),
		RefreshExpiry:   time.Now().Add(24 * time.Hour).UnixMilli(),
	}
	got := CanonicalToShopeeFields(app, conn)
	if got.ShopID != "530635055" {
		t.Errorf("ShopID = %q, want 530635055", got.ShopID)
	}
	if got.AccessToken != "at-live" {
		t.Errorf("AccessToken = %q, want at-live", got.AccessToken)
	}
	if got.RefreshToken != "rt-live" {
		t.Errorf("RefreshToken = %q, want rt-live", got.RefreshToken)
	}
	if got.LivePair.ID != 2011782 || got.LivePair.Key != "live-key" {
		t.Errorf("LivePair = %+v, want {2011782, live-key}", got.LivePair)
	}
	if got.TestPair.IsValid() {
		t.Errorf("TestPair should be zero-value; got %+v", got.TestPair)
	}
	if got.ActivePartnerEnv != "live" {
		t.Errorf("ActivePartnerEnv = %q, want live (default)", got.ActivePartnerEnv)
	}
}

// TestCanonicalToShopeeFields_WithTestPair — sandbox pair present.
func TestCanonicalToShopeeFields_WithTestPair(t *testing.T) {
	app := &models.CredentialAppConfig{
		PartnerID:        2011782,
		PartnerKey:       "live-key",
		TestPartnerID:    1187586,
		TestPartnerKey:   "test-key",
		ActivePartnerEnv: "test",
	}
	conn := &models.CredentialConnection{
		Platform:        "shopee",
		StoreIdentifier: "530635055",
		AccessToken:     "at",
	}
	got := CanonicalToShopeeFields(app, conn)
	if got.TestPair.ID != 1187586 || got.TestPair.Key != "test-key" {
		t.Errorf("TestPair = %+v, want {1187586, test-key}", got.TestPair)
	}
	if got.ActivePartnerEnv != "test" {
		t.Errorf("ActivePartnerEnv = %q, want test", got.ActivePartnerEnv)
	}
}

// TestCanonicalToShopeeFields_NilConnection — no store connected yet; still
// return the pairs so the client can be *partially* configured for OAuth
// initiation (which needs partner_id + partner_key but no shop_id).
func TestCanonicalToShopeeFields_NilConnection(t *testing.T) {
	app := &models.CredentialAppConfig{PartnerID: 2011782, PartnerKey: "live-key"}
	got := CanonicalToShopeeFields(app, nil)
	if got.ShopID != "" {
		t.Errorf("ShopID should be empty when no conn; got %q", got.ShopID)
	}
	if got.LivePair.ID != 2011782 {
		t.Errorf("LivePair still populated from app: got %+v", got.LivePair)
	}
}

// TestCanonicalToShopeeFields_NilApp — no app_config yet (fresh tenant);
// caller should treat all fields as zero and fall back to legacy load.
func TestCanonicalToShopeeFields_NilApp(t *testing.T) {
	got := CanonicalToShopeeFields(nil, nil)
	if got.LivePair.IsValid() {
		t.Errorf("expected zero pair; got %+v", got.LivePair)
	}
	if got.AccessToken != "" {
		t.Errorf("expected empty token; got %q", got.AccessToken)
	}
}

// TestCanonicalToShopeeFields_EmptyActivePartnerEnv_DefaultsLive is the
// safety default that mirrors pkg/shopee.SelectActivePair.
func TestCanonicalToShopeeFields_EmptyActivePartnerEnv_DefaultsLive(t *testing.T) {
	app := &models.CredentialAppConfig{
		PartnerID:  1,
		PartnerKey: "k",
		// ActivePartnerEnv: intentionally empty
	}
	got := CanonicalToShopeeFields(app, nil)
	if got.ActivePartnerEnv != "live" {
		t.Errorf("ActivePartnerEnv = %q, want live default", got.ActivePartnerEnv)
	}
}
