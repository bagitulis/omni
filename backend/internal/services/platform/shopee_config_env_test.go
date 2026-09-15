package platform

import (
	"testing"

	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// Phase 9 — wire pkg/shopee/SelectActivePair into ShopeeConfigManager.
//
// The credential_app_configs table now stores a live pair AND an optional
// test pair, plus an active_partner_env toggle. The config manager must
// respect the toggle when picking which pair signs outgoing API calls.

// TestShopeeConfigManager_ApplyActivePair_LiveDefault — when env is live or
// blank, live pair wins.
func TestShopeeConfigManager_ApplyActivePair_LiveDefault(t *testing.T) {
	m := NewShopeeConfigManager("t1")
	m.ApplyActivePair(
		shopeePkg.PartnerPair{ID: 200, Key: "live_key"},
		shopeePkg.PartnerPair{ID: 100, Key: "test_key"},
		"live",
	)
	if m.PartnerID != "200" {
		t.Errorf("PartnerID = %q, want 200", m.PartnerID)
	}
	if m.PartnerKey != "live_key" {
		t.Errorf("PartnerKey = %q, want live_key", m.PartnerKey)
	}
	// Empty env → same defaults.
	m2 := NewShopeeConfigManager("t2")
	m2.ApplyActivePair(
		shopeePkg.PartnerPair{ID: 200, Key: "live_key"},
		shopeePkg.PartnerPair{ID: 100, Key: "test_key"},
		"",
	)
	if m2.PartnerID != "200" || m2.PartnerKey != "live_key" {
		t.Errorf("empty env: got (id=%q, key=%q), want (200, live_key)", m2.PartnerID, m2.PartnerKey)
	}
}

// TestShopeeConfigManager_ApplyActivePair_TestExplicit.
func TestShopeeConfigManager_ApplyActivePair_TestExplicit(t *testing.T) {
	m := NewShopeeConfigManager("t1")
	m.ApplyActivePair(
		shopeePkg.PartnerPair{ID: 200, Key: "live_key"},
		shopeePkg.PartnerPair{ID: 100, Key: "test_key"},
		"test",
	)
	if m.PartnerID != "100" {
		t.Errorf("PartnerID = %q, want 100 (test)", m.PartnerID)
	}
	if m.PartnerKey != "test_key" {
		t.Errorf("PartnerKey = %q, want test_key", m.PartnerKey)
	}
}

// TestShopeeConfigManager_ApplyActivePair_TestMissing_FallsBackToLive — same
// safety-default guard as SelectActivePair: a mis-configured tenant hits
// production with real creds rather than an empty sandbox.
func TestShopeeConfigManager_ApplyActivePair_TestMissing_FallsBackToLive(t *testing.T) {
	m := NewShopeeConfigManager("t1")
	m.ApplyActivePair(
		shopeePkg.PartnerPair{ID: 200, Key: "live_key"},
		shopeePkg.PartnerPair{}, // empty test pair
		"test",
	)
	if m.PartnerID != "200" || m.PartnerKey != "live_key" {
		t.Errorf("test-missing fallback: got (id=%q, key=%q), want live", m.PartnerID, m.PartnerKey)
	}
}
