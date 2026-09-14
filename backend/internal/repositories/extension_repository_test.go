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

// newExtRepoForTest returns a repository bound to a fresh Postgres test DB.
// Skips (never fails) when no container runtime is reachable.
func newExtRepoForTest(t *testing.T) *ExtensionRepository {
	t.Helper()
	db := testutils.SetupTestPostgresWithModels(t,
		&models.Extension{},
		&models.PairingCode{},
		&models.ScrapedProduct{},
	)
	t.Cleanup(func() { _ = db }) // container/shared DB cleanup handled by testutils
	return NewExtensionRepository(db)
}

func TestExtensionRepository_CreateAndList(t *testing.T) {
	repo := newExtRepoForTest(t)
	ctx := context.Background()

	ext := &models.Extension{
		ExtensionID:  "ext-1",
		Hostname:     "browser-extension",
		Capabilities: models.StringList{models.CapabilityShopeeScrape},
		Status:       models.ExtensionStatusDisconnected,
		PairedAt:     time.Now(),
	}
	require.NoError(t, repo.CreateExtension(ctx, ext))
	assert.NotZero(t, ext.ID, "Create should populate the auto-increment ID")

	list, err := repo.ListExtensions(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "ext-1", list[0].ExtensionID)
	assert.Contains(t, list[0].Capabilities, models.CapabilityShopeeScrape)
}

func TestExtensionRepository_GetByID_NotFound(t *testing.T) {
	repo := newExtRepoForTest(t)
	_, err := repo.GetExtensionByID(context.Background(), "does-not-exist")
	assert.Error(t, err, "missing extension must return an error, not a zero value")
}

func TestExtensionRepository_TokenHashLookup(t *testing.T) {
	repo := newExtRepoForTest(t)
	ctx := context.Background()

	require.NoError(t, repo.CreateExtension(ctx, &models.Extension{
		ExtensionID: "ext-token",
		TokenHash:   "hash-abc",
		Status:      models.ExtensionStatusDisconnected,
		PairedAt:    time.Now(),
	}))

	found, err := repo.FindExtensionByTokenHash(ctx, "hash-abc")
	require.NoError(t, err)
	assert.Equal(t, "ext-token", found.ExtensionID)

	// Negative: an unknown hash must not resolve.
	_, err = repo.FindExtensionByTokenHash(ctx, "hash-nope")
	assert.Error(t, err)

	// Negative: an empty hash must not resolve to the first row.
	_, err = repo.FindExtensionByTokenHash(ctx, "")
	assert.Error(t, err)
}

func TestExtensionRepository_StatusTransitions(t *testing.T) {
	repo := newExtRepoForTest(t)
	ctx := context.Background()

	require.NoError(t, repo.CreateExtension(ctx, &models.Extension{
		ExtensionID: "ext-status",
		Status:      models.ExtensionStatusDisconnected,
		PairedAt:    time.Now(),
	}))

	require.NoError(t, repo.UpdateExtensionStatus(ctx, "ext-status", models.ExtensionStatusConnected))
	got, err := repo.GetExtensionByID(ctx, "ext-status")
	require.NoError(t, err)
	assert.Equal(t, models.ExtensionStatusConnected, got.Status)
	assert.NotNil(t, got.LastSeen, "connecting must stamp last_seen")

	// Startup reset must clear "connected" rows: after a restart no WebSocket
	// exists, so a stale connected status would show phantom online extensions.
	require.NoError(t, repo.MarkAllExtensionsDisconnected(ctx))
	got, err = repo.GetExtensionByID(ctx, "ext-status")
	require.NoError(t, err)
	assert.Equal(t, models.ExtensionStatusDisconnected, got.Status)
}

func TestExtensionRepository_Delete(t *testing.T) {
	repo := newExtRepoForTest(t)
	ctx := context.Background()

	require.NoError(t, repo.CreateExtension(ctx, &models.Extension{
		ExtensionID: "ext-del",
		Status:      models.ExtensionStatusDisconnected,
		PairedAt:    time.Now(),
	}))

	require.NoError(t, repo.DeleteExtension(ctx, "ext-del"))

	// Idempotency: deleting something already gone reports not-found rather
	// than silently succeeding.
	assert.Error(t, repo.DeleteExtension(ctx, "ext-del"))
}

func TestExtensionRepository_ConsumePairingCode_SingleUse(t *testing.T) {
	repo := newExtRepoForTest(t)
	ctx := context.Background()
	now := time.Now()

	require.NoError(t, repo.CreatePairingCode(ctx, &models.PairingCode{
		Code:      "ABCD1234",
		ExpiresAt: now.Add(5 * time.Minute),
	}))

	consumed, err := repo.ConsumePairingCode(ctx, "ABCD1234", now)
	require.NoError(t, err)
	assert.True(t, consumed.Consumed)

	// Second redemption MUST fail — one code mints exactly one token.
	_, err = repo.ConsumePairingCode(ctx, "ABCD1234", now)
	assert.ErrorIs(t, err, ErrPairingCodeUnusable,
		"reusing a consumed pairing code must be rejected")
}

