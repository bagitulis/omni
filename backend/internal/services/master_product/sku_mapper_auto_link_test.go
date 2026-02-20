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
		&models.TiktokProduct{},
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

func TestAutoMapAndLink_CrossPlatform(t *testing.T) {
	ctx := context.Background()
	db := setupAutoMapLinkTestDB(t)

	product := &models.MasterProduct{
		TenantID: "tenant-cross", Title: "Cross Platform Product",
		Status: models.MasterProductStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(product).Error)

	masterSku := &models.MasterProductSku{
		TenantID: "tenant-cross", MasterProductID: product.ID, SellerSku: "CROSS-001",
		Price: 5000, Stock: 20, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(masterSku).Error)

	modelID := int64(111)
	require.NoError(t, db.Create(&models.ShopeeSku{
		TenantID: "tenant-cross", ProductID: 1, ItemID: 900, ModelID: &modelID,
		SellerSku: "CROSS-001", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	require.NoError(t, db.Create(&models.TiktokProduct{
		ID: 800, TenantID: "tenant-cross", ProductID: "1729991138619656479",
		Name: "Cross Platform Product", Status: "ACTIVATE",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	require.NoError(t, db.Create(&models.TiktokSku{
		TenantID: "tenant-cross", ProductID: 800, SkuID: "TIK-800",
		SellerSku: "CROSS-001", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	mapper := NewSkuMapper(db, "tenant-cross")
	result, err := mapper.AutoMapAndLinkBySkus(ctx, []string{"CROSS-001"})
	require.NoError(t, err)

	// Shopee + TikTok matched → 2 links created
	assert.Equal(t, 2, result.MappedCount)
	assert.Equal(t, 0, result.SkippedCount)
	require.Len(t, result.Mappings, 2)

	platforms := map[string]bool{}
	for _, m := range result.Mappings {
		platforms[m.Platform] = true
	}
	assert.True(t, platforms[models.PlatformShopee], "expected Shopee link")
	assert.True(t, platforms[models.PlatformTiktok], "expected TikTok link")

	links, err := mapper.repo.FindPlatformLinks(ctx, product.ID)
	require.NoError(t, err)
	assert.Len(t, links, 2)
}

func TestAutoMapAndLink_Ambiguous(t *testing.T) {
	ctx := context.Background()
	db := setupAutoMapLinkTestDB(t)

	product := &models.MasterProduct{
		TenantID: "tenant-amb", Title: "Ambiguous Product",
		Status: models.MasterProductStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(product).Error)

	masterSku := &models.MasterProductSku{
		TenantID: "tenant-amb", MasterProductID: product.ID, SellerSku: "AMB-001",
		Price: 1000, Stock: 5, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(masterSku).Error)

	// Two Shopee SKUs with same seller_sku but different items → ambiguous
	modelID1 := int64(1)
	modelID2 := int64(2)
	require.NoError(t, db.Create(&models.ShopeeSku{
		TenantID: "tenant-amb", ProductID: 1, ItemID: 100, ModelID: &modelID1,
		SellerSku: "AMB-001", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)
	require.NoError(t, db.Create(&models.ShopeeSku{
		TenantID: "tenant-amb", ProductID: 2, ItemID: 200, ModelID: &modelID2,
		SellerSku: "AMB-001", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	mapper := NewSkuMapper(db, "tenant-amb")
	result, err := mapper.AutoMapAndLinkBySkus(ctx, []string{"AMB-001"})
	require.NoError(t, err)

	// Shopee ambiguous, TikTok/Lazada no match → skipped with ambiguous error
	assert.Equal(t, 0, result.MappedCount)
	assert.Equal(t, 1, result.SkippedCount)
	require.NotEmpty(t, result.Errors)

	hasAmbiguousError := false
	for _, e := range result.Errors {
		if len(e) > 0 && (e == fmt.Sprintf("ambiguous platform match for SKU %s", "AMB-001")) {
			hasAmbiguousError = true
		}
	}
	assert.True(t, hasAmbiguousError, "expected ambiguous error message")

	links, err := mapper.repo.FindPlatformLinks(ctx, product.ID)
	require.NoError(t, err)
	assert.Len(t, links, 0, "no links should be created for ambiguous SKU")
}

func TestAutoMapAndLink_CaseInsensitive(t *testing.T) {
	ctx := context.Background()
	db := setupAutoMapLinkTestDB(t)

	product := &models.MasterProduct{
		TenantID: "tenant-ci", Title: "Case Insensitive Product",
		Status: models.MasterProductStatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(product).Error)

	// Master SKU stored as lowercase
	masterSku := &models.MasterProductSku{
		TenantID: "tenant-ci", MasterProductID: product.ID, SellerSku: "bakiw5971",
		Price: 10000, Stock: 10, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(masterSku).Error)

	// Shopee SKU stored as lowercase
	modelID := int64(50001)
	require.NoError(t, db.Create(&models.ShopeeSku{
		TenantID: "tenant-ci", ProductID: 1, ItemID: 10001, ModelID: &modelID,
		SellerSku: "bakiw5971", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)

	mapper := NewSkuMapper(db, "tenant-ci")

	// Input is UPPERCASE → should still match lowercase master SKU and Shopee SKU
	result, err := mapper.AutoMapAndLinkBySkus(ctx, []string{"BAKIW5971"})
	require.NoError(t, err)

	assert.Equal(t, 1, result.MappedCount)
	assert.Equal(t, 0, result.SkippedCount)
	require.Len(t, result.Mappings, 1)
	assert.Equal(t, models.PlatformShopee, result.Mappings[0].Platform)
	assert.Equal(t, "10001", result.Mappings[0].PlatformItemID)

	links, err := mapper.repo.FindPlatformLinks(ctx, product.ID)
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, masterSku.ID, *links[0].MasterSkuID)
}
