package master_product

import (
	"context"
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshProductImages_RequiresTenantID(t *testing.T) {
	service := NewService(nil)

	result, err := service.RefreshProductImages(context.Background(), "", 1, true)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrTenantIDRequired)
}

func TestRefreshProductImages_ReturnsProductNotFound(t *testing.T) {
	db := setupMasterProductServiceTestDB(t)
	service := NewService(db)

	result, err := service.RefreshProductImages(context.Background(), "tenant-a", 999, false)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrProductNotFound)
}

func TestRefreshProductImages_ForceFalseSkipsAggregationWhenImagesExist(t *testing.T) {
	db := setupMasterProductServiceTestDB(t)
	service := NewService(db)
	ctx := context.Background()
	now := time.Now()

	product := &models.MasterProduct{
		TenantID:  "tenant-a",
		Title:     "Image Product",
		Status:    models.MasterProductStatusDraft,
		Images:    models.JSONArray{"/uploads/existing.webp"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, db.WithContext(ctx).Create(product).Error)

	require.NoError(t, db.WithContext(ctx).Create(&models.MasterProductPlatformLink{
		MasterProductID: product.ID,
		Platform:        models.PlatformShopee,
		PlatformItemID:  "12345",
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&models.ShopeeProduct{
		TenantID:    "tenant-a",
		ItemID:      12345,
		Name:        "Shopee Source",
		Status:      "NORMAL",
		LocalImages: models.JSONArray{"/uploads/new-1.webp", "/uploads/new-2.webp"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error)

	result, err := service.RefreshProductImages(ctx, "tenant-a", product.ID, false)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 1, result.PreviousImageCount)
	assert.Equal(t, 1, result.ImageCount)
	assert.False(t, result.Updated)
	assert.False(t, result.Forced)
	assert.Equal(t, []string{"/uploads/existing.webp"}, result.Images)
}

func TestRefreshProductImages_ForceTrueAggregatesPlatformImages(t *testing.T) {
	db := setupMasterProductServiceTestDB(t)
	service := NewService(db)
	ctx := context.Background()
	now := time.Now()

	product := &models.MasterProduct{
		TenantID:  "tenant-a",
		Title:     "Image Product",
		Status:    models.MasterProductStatusDraft,
		Images:    models.JSONArray{"/uploads/old.webp"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, db.WithContext(ctx).Create(product).Error)

	require.NoError(t, db.WithContext(ctx).Create(&models.MasterProductPlatformLink{
		MasterProductID: product.ID,
		Platform:        models.PlatformShopee,
		PlatformItemID:  "54321",
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&models.ShopeeProduct{
		TenantID:    "tenant-a",
		ItemID:      54321,
		Name:        "Shopee Source",
		Status:      "NORMAL",
		LocalImages: models.JSONArray{"/uploads/new-a.webp", "/uploads/new-b.webp"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error)

	result, err := service.RefreshProductImages(ctx, "tenant-a", product.ID, true)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 1, result.PreviousImageCount)
	assert.Equal(t, 2, result.ImageCount)
	assert.True(t, result.Updated)
	assert.True(t, result.Forced)
	assert.Equal(t, []string{"/uploads/new-a.webp", "/uploads/new-b.webp"}, result.Images)
}

func TestRefreshProductImages_UsesRemoteImageWhenLocalImagesEmpty(t *testing.T) {
	db := setupMasterProductServiceTestDB(t)
	service := NewService(db)
	ctx := context.Background()
	now := time.Now()

	product := &models.MasterProduct{
		TenantID:  "tenant-a",
		Title:     "Image Product",
		Status:    models.MasterProductStatusDraft,
		Images:    models.JSONArray{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, db.WithContext(ctx).Create(product).Error)

	require.NoError(t, db.WithContext(ctx).Create(&models.MasterProductPlatformLink{
		MasterProductID: product.ID,
		Platform:        models.PlatformShopee,
		PlatformItemID:  "99999",
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&models.ShopeeProduct{
		TenantID:    "tenant-a",
		ItemID:      99999,
		Name:        "Shopee Source",
		Status:      "NORMAL",
		Image:       "https://example.com/remote-image.webp",
		LocalImages: models.JSONArray{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error)

	result, err := service.RefreshProductImages(ctx, "tenant-a", product.ID, true)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 1, result.ImageCount)
	assert.True(t, result.Updated)
	assert.Equal(t, []string{"https://example.com/remote-image.webp"}, result.Images)
}