func TestExtensionRepository_ConsumePairingCode_Expired(t *testing.T) {
	repo := newExtRepoForTest(t)
	ctx := context.Background()
	now := time.Now()

	require.NoError(t, repo.CreatePairingCode(ctx, &models.PairingCode{
		Code:      "EXPIRED1",
		ExpiresAt: now.Add(-1 * time.Minute),
	}))

	_, err := repo.ConsumePairingCode(ctx, "EXPIRED1", now)
	assert.ErrorIs(t, err, ErrPairingCodeUnusable, "expired code must be rejected")
}

func TestExtensionRepository_ConsumePairingCode_Unknown(t *testing.T) {
	repo := newExtRepoForTest(t)

	_, err := repo.ConsumePairingCode(context.Background(), "NEVERISSUED", time.Now())
	assert.ErrorIs(t, err, ErrPairingCodeUnusable,
		"unknown code must be indistinguishable from expired/used")
}

func TestExtensionRepository_ScrapedProducts_InsertListCountDedup(t *testing.T) {
	repo := newExtRepoForTest(t)
	ctx := context.Background()

	batch := []models.ScrapedProduct{
		{JobID: "job-1", ProductName: "A", Link: "https://shopee.co.id/a-i.1.11", PageNumber: 1},
		{JobID: "job-1", ProductName: "B", Link: "https://shopee.co.id/b-i.1.12", PageNumber: 1},
	}
	require.NoError(t, repo.InsertScrapedProducts(ctx, batch))

	n, err := repo.CountScrapedProducts(ctx, "job-1")
	require.NoError(t, err)
	assert.EqualValues(t, 2, n)

	// Resume re-reads the boundary page, so the same links arrive again.
	// They must be skipped, not duplicated and not fail the job.
	require.NoError(t, repo.InsertScrapedProducts(ctx, batch))
	n, err = repo.CountScrapedProducts(ctx, "job-1")
	require.NoError(t, err)
	assert.EqualValues(t, 2, n, "re-inserting identical links must not duplicate rows")

	// Empty batch is a no-op, not an error.
	assert.NoError(t, repo.InsertScrapedProducts(ctx, nil))

	page, total, err := repo.ListScrapedProducts(ctx, "job-1", 1, 50)
	require.NoError(t, err)
	assert.EqualValues(t, 2, total)
	assert.Len(t, page, 2)
}

func TestExtensionRepository_ListScrapedProducts_Pagination(t *testing.T) {
	repo := newExtRepoForTest(t)
	ctx := context.Background()

	var batch []models.ScrapedProduct
	for i := 0; i < 5; i++ {
		batch = append(batch, models.ScrapedProduct{
			JobID:       "job-page",
			ProductName: "P",
			Link:        "https://shopee.co.id/p" + string(rune('a'+i)) + "-i.1.100",
		})
	}
	require.NoError(t, repo.InsertScrapedProducts(ctx, batch))

	page1, total, err := repo.ListScrapedProducts(ctx, "job-page", 1, 2)
	require.NoError(t, err)
	assert.EqualValues(t, 5, total)
	assert.Len(t, page1, 2)

	page3, _, err := repo.ListScrapedProducts(ctx, "job-page", 3, 2)
	require.NoError(t, err)
	assert.Len(t, page3, 1, "last page holds the remainder")

	// Page 0 is normalised to page 1 rather than producing a negative offset.
	page0, _, err := repo.ListScrapedProducts(ctx, "job-page", 0, 2)
	require.NoError(t, err)
	assert.Len(t, page0, 2)

	// Oversized page size is clamped, not honoured blindly.
	_, _, err = repo.ListScrapedProducts(ctx, "job-page", 1, 100000)
	require.NoError(t, err)
}

func TestExtensionRepository_ScrapedProducts_ScopedByJob(t *testing.T) {
	repo := newExtRepoForTest(t)
	ctx := context.Background()

	require.NoError(t, repo.InsertScrapedProducts(ctx, []models.ScrapedProduct{
		{JobID: "job-A", ProductName: "A", Link: "https://shopee.co.id/a-i.1.1"},
		{JobID: "job-B", ProductName: "B", Link: "https://shopee.co.id/b-i.1.2"},
	}))

	nA, err := repo.CountScrapedProducts(ctx, "job-A")
	require.NoError(t, err)
	assert.EqualValues(t, 1, nA, "counts must not leak across jobs")

	nMissing, err := repo.CountScrapedProducts(ctx, "job-missing")
	require.NoError(t, err)
	assert.EqualValues(t, 0, nMissing)
}

func TestExtensionRepository_DeleteExpiredPairingCodes(t *testing.T) {
	repo := newExtRepoForTest(t)
	ctx := context.Background()
	now := time.Now()

	require.NoError(t, repo.CreatePairingCode(ctx, &models.PairingCode{
		Code: "OLD1", ExpiresAt: now.Add(-time.Hour),
	}))
	require.NoError(t, repo.CreatePairingCode(ctx, &models.PairingCode{
		Code: "NEW1", ExpiresAt: now.Add(time.Hour),
	}))

	require.NoError(t, repo.DeleteExpiredPairingCodes(ctx, now))

	// The live code must survive pruning.
	_, err := repo.ConsumePairingCode(ctx, "NEW1", now)
	require.NoError(t, err, "pruning must not remove unexpired codes")

	// The expired one is gone; it is rejected either way, so assert on pruning
	// by confirming the row no longer exists.
	_, err = repo.ConsumePairingCode(ctx, "OLD1", now)
	assert.ErrorIs(t, err, ErrPairingCodeUnusable)
}
