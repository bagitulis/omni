package repositories

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTiktokProductRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.TiktokProduct{})
	repo := NewTiktokProductRepository(db)
	ctx := context.Background()

	// Helper to create test product
	createTestProduct := func(t *testing.T, productID string, name string) *models.TiktokProduct {
		product := &models.TiktokProduct{
			TenantID:    "test-tenant",
			ProductID:   productID,
			Name:        name,
			Description: "Test Description",
			Status:      "ACTIVE",
			Price:       100000,
			Quantity:    200,
		}
		err := db.Create(product).Error
		require.NoError(t, err)
		return product
	}

	t.Run("FindByProductID", func(t *testing.T) {
		product := createTestProduct(t, "TT-PROD-001", "Find Product")

		tests := []struct {
			name      string
			productID string
			wantErr   bool
		}{
			{
				name:      "finds existing product",
				productID: product.ProductID,
				wantErr:   false,
			},
			{
				name:      "returns error for non-existent product",
				productID: "NONEXISTENT",
				wantErr:   true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				found, err := repo.FindByProductID(ctx, tt.productID)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				require.NotNil(t, found)
				assert.Equal(t, tt.productID, found.ProductID)
			})
		}
	})

	t.Run("FindAll", func(t *testing.T) {
		// Create multiple products
		createTestProduct(t, "TT-PROD-FA01", "FindAll Product 1")
		createTestProduct(t, "TT-PROD-FA02", "FindAll Product 2")
		createTestProduct(t, "TT-PROD-FA03", "FindAll Product 3")

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
			newProduct := &models.TiktokProduct{
				TenantID:  "test-tenant",
				ProductID: "TT-PROD-UPSERT01",
				Name:      "Upsert New Product",
				Status:    "ACTIVE",
				Price:     75000,
			}
			err := repo.Upsert(ctx, newProduct)
			assert.NoError(t, err)
			assert.NotZero(t, newProduct.ID)

			// Verify created
			found, err := repo.FindByProductID(ctx, "TT-PROD-UPSERT01")
			assert.NoError(t, err)
			assert.Equal(t, "Upsert New Product", found.Name)
		})

		t.Run("updates existing product", func(t *testing.T) {
			product := createTestProduct(t, "TT-PROD-UPSERT02", "Upsert Existing")

			// Upsert with updated data
			updateProduct := &models.TiktokProduct{
				TenantID:  product.TenantID,
				ProductID: product.ProductID,
				Name:      "Upsert Updated Name",
				Status:    "INACTIVE",
				Price:     99999,
			}
			err := repo.Upsert(ctx, updateProduct)
			assert.NoError(t, err)

			// Verify updated
			found, err := repo.FindByProductID(ctx, product.ProductID)
			assert.NoError(t, err)
			assert.Equal(t, "Upsert Updated Name", found.Name)
			assert.Equal(t, "INACTIVE", found.Status)
		})
	})

	t.Run("Search", func(t *testing.T) {
		// Create products with searchable names
		createTestProduct(t, "TT-PROD-SRCH01", "Special Widget Product")
		createTestProduct(t, "TT-PROD-SRCH02", "Another Widget Item")
		createTestProduct(t, "TT-PROD-SRCH03", "Gadget Product")

		tests := []struct {
			name     string
			query    string
			page     int
			pageSize int
			wantMin  int
		}{
			{
				name:     "finds products by name",
				query:    "Widget",
				page:     1,
				pageSize: 10,
				wantMin:  2,
			},
			{
				name:     "finds products by partial name",
				query:    "Product",
				page:     1,
				pageSize: 10,
				wantMin:  2,
			},
			{
				name:     "returns empty for non-matching query",
				query:    "ZZZZNONEXISTENT",
				page:     1,
				pageSize: 10,
				wantMin:  0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				products, total, err := repo.Search(ctx, tt.query, tt.page, tt.pageSize)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, int(total), tt.wantMin)
				if tt.wantMin > 0 {
					assert.NotEmpty(t, products)
				}
			})
		}
	})
}
