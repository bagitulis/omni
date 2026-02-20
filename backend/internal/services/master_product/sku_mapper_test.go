package master_product

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupSkuMapperTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:sku_mapper_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.MasterProduct{},
		&models.MasterProductSku{},
		&models.MasterProductPlatformLink{},
		&models.ShopeeSku{},
		&models.TiktokProduct{},
		&models.TiktokSku{},
		&models.LazadaSku{},
	)
	require.NoError(t, err)

	return db
}

func seedSkuMappingData(t *testing.T, db *gorm.DB, tenantID string) {
	t.Helper()

	now := time.Now()
	product := &models.MasterProduct{
		TenantID:  tenantID,
		Title:     "Case Insensitive Product",
		Status:    models.MasterProductStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, db.Create(product).Error)

	masterSku := &models.MasterProductSku{
		TenantID:        tenantID,
		MasterProductID: product.ID,
		SellerSku:       "bakiw5971",
		VariantName:     "Default",
		Price:           10000,
		Stock:           10,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	require.NoError(t, db.Create(masterSku).Error)

	modelID := int64(50001)
	require.NoError(t, db.Create(&models.ShopeeSku{
		TenantID:    tenantID,
		ProductID:   1,
		ItemID:      10001,
		ModelID:     &modelID,
		SellerSku:   "bakiw5971",
		VariantName: "Default",
		Price:       10000,
		Quantity:    10,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error)

	require.NoError(t, db.Create(&models.TiktokProduct{
		ID:        20001,
		TenantID:  tenantID,
		ProductID: "1729991138619650001",
		Name:      "Case Insensitive Product",
		Status:    "ACTIVATE",
		CreatedAt: now,
		UpdatedAt: now,
	}).Error)

	require.NoError(t, db.Create(&models.TiktokSku{
		TenantID:    tenantID,
		ProductID:   20001,
		SkuID:       "TIK-50001",
		SellerSku:   "bakiw5971",
		VariantName: "Default",
		Price:       10000,
		Quantity:    10,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error)

	require.NoError(t, db.Create(&models.LazadaSku{
		TenantID:    tenantID,
		ItemID:      "LZD-30001",
		ProductID:   1,
		SkuID:       "LZD-SKU-30001",
		SellerSku:   "bakiw5971",
		VariantName: "Default",
		Price:       10000,
		Quantity:    10,
		Available:   10,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error)

	require.NoError(t, db.Create(&models.LazadaSku{
		TenantID:    tenantID,
		ItemID:      "LZD-30002",
		ProductID:   1,
		SkuID:       "LZD-SKU-30002",
		ShopSku:     "shop-only-sku",
		VariantName: "ShopSkuOnly",
		Price:       10000,
		Quantity:    10,
		Available:   10,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error)
}

func TestSkuMapper_AutoMapBySku_CaseInsensitive(t *testing.T) {
	ctx := context.Background()
	db := setupSkuMapperTestDB(t)
	seedSkuMappingData(t, db, "tenant-case")

	mapper := NewSkuMapper(db, "tenant-case")
	testCases := []string{"BAKIW5971", "Bakiw5971", " BAKIW5971 "}

	for _, input := range testCases {
		results, err := mapper.AutoMapBySku(ctx, input)
		require.NoError(t, err)
		require.Len(t, results, 3)

		assert.True(t, results[0].Found, "expected Shopee match for %q", input)
		assert.True(t, results[1].Found, "expected TikTok match for %q", input)
		assert.True(t, results[2].Found, "expected Lazada match for %q", input)
	}
}

func TestSkuMapper_AutoMapBySku_NoMatchForDifferentSku(t *testing.T) {
	ctx := context.Background()
	db := setupSkuMapperTestDB(t)
	seedSkuMappingData(t, db, "tenant-negative")

	mapper := NewSkuMapper(db, "tenant-negative")
	results, err := mapper.AutoMapBySku(ctx, "OTHER_SKU")
	require.NoError(t, err)
	require.Len(t, results, 3)

	for _, result := range results {
		assert.False(t, result.Found)
	}
}

func TestSkuMapper_AutoMapBySku_LazadaShopSkuCaseInsensitive(t *testing.T) {
	ctx := context.Background()
	db := setupSkuMapperTestDB(t)
	seedSkuMappingData(t, db, "tenant-lazada-shop-sku")

	mapper := NewSkuMapper(db, "tenant-lazada-shop-sku")
	results, err := mapper.AutoMapBySku(ctx, " SHOP-ONLY-SKU ")
	require.NoError(t, err)
	require.Len(t, results, 3)

	assert.False(t, results[0].Found)
	assert.False(t, results[1].Found)
	assert.True(t, results[2].Found)
	assert.Equal(t, "LZD-30002", results[2].PlatformItemID)
}
