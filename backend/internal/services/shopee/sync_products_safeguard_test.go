package shopee

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShopeeSyncProducts_SkipsDeleteWhenItemListEmpty(t *testing.T) {
	db := setupShopeeSyncCleanupTestDB(t)
	ctx := context.Background()

	service := NewProductSyncService(nil, db, "tenant1")

	require.NoError(t, db.WithContext(ctx).Create(&models.ShopeeProduct{
		TenantID: "tenant1",
		ItemID:   1001,
		Name:     "Existing Product",
		Status:   "NORMAL",
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&models.ShopeeSku{
		TenantID:  "tenant1",
		ProductID: "1",
		ItemID:    1001,
		SellerSku: "EXISTING-SKU",
		Price:     10000,
		Quantity:  3,
	}).Error)

	count, err := service.syncProductsByItemIDs(ctx, []int64{})
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	var productsCount int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.ShopeeProduct{}).
		Where("tenant_id = ?", "tenant1").
		Count(&productsCount).Error)
	assert.Equal(t, int64(1), productsCount)

	var skusCount int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.ShopeeSku{}).
		Where("tenant_id = ?", "tenant1").
		Count(&skusCount).Error)
	assert.Equal(t, int64(1), skusCount)
}
