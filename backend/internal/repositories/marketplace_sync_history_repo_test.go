package repositories

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarketplaceSyncHistoryRepo(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.MarketplaceSyncHistory{})
	repo := NewMarketplaceSyncHistoryRepo(db)
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		entry := &models.MarketplaceSyncHistory{
			TenantID:  "tenant-create-1",
			SKU:       "SKU-TEST-001",
			Platform:  "shopee",
			Operation: "stock_update",
			Status:    "success",
		}

		err := repo.Create(ctx, entry)
		require.NoError(t, err)
		assert.NotEmpty(t, entry.ID, "ID should be auto-generated")
		assert.False(t, entry.CreatedAt.IsZero(), "CreatedAt should be set")
	})

	t.Run("Create_WithOptionalFields", func(t *testing.T) {
		reqData := `{"stock": 100}`
		respData := `{"success": true}`
		errMsg := "partial update failed for variant 2"

		entry := &models.MarketplaceSyncHistory{
			TenantID:     "tenant-create-2",
			SKU:          "SKU-TEST-002",
			Platform:     "lazada",
			Operation:    "price_update",
			Status:       "partial",
			RequestData:  &reqData,
			ResponseData: &respData,
			ErrorMessage: &errMsg,
		}

		err := repo.Create(ctx, entry)
		require.NoError(t, err)
		assert.NotEmpty(t, entry.ID)
		assert.Equal(t, &reqData, entry.RequestData)
		assert.Equal(t, &respData, entry.ResponseData)
		assert.Equal(t, &errMsg, entry.ErrorMessage)
	})

	t.Run("CreateBatch", func(t *testing.T) {
		entries := []models.MarketplaceSyncHistory{
			{
				TenantID:  "tenant-batch-1",
				SKU:       "SKU-BATCH-001",
				Platform:  "shopee",
				Operation: "stock_update",
				Status:    "success",
			},
			{
				TenantID:  "tenant-batch-1",
				SKU:       "SKU-BATCH-002",
				Platform:  "shopee",
				Operation: "stock_update",
				Status:    "failed",
			},
		}

		err := repo.CreateBatch(ctx, entries)
		require.NoError(t, err)
		for i := range entries {
			assert.NotEmpty(t, entries[i].ID, "ID should be auto-generated for entry %d", i)
		}
	})

	t.Run("CreateBatch_Empty", func(t *testing.T) {
		err := repo.CreateBatch(ctx, []models.MarketplaceSyncHistory{})
		assert.NoError(t, err, "Empty batch should not error")
	})

	t.Run("List_BasicPagination", func(t *testing.T) {
		tenantID := "tenant-list-1"
		// Create 5 entries
		for i := 0; i < 5; i++ {
			entry := &models.MarketplaceSyncHistory{
				TenantID:  tenantID,
				SKU:       "SKU-LIST-001",
				Platform:  "tiktok",
				Operation: "wholesale_update",
				Status:    "success",
			}
			err := repo.Create(ctx, entry)
			require.NoError(t, err)
		}

		filter := models.MarketplaceSyncHistoryFilter{
			TenantID: tenantID,
			Page:     1,
			PageSize: 3,
		}

		result, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
		assert.Len(t, result.Entries, 3)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 3, result.PageSize)
	})

	t.Run("List_FilterByPlatform", func(t *testing.T) {
		tenantID := "tenant-filter-platform"
		for _, platform := range []string{"shopee", "lazada", "shopee"} {
			err := repo.Create(ctx, &models.MarketplaceSyncHistory{
				TenantID:  tenantID,
				SKU:       "SKU-PLAT-001",
				Platform:  platform,
				Operation: "stock_update",
				Status:    "success",
			})
			require.NoError(t, err)
		}

		filter := models.MarketplaceSyncHistoryFilter{
			TenantID: tenantID,
			Platform: "shopee",
			Page:     1,
			PageSize: 20,
		}

		result, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
		for _, e := range result.Entries {
			assert.Equal(t, "shopee", e.Platform)
		}
	})

	t.Run("List_FilterByOperation", func(t *testing.T) {
		tenantID := "tenant-filter-op"
		for _, op := range []string{"stock_update", "price_update", "stock_update"} {
			err := repo.Create(ctx, &models.MarketplaceSyncHistory{
				TenantID:  tenantID,
				SKU:       "SKU-OP-001",
				Platform:  "shopee",
				Operation: op,
				Status:    "success",
			})
			require.NoError(t, err)
		}

		filter := models.MarketplaceSyncHistoryFilter{
			TenantID:  tenantID,
			Operation: "price_update",
			Page:      1,
			PageSize:  20,
		}

		result, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "price_update", result.Entries[0].Operation)
	})

	t.Run("List_FilterByStatus", func(t *testing.T) {
		tenantID := "tenant-filter-status"
		for _, status := range []string{"success", "failed", "success"} {
			err := repo.Create(ctx, &models.MarketplaceSyncHistory{
				TenantID:  tenantID,
				SKU:       "SKU-STAT-001",
				Platform:  "tiktok",
				Operation: "mpq_update",
				Status:    status,
			})
			require.NoError(t, err)
		}

		filter := models.MarketplaceSyncHistoryFilter{
			TenantID: tenantID,
			Status:   "failed",
			Page:     1,
			PageSize: 20,
		}

		result, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "failed", result.Entries[0].Status)
	})

	t.Run("List_FilterBySKUSearch", func(t *testing.T) {
		tenantID := "tenant-filter-sku"
		for _, sku := range []string{"ABC-001", "ABC-002", "XYZ-001"} {
			err := repo.Create(ctx, &models.MarketplaceSyncHistory{
				TenantID:  tenantID,
				SKU:       sku,
				Platform:  "shopee",
				Operation: "clone",
				Status:    "success",
			})
			require.NoError(t, err)
		}

		filter := models.MarketplaceSyncHistoryFilter{
			TenantID:  tenantID,
			SKUSearch: "ABC",
			Page:      1,
			PageSize:  20,
		}

		result, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
		for _, e := range result.Entries {
			assert.Contains(t, e.SKU, "ABC")
		}
	})

	t.Run("List_TenantIsolation", func(t *testing.T) {
		// Entries in different tenants should not leak
		for _, tid := range []string{"tenant-iso-A", "tenant-iso-B"} {
			err := repo.Create(ctx, &models.MarketplaceSyncHistory{
				TenantID:  tid,
				SKU:       "SKU-ISO-001",
				Platform:  "shopee",
				Operation: "stock_update",
				Status:    "success",
			})
			require.NoError(t, err)
		}

		filter := models.MarketplaceSyncHistoryFilter{
			TenantID: "tenant-iso-A",
			Page:     1,
			PageSize: 20,
		}

		result, err := repo.List(ctx, filter)
		require.NoError(t, err)
		for _, e := range result.Entries {
			assert.Equal(t, "tenant-iso-A", e.TenantID, "Should not see other tenant's data")
		}
	})
}
