package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests run on in-memory SQLite so they provide signal even when no
// container runtime is available (the Postgres-backed suite skips in that case).
//
// Caveat: SQLite does not exercise Postgres-specific semantics. The postgres
// variants in extension_repository_test.go remain the source of truth for
// ON CONFLICT behaviour and jsonb handling.

func newExtRepoSqlite(t *testing.T) *ExtensionRepository {
	t.Helper()
	db := testutils.SetupTestSQLite(t,
		&models.Extension{},
		&models.PairingCode{},
		&models.ScrapedProduct{},
	)
	return NewExtensionRepository(db)
}

func TestSqlite_ExtensionCreateGetDelete(t *testing.T) {
	repo := newExtRepoSqlite(t)
	ctx := context.Background()

	ext := &models.Extension{
		ExtensionID:  "ext-1",
		Hostname:     "browser-extension",
		Capabilities: models.StringList{models.CapabilityShopeeScrape},
		Status:       models.ExtensionStatusDisconnected,
		PairedAt:     time.Now(),
	}
	require.NoError(t, repo.CreateExtension(ctx, ext))
	assert.NotZero(t, ext.ID)

	got, err := repo.GetExtensionByID(ctx, "ext-1")
	require.NoError(t, err)
	assert.Equal(t, "browser-extension", got.Hostname)
	assert.Equal(t, models.StringList{models.CapabilityShopeeScrape}, got.Capabilities,
		"capabilities must round-trip through JSON encoding")

	list, err := repo.ListExtensions(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	require.NoError(t, repo.DeleteExtension(ctx, "ext-1"))
	_, err = repo.GetExtensionByID(ctx, "ext-1")
	assert.Error(t, err, "deleted extension must no longer resolve")
}

func TestSqlite_ExtensionTokenHash(t *testing.T) {
	repo := newExtRepoSqlite(t)
	ctx := context.Background()

	require.NoError(t, repo.CreateExtension(ctx, &models.Extension{
		ExtensionID: "ext-tok",
		TokenHash:   "hash-1",
		PairedAt:    time.Now(),
	}))

	found, err := repo.FindExtensionByTokenHash(ctx, "hash-1")
	require.NoError(t, err)
	assert.Equal(t, "ext-tok", found.ExtensionID)

	// A wrong hash must not resolve.
	_, err = repo.FindExtensionByTokenHash(ctx, "hash-other")
	assert.Error(t, err)

	// An empty hash must not fall through to the first row. GORM treats a
	// zero-value string condition as "no condition", so this guards a real
	// footgun: empty hash silently returning an arbitrary extension would let
	// an unauthenticated caller impersonate one.
	_, err = repo.FindExtensionByTokenHash(ctx, "")
	assert.Error(t, err, "empty token hash must never resolve to an extension")
}

func TestSqlite_ExtensionStatusAndStartupReset(t *testing.T) {
	repo := newExtRepoSqlite(t)
	ctx := context.Background()

	require.NoError(t, repo.CreateExtension(ctx, &models.Extension{
		ExtensionID: "ext-s",
		Status:      models.ExtensionStatusDisconnected,
		PairedAt:    time.Now(),
	}))

	require.NoError(t, repo.UpdateExtensionStatus(ctx, "ext-s", models.ExtensionStatusConnected))
	got, err := repo.GetExtensionByID(ctx, "ext-s")
	require.NoError(t, err)
	assert.Equal(t, models.ExtensionStatusConnected, got.Status)
	assert.NotNil(t, got.LastSeen, "connecting must stamp last_seen")

	// Restart semantics: no live socket exists after boot, so every row must
	// be reset or the UI shows phantom online extensions.
	require.NoError(t, repo.MarkAllExtensionsDisconnected(ctx))
	got, err = repo.GetExtensionByID(ctx, "ext-s")
	require.NoError(t, err)
	assert.Equal(t, models.ExtensionStatusDisconnected, got.Status)
}

func TestSqlite_PairingCodeSingleUse(t *testing.T) {
	repo := newExtRepoSqlite(t)
	ctx := context.Background()
	now := time.Now()

	require.NoError(t, repo.CreatePairingCode(ctx, &models.PairingCode{
		Code: "CODE1", ExpiresAt: now.Add(5 * time.Minute),
	}))

	consumed, err := repo.ConsumePairingCode(ctx, "CODE1", now)
	require.NoError(t, err)
	assert.True(t, consumed.Consumed)
	assert.NotNil(t, consumed.ConsumedAt)

	// Single-use: a second redemption must be refused.
	_, err = repo.ConsumePairingCode(ctx, "CODE1", now)
	assert.ErrorIs(t, err, ErrPairingCodeUnusable)
}

func TestSqlite_PairingCodeExpiredAndUnknown(t *testing.T) {
	repo := newExtRepoSqlite(t)
	ctx := context.Background()
	now := time.Now()

	require.NoError(t, repo.CreatePairingCode(ctx, &models.PairingCode{
		Code: "EXPIRED", ExpiresAt: now.Add(-time.Minute),
	}))

	_, err := repo.ConsumePairingCode(ctx, "EXPIRED", now)
	assert.ErrorIs(t, err, ErrPairingCodeUnusable, "expired code must be rejected")

	_, err = repo.ConsumePairingCode(ctx, "NEVER-ISSUED", now)
	assert.ErrorIs(t, err, ErrPairingCodeUnusable,
		"unknown code must be indistinguishable from expired/used")
}

func TestSqlite_PairingCodePruning(t *testing.T) {
	repo := newExtRepoSqlite(t)
	ctx := context.Background()
	now := time.Now()

	require.NoError(t, repo.CreatePairingCode(ctx, &models.PairingCode{
		Code: "STALE", ExpiresAt: now.Add(-time.Hour),
	}))
	require.NoError(t, repo.CreatePairingCode(ctx, &models.PairingCode{
		Code: "LIVE", ExpiresAt: now.Add(time.Hour),
	}))

	require.NoError(t, repo.DeleteExpiredPairingCodes(ctx, now))

	// The live code must survive.
	_, err := repo.ConsumePairingCode(ctx, "LIVE", now)
	require.NoError(t, err, "pruning must not remove unexpired codes")
}

func TestSqlite_ScrapedProductsInsertCountDedupPage(t *testing.T) {
	repo := newExtRepoSqlite(t)
	ctx := context.Background()

	batch := []models.ScrapedProduct{
		{JobID: "job-1", ProductName: "A", Link: "https://shopee.co.id/a-i.1.11", PageNumber: 1},
		{JobID: "job-1", ProductName: "B", Link: "https://shopee.co.id/b-i.1.12", PageNumber: 1},
	}
	require.NoError(t, repo.InsertScrapedProducts(ctx, batch))

	n, err := repo.CountScrapedProducts(ctx, "job-1")
	require.NoError(t, err)
	assert.EqualValues(t, 2, n)

	// Resume re-reads the boundary page; duplicates must be skipped, not fail.
	require.NoError(t, repo.InsertScrapedProducts(ctx, batch))
	n, err = repo.CountScrapedProducts(ctx, "job-1")
	require.NoError(t, err)
	assert.EqualValues(t, 2, n, "identical links must not duplicate")

	// Empty batch is a no-op.
	assert.NoError(t, repo.InsertScrapedProducts(ctx, nil))

	page, total, err := repo.ListScrapedProducts(ctx, "job-1", 1, 50)
	require.NoError(t, err)
	assert.EqualValues(t, 2, total)
	assert.Len(t, page, 2)
}

func TestSqlite_ScrapedProductsPaginationAndClamping(t *testing.T) {
	repo := newExtRepoSqlite(t)
	ctx := context.Background()

	var batch []models.ScrapedProduct
	for i := 0; i < 5; i++ {
		batch = append(batch, models.ScrapedProduct{
			JobID:       "job-p",
			ProductName: "P",
			Link:        "https://shopee.co.id/p" + string(rune('a'+i)) + "-i.1.100",
		})
	}
	require.NoError(t, repo.InsertScrapedProducts(ctx, batch))

	p1, total, err := repo.ListScrapedProducts(ctx, "job-p", 1, 2)
	require.NoError(t, err)
	assert.EqualValues(t, 5, total)
	assert.Len(t, p1, 2)

	p3, _, err := repo.ListScrapedProducts(ctx, "job-p", 3, 2)
	require.NoError(t, err)
	assert.Len(t, p3, 1, "final page holds the remainder")

	// page=0 must normalise, not produce a negative offset.
	p0, _, err := repo.ListScrapedProducts(ctx, "job-p", 0, 2)
	require.NoError(t, err)
	assert.Len(t, p0, 2)

	// Oversized page size must be clamped rather than honoured.
	_, _, err = repo.ListScrapedProducts(ctx, "job-p", 1, 100000)
	require.NoError(t, err)
}

func TestSqlite_ScrapedProductsScopedByJob(t *testing.T) {
	repo := newExtRepoSqlite(t)
	ctx := context.Background()

	require.NoError(t, repo.InsertScrapedProducts(ctx, []models.ScrapedProduct{
		{JobID: "job-A", ProductName: "A", Link: "https://shopee.co.id/a-i.1.1"},
		{JobID: "job-B", ProductName: "B", Link: "https://shopee.co.id/b-i.1.2"},
	}))

	nA, err := repo.CountScrapedProducts(ctx, "job-A")
	require.NoError(t, err)
	assert.EqualValues(t, 1, nA, "counts must not leak across jobs")

	nMissing, err := repo.CountScrapedProducts(ctx, "job-none")
	require.NoError(t, err)
	assert.EqualValues(t, 0, nMissing)
}
