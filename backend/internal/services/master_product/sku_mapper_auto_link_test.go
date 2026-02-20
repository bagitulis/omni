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

func setupAutoMapLinkTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:sku_mapper_auto_link_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.MasterProduct{},
		&models.MasterProductSku{},
		&models.MasterProductPlatformLink{},
		&models.ShopeeSku{},
		&models.TiktokSku{},
		&models.LazadaSku{},
	)
	require.NoError(t, err)

	return db
}

func TestSkuMapper_AutoMapAndLinkBySkus(t *testing.T) {
	ctx := context.Background()
	db := setupAutoMapLinkTestDB(t)

	product := &models.MasterProduct{
		TenantID:  "tenant-a",
		Title:     "Test Product",
		Status:    models.MasterProductStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(product).Error)

	masterSku := &models.MasterProductSku{
		TenantID:        "tenant-a",
		MasterProductID: product.ID,
		SellerSku:       "SKU-001",
		Price:           10000,
		Stock:           10,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	require.NoError(t, db.Create(masterSku).Error)

	modelID := int64(991122)
	shopeeSku := &models.ShopeeSku{
		TenantID:  "tenant-a",
		ProductID: 1,
		ItemID:    778899,
		ModelID:   &modelID,
		SellerSku: "SKU-001",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(shopeeSku).Error)

	mapper := NewSkuMapper(db, "tenant-a")
	result, err := mapper.AutoMapAndLinkBySkus(ctx, []string{"SKU-001", "SKU-404"})
	require.NoError(t, err)

	assert.Equal(t, 1, result.MappedCount)
	assert.Equal(t, 1, result.SkippedCount)
	require.Len(t, result.Mappings, 1)
	assert.Equal(t, "SKU-001", result.Mappings[0].SellerSku)
	assert.Equal(t, models.PlatformShopee, result.Mappings[0].Platform)
	assert.Equal(t, models.SyncStatusSynced, result.Mappings[0].SyncStatus)

	links, err := mapper.repo.FindPlatformLinks(ctx, product.ID)
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.NotNil(t, links[0].MasterSkuID)
	assert.Equal(t, masterSku.ID, *links[0].MasterSkuID)
	assert.Equal(t, models.PlatformShopee, links[0].Platform)
	assert.Equal(t, models.SyncStatusSynced, links[0].SyncStatus)
	assert.Equal(t, "778899", links[0].PlatformItemID)
}
