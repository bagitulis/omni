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

func TestWebhookRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t,
		&models.WebhookLog{},
		&models.WebhookOrderEvent{},
		&models.WebhookProductEvent{},
		&models.WebhookReturnEvent{},
	)
	repo := NewWebhookRepository(db)
	ctx := context.Background()

	t.Run("CreateLog", func(t *testing.T) {
		payload := map[string]interface{}{"order_sn": "123456"}
		headers := map[string]interface{}{"Authorization": "Bearer token"}

		log, err := repo.CreateLog(ctx, "tenant-webhook-1", "shopee", "order.status", payload, headers)
		assert.NoError(t, err)
		require.NotNil(t, log)
		assert.NotEmpty(t, log.ID)
		assert.Equal(t, "shopee", log.Platform)
		assert.Equal(t, "order.status", log.EventType)
		assert.Equal(t, models.WebhookStatusReceived, log.Status)
	})

	t.Run("UpdateLogStatus", func(t *testing.T) {
		log, err := repo.CreateLog(ctx, "tenant-webhook-2", "lazada", "order.create", nil, nil)
		require.NoError(t, err)

		// Update to processed
		err = repo.UpdateLogStatus(ctx, log.ID, models.WebhookStatusProcessed, "")
		assert.NoError(t, err)

		// Verify
		var updated models.WebhookLog
		db.Where("id = ?", log.ID).First(&updated)
		assert.Equal(t, models.WebhookStatusProcessed, updated.Status)
		assert.NotNil(t, updated.ProcessedAt)

		// Update to failed
		log2, _ := repo.CreateLog(ctx, "tenant-webhook-3", "tiktok", "order.fail", nil, nil)
		err = repo.UpdateLogStatus(ctx, log2.ID, models.WebhookStatusFailed, "Processing error")
		assert.NoError(t, err)

		var failed models.WebhookLog
		db.Where("id = ?", log2.ID).First(&failed)
		assert.Equal(t, models.WebhookStatusFailed, failed.Status)
		assert.Equal(t, "Processing error", failed.ErrorMsg)
	})

	t.Run("CreateOrderEvent", func(t *testing.T) {
		event := &models.WebhookOrderEvent{
			TenantID:  "tenant-event-1",
			Platform:  "shopee",
			EventType: "order.status.update",
			OrderSN:   "ORDER123",
			ShopID:    "shop-456",
			OldStatus: "pending",
			NewStatus: "shipped",
			Payload:   `{"test": "data"}`,
		}

		err := repo.CreateOrderEvent(ctx, event)
		assert.NoError(t, err)
		assert.NotZero(t, event.ID)
		assert.False(t, event.CreatedAt.IsZero())
	})

	t.Run("CreateProductEvent", func(t *testing.T) {
		event := &models.WebhookProductEvent{
			TenantID:  "tenant-product-1",
			EventType: "product.update",
			ShopID:    "shop-789",
			ItemID:    "item-123",
			Action:    "update",
		}

		err := repo.CreateProductEvent(ctx, event)
		assert.NoError(t, err)
		assert.NotZero(t, event.ID)
	})

	t.Run("CreateReturnEvent", func(t *testing.T) {
		event := &models.WebhookReturnEvent{
			TenantID:  "tenant-return-1",
			EventType: "return.create",
			ShopID:    "shop-999",
			OrderSN:   "ORDER999",
			ReturnSN:  "RETURN123",
			Status:    "pending",
			Reason:    "Defective product",
		}

		err := repo.CreateReturnEvent(ctx, event)
		assert.NoError(t, err)
		assert.NotZero(t, event.ID)
	})

	t.Run("FindLogsByTenant", func(t *testing.T) {
		tenantID := "tenant-findlogs-1"
		for i := 0; i < 3; i++ {
			_, err := repo.CreateLog(ctx, tenantID, "shopee", "order.test", nil, nil)
			require.NoError(t, err)
		}

		logs, total, err := repo.FindLogsByTenant(ctx, tenantID, 10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(3))
		assert.GreaterOrEqual(t, len(logs), 3)
	})

	t.Run("FindOrderEventsByOrder", func(t *testing.T) {
		tenantID := "tenant-findorder-1"
		orderSN := "ORDER-FIND-123"
		event := &models.WebhookOrderEvent{
			TenantID:  tenantID,
			Platform:  "shopee",
			EventType: "order.status",
			OrderSN:   orderSN,
		}
		err := repo.CreateOrderEvent(ctx, event)
		require.NoError(t, err)

		events, err := repo.FindOrderEventsByOrder(ctx, tenantID, orderSN)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(events), 1)
		assert.Equal(t, orderSN, events[0].OrderSN)
	})

	t.Run("FindOrderEventsByPlatform", func(t *testing.T) {
		tenantID := "tenant-findplatform-1"
		event := &models.WebhookOrderEvent{
			TenantID:  tenantID,
			Platform:  "tiktok",
			EventType: "order.create",
			OrderSN:   "TIKTOK-ORDER-1",
		}
		err := repo.CreateOrderEvent(ctx, event)
		require.NoError(t, err)

		events, total, err := repo.FindOrderEventsByPlatform(ctx, tenantID, "tiktok", 10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))
		assert.GreaterOrEqual(t, len(events), 1)
	})

	t.Run("FindRecentOrderEvents", func(t *testing.T) {
		tenantID := "tenant-recent-1"
		event := &models.WebhookOrderEvent{
			TenantID:  tenantID,
			Platform:  "lazada",
			EventType: "order.ship",
			OrderSN:   "RECENT-ORDER-1",
		}
		err := repo.CreateOrderEvent(ctx, event)
		require.NoError(t, err)

		since := time.Now().Add(-1 * time.Hour)
		events, err := repo.FindRecentOrderEvents(ctx, tenantID, since)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(events), 1)
	})

	t.Run("DeleteOldLogs", func(t *testing.T) {
		// Create old log
		oldLog := &models.WebhookLog{
			ID:        "old-log-id",
			TenantID:  "tenant-oldlog",
			Platform:  "shopee",
			EventType: "old.event",
			Status:    models.WebhookStatusProcessed,
			CreatedAt: time.Now().Add(-48 * time.Hour), // 2 days ago
		}
		err := db.WithContext(ctx).Create(oldLog).Error
		require.NoError(t, err)

		// Delete logs older than 1 hour
		err = repo.DeleteOldLogs(ctx, 1*time.Hour)
		assert.NoError(t, err)

		// Verify old log is deleted
		var count int64
		db.Model(&models.WebhookLog{}).Where("id = ?", "old-log-id").Count(&count)
		assert.Equal(t, int64(0), count)
	})
}
