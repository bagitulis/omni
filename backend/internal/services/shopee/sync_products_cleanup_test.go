package shopee

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupShopeeSyncCleanupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	err = db.AutoMigrate(&models.ShopeeProduct{}, &models.ShopeeSku{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func TestProductSyncService_ClearShopeeProductCache(t *testing.T) {
	db := setupShopeeSyncCleanupTestDB(t)
	ctx := context.Background()

	service := NewProductSyncService(nil, db, "tenant1")

	require.NoError(t, db.WithContext(ctx).Create(&models.ShopeeProduct{
		TenantID: "tenant1",
		ItemID:   1001,
		Name:     "Tenant 1 Product",
		Status:   "NORMAL",
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&models.ShopeeSku{
		TenantID:  "tenant1",
		ProductID: 1,
		ItemID:    1001,
		SellerSku: "TENANT1-SKU",
		Price:     10000,
		Quantity:  5,
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&models.ShopeeProduct{
		TenantID: "tenant2",
		ItemID:   2002,
		Name:     "Tenant 2 Product",
		Status:   "NORMAL",
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&models.ShopeeSku{
		TenantID:  "tenant2",
		ProductID: 2,
		ItemID:    2002,
		SellerSku: "TENANT2-SKU",
		Price:     20000,
		Quantity:  7,
	}).Error)

	require.NoError(t, service.clearShopeeProductCache(ctx))

	var tenant1Products int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.ShopeeProduct{}).
		Where("tenant_id = ?", "tenant1").
		Count(&tenant1Products).Error)
	assert.Equal(t, int64(0), tenant1Products)

	var tenant1Skus int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.ShopeeSku{}).
		Where("tenant_id = ?", "tenant1").
		Count(&tenant1Skus).Error)
	assert.Equal(t, int64(0), tenant1Skus)

	var tenant2Products int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.ShopeeProduct{}).
		Where("tenant_id = ?", "tenant2").
		Count(&tenant2Products).Error)
	assert.Equal(t, int64(1), tenant2Products)

	var tenant2Skus int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.ShopeeSku{}).
		Where("tenant_id = ?", "tenant2").
		Count(&tenant2Skus).Error)
	assert.Equal(t, int64(1), tenant2Skus)
}
