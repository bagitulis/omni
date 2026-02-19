package master_product

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStagingImportService_ImportFromLazadaStaging_MatchesByExistingSellerSkuWhenNameEmpty(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t,
		&models.MasterProduct{},
		&models.MasterProductSku{},
		&models.MasterProductPlatformLink{},
		&models.LazadaProduct{},
		&models.LazadaSku{},
	)
	ctx := context.Background()
	tenantID := "test-tenant"

	existingProduct := &models.MasterProduct{
		TenantID: tenantID,
		Title:    "Frisian Flag Bendera Kental Manis Pouch 240gr",
		Status:   models.MasterProductStatusActive,
	}
	require.NoError(t, db.WithContext(ctx).Create(existingProduct).Error)

	existingSku := &models.MasterProductSku{
		TenantID:        tenantID,
		MasterProductID: existingProduct.ID,
		SellerSku:       "FFBSK5993",
		VariantName:     "Putih 240gr",
		Price:           55400,
		Stock:           130,
	}
	require.NoError(t, db.WithContext(ctx).Create(existingSku).Error)

	require.NoError(t, db.WithContext(ctx).Create(&models.LazadaProduct{
		TenantID: tenantID,
		ItemID:   "6361286052",
		Name:     "",
		Status:   "live",
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&models.LazadaSku{
		TenantID:    tenantID,
		ItemID:      "6361286052",
		SkuID:       "12060414299",
		SellerSku:   "FFBSK5993",
		Name:        "FFBSK5993",
		VariantName: "",
		Price:       55400,
		Quantity:    130,
	}).Error)

	service := NewStagingImportService(db)
	result, err := service.ImportFromLazadaStaging(ctx, tenantID)
	require.NoError(t, err)
	require.Empty(t, result.Errors)

	assert.Equal(t, 1, result.ProductsMatched)
	assert.Equal(t, 0, result.ProductsCreated)

	var emptyTitleProducts int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.MasterProduct{}).
		Where("tenant_id = ? AND TRIM(title) = ''", tenantID).
		Count(&emptyTitleProducts).Error)
	assert.Equal(t, int64(0), emptyTitleProducts)

	var updatedSku models.MasterProductSku
	require.NoError(t, db.WithContext(ctx).
		Where("tenant_id = ? AND master_product_id = ? AND seller_sku = ?", tenantID, existingProduct.ID, "FFBSK5993").
		First(&updatedSku).Error)
	assert.Equal(t, "Putih 240gr", updatedSku.VariantName)

	var lazadaLink models.MasterProductPlatformLink
	require.NoError(t, db.WithContext(ctx).
		Where("master_product_id = ? AND platform = ? AND platform_item_id = ?", existingProduct.ID, "lazada", "6361286052").
		First(&lazadaLink).Error)
	assert.Equal(t, "12060414299", lazadaLink.PlatformSkuID)
}

func TestStagingImportService_ImportFromLazadaStaging_UsesFallbackTitleForUnnamedProducts(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t,
		&models.MasterProduct{},
		&models.MasterProductSku{},
		&models.MasterProductPlatformLink{},
		&models.LazadaProduct{},
		&models.LazadaSku{},
	)
	ctx := context.Background()
	tenantID := "test-tenant"

	require.NoError(t, db.WithContext(ctx).Create(&models.LazadaProduct{
		TenantID: tenantID,
		ItemID:   "777888999",
		Name:     "",
		Status:   "live",
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&models.LazadaSku{
		TenantID:    tenantID,
		ItemID:      "777888999",
		SkuID:       "SKU-777",
		SellerSku:   "FFBSK777",
		Name:        "FFBSK777",
		VariantName: "",
		Price:       10000,
		Quantity:    5,
	}).Error)

	service := NewStagingImportService(db)
	result, err := service.ImportFromLazadaStaging(ctx, tenantID)
	require.NoError(t, err)
	require.Empty(t, result.Errors)

	assert.Equal(t, 1, result.ProductsCreated)

	var createdProduct models.MasterProduct
	require.NoError(t, db.WithContext(ctx).
		Where("tenant_id = ? AND title = ?", tenantID, "Lazada Item 777888999").
		First(&createdProduct).Error)

	var createdSku models.MasterProductSku
	require.NoError(t, db.WithContext(ctx).
		Where("tenant_id = ? AND master_product_id = ? AND seller_sku = ?", tenantID, createdProduct.ID, "FFBSK777").
		First(&createdSku).Error)
	assert.Equal(t, "", createdSku.VariantName)
}
