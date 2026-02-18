package master_product

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceProductSKUs_ReplacesRowsAndRemovesDeletedLinks(t *testing.T) {
	db := setupMasterProductServiceTestDB(t)
	service := NewService(db)
	ctx := context.Background()
	now := time.Now()

	product := &models.MasterProduct{
		TenantID:  "tenant-a",
		Title:     "Product A",
		Status:    models.MasterProductStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, db.WithContext(ctx).Create(product).Error)

	skuKeep := &models.MasterProductSku{
		TenantID:        "tenant-a",
		MasterProductID: product.ID,
		SellerSku:       "SKU-KEEP",
		VariantName:     "Old",
		Price:           10000,
		Stock:           5,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	skuDelete := &models.MasterProductSku{
		TenantID:        "tenant-a",
		MasterProductID: product.ID,
		SellerSku:       "SKU-DELETE",
		Price:           12000,
		Stock:           7,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	require.NoError(t, db.WithContext(ctx).Create(skuKeep).Error)
	require.NoError(t, db.WithContext(ctx).Create(skuDelete).Error)

	linkKeep := &models.MasterProductPlatformLink{
		MasterProductID: product.ID,
		MasterSkuID:     &skuKeep.ID,
		Platform:        models.PlatformShopee,
		PlatformItemID:  "10001",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	linkDelete := &models.MasterProductPlatformLink{
		MasterProductID: product.ID,
		MasterSkuID:     &skuDelete.ID,
		Platform:        models.PlatformLazada,
		PlatformItemID:  "20002",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	require.NoError(t, db.WithContext(ctx).Create(linkKeep).Error)
	require.NoError(t, db.WithContext(ctx).Create(linkDelete).Error)

	keepID := skuKeep.ID
	err := service.replaceProductSKUs(ctx, "tenant-a", product.ID, []UpdateSkuInput{
		{
			ID:          &keepID,
			SellerSku:   "SKU-KEEP-UPDATED",
			VariantName: "New",
			VariantData: map[string]interface{}{"color": "red"},
			Price:       15000,
			Stock:       11,
		},
		{
			SellerSku:   "SKU-NEW",
			VariantName: "Variant B",
			VariantData: map[string]interface{}{"size": "L"},
			Price:       18000,
			Stock:       9,
		},
	})
	require.NoError(t, err)

	skus, err := service.repo.FindSkusByProductID(ctx, product.ID)
	require.NoError(t, err)
	require.Len(t, skus, 2)

	var updatedSKU *models.MasterProductSku
	var newSKU *models.MasterProductSku
	for i := range skus {
		sku := &skus[i]
		switch sku.SellerSku {
		case "SKU-KEEP-UPDATED":
			updatedSKU = sku
		case "SKU-NEW":
			newSKU = sku
		}
	}

	require.NotNil(t, updatedSKU)
	require.NotNil(t, newSKU)
	assert.Equal(t, "New", updatedSKU.VariantName)
	assert.Equal(t, 15000.0, updatedSKU.Price)
	assert.Equal(t, 11, updatedSKU.Stock)

	links, err := service.repo.FindPlatformLinks(ctx, product.ID)
	require.NoError(t, err)
	require.Len(t, links, 1)
	require.NotNil(t, links[0].MasterSkuID)
	assert.Equal(t, skuKeep.ID, *links[0].MasterSkuID)
}

func TestReplaceProductSKUs_TooManySKUs(t *testing.T) {
	db := setupMasterProductServiceTestDB(t)
	service := NewService(db)

	inputs := make([]UpdateSkuInput, models.MasterProductMaxSKUs+1)
	for i := range inputs {
		inputs[i] = UpdateSkuInput{SellerSku: fmt.Sprintf("SKU-%d", i+1)}
	}

	err := service.replaceProductSKUs(context.Background(), "tenant-a", 1, inputs)
	assert.ErrorIs(t, err, ErrTooManySKUs)
}

func TestReplaceProductSKUs_SellerSkuRequired(t *testing.T) {
	db := setupMasterProductServiceTestDB(t)
	service := NewService(db)

	err := service.replaceProductSKUs(context.Background(), "tenant-a", 1, []UpdateSkuInput{
		{SellerSku: "   "},
	})
	assert.ErrorIs(t, err, ErrSellerSkuRequired)
}

func TestReplaceProductSKUs_ReturnsSkuNotFoundWhenInputIDNotOwnedByProduct(t *testing.T) {
	db := setupMasterProductServiceTestDB(t)
	service := NewService(db)
	ctx := context.Background()
	now := time.Now()

	productA := &models.MasterProduct{TenantID: "tenant-a", Title: "A", Status: models.MasterProductStatusDraft, CreatedAt: now, UpdatedAt: now}
	productB := &models.MasterProduct{TenantID: "tenant-a", Title: "B", Status: models.MasterProductStatusDraft, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.WithContext(ctx).Create(productA).Error)
	require.NoError(t, db.WithContext(ctx).Create(productB).Error)

	skuB := &models.MasterProductSku{
		TenantID:        "tenant-a",
		MasterProductID: productB.ID,
		SellerSku:       "SKU-B",
		Price:           5000,
		Stock:           2,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	require.NoError(t, db.WithContext(ctx).Create(skuB).Error)

	err := service.replaceProductSKUs(ctx, "tenant-a", productA.ID, []UpdateSkuInput{{
		ID:        &skuB.ID,
		SellerSku: "SKU-B-UPDATED",
		Price:     6000,
		Stock:     3,
	}})

	assert.ErrorIs(t, err, ErrSkuNotFound)
}
