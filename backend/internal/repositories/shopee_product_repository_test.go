package repositories

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShopeeProductRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.ShopeeProduct{}, &models.ShopeeSku{})
	repo := NewShopeeProductRepository(db)
	ctx := context.Background()

	// Helper to create test product
	createTestProduct := func(t *testing.T, itemID int64, name string) *models.ShopeeProduct {
		product := &models.ShopeeProduct{
			TenantID:    "test-tenant",
			ItemID:      itemID,
			Name:        name,
			Description: "Test Description",
			Status:      "NORMAL",
			Price:       50000,
			Quantity:    100,
		}
		err := repo.Create(ctx, product)
		require.NoError(t, err)
		return product
	}

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name    string
			product *models.ShopeeProduct
			wantErr bool
		}{
			{
				name: "creates product successfully",
				product: &models.ShopeeProduct{
					TenantID: "test-tenant",
					ItemID:   1001,
					Name:     "Test Product 1",
					Status:   "NORMAL",
					Price:    25000,
					Quantity: 50,
				},
				wantErr: false,
			},
			{
				name: "creates product with all fields",
				product: &models.ShopeeProduct{
					TenantID:    "test-tenant",
					ItemID:      1002,
					Name:        "Test Product 2",
					Description: "Full description",
					Status:      "BANNED",
					Price:       75000,
					Quantity:    200,
					Image:       "https://example.com/image.jpg",
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := repo.Create(ctx, tt.product)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				assert.NotZero(t, tt.product.ID)
				assert.False(t, tt.product.CreatedAt.IsZero())
			})
		}
	})

	t.Run("FindByItemID", func(t *testing.T) {
		product := createTestProduct(t, 2001, "Find By ItemID Product")

		tests := []struct {
			name    string
			itemID  int64
			wantErr bool
		}{
			{
				name:    "finds existing product",
				itemID:  product.ItemID,
				wantErr: false,
			},
			{
				name:    "returns error for non-existent product",
				itemID:  999999,
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
		createTestProduct(t, 3001, "FindAll Product 1")
		createTestProduct(t, 3002, "FindAll Product 2")
		createTestProduct(t, 3003, "FindAll Product 3")

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

	t.Run("Update", func(t *testing.T) {
		product := createTestProduct(t, 4001, "Update Product")

		product.Name = "Updated Name"
		product.Price = 99999
		product.Status = "DELETED"

		err := repo.Update(ctx, product)
		assert.NoError(t, err)

		// Verify update
		found, err := repo.FindByItemID(ctx, product.ItemID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, float64(99999), found.Price)
		assert.Equal(t, "DELETED", found.Status)
	})

	t.Run("Upsert", func(t *testing.T) {
		t.Run("creates new product", func(t *testing.T) {
			newProduct := &models.ShopeeProduct{
				TenantID: "test-tenant",
				ItemID:   5001,
				Name:     "Upsert New Product",
				Status:   "NORMAL",
				Price:    10000,
			}
			err := repo.Upsert(ctx, newProduct)
			assert.NoError(t, err)
			assert.NotZero(t, newProduct.ID)

			// Verify created
			found, err := repo.FindByItemID(ctx, 5001)
			assert.NoError(t, err)
			assert.Equal(t, "Upsert New Product", found.Name)
		})

		t.Run("updates existing product", func(t *testing.T) {
			product := createTestProduct(t, 5002, "Upsert Existing")

			// Upsert with updated data
			updateProduct := &models.ShopeeProduct{
				TenantID: product.TenantID,
				ItemID:   product.ItemID,
				Name:     "Upsert Updated Name",
				Status:   "BANNED",
				Price:    88888,
			}
			err := repo.Upsert(ctx, updateProduct)
			assert.NoError(t, err)

			// Verify updated
			found, err := repo.FindByItemID(ctx, product.ItemID)
			assert.NoError(t, err)
			assert.Equal(t, "Upsert Updated Name", found.Name)
			assert.Equal(t, "BANNED", found.Status)
		})
	})

	t.Run("FindBySKU", func(t *testing.T) {
		// Create a product and SKU
		product := createTestProduct(t, 6001, "FindBySKU Product")

		skuRepo := NewShopeeSkuRepository(db)
		testSku := &models.ShopeeSku{
			TenantID:  "test-tenant",
			ProductID: product.ID,
			ItemID:    product.ItemID,
			SellerSku: "TEST-SKU-001",
			Price:     10000,
			Quantity:  5,
		}
		err := skuRepo.Create(ctx, testSku)
		require.NoError(t, err)

		t.Run("finds product by seller_sku", func(t *testing.T) {
			found, err := repo.FindBySKU(ctx, "TEST-SKU-001")
			assert.NoError(t, err)
			require.NotNil(t, found)
			assert.Equal(t, product.ItemID, found.ItemID)
		})

		t.Run("returns error for non-existent SKU", func(t *testing.T) {
			_, err := repo.FindBySKU(ctx, "NON-EXISTENT")
			assert.Error(t, err)
		})
	})
}
