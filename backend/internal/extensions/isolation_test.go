package extensions

import (
	"testing"
	"time"
)

// Tenant isolation is a data-leak class of bug, so it gets an explicit negative
// test rather than being assumed from the fact that queries take a *gorm.DB.
//
// The design's core claim: a tenant is derived server-side from a pairing token
// and is NEVER accepted from the client. These tests pin that property at the
// unit level, where they run without a database.

// TestPairRequest_HasNoClientSuppliedTenant guards the central invariant.
//
// If a TenantID field is ever added to the inbound pairing request, a client
// could choose which tenant it is written into. The field's absence is the
// defence, so its absence is asserted.
func TestPairRequest_HasNoClientSuppliedTenant(t *testing.T) {
	if hasField(PairRequest{}, "TenantID") {
		t.Fatal("PairRequest must NOT accept a tenant_id: the tenant is derived " +
			"server-side from the pairing token, never supplied by the client")
	}
}

// TestConnectRequest_HasNoClientSuppliedTenant is the same guard for the
// WebSocket handshake.
func TestConnectRequest_HasNoClientSuppliedTenant(t *testing.T) {
	if hasField(ConnectRequest{}, "TenantID") {
		t.Fatal("ConnectRequest must NOT accept a tenant_id: a client that could " +
			"name its own tenant could read another tenant's data")
	}
}

// TestScrapeRequest_HasNoClientSuppliedTenant is the same guard for job
// dispatch — the highest-impact place to get this wrong, since scraped rows are
// written with whatever tenant the request carries.
func TestScrapeRequest_HasNoClientSuppliedTenant(t *testing.T) {
	if hasField(ScrapeRequest{}, "TenantID") {
		t.Fatal("ScrapeRequest must NOT accept a tenant_id: results are written " +
			"into the schema of the tenant bound to the pairing token")
	}
}

// TestTokenBinding_CannotRepresentTwoTenants documents that a token maps to
// exactly one tenant, so there is no representation for a cross-tenant token.
func TestTokenBinding_SingleTenant(t *testing.T) {
	// A binding is created from a token hash plus the tenant resolved at pair
	// time. There is deliberately no constructor that takes two tenants, and no
	// setter that can retarget an existing binding.
	token, hash, err := GeneratePairingToken()
	if err != nil {
		t.Fatalf("GeneratePairingToken: %v", err)
	}

	b := TokenBinding{
		TokenHash: hash,
		TenantID:  "tenant-a",
	}
	if !b.Matches(TokenBinding{TokenHash: hash, TenantID: "tenant-a"}) {
		t.Error("a binding must match itself")
	}
	// Same token hash, different tenant: must NOT match. This is the cross-tenant
	// case; a token minted for tenant A must never act as tenant B.
	if b.Matches(TokenBinding{TokenHash: hash, TenantID: "tenant-b"}) {
		t.Fatal("a token bound to tenant-a must never match tenant-b")
	}
	// A different token must not match, even for the same tenant.
	other, otherHash, _ := GeneratePairingToken()
	_ = other
	if b.Matches(TokenBinding{TokenHash: otherHash, TenantID: "tenant-a"}) {
		t.Error("a different token must not match")
	}
	if token == other {
		t.Error("two generated tokens must differ")
	}
}

// TestTokenBinding_EmptyTenantIsRejected mirrors the platform rule that there is
// no default tenant: an empty tenant must fail closed, never fall back.
func TestTokenBinding_EmptyTenantIsRejected(t *testing.T) {
	if (TokenBinding{TenantID: ""}).Valid(time.Now()) {
		t.Error("a binding with an empty tenant_id must be invalid: there is no default tenant")
	}
	if (TokenBinding{TenantID: "tenant-a"}).Valid(time.Now()) {
		t.Error("a binding without a token hash must be invalid")
	}
}

// TestTokenBinding_ExpiryFailsClosed checks that an expired binding is invalid
// rather than silently accepted.
func TestTokenBinding_ExpiryFailsClosed(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)

	expired := TokenBinding{
		TokenHash: "h",
		TenantID:  "tenant-a",
		ExpiresAt: now.Add(-time.Second),
	}
	if expired.Valid(now) {
		t.Error("an expired binding must be invalid")
	}

	live := TokenBinding{
		TokenHash: "h",
		TenantID:  "tenant-a",
		ExpiresAt: now.Add(time.Hour),
	}
	if !live.Valid(now) {
		t.Error("a live binding must be valid")
	}

	// No expiry set means non-expiring, which is only valid with a tenant+hash.
	noExpiry := TokenBinding{TokenHash: "h", TenantID: "tenant-a"}
	if !noExpiry.Valid(now) {
		t.Error("a binding with no expiry is non-expiring and must be valid")
	}
}
