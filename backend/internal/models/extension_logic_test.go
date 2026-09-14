package models

import (
	"encoding/json"
	"testing"
	"time"
)

// These are pure unit tests: they run with no database and no container
// runtime, so the pairing/expiry logic has executable coverage even on a
// machine where the Postgres-backed repository tests skip.

func TestPairingCode_IsUsable(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name   string
		code   PairingCode
		want   bool
		reason string
	}{
		{
			name:   "fresh code is usable",
			code:   PairingCode{ExpiresAt: now.Add(5 * time.Minute)},
			want:   true,
			reason: "unconsumed and unexpired",
		},
		{
			name:   "consumed code is not usable",
			code:   PairingCode{Consumed: true, ExpiresAt: now.Add(5 * time.Minute)},
			want:   false,
			reason: "single-use: one code mints exactly one token",
		},
		{
			name:   "expired code is not usable",
			code:   PairingCode{ExpiresAt: now.Add(-1 * time.Second)},
			want:   false,
			reason: "TTL elapsed",
		},
		{
			name:   "consumed and expired is not usable",
			code:   PairingCode{Consumed: true, ExpiresAt: now.Add(-1 * time.Hour)},
			want:   false,
			reason: "both conditions fail",
		},
		{
			name:   "expiry exactly at now is not usable",
			code:   PairingCode{ExpiresAt: now},
			want:   false,
			reason: "boundary is exclusive so a code cannot be redeemed at its deadline",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.code.IsUsable(now); got != tc.want {
				t.Errorf("IsUsable() = %v, want %v (%s)", got, tc.want, tc.reason)
			}
		})
	}
}

func TestPairingCode_IsExpired(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)

	if !(&PairingCode{ExpiresAt: now.Add(-time.Minute)}).IsExpired(now) {
		t.Error("code past its expiry must report expired")
	}
	if (&PairingCode{ExpiresAt: now.Add(time.Minute)}).IsExpired(now) {
		t.Error("code before its expiry must not report expired")
	}
}

func TestStringList_ValueAndScan(t *testing.T) {
	original := StringList{CapabilityTabDiscovery, CapabilityShopeeScrape}

	// Value renders as JSON for the jsonb column.
	raw, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	asBytes, ok := raw.([]byte)
	if !ok {
		t.Fatalf("Value() returned %T, want []byte", raw)
	}

	var decoded []string
	if err := json.Unmarshal(asBytes, &decoded); err != nil {
		t.Fatalf("Value() produced invalid JSON: %v", err)
	}
	if len(decoded) != 2 || decoded[0] != CapabilityTabDiscovery {
		t.Errorf("round-trip mismatch: %v", decoded)
	}

	// Scan accepts the []byte form Postgres returns.
	var scanned StringList
	if err := scanned.Scan(asBytes); err != nil {
		t.Fatalf("Scan([]byte) error: %v", err)
	}
	if len(scanned) != 2 || scanned[1] != CapabilityShopeeScrape {
		t.Errorf("Scan([]byte) mismatch: %v", scanned)
	}

	// Scan also accepts a string (some drivers hand back string).
	var fromString StringList
	if err := fromString.Scan(`["tab_discovery"]`); err != nil {
		t.Fatalf("Scan(string) error: %v", err)
	}
	if len(fromString) != 1 {
		t.Errorf("Scan(string) mismatch: %v", fromString)
	}

	// NULL column scans to nil without error.
	var fromNil StringList
	if err := fromNil.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error: %v", err)
	}
	if fromNil != nil {
		t.Errorf("Scan(nil) = %v, want nil", fromNil)
	}

	// Unsupported types must error, not silently produce garbage.
	var bad StringList
	if err := bad.Scan(12345); err == nil {
		t.Error("Scan(int) must return an error")
	}
}

func TestStringList_Has(t *testing.T) {
	list := StringList{CapabilityTabDiscovery}

	if !list.Has(CapabilityTabDiscovery) {
		t.Error("Has() must find a present capability")
	}
	if list.Has(CapabilityShopeeScrape) {
		t.Error("Has() must not report an absent capability")
	}
	if StringList(nil).Has("anything") {
		t.Error("Has() on a nil list must be false")
	}
}

func TestStringList_GormDataType(t *testing.T) {
	var list StringList
	if got := list.GormDataType(); got != "jsonb" {
		t.Errorf("GormDataType() = %q, want jsonb", got)
	}
}

func TestScrapeModeAndSourceConstants(t *testing.T) {
	// These values are persisted and exposed via the API, so a rename is a
	// breaking change. Pin them explicitly.
	if ScrapeModeSearch != "search" || ScrapeModeShop != "shop" || ScrapeModeProduct != "product" {
		t.Error("scrape mode constants changed; this is API-visible")
	}
	if ScrapeSourceNetwork != "network" || ScrapeSourceDOM != "dom" {
		t.Error("scrape source constants changed; this is API-visible")
	}
	if ExtensionStatusConnected != "connected" || ExtensionStatusDisconnected != "disconnected" {
		t.Error("extension status constants changed; this is API-visible")
	}
}
