package lazada

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/image"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// SyncService handles syncing data from Lazada API
type SyncService struct {
	client    *lazadaPkg.Client
	db        *gorm.DB
	orderRepo *repositories.LazadaOrderRepository
	prodRepo  *repositories.LazadaProductRepository
	imgMgr    image.Manager
	tenantID  string
}

// NewSyncService creates a new sync service
func NewSyncService(client *lazadaPkg.Client, db *gorm.DB) *SyncService {
	return &SyncService{
		client:    client,
		db:        db,
		orderRepo: repositories.NewLazadaOrderRepository(db),
		prodRepo:  repositories.NewLazadaProductRepository(db),
		imgMgr:    image.NewManager(db, ""),
	}
}

// NewSyncServiceWithTenant creates a new sync service with tenant ID
func NewSyncServiceWithTenant(client *lazadaPkg.Client, db *gorm.DB, tenantID string) *SyncService {
	return &SyncService{
		client:    client,
		db:        db,
		orderRepo: repositories.NewLazadaOrderRepository(db),
		prodRepo:  repositories.NewLazadaProductRepository(db),
		imgMgr:    image.NewManager(db, ""),
		tenantID:  tenantID,
	}
}

// SyncOrders fetches orders from Lazada API and saves to database
func (s *SyncService) SyncOrders(ctx context.Context, status string) (int, error) {
	resp, err := s.client.GetOrders(status, 0, 100)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, order := range resp.Data.Orders {
		// Parse ship by date from promised_shipping_times (ISO date string)
		var shipByDate *int64
		if order.PromisedShipDate != "" {
			// Try parsing various date formats
			for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
				if t, err := time.Parse(layout, order.PromisedShipDate); err == nil {
					ts := t.Unix()
					shipByDate = &ts
					break
				}
			}
		}

		dbOrder := &models.LazadaOrder{
			TenantID:        s.tenantID,
			OrderSN:         order.OrderID, // API returns OrderID, we store as OrderSN
			OrderStatus:     order.Status,
			BuyerUsername:   order.CustomerName,
			ShippingCarrier: order.ShippingType,
			ShipByDate:      shipByDate,
		}

		if err := s.orderRepo.Upsert(ctx, dbOrder); err == nil {
			count++
		}
	}

	return count, nil
}

// SyncProducts fetches products from Lazada API and saves to database
func (s *SyncService) SyncProducts(ctx context.Context) (int, error) {
	// Lazada API max limit per page is 50
	resp, err := s.client.GetProducts(0, 50)
	if err != nil {
		return 0, err
	}

	log.Printf("[Lazada Sync] SyncProducts: Got %d products from API", len(resp.Data.Products))

	count := 0
	for _, prod := range resp.Data.Products {
		itemID := prod.ItemID.String()

		// Extract product name from attributes first, fallback to name field
		// Lazada API returns name in attributes.name, NOT in top-level name field
		productName := prod.Name
		if prod.Attributes.Name != "" {
			productName = prod.Attributes.Name
		}

		// Extract description from attributes first, fallback to description field
		description := prod.Description
		if prod.Attributes.Description != "" {
			description = prod.Attributes.Description
		}

		// Extract brand from attributes first, fallback to brand field
		brand := prod.Brand
		if prod.Attributes.Brand != "" {
			brand = prod.Attributes.Brand
		}

		// Extract first image URL for display
		var imageURL string
		if len(prod.Images) > 0 {
			imageURL = prod.Images[0]
		}

		// Debug log first product
		if count == 0 {
			log.Printf("[Lazada Sync] First product: itemID=%s, name='%s', attrName='%s', finalName='%s', images=%d, imageURL='%s'",
				itemID, prod.Name, prod.Attributes.Name, productName, len(prod.Images), imageURL)
		}

		dbProd := &models.LazadaProduct{
			TenantID:    s.tenantID,
			ItemID:      itemID,
			Name:        productName,
			Description: description,
			Brand:       brand,
			Price:       prod.Price,
			Status:      prod.Status,
			Image:       imageURL,
		}

		if err := s.prodRepo.Upsert(ctx, dbProd); err == nil {
			count++

			// Download and save product images locally
			if len(prod.Images) > 0 {
				// Get saved product to get its ID
				savedProd, _ := s.prodRepo.FindByItemID(ctx, itemID)
				if savedProd != nil {
					localPaths := s.downloadAndSaveProductImages(ctx, itemID, prod.Images)
					if len(localPaths) > 0 {
						s.updateProductLocalImages(ctx, savedProd.ID, localPaths)
					}
				}
			}
		}
	}

	return count, nil
}

