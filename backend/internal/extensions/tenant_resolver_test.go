package extensions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/testutils"
	"gorm.io/gorm"
)

// The resolvers search every tenant. These tests use in-memory SQLite so the
// search behaviour — including the negative cases — is verifiable without a
// container runtime.

// tenantFixture builds two independent tenant databases, each a separate schema
// stand-in, so a lookup can be shown to find the right one and no other.
//
// The databases are explicitly named: SetupTestSQLite keys off the test name, so
// calling it twice in one test would hand back the SAME database and every
// isolation assertion below would pass vacuously.
func tenantFixture(t *testing.T) (TenantDBFunc, TenantLister, map[string]*gorm.DB) {
	t.Helper()

	dbs := map[string]*gorm.DB{
		"tenant-a": testutils.SetupNamedTestSQLite(t, "tenant_a_"+sanitize(t.Name()),
			&models.Extension{}, &models.PairingCode{}, &models.ScrapedProduct{}),
		"tenant-b": testutils.SetupNamedTestSQLite(t, "tenant_b_"+sanitize(t.Name()),
			&models.Extension{}, &models.PairingCode{}, &models.ScrapedProduct{}),
	}

	tenantDB := func(tenantID string) (*gorm.DB, error) {
		db, ok := dbs[tenantID]
		if !ok {
			return nil, errors.New("unknown tenant")
		}
		return db, nil
	}
	list := func() []string { return []string{"tenant-a", "tenant-b"} }

	return tenantDB, list, dbs
}

// sanitize makes a test name safe for use in a sqlite DSN.
func sanitize(name string) string {
	out := make([]rune, 0, len(name))
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			out = append(out, r)
		default:
			out = append(out, '_')
		}
	}
	return string(out)
}

func TestResolveTenantForPairingCode_FindsOwningTenant(t *testing.T) {
	tenantDB, list, dbs := tenantFixture(t)
	ctx := context.Background()
	now := time.Now()

	// Code lives only in tenant-b.
	repoB := repositories.NewExtensionRepository(dbs["tenant-b"])
	if err := repoB.CreatePairingCode(ctx, &models.PairingCode{
		Code: "CODEB123", ExpiresAt: now.Add(5 * time.Minute),
	}); err != nil {
		t.Fatalf("seed code: %v", err)
	}

	got, err := ResolveTenantForPairingCode(ctx, "CODEB123", tenantDB, list)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != "tenant-b" {
		t.Errorf("resolved tenant = %q, want tenant-b", got)
	}
}

func TestResolveTenantForPairingCode_UnknownCode(t *testing.T) {
	tenantDB, list, _ := tenantFixture(t)

	_, err := ResolveTenantForPairingCode(context.Background(), "NOPE1234", tenantDB, list)
	if !errors.Is(err, ErrTenantNotFoundForCredential) {
		t.Errorf("unknown code error = %v, want ErrTenantNotFoundForCredential", err)
	}
}

func TestResolveTenantForPairingCode_ExpiredCodeDoesNotResolve(t *testing.T) {
	tenantDB, list, dbs := tenantFixture(t)
	ctx := context.Background()
	now := time.Now()

	repoA := repositories.NewExtensionRepository(dbs["tenant-a"])
	if err := repoA.CreatePairingCode(ctx, &models.PairingCode{
		Code: "EXPIRED1", ExpiresAt: now.Add(-time.Minute),
	}); err != nil {
		t.Fatalf("seed expired code: %v", err)
	}

	_, err := ResolveTenantForPairingCode(ctx, "EXPIRED1", tenantDB, list)
	if !errors.Is(err, ErrTenantNotFoundForCredential) {
		t.Errorf("expired code must not resolve, got err = %v", err)
	}
}

func TestResolveTenantForPairingCode_EmptyCode(t *testing.T) {
	tenantDB, list, _ := tenantFixture(t)

	_, err := ResolveTenantForPairingCode(context.Background(), "", tenantDB, list)
	if !errors.Is(err, ErrTenantNotFoundForCredential) {
		t.Errorf("empty code must not resolve, got err = %v", err)
	}
}

