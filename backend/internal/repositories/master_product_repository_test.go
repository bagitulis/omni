package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMasterProductRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t,
		&models.MasterProduct{},
		&models.MasterProductSku{},
		&models.MasterProductPlatformLink{},
	)
	repo := NewMasterProductRepository(db)
	ctx := context.Background()

	tenantID := "test-tenant"

	// Helper to create test product
	createTestProduct := func(t *testing.T, title string) *models.MasterProduct {
		product := &models.MasterProduct{
			TenantID:    tenantID,
			Title:       title,
			Description: "Test Description",
			Status:      models.MasterProductStatusActive,
		}
		err := repo.Create(ctx, product)
		require.NoError(t, err)
		return product
	}

	// Helper to create test SKU
	createTestSku := func(t *testing.T, product *models.MasterProduct, sellerSku string) *models.MasterProductSku {
		sku := &models.MasterProductSku{
			TenantID:        tenantID,
			MasterProductID: product.ID,
			SellerSku:       sellerSku,
			VariantName:     "Size: M",
			Price:           50000,
			Stock:           100,
		}
		err := repo.CreateSku(ctx, sku)
		require.NoError(t, err)
		return sku
	}

	// ================== Product CRUD Tests ==================

	t.Run("Create", func(t *testing.T) {
		product := &models.MasterProduct{
			TenantID:    tenantID,
			Title:       "Create Test Product",
			Description: "Description",
			Status:      models.MasterProductStatusDraft,
		}
		err := repo.Create(ctx, product)
		assert.NoError(t, err)
		assert.NotZero(t, product.ID)
		assert.False(t, product.CreatedAt.IsZero())
	})

	t.Run("FindByID", func(t *testing.T) {
		product := createTestProduct(t, "FindByID Product")

		tests := []struct {
			name    string
			id      uint
			wantErr bool
		}{
			{
				name:    "finds existing product",
				id:      product.ID,
				wantErr: false,
			},
			{
				name:    "returns error for non-existent product",
				id:      999999,
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				found, err := repo.FindByID(ctx, tt.id)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				require.NotNil(t, found)
				assert.Equal(t, tt.id, found.ID)
			})
		}
	})

	t.Run("FindAll", func(t *testing.T) {
		createTestProduct(t, "FindAll Product 1")
		createTestProduct(t, "FindAll Product 2")
		createTestProduct(t, "FindAll Product 3")

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
				products, total, err := repo.FindAll(ctx, tenantID, tt.page, tt.pageSize)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, int(total), tt.wantMin)
				assert.LessOrEqual(t, len(products), tt.pageSize)
			})
		}
	})

	t.Run("FindByTenantAndID", func(t *testing.T) {
		product := createTestProduct(t, "FindByTenantAndID Product")

		tests := []struct {
			name     string
			tenantID string
			id       uint
			wantErr  bool
		}{
			{
				name:     "finds product with correct tenant",
				tenantID: tenantID,
				id:       product.ID,
				wantErr:  false,
			},
			{
				name:     "returns error for wrong tenant",
				tenantID: "wrong-tenant",
				id:       product.ID,
				wantErr:  true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				found, err := repo.FindByTenantAndID(ctx, tt.tenantID, tt.id)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				require.NotNil(t, found)
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		product := createTestProduct(t, "Update Product")

		product.Title = "Updated Title"
		product.Status = models.MasterProductStatusArchived

		err := repo.Update(ctx, product)
		assert.NoError(t, err)

		// Verify update
		found, err := repo.FindByID(ctx, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", found.Title)
		assert.Equal(t, models.MasterProductStatusArchived, found.Status)
	})

	t.Run("Delete", func(t *testing.T) {
		product := createTestProduct(t, "Delete Product")

		err := repo.Delete(ctx, product.ID)
		assert.NoError(t, err)

		// Verify deleted
		_, err = repo.FindByID(ctx, product.ID)
		assert.Error(t, err)
	})

	// ================== SKU Operations Tests ==================

	t.Run("CreateSku", func(t *testing.T) {
		product := createTestProduct(t, "Product for SKU")

		sku := &models.MasterProductSku{
			TenantID:        tenantID,
			MasterProductID: product.ID,
			SellerSku:       "SKU-CREATE-001",
			VariantName:     "Color: Red",
			Price:           35000,
			Stock:           50,
		}
		err := repo.CreateSku(ctx, sku)
		assert.NoError(t, err)
		assert.NotZero(t, sku.ID)
	})

	t.Run("FindBySku", func(t *testing.T) {
		product := createTestProduct(t, "Product for FindBySku")
		sku := createTestSku(t, product, "SKU-FIND-001")

		tests := []struct {
			name      string
			sellerSku string
			wantErr   bool
		}{
			{
				name:      "finds existing SKU",
				sellerSku: sku.SellerSku,
				wantErr:   false,
			},
			{
				name:      "returns error for non-existent SKU",
				sellerSku: "NONEXISTENT",
				wantErr:   true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				found, err := repo.FindBySku(ctx, tenantID, tt.sellerSku)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				require.NotNil(t, found)
				assert.Equal(t, tt.sellerSku, found.SellerSku)
			})
		}
	})

	t.Run("UpdateSku", func(t *testing.T) {
		product := createTestProduct(t, "Product for UpdateSku")
		sku := createTestSku(t, product, "SKU-UPDATE-001")

		sku.Price = 99999
		sku.Stock = 500

		err := repo.UpdateSku(ctx, sku)
		assert.NoError(t, err)

		// Verify update
		found, err := repo.FindBySku(ctx, tenantID, sku.SellerSku)
		assert.NoError(t, err)
		assert.Equal(t, float64(99999), found.Price)
		assert.Equal(t, 500, found.Stock)
	})

	t.Run("DeleteSku", func(t *testing.T) {
		product := createTestProduct(t, "Product for DeleteSku")
		sku := createTestSku(t, product, "SKU-DELETE-001")

		err := repo.DeleteSku(ctx, sku.ID)
		assert.NoError(t, err)

		// Verify deleted
		_, err = repo.FindBySku(ctx, tenantID, sku.SellerSku)
		assert.Error(t, err)
	})

	t.Run("FindSkusByProductID", func(t *testing.T) {
		product := createTestProduct(t, "Product with Multiple SKUs")
		createTestSku(t, product, "MULTI-SKU-001")
		createTestSku(t, product, "MULTI-SKU-002")
		createTestSku(t, product, "MULTI-SKU-003")

		skus, err := repo.FindSkusByProductID(ctx, product.ID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(skus), 3)
	})

	// ================== Platform Link Tests ==================

	t.Run("CreatePlatformLink", func(t *testing.T) {
		product := createTestProduct(t, "Product for Link")

		link := &models.MasterProductPlatformLink{
			MasterProductID:   product.ID,
			Platform:          "shopee",
			PlatformProductID: "SHOPEE-123",
			PlatformItemID:    "ITEM-123",
			SyncStatus:        models.SyncStatusPending,
		}
		err := repo.CreatePlatformLink(ctx, link)
		assert.NoError(t, err)
		assert.NotZero(t, link.ID)
	})

	t.Run("FindPlatformLinks", func(t *testing.T) {
		product := createTestProduct(t, "Product with Links")

		// Create multiple links
		for _, platform := range []string{"shopee", "lazada", "tiktok"} {
			link := &models.MasterProductPlatformLink{
				MasterProductID:   product.ID,
				Platform:          platform,
				PlatformProductID: platform + "-PROD-123",
				SyncStatus:        models.SyncStatusPending,
			}
			err := repo.CreatePlatformLink(ctx, link)
			require.NoError(t, err)
		}

		links, err := repo.FindPlatformLinks(ctx, product.ID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(links), 3)
	})

	t.Run("FindPlatformLinkByPlatform", func(t *testing.T) {
		product := createTestProduct(t, "Product for FindLinkByPlatform")

		link := &models.MasterProductPlatformLink{
			MasterProductID:   product.ID,
			Platform:          "tiktok",
			PlatformProductID: "TIKTOK-PROD-456",
			SyncStatus:        models.SyncStatusSynced,
		}
		err := repo.CreatePlatformLink(ctx, link)
		require.NoError(t, err)

		found, err := repo.FindPlatformLinkByPlatform(ctx, product.ID, "tiktok")
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "tiktok", found.Platform)
	})

	t.Run("UpdateLinkStatus", func(t *testing.T) {
		product := createTestProduct(t, "Product for UpdateLinkStatus")

		link := &models.MasterProductPlatformLink{
			MasterProductID:   product.ID,
			Platform:          "shopee",
			PlatformProductID: "SHOPEE-STATUS-123",
			SyncStatus:        models.SyncStatusPending,
		}
		err := repo.CreatePlatformLink(ctx, link)
		require.NoError(t, err)

		err = repo.UpdateLinkStatus(ctx, link.ID, models.SyncStatusSynced)
		assert.NoError(t, err)

		// Verify update
		found, err := repo.FindPlatformLinkByPlatform(ctx, product.ID, "shopee")
		assert.NoError(t, err)
		assert.Equal(t, models.SyncStatusSynced, found.SyncStatus)
		assert.NotNil(t, found.LastSyncedAt)
	})

	t.Run("DeletePlatformLink", func(t *testing.T) {
		product := createTestProduct(t, "Product for DeleteLink")

		link := &models.MasterProductPlatformLink{
			MasterProductID:   product.ID,
			Platform:          "lazada",
			PlatformProductID: "LAZADA-DELETE-123",
			SyncStatus:        models.SyncStatusPending,
		}
		err := repo.CreatePlatformLink(ctx, link)
		require.NoError(t, err)

		err = repo.DeletePlatformLink(ctx, link.ID)
		assert.NoError(t, err)

		// Verify deleted
		_, err = repo.FindPlatformLinkByPlatform(ctx, product.ID, "lazada")
		assert.Error(t, err)
	})

	t.Run("UpsertPlatformLink", func(t *testing.T) {
		product := createTestProduct(t, "Product for UpsertLink")

		// Create the unique index that UpsertPlatformLink relies on
		tableName := models.MasterProductPlatformLink{}.TableName()
		err := db.Exec(fmt.Sprintf(
			"CREATE UNIQUE INDEX IF NOT EXISTS idx_platform_links_unique ON %s (platform, platform_product_id, platform_sku_id)",
			tableName,
		)).Error
		require.NoError(t, err)

		t.Run("creates new link", func(t *testing.T) {
			link := &models.MasterProductPlatformLink{
				MasterProductID:   product.ID,
				Platform:          "shopee",
				PlatformProductID: "UPSERT-NEW-123",
				SyncStatus:        models.SyncStatusPending,
			}
			err := repo.UpsertPlatformLink(ctx, link)
			assert.NoError(t, err)
			assert.NotZero(t, link.ID)
		})

		t.Run("updates existing link", func(t *testing.T) {
			// Create initial link
			link := &models.MasterProductPlatformLink{
				MasterProductID:   product.ID,
				Platform:          "tiktok",
				PlatformProductID: "UPSERT-EXIST-123",
				SyncStatus:        models.SyncStatusPending,
			}
			err := repo.UpsertPlatformLink(ctx, link)
			require.NoError(t, err)

			// Upsert with same conflict key but updated fields
			updateLink := &models.MasterProductPlatformLink{
				MasterProductID:   product.ID,
				Platform:          "tiktok",
				PlatformProductID: "UPSERT-EXIST-123",
				PlatformItemID:    "UPDATED-ITEM-ID",
				SyncStatus:        models.SyncStatusSynced,
			}
			err = repo.UpsertPlatformLink(ctx, updateLink)
			assert.NoError(t, err)

			// Verify updated
			found, err := repo.FindPlatformLinkByPlatform(ctx, product.ID, "tiktok")
			assert.NoError(t, err)
			assert.Equal(t, "UPSERT-EXIST-123", found.PlatformProductID)
			assert.Equal(t, "UPDATED-ITEM-ID", found.PlatformItemID)
			assert.Equal(t, models.SyncStatusSynced, found.SyncStatus)
		})
	})

	// ================== Query Helper Tests ==================

	t.Run("CountByTenant", func(t *testing.T) {
		createTestProduct(t, "Count Product 1")
		createTestProduct(t, "Count Product 2")

		count, err := repo.CountByTenant(ctx, tenantID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))
	})

	t.Run("FindByStatus", func(t *testing.T) {
		// Create products with specific status
		for i := 0; i < 3; i++ {
			product := &models.MasterProduct{
				TenantID: tenantID,
				Title:    "Status Draft Product",
				Status:   models.MasterProductStatusDraft,
			}
			err := repo.Create(ctx, product)
			require.NoError(t, err)
		}

		products, total, err := repo.FindByStatus(ctx, tenantID, models.MasterProductStatusDraft, 1, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, int(total), 3)
		for _, p := range products {
			assert.Equal(t, models.MasterProductStatusDraft, p.Status)
		}
	})

	t.Run("SearchByTitle", func(t *testing.T) {
		createTestProduct(t, "Searchable Widget Product")
		createTestProduct(t, "Another Widget Item")
		createTestProduct(t, "Gadget Thing")

		tests := []struct {
			name    string
			search  string
			wantMin int
		}{
			{
				name:    "finds by partial title",
				search:  "Widget",
				wantMin: 2,
			},
			{
				name:    "case insensitive search",
				search:  "widget",
				wantMin: 2,
			},
			{
				name:    "returns empty for non-matching",
				search:  "ZZZZNONEXISTENT",
				wantMin: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				products, total, err := repo.SearchByTitle(ctx, tenantID, tt.search, 1, 10)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, int(total), tt.wantMin)
				if tt.wantMin > 0 {
					assert.NotEmpty(t, products)
				}
			})
		}
	})

	t.Run("FindLinked", func(t *testing.T) {
		linkedShopee := createTestProduct(t, "Linked Shopee Active Product")
		linkedShopeeSKU := createTestSku(t, linkedShopee, "LINK-SHOPEE-SKU")
		err := repo.CreatePlatformLink(ctx, &models.MasterProductPlatformLink{
			MasterProductID:   linkedShopee.ID,
			MasterSkuID:       &linkedShopeeSKU.ID,
			Platform:          "shopee",
			PlatformProductID: "SHOPEE-PROD-1",
			PlatformItemID:    "SHOPEE-ITEM-1",
			SyncStatus:        models.SyncStatusSynced,
		})
		require.NoError(t, err)

		linkedOutdated := createTestProduct(t, "Linked TikTok Outdated Product")
		linkedOutdatedSKU := createTestSku(t, linkedOutdated, "LINK-TIKTOK-SKU")
		err = repo.CreatePlatformLink(ctx, &models.MasterProductPlatformLink{
			MasterProductID:   linkedOutdated.ID,
			MasterSkuID:       &linkedOutdatedSKU.ID,
			Platform:          "tiktok",
			PlatformProductID: "TIKTOK-PROD-1",
			PlatformItemID:    "TIKTOK-ITEM-1",
			SyncStatus:        models.SyncStatusOutdated,
		})
		require.NoError(t, err)

		pendingOnly := createTestProduct(t, "Pending Link Product")
		pendingOnlySKU := createTestSku(t, pendingOnly, "PENDING-SKU")
		err = repo.CreatePlatformLink(ctx, &models.MasterProductPlatformLink{
			MasterProductID:   pendingOnly.ID,
			MasterSkuID:       &pendingOnlySKU.ID,
			Platform:          "shopee",
			PlatformProductID: "SHOPEE-PROD-PENDING",
			PlatformItemID:    "SHOPEE-ITEM-PENDING",
			SyncStatus:        models.SyncStatusPending,
		})
		require.NoError(t, err)

		draftLinked := &models.MasterProduct{
			TenantID: tenantID,
			Title:    "Draft Linked Product",
			Status:   models.MasterProductStatusDraft,
		}
		err = repo.Create(ctx, draftLinked)
		require.NoError(t, err)
		draftLinkedSKU := createTestSku(t, draftLinked, "DRAFT-LINKED-SKU")
		err = repo.CreatePlatformLink(ctx, &models.MasterProductPlatformLink{
			MasterProductID:   draftLinked.ID,
			MasterSkuID:       &draftLinkedSKU.ID,
			Platform:          "lazada",
			PlatformProductID: "LAZADA-PROD-1",
			PlatformItemID:    "LAZADA-ITEM-1",
			SyncStatus:        models.SyncStatusSynced,
		})
		require.NoError(t, err)

		activeLinked, total, err := repo.FindLinked(
			ctx,
			tenantID,
			1,
			100,
			models.MasterProductStatusActive,
			"",
			"",
		)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2))

		activeLinkedIDs := make([]uint, 0, len(activeLinked))
		for _, product := range activeLinked {
			assert.Equal(t, models.MasterProductStatusActive, product.Status)
			activeLinkedIDs = append(activeLinkedIDs, product.ID)
		}
		assert.Contains(t, activeLinkedIDs, linkedShopee.ID)
		assert.Contains(t, activeLinkedIDs, linkedOutdated.ID)
		assert.Contains(t, activeLinkedIDs, pendingOnly.ID) // pending IS linked after Task 7 fix
		assert.NotContains(t, activeLinkedIDs, draftLinked.ID)

		shopeeLinked, _, err := repo.FindLinked(
			ctx,
			tenantID,
			1,
			100,
			models.MasterProductStatusActive,
			"",
			"shopee",
		)
		assert.NoError(t, err)

		shopeeLinkedIDs := make([]uint, 0, len(shopeeLinked))
		for _, product := range shopeeLinked {
			shopeeLinkedIDs = append(shopeeLinkedIDs, product.ID)
		}
		assert.Contains(t, shopeeLinkedIDs, linkedShopee.ID)
		assert.NotContains(t, shopeeLinkedIDs, linkedOutdated.ID)
	})
}

