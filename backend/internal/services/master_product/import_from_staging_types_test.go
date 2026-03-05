package master_product

import (
	"context"
	"fmt"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStagingImportService_UpsertMasterSku(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t,
		&models.MasterProduct{},
		&models.MasterProductSku{},
	)
	service := NewStagingImportService(db)
	ctx := context.Background()

	product := &models.MasterProduct{
		TenantID: "test-tenant",
		Title:    "Test Product",
		Status:   models.MasterProductStatusActive,
	}
	require.NoError(t, db.WithContext(ctx).Create(product).Error)

	createdSku, created, err := service.upsertMasterSku(ctx, &models.MasterProductSku{
		TenantID:        "test-tenant",
		MasterProductID: product.ID,
		SellerSku:       "SKU-123",
		VariantName:     "Hitam",
		Price:           10000,
		Stock:           10,
	})
	require.NoError(t, err)
	assert.True(t, created)
	require.NotZero(t, createdSku.ID)

	updatedSku, created, err := service.upsertMasterSku(ctx, &models.MasterProductSku{
		TenantID:        "test-tenant",
		MasterProductID: product.ID,
		SellerSku:       "SKU-123",
		VariantName:     "Coklat",
		Price:           12000,
		Stock:           22,
	})
	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, createdSku.ID, updatedSku.ID)
	assert.Equal(t, "Coklat", updatedSku.VariantName)
	// Price is NOT overwritten when existing price > 0
	assert.Equal(t, float64(10000), updatedSku.Price)
	// Stock is NOT overwritten on staging import (per-platform values stay in staging tables)
	assert.Equal(t, 10, updatedSku.Stock)

	var skuCount int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.MasterProductSku{}).
		Where("tenant_id = ? AND master_product_id = ? AND seller_sku = ?", "test-tenant", product.ID, "SKU-123").
		Count(&skuCount).Error)
	assert.Equal(t, int64(1), skuCount)
}

func TestStagingImportService_ImportFromShopeeStaging_Idempotent(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t,
		&models.MasterProduct{},
		&models.MasterProductSku{},
		&models.MasterProductPlatformLink{},
		&models.ShopeeProduct{},
		&models.ShopeeSku{},
	)
	ctx := context.Background()

	linkTable := models.MasterProductPlatformLink{}.TableName()
	err := db.Exec(fmt.Sprintf(
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_platform_links_unique ON %s (platform, platform_product_id, COALESCE(platform_sku_id, '')) WHERE (platform_product_id IS NOT NULL)",
		linkTable,
	)).Error
	require.NoError(t, err)

	tenantID := "test-tenant"
	itemID := int64(901001)
	modelID := int64(9001)

	require.NoError(t, db.WithContext(ctx).Create(&models.ShopeeProduct{
		TenantID: tenantID,
		ItemID:   itemID,
		Name:     "Kiwi Chia Smoothie",
		Price:    40600,
		Quantity: 50,
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&models.ShopeeSku{
		TenantID:    tenantID,
		ItemID:      itemID,
		ModelID:     &modelID,
		SellerSku:   "KIWI-CHIA-40600",
		VariantName: "Hitam",
		Price:       40600,
		Quantity:    50,
	}).Error)

	service := NewStagingImportService(db)

	firstRun, err := service.ImportFromShopeeStaging(ctx, tenantID)
	require.NoError(t, err)
	require.Empty(t, firstRun.Errors)
	require.Equal(t, 1, firstRun.ProductsCreated)
	require.Equal(t, 1, firstRun.SkusCreated)

	secondRun, err := service.ImportFromShopeeStaging(ctx, tenantID)
	require.NoError(t, err)
	require.Empty(t, secondRun.Errors)
	assert.Equal(t, 1, secondRun.ProductsMatched)
	assert.Equal(t, 0, secondRun.ProductsCreated)
	assert.Equal(t, 0, secondRun.SkusCreated)

	var productsCount int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.MasterProduct{}).
		Where("tenant_id = ? AND title = ?", tenantID, "Kiwi Chia Smoothie").
		Count(&productsCount).Error)
	assert.Equal(t, int64(1), productsCount)

	var skusCount int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.MasterProductSku{}).
		Where("tenant_id = ? AND seller_sku = ?", tenantID, "KIWI-CHIA-40600").
		Count(&skusCount).Error)
	assert.Equal(t, int64(1), skusCount)

	var linksCount int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.MasterProductPlatformLink{}).
		Where("platform = ? AND platform_product_id = ?", "shopee", "901001").
		Count(&linksCount).Error)
	assert.Equal(t, int64(1), linksCount)
}