// A tenant whose database is unreachable must not break the search for the
// others: the code may belong to a healthy tenant.
func TestResolveTenantForPairingCode_SkipsUnreachableTenant(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	good := testutils.SetupTestSQLite(t, &models.Extension{}, &models.PairingCode{})
	repo := repositories.NewExtensionRepository(good)
	if err := repo.CreatePairingCode(ctx, &models.PairingCode{
		Code: "HEALTHY1", ExpiresAt: now.Add(5 * time.Minute),
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	tenantDB := func(tenantID string) (*gorm.DB, error) {
		if tenantID == "broken" {
			return nil, errors.New("database unreachable")
		}
		return good, nil
	}
	list := func() []string { return []string{"broken", "healthy"} }

	got, err := ResolveTenantForPairingCode(ctx, "HEALTHY1", tenantDB, list)
	if err != nil {
		t.Fatalf("resolve must tolerate an unreachable tenant: %v", err)
	}
	if got != "healthy" {
		t.Errorf("resolved tenant = %q, want healthy", got)
	}
}

func TestResolveExtensionByToken_FindsOwningTenant(t *testing.T) {
	tenantDB, list, dbs := tenantFixture(t)
	ctx := context.Background()

	// Mint a token, then store only its hash in tenant-a.
	token, hash, err := GeneratePairingToken()
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	repoA := repositories.NewExtensionRepository(dbs["tenant-a"])
	if err := repoA.CreateExtension(ctx, &models.Extension{
		ExtensionID: "ext-a1",
		TokenHash:   hash,
		Status:      models.ExtensionStatusDisconnected,
		PairedAt:    time.Now(),
	}); err != nil {
		t.Fatalf("seed extension: %v", err)
	}

	tenantID, ext, err := ResolveExtensionByToken(ctx, token, tenantDB, list)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if tenantID != "tenant-a" {
		t.Errorf("resolved tenant = %q, want tenant-a", tenantID)
	}
	if ext == nil || ext.ExtensionID != "ext-a1" {
		t.Errorf("resolved extension = %+v, want ext-a1", ext)
	}
}

func TestResolveExtensionByToken_UnknownToken(t *testing.T) {
	tenantDB, list, _ := tenantFixture(t)

	_, _, err := ResolveExtensionByToken(context.Background(), "not-a-real-token", tenantDB, list)
	if !errors.Is(err, ErrTenantNotFoundForCredential) {
		t.Errorf("unknown token error = %v, want ErrTenantNotFoundForCredential", err)
	}
}

func TestResolveExtensionByToken_EmptyToken(t *testing.T) {
	tenantDB, list, dbs := tenantFixture(t)
	ctx := context.Background()

	// Seed an extension with an EMPTY token hash to prove an empty token cannot
	// match it. GORM treats a zero-value string condition as "no condition", so
	// without an explicit guard an empty token would return this row and let an
	// unauthenticated caller impersonate the extension.
	if err := repositories.NewExtensionRepository(dbs["tenant-a"]).CreateExtension(ctx, &models.Extension{
		ExtensionID: "ext-emptyhash",
		TokenHash:   "",
		Status:      models.ExtensionStatusDisconnected,
		PairedAt:    time.Now(),
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, _, err := ResolveExtensionByToken(ctx, "", tenantDB, list)
	if !errors.Is(err, ErrTenantNotFoundForCredential) {
		t.Errorf("empty token must never resolve, got err = %v", err)
	}
}

// A token for tenant-a must never resolve to tenant-b, even when both tenants
// have extensions. This is the cross-tenant negative case.
func TestResolveExtensionByToken_DoesNotCrossTenants(t *testing.T) {
	tenantDB, list, dbs := tenantFixture(t)
	ctx := context.Background()

	tokenA, hashA, _ := GeneratePairingToken()
	tokenB, hashB, _ := GeneratePairingToken()

	seed := func(tenantID, extID, hash string) {
		if err := repositories.NewExtensionRepository(dbs[tenantID]).CreateExtension(ctx, &models.Extension{
			ExtensionID: extID,
			TokenHash:   hash,
			Status:      models.ExtensionStatusDisconnected,
			PairedAt:    time.Now(),
		}); err != nil {
			t.Fatalf("seed %s: %v", extID, err)
		}
	}
	seed("tenant-a", "ext-a", hashA)
	seed("tenant-b", "ext-b", hashB)

	tenantForA, extForA, err := ResolveExtensionByToken(ctx, tokenA, tenantDB, list)
	if err != nil {
		t.Fatalf("resolve tokenA: %v", err)
	}
	if tenantForA != "tenant-a" || extForA.ExtensionID != "ext-a" {
		t.Fatalf("token A resolved to (%s, %s), want (tenant-a, ext-a)", tenantForA, extForA.ExtensionID)
	}

	tenantForB, extForB, err := ResolveExtensionByToken(ctx, tokenB, tenantDB, list)
	if err != nil {
		t.Fatalf("resolve tokenB: %v", err)
	}
	if tenantForB != "tenant-b" || extForB.ExtensionID != "ext-b" {
		t.Fatalf("token B resolved to (%s, %s), want (tenant-b, ext-b)", tenantForB, extForB.ExtensionID)
	}

	// The critical assertion: each token saw only its own tenant's extension.
	if extForA.ExtensionID == extForB.ExtensionID {
		t.Fatal("different tokens resolved to the same extension: tenant isolation is broken")
	}
}