// TestMasterProductRepository_FindLinked is the dedicated test matrix for FindLinked()
// covering all four sync statuses, unlinked exclusion, and platform scoping.
func TestMasterProductRepository_FindLinked(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t,
		&models.MasterProduct{},
		&models.MasterProductSku{},
		&models.MasterProductPlatformLink{},
	)
	repo := NewMasterProductRepository(db)
	ctx := context.Background()

	tenantID := "tenant-find-linked"

	// createProduct creates an active master product for the test tenant.
	createProduct := func(t *testing.T, title string) *models.MasterProduct {
		t.Helper()
		product := &models.MasterProduct{
			TenantID:    tenantID,
			Title:       title,
			Description: "Test Description",
			Status:      models.MasterProductStatusActive,
		}
		err := repo.Create(ctx, product)
		require.NoError(t, err)
		return product
	}

	// createSku creates a SKU for the given product.
	createSku := func(t *testing.T, product *models.MasterProduct, sellerSku string) *models.MasterProductSku {
		t.Helper()
		sku := &models.MasterProductSku{
			TenantID:        tenantID,
			MasterProductID: product.ID,
			SellerSku:       sellerSku,
			VariantName:     "Default",
			Price:           10000,
			Stock:           10,
		}
		err := repo.CreateSku(ctx, sku)
		require.NoError(t, err)
		return sku
	}

	// createLink creates a MasterProductPlatformLink with the given sync status.
	createLink := func(t *testing.T, productID uint, skuID *uint, platform, syncStatus string) {
		t.Helper()
		link := &models.MasterProductPlatformLink{
			MasterProductID: productID,
			MasterSkuID:     skuID,
			Platform:        platform,
			PlatformItemID:  platform + "-" + syncStatus + "-item",
			SyncStatus:      syncStatus,
		}
		err := repo.CreatePlatformLink(ctx, link)
		require.NoError(t, err)
	}

	// collectIDs extracts product IDs from a result slice for assertion.
	collectIDs := func(products []models.MasterProduct) []uint {
		ids := make([]uint, 0, len(products))
		for _, p := range products {
			ids = append(ids, p.ID)
		}
		return ids
	}

	t.Run("includes synced linked products", func(t *testing.T) {
		p := createProduct(t, "FL-Synced Product")
		sku := createSku(t, p, "FL-SYNCED-SKU-001")
		createLink(t, p.ID, &sku.ID, "shopee", models.SyncStatusSynced)

		results, _, err := repo.FindLinked(ctx, tenantID, 1, 100, "", "", "")
		require.NoError(t, err)
		assert.Contains(t, collectIDs(results), p.ID)
	})

	t.Run("includes outdated linked products", func(t *testing.T) {
		p := createProduct(t, "FL-Outdated Product")
		sku := createSku(t, p, "FL-OUTDATED-SKU-001")
		createLink(t, p.ID, &sku.ID, "shopee", models.SyncStatusOutdated)

		results, _, err := repo.FindLinked(ctx, tenantID, 1, 100, "", "", "")
		require.NoError(t, err)
		assert.Contains(t, collectIDs(results), p.ID)
	})

	t.Run("includes pending linked products", func(t *testing.T) {
		p := createProduct(t, "FL-Pending Product")
		sku := createSku(t, p, "FL-PENDING-SKU-001")
		createLink(t, p.ID, &sku.ID, "shopee", models.SyncStatusPending)

		results, _, err := repo.FindLinked(ctx, tenantID, 1, 100, "", "", "")
		require.NoError(t, err)
		assert.Contains(t, collectIDs(results), p.ID)
	})

	t.Run("includes error linked products", func(t *testing.T) {
		p := createProduct(t, "FL-Error Product")
		sku := createSku(t, p, "FL-ERROR-SKU-001")
		createLink(t, p.ID, &sku.ID, "shopee", models.SyncStatusError)

		results, _, err := repo.FindLinked(ctx, tenantID, 1, 100, "", "", "")
		require.NoError(t, err)
		assert.Contains(t, collectIDs(results), p.ID)
	})

	t.Run("excludes unlinked products", func(t *testing.T) {
		// Create a product with NO platform link — it must not appear in FindLinked results.
		p := createProduct(t, "FL-Unlinked Product")

		results, _, err := repo.FindLinked(ctx, tenantID, 1, 100, "", "", "")
		require.NoError(t, err)
		assert.NotContains(t, collectIDs(results), p.ID)
	})

	t.Run("platform scoping works", func(t *testing.T) {
		shopeeProduct := createProduct(t, "FL-Shopee Scoped Product")
		shopeeSku := createSku(t, shopeeProduct, "FL-SCOPE-SHOPEE-SKU")
		createLink(t, shopeeProduct.ID, &shopeeSku.ID, "shopee", models.SyncStatusSynced)

		tiktokProduct := createProduct(t, "FL-TikTok Scoped Product")
		tiktokSku := createSku(t, tiktokProduct, "FL-SCOPE-TIKTOK-SKU")
		createLink(t, tiktokProduct.ID, &tiktokSku.ID, "tiktok", models.SyncStatusSynced)

		results, _, err := repo.FindLinked(ctx, tenantID, 1, 100, "", "", "shopee")
		require.NoError(t, err)

		ids := collectIDs(results)
		assert.Contains(t, ids, shopeeProduct.ID)
		assert.NotContains(t, ids, tiktokProduct.ID)
	})
}
