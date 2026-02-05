package repositories

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTiktokSkuRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.TiktokProduct{}, &models.TiktokSku{})
	repo := NewTiktokSkuRepository(db)
	ctx := context.Background()

	// Helper to create test product first
	createTestProduct := func(t *testing.T, productID string) *models.TiktokProduct {
		product := &models.TiktokProduct{
			TenantID:  "test-tenant",
			ProductID: productID,
			Name:      "Test Product",
			Status:    "ACTIVE",
			Price:     100000,
			Quantity:  200,
		}
		err := db.Create(product).Error
		require.NoError(t, err)
		return product
	}

	// Helper to create test SKU
	createTestSku := func(t *testing.T, product *models.TiktokProduct, skuID string, sellerSku string) *models.TiktokSku {
		sku := &models.TiktokSku{
			TenantID:    "test-tenant",
			ProductID:   product.ID,
			SkuID:       skuID,
			SellerSku:   sellerSku,
			VariantName: "Size: M",
			Price:       85000,
			Quantity:    50,
		}
		err := db.Create(sku).Error
		require.NoError(t, err)
		return sku
	}

	t.Run("Upsert", func(t *testing.T) {
		product := createTestProduct(t, "TT-SKU-PROD-001")

		t.Run("creates new SKU", func(t *testing.T) {
			newSku := &models.TiktokSku{
				TenantID:    "test-tenant",
				ProductID:   product.ID,
				SkuID:       "TT-SKU-NEW001",
				SellerSku:   "SELLER-NEW-001",
				VariantName: "Color: Red",
				Price:       75000,
				Quantity:    100,
			}
			err := repo.Upsert(ctx, newSku)
			assert.NoError(t, err)
			assert.NotZero(t, newSku.ID)

			// Verify created
			found, err := repo.FindBySkuID(ctx, "TT-SKU-NEW001")
			assert.NoError(t, err)
			require.NotNil(t, found)
			assert.Equal(t, "SELLER-NEW-001", found.SellerSku)
		})

		t.Run("updates existing SKU", func(t *testing.T) {
			sku := createTestSku(t, product, "TT-SKU-EXIST001", "SELLER-EXIST-001")

			// Upsert with updated data
			updateSku := &models.TiktokSku{
				TenantID:    sku.TenantID,
				ProductID:   sku.ProductID,
				SkuID:       sku.SkuID,
				SellerSku:   "SELLER-UPDATED",
				VariantName: "Size: XL",
				Price:       99999,
				Quantity:    999,
			}
			err := repo.Upsert(ctx, updateSku)
			assert.NoError(t, err)

			// Verify updated
			found, err := repo.FindBySkuID(ctx, sku.SkuID)
			assert.NoError(t, err)
			require.NotNil(t, found)
			assert.Equal(t, "SELLER-UPDATED", found.SellerSku)
			assert.Equal(t, float64(99999), found.Price)
		})
	})

	t.Run("FindBySkuID", func(t *testing.T) {
		product := createTestProduct(t, "TT-SKU-PROD-002")
		sku := createTestSku(t, product, "TT-SKU-FIND001", "SELLER-FIND")

		tests := []struct {
			name    string
			skuID   string
			wantNil bool
		}{
			{
				name:    "finds existing SKU",
				skuID:   sku.SkuID,
				wantNil: false,
			},
			{
				name:    "returns nil for non-existent SKU",
				skuID:   "NONEXISTENT",
				wantNil: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				found, err := repo.FindBySkuID(ctx, tt.skuID)
				assert.NoError(t, err)
				if tt.wantNil {
					assert.Nil(t, found)
				} else {
					require.NotNil(t, found)
					assert.Equal(t, tt.skuID, found.SkuID)
				}
			})
		}
	})

	t.Run("FindByProductID", func(t *testing.T) {
		product := createTestProduct(t, "TT-SKU-PROD-003")
		createTestSku(t, product, "TT-SKU-MULTI001", "MULTI-001")
		createTestSku(t, product, "TT-SKU-MULTI002", "MULTI-002")
		createTestSku(t, product, "TT-SKU-MULTI003", "MULTI-003")

		t.Run("finds all SKUs for product", func(t *testing.T) {
			skus, err := repo.FindByProductID(ctx, product.ID)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(skus), 3)
		})

		t.Run("returns empty for non-existent product", func(t *testing.T) {
			skus, err := repo.FindByProductID(ctx, 999999)
			assert.NoError(t, err)
			assert.Empty(t, skus)
		})
	})

	t.Run("DeleteByProductID", func(t *testing.T) {
		product := createTestProduct(t, "TT-SKU-PROD-004")
		createTestSku(t, product, "TT-SKU-DEL001", "DEL-001")
		createTestSku(t, product, "TT-SKU-DEL002", "DEL-002")

		// Verify SKUs exist
		skus, err := repo.FindByProductID(ctx, product.ID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(skus), 2)

		// Delete all SKUs for product
		err = repo.DeleteByProductID(ctx, product.ID)
		assert.NoError(t, err)

		// Verify deleted
		skus, err = repo.FindByProductID(ctx, product.ID)
		assert.NoError(t, err)
		assert.Empty(t, skus)
	})

	t.Run("Count", func(t *testing.T) {
		// Create some SKUs
		product := createTestProduct(t, "TT-SKU-PROD-005")
		createTestSku(t, product, "TT-SKU-COUNT001", "COUNT-001")
		createTestSku(t, product, "TT-SKU-COUNT002", "COUNT-002")

		count, err := repo.Count(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))
	})
}
