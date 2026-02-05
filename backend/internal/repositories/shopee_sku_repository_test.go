package repositories

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShopeeSkuRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.ShopeeProduct{}, &models.ShopeeSku{})
	repo := NewShopeeSkuRepository(db)
	ctx := context.Background()

	// Helper to create test product first
	createTestProduct := func(t *testing.T, itemID int64) *models.ShopeeProduct {
		product := &models.ShopeeProduct{
			TenantID: "test-tenant",
			ItemID:   itemID,
			Name:     "Test Product",
			Status:   "NORMAL",
			Price:    50000,
			Quantity: 100,
		}
		err := db.Create(product).Error
		require.NoError(t, err)
		return product
	}

	// Helper to create test SKU
	createTestSku := func(t *testing.T, product *models.ShopeeProduct, modelID int64, sellerSku string) *models.ShopeeSku {
		sku := &models.ShopeeSku{
			TenantID:    "test-tenant",
			ProductID:   product.ID,
			ItemID:      product.ItemID,
			ModelID:     &modelID,
			SellerSku:   sellerSku,
			VariantName: "Size: M",
			Price:       45000,
			Quantity:    25,
		}
		err := repo.Create(ctx, sku)
		require.NoError(t, err)
		return sku
	}

	t.Run("Create", func(t *testing.T) {
		product := createTestProduct(t, 1001)
		modelID := int64(10010)

		tests := []struct {
			name    string
			sku     *models.ShopeeSku
			wantErr bool
		}{
			{
				name: "creates SKU successfully",
				sku: &models.ShopeeSku{
					TenantID:    "test-tenant",
					ProductID:   product.ID,
					ItemID:      product.ItemID,
					ModelID:     &modelID,
					SellerSku:   "SKU-CREATE-001",
					VariantName: "Color: Red",
					Price:       30000,
					Quantity:    50,
				},
				wantErr: false,
			},
			{
				name: "creates SKU without model ID (non-variant)",
				sku: &models.ShopeeSku{
					TenantID:  "test-tenant",
					ProductID: product.ID,
					ItemID:    product.ItemID,
					SellerSku: "SKU-CREATE-002",
					Price:     35000,
					Quantity:  30,
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := repo.Create(ctx, tt.sku)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				assert.NotZero(t, tt.sku.ID)
				assert.False(t, tt.sku.CreatedAt.IsZero())
			})
		}
	})

	t.Run("FindByModelID", func(t *testing.T) {
		product := createTestProduct(t, 2001)
		modelID := int64(20010)
		sku := createTestSku(t, product, modelID, "SKU-FIND-MODEL")

		tests := []struct {
			name    string
			modelID int64
			wantErr bool
		}{
			{
				name:    "finds existing SKU by model ID",
				modelID: modelID,
				wantErr: false,
			},
			{
				name:    "returns error for non-existent model ID",
				modelID: 999999,
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				found, err := repo.FindByModelID(ctx, tt.modelID)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				require.NotNil(t, found)
				assert.Equal(t, sku.SellerSku, found.SellerSku)
			})
		}
	})

	t.Run("FindByItemID", func(t *testing.T) {
		product := createTestProduct(t, 3001)
		modelID1 := int64(30010)
		modelID2 := int64(30020)
		createTestSku(t, product, modelID1, "SKU-ITEM-001")
		createTestSku(t, product, modelID2, "SKU-ITEM-002")

		t.Run("finds all SKUs for item", func(t *testing.T) {
			skus, err := repo.FindByItemID(ctx, product.ItemID)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(skus), 2)
		})

		t.Run("returns empty for non-existent item", func(t *testing.T) {
			skus, err := repo.FindByItemID(ctx, 999999)
			assert.NoError(t, err)
			assert.Empty(t, skus)
		})
	})

	t.Run("Update", func(t *testing.T) {
		product := createTestProduct(t, 4001)
		modelID := int64(40010)
		sku := createTestSku(t, product, modelID, "SKU-UPDATE")

		sku.Price = 99999
		sku.Quantity = 500
		sku.VariantName = "Size: XL"

		err := repo.Update(ctx, sku)
		assert.NoError(t, err)

		// Verify update
		found, err := repo.FindByModelID(ctx, modelID)
		assert.NoError(t, err)
		assert.Equal(t, float64(99999), found.Price)
		assert.Equal(t, 500, found.Quantity)
		assert.Equal(t, "Size: XL", found.VariantName)
	})

	t.Run("Upsert", func(t *testing.T) {
		product := createTestProduct(t, 5001)

		t.Run("creates new SKU with model ID", func(t *testing.T) {
			modelID := int64(50010)
			newSku := &models.ShopeeSku{
				TenantID:    "test-tenant",
				ProductID:   product.ID,
				ItemID:      product.ItemID,
				ModelID:     &modelID,
				SellerSku:   "SKU-UPSERT-NEW",
				VariantName: "Color: Blue",
				Price:       40000,
				Quantity:    60,
			}
			err := repo.Upsert(ctx, newSku)
			assert.NoError(t, err)
			assert.NotZero(t, newSku.ID)

			// Verify created
			found, err := repo.FindByModelID(ctx, modelID)
			assert.NoError(t, err)
			assert.Equal(t, "SKU-UPSERT-NEW", found.SellerSku)
		})

		t.Run("updates existing SKU by model ID", func(t *testing.T) {
			modelID := int64(50020)
			sku := createTestSku(t, product, modelID, "SKU-UPSERT-EXIST")

			// Upsert with updated data
			updateSku := &models.ShopeeSku{
				TenantID:    sku.TenantID,
				ProductID:   sku.ProductID,
				ItemID:      sku.ItemID,
				ModelID:     &modelID,
				SellerSku:   "SKU-UPSERT-UPDATED",
				VariantName: "Size: XXL",
				Price:       88888,
				Quantity:    999,
			}
			err := repo.Upsert(ctx, updateSku)
			assert.NoError(t, err)

			// Verify updated
			found, err := repo.FindByModelID(ctx, modelID)
			assert.NoError(t, err)
			assert.Equal(t, "SKU-UPSERT-UPDATED", found.SellerSku)
			assert.Equal(t, float64(88888), found.Price)
		})

		t.Run("upserts non-variant SKU by item ID", func(t *testing.T) {
			nonVariantProduct := createTestProduct(t, 5002)
			nonVariantSku := &models.ShopeeSku{
				TenantID:  "test-tenant",
				ProductID: nonVariantProduct.ID,
				ItemID:    nonVariantProduct.ItemID,
				ModelID:   nil,
				SellerSku: "SKU-NONVARIANT",
				Price:     55000,
				Quantity:  100,
			}
			err := repo.Upsert(ctx, nonVariantSku)
			assert.NoError(t, err)
			assert.NotZero(t, nonVariantSku.ID)
		})
	})

	t.Run("DeleteByItemID", func(t *testing.T) {
		product := createTestProduct(t, 6001)
		modelID1 := int64(60010)
		modelID2 := int64(60020)
		createTestSku(t, product, modelID1, "SKU-DELETE-001")
		createTestSku(t, product, modelID2, "SKU-DELETE-002")

		// Verify SKUs exist
		skus, err := repo.FindByItemID(ctx, product.ItemID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(skus), 2)

		// Delete all SKUs for item
		err = repo.DeleteByItemID(ctx, product.ItemID)
		assert.NoError(t, err)

		// Verify deleted
		skus, err = repo.FindByItemID(ctx, product.ItemID)
		assert.NoError(t, err)
		assert.Empty(t, skus)
	})
}
