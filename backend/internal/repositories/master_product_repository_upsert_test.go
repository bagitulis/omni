package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMasterProductRepository_UpsertPlatformLink_ByPlatformProductID(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t,
		&models.MasterProduct{},
		&models.MasterProductSku{},
		&models.MasterProductPlatformLink{},
	)
	repo := NewMasterProductRepository(db)
	ctx := context.Background()

	tableName := models.MasterProductPlatformLink{}.TableName()
	err := db.Exec(fmt.Sprintf(
	"CREATE UNIQUE INDEX IF NOT EXISTS idx_platform_links_unique ON %s (platform, platform_product_id, platform_sku_id)",
		tableName,
	)).Error
	require.NoError(t, err)

	product := &models.MasterProduct{
		TenantID: "test-tenant",
		Title:    "Upsert Product",
		Status:   models.MasterProductStatusActive,
	}
	require.NoError(t, repo.Create(ctx, product))

	firstSku := &models.MasterProductSku{
		TenantID:        "test-tenant",
		MasterProductID: product.ID,
		SellerSku:       "SKU-001",
		VariantName:     "Black",
		Price:           10000,
		Stock:           10,
	}
	secondSku := &models.MasterProductSku{
		TenantID:        "test-tenant",
		MasterProductID: product.ID,
		SellerSku:       "SKU-002",
		VariantName:     "Brown",
		Price:           12000,
		Stock:           12,
	}
	require.NoError(t, repo.CreateSku(ctx, firstSku))
	require.NoError(t, repo.CreateSku(ctx, secondSku))

	firstLink := &models.MasterProductPlatformLink{
		MasterProductID:   product.ID,
		MasterSkuID:       &firstSku.ID,
		Platform:          "shopee",
		PlatformProductID: "ITEM-123",
		PlatformItemID:    "ITEM-123",
		PlatformSkuID:     "MODEL-1",
		SyncStatus:        models.SyncStatusSynced,
	}
	require.NoError(t, repo.UpsertPlatformLink(ctx, firstLink))

	secondLink := &models.MasterProductPlatformLink{
		MasterProductID:   product.ID,
		MasterSkuID:       &secondSku.ID,
		Platform:          "shopee",
		PlatformProductID: "ITEM-123",
		PlatformItemID:    "ITEM-123",
		PlatformSkuID:     "MODEL-2",
		SyncStatus:        models.SyncStatusSynced,
	}
	require.NoError(t, repo.UpsertPlatformLink(ctx, secondLink))

	var links []models.MasterProductPlatformLink
	err = db.WithContext(ctx).
		Where("platform = ? AND platform_product_id = ?", "shopee", "ITEM-123").
		Find(&links).Error
	require.NoError(t, err)
	require.Len(t, links, 2)

	linkedSkus := make(map[string]uint, len(links))
	for _, existing := range links {
		require.NotNil(t, existing.MasterSkuID)
		linkedSkus[existing.PlatformSkuID] = *existing.MasterSkuID
	}

	assert.Equal(t, firstSku.ID, linkedSkus["MODEL-1"])
	assert.Equal(t, secondSku.ID, linkedSkus["MODEL-2"])
}
