package shopee

import (
	"context"
	"encoding/json"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/image"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ProductSyncService handles syncing products from Shopee API to database
type ProductSyncService struct {
	client   *shopeePkg.Client
	db       *gorm.DB
	prodRepo *repositories.ShopeeProductRepository
	skuRepo  *repositories.ShopeeSkuRepository
	imgMgr   image.Manager
	tenantID string
}

// NewProductSyncService creates a new product sync service with tenant ID
func NewProductSyncService(client *shopeePkg.Client, db *gorm.DB, tenantID string) *ProductSyncService {
	return &ProductSyncService{
		client:   client,
		db:       db,
		prodRepo: repositories.NewShopeeProductRepository(db),
		skuRepo:  repositories.NewShopeeSkuRepository(db),
		imgMgr:   image.NewManager(db, ""),
		tenantID: tenantID,
	}
}

// SyncProducts fetches products from Shopee API and saves to database
// Implements pagination using has_next_page and next_offset from API
func (s *ProductSyncService) SyncProducts(ctx context.Context) (int, error) {
	allItemIDs := []int64{}
	offset := 0
	pageSize := 100 // Max allowed by Shopee API

	// Loop through all pages until has_next_page is false
	for {
		listResp, err := s.client.GetProductList(offset, pageSize)
		if err != nil {
			return 0, err
		}

		if len(listResp.Response.Item) == 0 {
			break
		}

		// Collect item IDs from this page
		for _, item := range listResp.Response.Item {
			allItemIDs = append(allItemIDs, item.ItemID)
		}

		// Check if there's more pages using has_next_page
		if !listResp.Response.HasNextPage {
			break
		}

		// Use next_offset from API response for next page
		offset = listResp.Response.NextOffset
	}

	if len(allItemIDs) == 0 {
		if err := s.clearShopeeProductCache(ctx); err != nil {
			return 0, err
		}
		return 0, nil
	}

	if err := s.clearShopeeProductCache(ctx); err != nil {
		return 0, err
	}

	// Get product details in batches of 50 (API limit)
	count := 0
	batchSize := 50
	for i := 0; i < len(allItemIDs); i += batchSize {
		end := i + batchSize
		if end > len(allItemIDs) {
			end = len(allItemIDs)
		}
		batchIDs := allItemIDs[i:end]

		detailResp, err := s.client.GetProductDetail(batchIDs)
		if err != nil {
			continue // Skip failed batch, continue with next
		}

		// Save products and their SKUs to database
		for _, prod := range detailResp.Response.ItemList {
			dbProd := &models.ShopeeProduct{
				TenantID:    s.tenantID,
				ItemID:      prod.ItemID,
				Name:        prod.ItemName,
				Description: prod.Description,
				Status:      prod.ItemStatus,
				Price:       prod.CurrentPrice,
				Quantity:    prod.Stock,
			}

			if err := s.prodRepo.Upsert(ctx, dbProd); err == nil {
				count++

				// Get saved product to get its ID
				savedProd, _ := s.prodRepo.FindByItemID(ctx, prod.ItemID)
				if savedProd != nil {
					// Sync SKUs for this product
					s.syncProductSKUs(ctx, savedProd, prod.ItemID)

					// Download and save product images locally
					if len(prod.Images) > 0 {
						localPaths := s.downloadAndSaveProductImages(ctx, prod.ItemID, prod.Images)
						if len(localPaths) > 0 {
							s.updateProductLocalImages(ctx, savedProd.ID, localPaths)
						}
					}
				}
			}
		}
	}

	return count, nil
}

// syncProductSKUs fetches and saves SKUs (models) for a product
func (s *ProductSyncService) syncProductSKUs(ctx context.Context, product *models.ShopeeProduct, itemID int64) {
	modelResp, err := s.client.GetModelList(itemID)
	if err != nil {
		return
	}

	tierVariations := modelResp.Response.TierVariation
	modelList := modelResp.Response.Model

	if len(modelList) == 0 {
		// Non-variant product - create single SKU
		sku := &models.ShopeeSku{
			TenantID:    s.tenantID,
			ProductID:   product.ID,
			ItemID:      itemID,
			SellerSku:   "",
			Price:       product.Price,
			Quantity:    product.Quantity,
			VariantName: "",
		}
		s.skuRepo.Upsert(ctx, sku)
		return
	}

	// Product has variants - sync each model as SKU
	for _, m := range modelList {
		// Build variant name from tier_index
		variantName := buildVariantName(m.TierIndex, tierVariations)

		// Get price and stock
		price := float64(0)
		if len(m.PriceInfo) > 0 {
			price = m.PriceInfo[0].CurrentPrice
			if price == 0 {
				price = m.PriceInfo[0].OriginalPrice
			}
		}
		quantity := m.StockInfoV2.SummaryInfo.TotalAvailableStock

		modelID := m.ModelID
		sku := &models.ShopeeSku{
			TenantID:    s.tenantID,
			ProductID:   product.ID,
			ItemID:      itemID,
			ModelID:     &modelID,
			SellerSku:   m.ModelSKU,
			Price:       price,
			Quantity:    quantity,
			VariantName: variantName,
		}
		s.skuRepo.Upsert(ctx, sku)
	}
}

// downloadAndSaveProductImages downloads product images using unified ImageManager
// Returns slice of local paths (thumb size for display). Errors are logged but don't break sync.
func (s *ProductSyncService) downloadAndSaveProductImages(ctx context.Context, itemID int64, imageURLs []string) []string {
	var localPaths []string

	for i, url := range imageURLs {
		if url == "" {
			continue
		}

		// Use unified ImageManager - handles deduplication and thumbnails
		img, err := s.imgMgr.CacheImage(ctx, s.tenantID, url)
		if err != nil {
			log.Warn().
				Str("service", "shopee_sync").
				Int("image_index", i).
				Int64("item_id", itemID).
				Err(err).
				Msg("Failed to cache image")
			continue
		}

		// Get paths and use thumb for display in tables
		paths := s.imgMgr.GetPaths(img)
		localPaths = append(localPaths, paths.Thumb)
	}

	return localPaths
}

// updateProductLocalImages updates the product with local image paths
func (s *ProductSyncService) updateProductLocalImages(ctx context.Context, productID uint, localPaths []string) {
	if len(localPaths) == 0 {
		return
	}

	// Convert to JSON
	pathsJSON, err := json.Marshal(localPaths)
	if err != nil {
		log.Warn().
			Str("service", "shopee_sync").
			Err(err).
			Msg("Failed to marshal local paths")
		return
	}

	// Update product
	err = s.db.Model(&models.ShopeeProduct{}).
		Where("id = ?", productID).
		Update("local_images", pathsJSON).Error
	if err != nil {
		log.Warn().
			Str("service", "shopee_sync").
			Uint("product_id", productID).
			Err(err).
			Msg("Failed to update local_images")
	}
}

// SyncProductsWithDetails fetches products from Shopee API, saves to DB, and returns them
// Used by GET /api/shopee/products to sync and return products
func (s *ProductSyncService) SyncProductsWithDetails(ctx context.Context, itemStatus string, offset, limit int) ([]map[string]interface{}, int, error) {
	allItemIDs := []int64{}
	pageOffset := 0
	pageSize := 100

	log.Debug().
		Str("service", "shopee_sync").
		Str("item_status", itemStatus).
		Msg("Starting product sync")

	// Loop through all pages to collect item IDs
	for {
		listResp, err := s.client.GetProductList(pageOffset, pageSize)
		if err != nil {
			log.Error().
				Str("service", "shopee_sync").
				Err(err).
				Msg("GetProductList error")
			return nil, 0, err
		}

		log.Debug().
			Str("service", "shopee_sync").
			Int("items_count", len(listResp.Response.Item)).
			Bool("has_next_page", listResp.Response.HasNextPage).
			Msg("GetProductList returned")

		if len(listResp.Response.Item) == 0 {
			break
		}

		// Collect item IDs (filtering happens at detail level)
		for _, item := range listResp.Response.Item {
			allItemIDs = append(allItemIDs, item.ItemID)
		}

		if !listResp.Response.HasNextPage {
			break
		}

		pageOffset = listResp.Response.NextOffset
	}

	log.Debug().
		Str("service", "shopee_sync").
		Int("total_items", len(allItemIDs)).
		Msg("Total item IDs collected")

	if len(allItemIDs) == 0 {
		log.Debug().
			Str("service", "shopee_sync").
			Msg("No products found from Shopee API")
		if err := s.clearShopeeProductCache(ctx); err != nil {
			return nil, 0, err
		}
		return []map[string]interface{}{}, 0, nil
	}

	if err := s.clearShopeeProductCache(ctx); err != nil {
		return nil, 0, err
	}

	// Get product details in batches
	var products []map[string]interface{}
	savedCount := 0
	batchSize := 50

	for i := 0; i < len(allItemIDs); i += batchSize {
		end := i + batchSize
		if end > len(allItemIDs) {
			end = len(allItemIDs)
		}
		batchIDs := allItemIDs[i:end]

		detailResp, err := s.client.GetProductDetail(batchIDs)
		if err != nil {
			log.Warn().
				Str("service", "shopee_sync").
				Err(err).
				Msg("GetProductDetail batch error")
			continue
		}

		log.Debug().
			Str("service", "shopee_sync").
			Int("items_count", len(detailResp.Response.ItemList)).
			Msg("GetProductDetail returned for batch")

		for _, prod := range detailResp.Response.ItemList {
			// Filter by status if provided
			if itemStatus != "" && prod.ItemStatus != itemStatus {
				continue
			}

			// Save to database with tenant_id
			dbProd := &models.ShopeeProduct{
				TenantID:    s.tenantID,
				ItemID:      prod.ItemID,
				Name:        prod.ItemName,
				Description: prod.Description,
				Status:      prod.ItemStatus,
				Price:       prod.CurrentPrice,
				Quantity:    prod.Stock,
			}

			if err := s.prodRepo.Upsert(ctx, dbProd); err != nil {
				log.Warn().
					Str("service", "shopee_sync").
					Int64("item_id", prod.ItemID).
					Err(err).
					Msg("Failed to upsert product")
			} else {
				savedCount++

				// Sync SKUs
				savedProd, _ := s.prodRepo.FindByItemID(ctx, prod.ItemID)
				if savedProd != nil {
					s.syncProductSKUs(ctx, savedProd, prod.ItemID)

					// Download and save product images locally
					if len(prod.Images) > 0 {
						localPaths := s.downloadAndSaveProductImages(ctx, prod.ItemID, prod.Images)
						if len(localPaths) > 0 {
							s.updateProductLocalImages(ctx, savedProd.ID, localPaths)
						}
					}
				}
			}

			// Build response item matching Node.js format
			productItem := map[string]interface{}{
				"item_id":        prod.ItemID,
				"item_name":      prod.ItemName,
				"item_status":    prod.ItemStatus,
				"description":    prod.Description,
				"category_id":    prod.CategoryID,
				"original_price": prod.OriginalPrice,
				"current_price":  prod.CurrentPrice,
				"stock":          prod.Stock,
			}

			// Add image if available
			if len(prod.Images) > 0 {
				productItem["image"] = prod.Images[0]
			}

			products = append(products, productItem)
		}
	}

	log.Info().
		Str("service", "shopee_sync").
		Int("total_products", len(products)).
		Int("saved_count", savedCount).
		Msg("Product sync completed")

	return products, savedCount, nil
}
