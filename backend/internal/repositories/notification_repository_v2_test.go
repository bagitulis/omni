package repositories

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Phase A tests for the V2 notification repository behaviour: per-user reads,
// dedup UPSERT, bulk ops, counts endpoint, snooze filter.

func setupV2Repo(t *testing.T) (*NotificationRepository, context.Context) {
	db := testutils.SetupTestPostgresWithModels(t,
		&models.Notification{},
		&models.NotificationRead{},
		&models.NotificationSettings{},
	)
	// The application-level partial unique index for dedup is created by the
	// migration; mirror that here so tests exercise the real UPSERT path.
	require.NoError(t, db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS ix_notif_dedup_active
		  ON notifications (dedup_key)
		  WHERE dedup_key IS NOT NULL`).Error)
	return NewNotificationRepository(db), context.Background()
}

func TestNotificationRepositoryV2_UpsertDedup(t *testing.T) {
	repo, ctx := setupV2Repo(t)

	first, err := repo.UpsertDedup(ctx, &models.Notification{
		Type:      models.NotifTypeError,
		Category:  models.CatSync,
		Severity:  models.SeverityHigh,
		Title:     "Shopee Sync Failed",
		Message:   "boom",
		DedupKey:  strPtr("shopee_sync:tenant-a"),
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)
	require.NotNil(t, first)
	assert.Equal(t, 1, first.DedupCount)

	second, err := repo.UpsertDedup(ctx, &models.Notification{
		Type:      models.NotifTypeError,
		Category:  models.CatSync,
		Severity:  models.SeverityHigh,
		Title:     "Shopee Sync Failed",
		Message:   "boom again",
		DedupKey:  strPtr("shopee_sync:tenant-a"),
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID, "same dedup_key must reuse the same row")
	assert.Equal(t, 2, second.DedupCount)
	assert.Equal(t, "boom again", second.Message, "latest message wins")
}

func TestNotificationRepositoryV2_UpsertDedupRaceSafe(t *testing.T) {
	repo, ctx := setupV2Repo(t)
	key := "cron:refresh:tenant-a"

	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.UpsertDedup(ctx, &models.Notification{
				Type:      models.NotifTypeError,
				Category:  models.CatSystem,
				Severity:  models.SeverityMedium,
				Title:     "Credential Expiring",
				Message:   "less than 3 days",
				DedupKey:  strPtr(key),
				CreatedAt: time.Now(),
			})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	items, err := repo.ListActive(ctx, ListFilter{Limit: 50})
	require.NoError(t, err)
	matched := 0
	for _, it := range items {
		if it.DedupKey != nil && *it.DedupKey == key {
			matched++
			assert.Equal(t, 20, it.DedupCount)
		}
	}
	assert.Equal(t, 1, matched, "exactly one row per dedup key")
}

func TestNotificationRepositoryV2_PerUserReads(t *testing.T) {
	repo, ctx := setupV2Repo(t)

	n, err := repo.UpsertDedup(ctx, &models.Notification{
		Type:      models.NotifTypeInfo,
		Category:  models.CatOrder,
		Severity:  models.SeverityLow,
		Title:     "Order updated",
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	// User A marks read; user B still unread.
	require.NoError(t, repo.MarkReadForUser(ctx, n.ID, "user-a"))

	countA, err := repo.CountsForUser(ctx, "user-a")
	require.NoError(t, err)
	assert.Equal(t, int64(0), countA.Unread)

	countB, err := repo.CountsForUser(ctx, "user-b")
	require.NoError(t, err)
	assert.Equal(t, int64(1), countB.Unread)
}

func TestNotificationRepositoryV2_MarkAllForUser(t *testing.T) {
	repo, ctx := setupV2Repo(t)

	for i := range 3 {
		_, err := repo.UpsertDedup(ctx, &models.Notification{
			Type:      models.NotifTypeInfo,
			Category:  models.CatSystem,
			Severity:  models.SeverityLow,
			Title:     "n",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Millisecond),
		})
		require.NoError(t, err)
	}

	affected, err := repo.MarkAllReadForUser(ctx, "u42")
	require.NoError(t, err)
	assert.Equal(t, int64(3), affected)

	c, err := repo.CountsForUser(ctx, "u42")
	require.NoError(t, err)
	assert.Equal(t, int64(0), c.Unread)
}

func TestNotificationRepositoryV2_BulkReadReturnsAffected(t *testing.T) {
	repo, ctx := setupV2Repo(t)

	ids := make([]int64, 0, 3)
	for range 3 {
		n, err := repo.UpsertDedup(ctx, &models.Notification{
			Type:     models.NotifTypeInfo,
			Category: models.CatSystem,
			Severity: models.SeverityLow,
			Title:    "n",
		})
		require.NoError(t, err)
		ids = append(ids, n.ID)
	}
	ids = append(ids, 9_999_999) // non-existent

	affected, err := repo.BulkMarkReadForUser(ctx, "u7", ids)
	require.NoError(t, err)
	assert.Equal(t, int64(3), affected, "only real IDs count")
}

func TestNotificationRepositoryV2_BulkDeleteReturnsAffected(t *testing.T) {
	repo, ctx := setupV2Repo(t)

	ids := make([]int64, 0, 2)
	for range 2 {
		n, err := repo.UpsertDedup(ctx, &models.Notification{
			Type:     models.NotifTypeInfo,
			Category: models.CatSystem,
			Severity: models.SeverityLow,
			Title:    "n",
		})
		require.NoError(t, err)
		ids = append(ids, n.ID)
	}
	affected, err := repo.BulkDelete(ctx, ids)
	require.NoError(t, err)
	assert.Equal(t, int64(2), affected)
}

func TestNotificationRepositoryV2_CountsBySeverity(t *testing.T) {
	repo, ctx := setupV2Repo(t)

	seed := []int16{
		models.SeverityCritical,
		models.SeverityCritical,
		models.SeverityHigh,
		models.SeverityLow,
	}
	for i, sev := range seed {
		_, err := repo.UpsertDedup(ctx, &models.Notification{
			Type:     models.NotifTypeInfo,
			Category: models.CatSystem,
			Severity: sev,
			Title:    "n",
			// Different dedup keys per row to force distinct records.
			DedupKey: strPtr("sev-test-" + time.Now().Format("150405.000000") + "-" + strconv.Itoa(i)),
		})
		require.NoError(t, err)
	}

	c, err := repo.CountsForUser(ctx, "u1")
	require.NoError(t, err)
	assert.Equal(t, int64(4), c.Total)
	assert.Equal(t, int64(4), c.Unread)
	assert.Equal(t, int64(2), c.BySeverity[models.SeverityCritical])
	assert.Equal(t, int64(1), c.BySeverity[models.SeverityHigh])
	assert.Equal(t, int64(1), c.BySeverity[models.SeverityLow])
}

func TestNotificationRepositoryV2_SnoozeHidesFromUnread(t *testing.T) {
	repo, ctx := setupV2Repo(t)

	n, err := repo.UpsertDedup(ctx, &models.Notification{
		Type:     models.NotifTypeWarning,
		Category: models.CatSystem,
		Severity: models.SeverityMedium,
		Title:    "snoozable",
	})
	require.NoError(t, err)

	// Snooze for 1 hour.
	until := time.Now().Add(time.Hour)
	require.NoError(t, repo.Snooze(ctx, n.ID, until))

	items, err := repo.ListActive(ctx, ListFilter{Limit: 50, UnreadOnly: true, UserID: "u1"})
	require.NoError(t, err)
	for _, it := range items {
		assert.NotEqual(t, n.ID, it.ID, "snoozed notif must not appear in unread")
	}
}

func TestNotificationRepositoryV2_SearchAndFilter(t *testing.T) {
	repo, ctx := setupV2Repo(t)

	_, _ = repo.UpsertDedup(ctx, &models.Notification{
		Type: models.NotifTypeError, Category: models.CatSync, Severity: models.SeverityHigh,
		Title: "Shopee Sync Failed", Message: "network error"})
	_, _ = repo.UpsertDedup(ctx, &models.Notification{
		Type: models.NotifTypeInfo, Category: models.CatOrder, Severity: models.SeverityLow,
		Title: "Order arrived", Message: "hello"})

	// Search by term in title.
	items, err := repo.ListActive(ctx, ListFilter{Limit: 50, UserID: "u1", Search: "shopee"})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.True(t, strings.Contains(strings.ToLower(items[0].Title), "shopee"))

	// Filter by category.
	items, err = repo.ListActive(ctx, ListFilter{Limit: 50, UserID: "u1", Category: models.CatOrder})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, models.CatOrder, items[0].Category)

	// Filter by min severity.
	items, err = repo.ListActive(ctx, ListFilter{Limit: 50, UserID: "u1", MinSeverity: models.SeverityHigh})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, models.NotifTypeError, items[0].Type)
}

// Small helper scoped to this test file.
func strPtr(s string) *string { return &s }
