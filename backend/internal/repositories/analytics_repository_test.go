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

func TestAnalyticsRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t,
		&models.AnalyticsSettings{},
		&models.ShopeeEscrowSync{},
		&models.TiktokEscrowSync{},
		&models.ShopeeOrder{},
	)
	repo := NewAnalyticsRepository(db)
	ctx := context.Background()

	t.Run("GetSettings", func(t *testing.T) {
		tests := []struct {
			name     string
			tenantID string
			platform string
			wantNil  bool
		}{
			{
				name:     "returns nil for non-existent settings",
				tenantID: "tenant-nonexistent",
				platform: "shopee",
				wantNil:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				settings, err := repo.GetSettings(ctx, tt.tenantID, tt.platform)
				assert.NoError(t, err)
				if tt.wantNil {
					assert.Nil(t, settings)
				}
			})
		}
	})

	t.Run("CreateOrUpdateSettings", func(t *testing.T) {
		// Create new settings
		settings := &models.AnalyticsSettings{
			TenantID:          "tenant-settings-1",
			Platform:          "shopee",
			PriceColumn:       "HARGA",
			FormulaDeduction:  1500,
			FormulaMultiplier: 0.84,
		}

		err := repo.CreateOrUpdateSettings(ctx, settings)
		assert.NoError(t, err)
		assert.NotEmpty(t, settings.ID)

		// Verify created
		found, err := repo.GetSettings(ctx, "tenant-settings-1", "shopee")
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "HARGA", found.PriceColumn)
		assert.Equal(t, float64(1500), found.FormulaDeduction)

		// Update existing settings
		settings.PriceColumn = "PRICE"
		settings.FormulaDeduction = 2000
		err = repo.CreateOrUpdateSettings(ctx, settings)
		assert.NoError(t, err)

		// Verify updated
		updated, err := repo.GetSettings(ctx, "tenant-settings-1", "shopee")
		assert.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "PRICE", updated.PriceColumn)
		assert.Equal(t, float64(2000), updated.FormulaDeduction)
	})

	t.Run("GetShopeeEscrowSync", func(t *testing.T) {
		// Non-existent
		sync, err := repo.GetShopeeEscrowSync(ctx, "tenant-escrow-1", 1, 2024)
		assert.NoError(t, err)
		assert.Nil(t, sync)
	})

	t.Run("CreateOrUpdateShopeeEscrowSync", func(t *testing.T) {
		// Create
		err := repo.CreateOrUpdateShopeeEscrowSync(ctx, "tenant-escrow-2", 1, 2024, 100)
		assert.NoError(t, err)

		// Verify
		sync, err := repo.GetShopeeEscrowSync(ctx, "tenant-escrow-2", 1, 2024)
		assert.NoError(t, err)
		require.NotNil(t, sync)
		assert.Equal(t, 100, sync.TotalOrders)

		// Update
		err = repo.CreateOrUpdateShopeeEscrowSync(ctx, "tenant-escrow-2", 1, 2024, 200)
		assert.NoError(t, err)

		// Verify update
		updated, err := repo.GetShopeeEscrowSync(ctx, "tenant-escrow-2", 1, 2024)
		assert.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, 200, updated.TotalOrders)
	})

	t.Run("GetTiktokEscrowSync", func(t *testing.T) {
		sync, err := repo.GetTiktokEscrowSync(ctx, "tenant-tiktok-1", 2, 2024)
		assert.NoError(t, err)
		assert.Nil(t, sync)
	})

	t.Run("GetOrderCountByPlatform", func(t *testing.T) {
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()

		counts, err := repo.GetOrderCountByPlatform(ctx, "tenant-orders-1", startDate, endDate)
		assert.NoError(t, err)
		assert.Contains(t, counts, "shopee")
	})

	t.Run("GetTotalSalesByPlatform", func(t *testing.T) {
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()

		sales, err := repo.GetTotalSalesByPlatform(ctx, "tenant-sales-1", startDate, endDate)
		assert.NoError(t, err)
		assert.Contains(t, sales, "shopee")
	})
}
