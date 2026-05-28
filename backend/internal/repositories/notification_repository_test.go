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

func TestNotificationRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t,
		&models.Notification{},
		&models.NotificationSettings{},
	)
	repo := NewNotificationRepository(db)
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
			notif := &models.Notification{
				Type:      "success",
				Category:  "sync",
				Title:     "Test Notification",
				Message:   "This is a test notification",
				ActionURL: "",
				Metadata:  "{}",
				Read:      false,
				CreatedAt: time.Now(),
			}

		err := repo.Create(ctx, notif)
		assert.NoError(t, err)
		assert.NotZero(t, notif.ID)
	})

	t.Run("List", func(t *testing.T) {
		// Create a few notifications
		for range 3 {
				notif := &models.Notification{
					Type:      "info",
					Category:  "order",
					Title:     "Order update",
					Message:   "Order updated notification",
					Metadata:  "{}",
					Read:      false,
					CreatedAt: time.Now(),
				}
			err := repo.Create(ctx, notif)
			require.NoError(t, err)
		}

		items, err := repo.List(ctx, 10, 0, false)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(items), 3)
	})

	t.Run("ListUnreadOnly", func(t *testing.T) {
		// Create a read notification and an unread one
		readNotif := &models.Notification{
			Type:      "info",
			Category:  "test",
			Title:     "Read notification",
			Message:   "Already read",
			Metadata:  "{}",
			Read:      true,
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, readNotif)
		require.NoError(t, err)

		unreadNotif := &models.Notification{
			Type:      "warning",
			Category:  "test",
			Title:     "Unread notification",
			Message:   "Not yet read",
			Metadata:  "{}",
			Read:      false,
			CreatedAt: time.Now(),
		}
		err = repo.Create(ctx, unreadNotif)
		require.NoError(t, err)

		items, err := repo.List(ctx, 10, 0, true)
		assert.NoError(t, err)
		for _, item := range items {
			assert.False(t, item.Read, "all returned items should be unread")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		notif := &models.Notification{
			Type:      "error",
			Category:  "product",
			Title:     "GetByID test",
			Message:   "Testing GetByID",
			Metadata:  "{}",
			Read:      false,
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, notif)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, notif.ID)
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, notif.Title, found.Title)

		// Non-existent ID
		_, err = repo.GetByID(ctx, 999999)
		assert.Error(t, err)
	})

	t.Run("UnreadCount", func(t *testing.T) {
		// Create an unread notification
		notif := &models.Notification{
			Type:      "info",
			Category:  "test",
			Title:     "Unread count test",
			Message:   "Testing unread count",
			Metadata:  "{}",
			Read:      false,
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, notif)
		require.NoError(t, err)

		count, err := repo.UnreadCount(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})

	t.Run("MarkAsRead", func(t *testing.T) {
		notif := &models.Notification{
			Type:      "success",
			Category:  "sync",
			Title:     "MarkAsRead test",
			Message:   "Will be marked as read",
			Metadata:  "{}",
			Read:      false,
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, notif)
		require.NoError(t, err)

		err = repo.MarkAsRead(ctx, notif.ID)
		assert.NoError(t, err)

		found, err := repo.GetByID(ctx, notif.ID)
		require.NoError(t, err)
		assert.True(t, found.Read)
	})

	t.Run("MarkAllAsRead", func(t *testing.T) {
		// Create two unread notifications
		for range 2 {
			notif := &models.Notification{
				Type:      "warning",
				Category:  "test",
				Title:     "MarkAll test",
				Message:   "Will be marked all as read",
				Metadata:  "{}",
				Read:      false,
				CreatedAt: time.Now(),
			}
			err := repo.Create(ctx, notif)
			require.NoError(t, err)
		}

		err := repo.MarkAllAsRead(ctx)
		assert.NoError(t, err)

		unreadCount, err := repo.UnreadCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), unreadCount)
	})

	t.Run("Delete", func(t *testing.T) {
		notif := &models.Notification{
			Type:      "info",
			Category:  "test",
			Title:     "Delete test",
			Message:   "Will be deleted",
			Metadata:  "{}",
			Read:      false,
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, notif)
		require.NoError(t, err)

		err = repo.Delete(ctx, notif.ID)
		assert.NoError(t, err)

		_, err = repo.GetByID(ctx, notif.ID)
		assert.Error(t, err)
	})

	t.Run("DeleteAll", func(t *testing.T) {
		// Create a notification
		notif := &models.Notification{
			Type:      "info",
			Category:  "test",
			Title:     "DeleteAll test",
			Message:   "Will be deleted with all",
			Metadata:  "{}",
			Read:      false,
			CreatedAt: time.Now(),
		}
		err := repo.Create(ctx, notif)
		require.NoError(t, err)

		err = repo.DeleteAll(ctx)
		assert.NoError(t, err)

		items, err := repo.List(ctx, 10, 0, false)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(items))
	})

	t.Run("DeleteAll_TenantScoped", func(t *testing.T) {
		// Create a second DB instance simulating a different tenant
		// (tenant isolation is schema-based, so separate DB = separate tenant)
		db2 := testutils.SetupTestPostgresWithModels(t,
			&models.Notification{},
			&models.NotificationSettings{},
		)
		repoA := NewNotificationRepository(db)
		repoB := NewNotificationRepository(db2)

		// Seed a notification in tenant A
		notifA := &models.Notification{
			Type:      "info",
			Category:  "tenant-a",
			Title:     "Tenant A notification",
			Message:   "Should be deleted by DeleteAll",
			Metadata:  "{}",
			Read:      false,
			CreatedAt: time.Now(),
		}
		err := repoA.Create(ctx, notifA)
		require.NoError(t, err)

		// Seed a notification in tenant B
		notifB := &models.Notification{
			Type:      "info",
			Category:  "tenant-b",
			Title:     "Tenant B notification",
			Message:   "Should survive DeleteAll on tenant A",
			Metadata:  "{}",
			Read:      false,
			CreatedAt: time.Now(),
		}
		err = repoB.Create(ctx, notifB)
		require.NoError(t, err)

		// Call DeleteAll on tenant A only
		err = repoA.DeleteAll(ctx)
		assert.NoError(t, err)

		// Verify tenant A has 0 notifications
		itemsA, err := repoA.List(ctx, 10, 0, false)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(itemsA), "Tenant A should have 0 notifications after DeleteAll")

		// Verify tenant B still has its notifications (tenant isolation)
		itemsB, err := repoB.List(ctx, 10, 0, false)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(itemsB), "Tenant B should still have 1 notification")
		assert.Equal(t, "Tenant B notification", itemsB[0].Title, "Tenant B notification title should match")
	})

	t.Run("GetSettings", func(t *testing.T) {
		settings, err := repo.GetSettings(ctx)
		assert.NoError(t, err)
		require.NotNil(t, settings)
		assert.Equal(t, 30, settings.RetentionDays) // default
	})

	t.Run("SaveSettings", func(t *testing.T) {
		err := repo.SaveSettings(ctx, 90)
		assert.NoError(t, err)

		settings, err := repo.GetSettings(ctx)
		assert.NoError(t, err)
		require.NotNil(t, settings)
		assert.Equal(t, 90, settings.RetentionDays)
	})

	t.Run("CleanupOlderThan", func(t *testing.T) {
		// Create an old notification
		oldNotif := &models.Notification{
			Type:      "info",
			Category:  "cleanup",
			Title:     "Old notification",
			Message:   "This should be cleaned up",
			Metadata:  "{}",
			Read:      false,
			CreatedAt: time.Now().Add(-48 * time.Hour), // 2 days ago
		}
		err := repo.Create(ctx, oldNotif)
		require.NoError(t, err)

		rows, err := repo.CleanupOlderThan(ctx, 1) // cleanup older than 1 day
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, rows, int64(1))
	})
}