// SyncProductsWithDetails fetches ALL products from Lazada API with pagination, saves to DB, and returns them
// Used by GET /api/lazada/products to sync and return products
func (s *SyncService) SyncProductsWithDetails(ctx context.Context, offset, limit int) ([]map[string]interface{}, int, error) {
	log.Printf("[Lazada Sync] Starting product sync with pagination")

	var allProducts []map[string]interface{}
	savedProductCount := 0
	savedSkuCount := 0
	pageOffset := 0
	pageLimit := 50 // Lazada max per page is 50

	for {
		log.Printf("[Lazada Sync] Fetching page offset=%d, limit=%d", pageOffset, pageLimit)

		resp, err := s.client.GetProducts(pageOffset, pageLimit)
		if err != nil {
			log.Printf("[Lazada Sync] GetProducts error at offset %d: %v", pageOffset, err)
			return nil, 0, err
		}

		totalProducts := resp.Data.TotalProducts
		products := resp.Data.Products

		log.Printf("[Lazada Sync] Got %d products, total=%d", len(products), totalProducts)

		if len(products) == 0 {
			break
		}

		for _, prod := range products {
			// Convert FlexibleString to string for storage
			itemID := prod.ItemID.String()

			// Extract product name from attributes first, fallback to name field
			productName := prod.Name
			if prod.Attributes.Name != "" {
				productName = prod.Attributes.Name
			}

			// Extract description from attributes first, fallback to description field
			description := prod.Description
			if prod.Attributes.Description != "" {
				description = prod.Attributes.Description
			}

			// Extract brand
			brand := prod.Brand
			if prod.Attributes.Brand != "" {
				brand = prod.Attributes.Brand
			}

			// Extract first image URL for display
			var imageURL string
			if len(prod.Images) > 0 {
				imageURL = prod.Images[0]
			}

			// Save product to database
			dbProd := &models.LazadaProduct{
				TenantID:    s.tenantID,
				ItemID:      itemID,
				Name:        productName,
				Description: description,
				Brand:       brand,
				Price:       prod.Price,
				Status:      prod.Status,
				Image:       imageURL,
			}

			if err := s.prodRepo.Upsert(ctx, dbProd); err != nil {
				log.Printf("[Lazada Sync] Failed to upsert product %s: %v", itemID, err)
			} else {
				savedProductCount++

				// Download and save product images locally
				if len(prod.Images) > 0 {
					// Get saved product to get its ID
					savedProd, _ := s.prodRepo.FindByItemID(ctx, itemID)
					if savedProd != nil {
						localPaths := s.downloadAndSaveProductImages(ctx, itemID, prod.Images)
						if len(localPaths) > 0 {
							s.updateProductLocalImages(ctx, savedProd.ID, localPaths)
						}
					}
				}
			}

			// Save SKUs to database
			for _, sku := range prod.Skus {
				skuID := sku.SkuID.String()
				skuName := sku.SellerSku
				if skuName == "" {
					skuName = sku.ShopSku
				}

				// Extract variation name from saleProp or Variation field
				variantName := sku.Variation
				if variantName == "" && sku.Pilihan != "" {
					variantName = sku.Pilihan
				}

				dbSku := &models.LazadaSku{
					TenantID:     s.tenantID,
					ItemID:       itemID,
					SkuID:        skuID,
					ShopSku:      sku.ShopSku,
					SellerSku:    sku.SellerSku,
					Name:         skuName,
					VariantName:  variantName,
					Price:        sku.Price,
					SpecialPrice: sku.SpecialPrice,
					Quantity:     sku.Quantity,
					Available:    sku.Available,
				}

				if err := s.prodRepo.UpsertSku(ctx, dbSku); err != nil {
					log.Printf("[Lazada Sync] Failed to upsert SKU %s: %v", skuID, err)
				} else {
					savedSkuCount++
				}

				// Add each SKU as a row (matching frontend expectation)
				productItem := map[string]interface{}{
					"item_id":      itemID,
					"sku_id":       skuID,
					"sku_name":     skuName,
					"product_name": productName,
					"variant_name": variantName,
					"price":        sku.Price,
					"quantity":     sku.Quantity,
					"status":       prod.Status,
				}
				allProducts = append(allProducts, productItem)
			}

			// If no SKUs, still add product row
			if len(prod.Skus) == 0 {
				productItem := map[string]interface{}{
					"item_id":      itemID,
					"sku_id":       "",
					"sku_name":     productName,
					"product_name": productName,
					"variant_name": "",
					"price":        prod.Price,
					"quantity":     0,
					"status":       prod.Status,
				}
				allProducts = append(allProducts, productItem)
			}
		}

		// Check if we've fetched all products
		pageOffset += pageLimit
		if pageOffset >= totalProducts {
			break
		}
	}

	log.Printf("[Lazada Sync] Completed - products: %d, skus: %d, total rows: %d", savedProductCount, savedSkuCount, len(allProducts))

	return allProducts, savedProductCount, nil
}

// downloadAndSaveProductImages downloads product images using unified ImageManager
// Returns slice of local paths (thumb size for display). Errors are logged but don't break sync.
func (s *SyncService) downloadAndSaveProductImages(ctx context.Context, itemID string, imageURLs []string) []string {
	var localPaths []string

	zlog := zerolog.Ctx(ctx)
	if zlog == nil {
		zlog = &zerolog.Logger{}
	}

	for i, url := range imageURLs {
		if url == "" {
			continue
		}

		// Use unified ImageManager - handles deduplication and thumbnails
		img, err := s.imgMgr.CacheImage(ctx, s.tenantID, url)
		if err != nil {
			zlog.Warn().
				Str("service", "lazada_sync").
				Int("image_index", i).
				Str("item_id", itemID).
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
func (s *SyncService) updateProductLocalImages(ctx context.Context, productID uint, localPaths []string) {
	if len(localPaths) == 0 {
		return
	}

	zlog := zerolog.Ctx(ctx)
	if zlog == nil {
		zlog = &zerolog.Logger{}
	}

	// Convert to JSON
	pathsJSON, err := json.Marshal(localPaths)
	if err != nil {
		zlog.Warn().
			Str("service", "lazada_sync").
			Err(err).
			Msg("Failed to marshal local paths")
		return
	}

	// Update product
	err = s.db.Model(&models.LazadaProduct{}).
		Where("id = ?", productID).
		Update("local_images", pathsJSON).Error
	if err != nil {
		zlog.Warn().
			Str("service", "lazada_sync").
			Uint("product_id", productID).
			Err(err).
			Msg("Failed to update local_images")
	}
}
