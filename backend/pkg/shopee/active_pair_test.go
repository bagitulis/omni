package shopee

import "testing"

// Phase 8 — runtime switch between Shopee Live and Test partner pairs.
//
// Callers pass the DB row's active_partner_env + the two pairs; the helper
// returns the pair that should be used for signing. Keeping the switch out of
// the Client constructor lets the same Client type serve either environment
// without duplicating client creation logic.

// TestSelectActivePair_LiveDefault — empty env or explicit "live" returns
// the live pair. Live is the safe default: production traffic must never
// silently fall through to the sandbox.
func TestSelectActivePair_LiveDefault(t *testing.T) {
	got := SelectActivePair(PartnerPair{ID: 200, Key: "live"}, PartnerPair{ID: 100, Key: "test"}, "live")
	if got.ID != 200 || got.Key != "live" {
		t.Errorf("live env: got %+v, want {ID:200,Key:live}", got)
	}
	got = SelectActivePair(PartnerPair{ID: 200, Key: "live"}, PartnerPair{ID: 100, Key: "test"}, "")
	if got.ID != 200 || got.Key != "live" {
		t.Errorf("empty env: got %+v, want live default", got)
	}
}

// TestSelectActivePair_TestExplicit — explicit "test" returns the test pair.
func TestSelectActivePair_TestExplicit(t *testing.T) {
	got := SelectActivePair(PartnerPair{ID: 200, Key: "live"}, PartnerPair{ID: 100, Key: "test"}, "test")
	if got.ID != 100 || got.Key != "test" {
		t.Errorf("test env: got %+v, want {ID:100,Key:test}", got)
	}
}

// TestSelectActivePair_TestRequestedButMissing_FallsBackToLive — the
// safety-net path: a caller flipped active_partner_env to "test" but the
// sandbox pair was never populated. Fall back to live rather than sending
// a request with empty credentials that would 401 anyway.
func TestSelectActivePair_TestRequestedButMissing_FallsBackToLive(t *testing.T) {
	got := SelectActivePair(PartnerPair{ID: 200, Key: "live"}, PartnerPair{}, "test")
	if got.ID != 200 || got.Key != "live" {
		t.Errorf("test-missing fallback: got %+v, want live", got)
	}
	// Also true when only the ID is present but the key is empty (invalid pair).
	got = SelectActivePair(PartnerPair{ID: 200, Key: "live"}, PartnerPair{ID: 100, Key: ""}, "test")
	if got.ID != 200 {
		t.Errorf("partial-test pair should fall back to live; got %+v", got)
	}
}

// TestSelectActivePair_UnknownEnv_DefaultsLive — defensive: unknown string
// (typo, injection) does not silently pick test.
func TestSelectActivePair_UnknownEnv_DefaultsLive(t *testing.T) {
	got := SelectActivePair(PartnerPair{ID: 200, Key: "live"}, PartnerPair{ID: 100, Key: "test"}, "GARBAGE")
	if got.ID != 200 {
		t.Errorf("garbage env: got %+v, want live", got)
	}
}
