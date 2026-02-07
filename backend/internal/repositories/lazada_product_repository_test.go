package repositories

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLazadaProductRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.LazadaProduct{}, &models.LazadaSku{})
	repo := NewLazadaProductRepository(db)
	ctx := context.Background()

	// Helper to create test product
	createTestProduct := func(t *testing.T, itemID string, name string) *models.LazadaProduct {
		product := &models.LazadaProduct{
			TenantID:    "test-tenant",
			ItemID:      itemID,
			Name:        name,
			Description: "Test Description",
			Status:      "active",
			Price:       75000,
			Quantity:    150,
		}
		err := db.Create(product).Error
		require.NoError(t, err)
		return product
	}

	// Helper to create test SKU
	createTestSku := func(t *testing.T, product *models.LazadaProduct, skuID string, sellerSku string) *models.LazadaSku {
		sku := &models.LazadaSku{
			TenantID:  "test-tenant",
			ItemID:    product.ItemID,
			ProductID: product.ID,
			SkuID:     skuID,
			SellerSku: sellerSku,
			Name:      "SKU Name",
			Price:     65000,
			Quantity:  50,
		}
		err := db.Create(sku).Error
		require.NoError(t, err)
		return sku
	}

	t.Run("FindByItemID", func(t *testing.T) {
		product := createTestProduct(t, "LZD-ITEM-001", "Find Product")

		tests := []struct {
			name    string
			itemID  string
			wantErr bool
		}{
			{
				name:    "finds existing product",
				itemID:  product.ItemID,
				wantErr: false,
			},
			{
				name:    "returns error for non-existent product",
				itemID:  "NONEXISTENT",
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				found, err := repo.FindByItemID(ctx, tt.itemID)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				require.NotNil(t, found)
				assert.Equal(t, tt.itemID, found.ItemID)
			})
		}
	})

	t.Run("FindAll", func(t *testing.T) {
		// Create multiple products
		createTestProduct(t, "LZD-ITEM-FA01", "FindAll Product 1")
		createTestProduct(t, "LZD-ITEM-FA02", "FindAll Product 2")
		createTestProduct(t, "LZD-ITEM-FA03", "FindAll Product 3")

		tests := []struct {
			name     string
			page     int
			pageSize int
			wantMin  int
		}{
			{
				name:     "returns paginated results",
				page:     1,
				pageSize: 10,
				wantMin:  3,
			},
			{
				name:     "handles page size limit",
				page:     1,
				pageSize: 2,
				wantMin:  2,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				products, total, err := repo.FindAll(ctx, tt.page, tt.pageSize)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, int(total), tt.wantMin)
				assert.LessOrEqual(t, len(products), tt.pageSize)
			})
		}
	})

	t.Run("Upsert", func(t *testing.T) {
		t.Run("creates new product", func(t *testing.T) {
			newProduct := &models.LazadaProduct{
				TenantID: "test-tenant",
				ItemID:   "LZD-ITEM-UPSERT01",
				Name:     "Upsert New Product",
				Status:   "active",
				Price:    55000,
			}
			err := repo.Upsert(ctx, newProduct)
			assert.NoError(t, err)
			assert.NotZero(t, newProduct.ID)

			// Verify created
			found, err := repo.FindByItemID(ctx, "LZD-ITEM-UPSERT01")
			assert.NoError(t, err)
			assert.Equal(t, "Upsert New Product", found.Name)
		})

		t.Run("updates existing product", func(t *testing.T) {
			product := createTestProduct(t, "LZD-ITEM-UPSERT02", "Upsert Existing")

			// Upsert with updated data
			updateProduct := &models.LazadaProduct{
				TenantID: product.TenantID,
				ItemID:   product.ItemID,
				Name:     "Upsert Updated Name",
				Status:   "inactive",
				Price:    99999,
			}
			err := repo.Upsert(ctx, updateProduct)
			assert.NoError(t, err)

			// Verify updated
			found, err := repo.FindByItemID(ctx, product.ItemID)
			assert.NoError(t, err)
			assert.Equal(t, "Upsert Updated Name", found.Name)
			assert.Equal(t, "inactive", found.Status)
		})

		t.Run("does not wipe existing name/image when new data is empty", func(t *testing.T) {
			product := createTestProduct(t, "LZD-ITEM-UPSERT03", "Existing Name")
			err := db.Model(&models.LazadaProduct{}).
				Where("id = ?", product.ID).
				Updates(map[string]interface{}{"image": "https://example.com/existing.jpg"}).Error
			require.NoError(t, err)

			updateProduct := &models.LazadaProduct{
				TenantID: product.TenantID,
				ItemID:   product.ItemID,
				Name:     "",
				Image:    "",
				Status:   "active",
				Price:    12345,
			}
			err = repo.Upsert(ctx, updateProduct)
			assert.NoError(t, err)

			found, err := repo.FindByItemID(ctx, product.ItemID)
			require.NoError(t, err)
			require.NotNil(t, found)
			assert.Equal(t, "Existing Name", found.Name)
			assert.Equal(t, "https://example.com/existing.jpg", found.Image)
		})
	})

	t.Run("UpsertSku", func(t *testing.T) {
		product := createTestProduct(t, "LZD-ITEM-SKU01", "Product for SKU Test")

		t.Run("creates new SKU", func(t *testing.T) {
			newSku := &models.LazadaSku{
				TenantID:  "test-tenant",
				ItemID:    product.ItemID,
				ProductID: product.ID,
				SkuID:     "LZD-SKU-NEW01",
				SellerSku: "SELLER-SKU-001",
				Name:      "New SKU",
				Price:     45000,
				Quantity:  30,
			}
			err := repo.UpsertSku(ctx, newSku)
			assert.NoError(t, err)
			assert.NotZero(t, newSku.ID)
		})

		t.Run("updates existing SKU", func(t *testing.T) {
			sku := createTestSku(t, product, "LZD-SKU-EXIST01", "SELLER-SKU-002")

			// Upsert with updated data
			updateSku := &models.LazadaSku{
				TenantID:  sku.TenantID,
				ItemID:    sku.ItemID,
				ProductID: sku.ProductID,
				SkuID:     sku.SkuID,
				SellerSku: "SELLER-SKU-UPDATED",
				Name:      "Updated SKU Name",
				Price:     88888,
				Quantity:  999,
			}
			err := repo.UpsertSku(ctx, updateSku)
			assert.NoError(t, err)

			// Verify updated
			skus, err := repo.FindSkusByItemID(ctx, product.ItemID)
			assert.NoError(t, err)
			var foundSku *models.LazadaSku
			for i := range skus {
				if skus[i].SkuID == sku.SkuID {
					foundSku = &skus[i]
					break
				}
			}
			require.NotNil(t, foundSku)
			assert.Equal(t, "Updated SKU Name", foundSku.Name)
			assert.Equal(t, float64(88888), foundSku.Price)
		})

		t.Run("does not wipe existing seller_sku/name when new data is empty", func(t *testing.T) {
			sku := createTestSku(t, product, "LZD-SKU-EXIST02", "SELLER-SKU-EXIST")
			updateSku := &models.LazadaSku{
				TenantID:  sku.TenantID,
				ItemID:    sku.ItemID,
				ProductID: sku.ProductID,
				SkuID:     sku.SkuID,
				SellerSku: "",
				Name:      "",
				Price:     77777,
				Quantity:  0,
			}
			err := repo.UpsertSku(ctx, updateSku)
			assert.NoError(t, err)

			skus, err := repo.FindSkusByItemID(ctx, product.ItemID)
			require.NoError(t, err)
			var foundSku *models.LazadaSku
			for i := range skus {
				if skus[i].SkuID == sku.SkuID {
					foundSku = &skus[i]
					break
				}
			}
			require.NotNil(t, foundSku)
			assert.Equal(t, "SELLER-SKU-EXIST", foundSku.SellerSku)
			assert.Equal(t, "SKU Name", foundSku.Name)
			assert.Equal(t, float64(77777), foundSku.Price)
		})
	})

	t.Run("FindSkusByItemID", func(t *testing.T) {
		product := createTestProduct(t, "LZD-ITEM-SKUS01", "Product with Multiple SKUs")
		createTestSku(t, product, "LZD-SKU-MULTI01", "MULTI-001")
		createTestSku(t, product, "LZD-SKU-MULTI02", "MULTI-002")
		createTestSku(t, product, "LZD-SKU-MULTI03", "MULTI-003")

		t.Run("finds all SKUs for item", func(t *testing.T) {
			skus, err := repo.FindSkusByItemID(ctx, product.ItemID)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(skus), 3)
		})

		t.Run("returns empty for non-existent item", func(t *testing.T) {
			skus, err := repo.FindSkusByItemID(ctx, "NONEXISTENT-ITEM")
			assert.NoError(t, err)
			assert.Empty(t, skus)
		})
	})

	t.Run("GetProductsWithSkus", func(t *testing.T) {
		// Create product with SKUs
		product := createTestProduct(t, "LZD-ITEM-JOINED01", "Product with Joined SKUs")
		createTestSku(t, product, "LZD-SKU-JOINED01", "JOINED-001")
		createTestSku(t, product, "LZD-SKU-JOINED02", "JOINED-002")

		results, total, err := repo.GetProductsWithSkus(ctx, 1, 20)
		assert.NoError(t, err)
		assert.Greater(t, int(total), 0)
		assert.NotEmpty(t, results)

		// Verify result structure
		for _, result := range results {
			assert.Contains(t, result, "item_id")
			assert.Contains(t, result, "product_name")
			assert.Contains(t, result, "price")
			assert.Contains(t, result, "quantity")
		}
	})
}
