package tiktok

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/image"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// SyncService handles syncing data from TikTok API
type SyncService struct {
	client    *tiktokPkg.Client
	db        *gorm.DB
	orderRepo *repositories.TiktokOrderRepository
	prodRepo  *repositories.TiktokProductRepository
	skuRepo   *repositories.TiktokSkuRepository
	imgMgr    image.Manager
	tenantID  string
}

// NewSyncService creates a new sync service
func NewSyncService(client *tiktokPkg.Client, db *gorm.DB) *SyncService {
	return &SyncService{
		client:    client,
		db:        db,
		orderRepo: repositories.NewTiktokOrderRepository(db),
		prodRepo:  repositories.NewTiktokProductRepository(db),
		skuRepo:   repositories.NewTiktokSkuRepository(db),
		imgMgr:    image.NewManager(db, ""),
	}
}

// NewSyncServiceWithTenant creates a new sync service with tenant ID
func NewSyncServiceWithTenant(client *tiktokPkg.Client, db *gorm.DB, tenantID string) *SyncService {
	return &SyncService{
		client:    client,
		db:        db,
		orderRepo: repositories.NewTiktokOrderRepository(db),
		prodRepo:  repositories.NewTiktokProductRepository(db),
		skuRepo:   repositories.NewTiktokSkuRepository(db),
		imgMgr:    image.NewManager(db, ""),
		tenantID:  tenantID,
	}
}

// SyncOrders fetches orders from TikTok API and saves to database
func (s *SyncService) SyncOrders(ctx context.Context, status string) (int, error) {
	resp, err := s.client.GetOrders(status, 100)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, order := range resp.Data.OrderList {
		// Get ship by date - prefer rts_sla_time, fallback to shipping_due_time
		var shipByDate *int64
		if order.RtsSlaTime > 0 {
			shipByDate = &order.RtsSlaTime
		} else if order.ShippingDueTime > 0 {
			shipByDate = &order.ShippingDueTime
		}

		dbOrder := &models.TiktokOrder{
			TenantID:        s.tenantID,
			OrderSN:         order.OrderID, // API returns OrderID, we store as OrderSN
			OrderStatus:     order.OrderStatus,
			ShippingCarrier: order.ShippingProvider,
			TrackingNumber:  order.TrackingNumber,
			BuyerUsername:   order.BuyerEmail,
			BuyerMessage:    order.BuyerMessage,
			ShipByDate:      shipByDate,
		}

		if err := s.orderRepo.Upsert(ctx, dbOrder); err == nil {
			count++
		}
	}

	return count, nil
}

// SyncProducts fetches products from TikTok API and saves to database
func (s *SyncService) SyncProducts(ctx context.Context) (int, error) {
	// Use v202502 API with POST request (more reliable than legacy GET API)
	resp, err := s.client.SearchProductsV202502("", 100, "")
	if err != nil {
		log.Error().
			Str("service", "tiktok_sync").
			Err(err).
			Msg("Failed to search products from TikTok API")
		return 0, err
	}

	log.Info().
		Str("service", "tiktok_sync").
		Int("total_products", len(resp.Data.Products)).
		Int("total_count", resp.Data.TotalCount).
		Msg("Fetched products from TikTok API")

	if err := s.clearTiktokProductCache(ctx); err != nil {
		return 0, err
	}

	count := 0
	for _, prod := range resp.Data.Products {
		// v202502 API returns product ID in "id" field
		dbProd := &models.TiktokProduct{
			TenantID:    s.tenantID,
			ProductID:   prod.ID,
			Name:        prod.Title,
			Description: prod.Description,
			Status:      prod.Status,
		}

		if err := s.prodRepo.Upsert(ctx, dbProd); err == nil {
			count++

			// Get saved product to get its ID
			savedProd, _ := s.prodRepo.FindByProductID(ctx, prod.ID)
			if savedProd != nil {
				// Fetch product detail to get images
				detailResp, err := s.client.GetProductDetail(prod.ID)
				if err != nil {
					log.Warn().
						Str("service", "tiktok_sync").
						Str("product_id", prod.ID).
						Err(err).
						Msg("Failed to get product detail for images")
					continue
				}

				// Extract image URLs from MainImages - prefer WebP if available
				var imageURLs []string
				for _, img := range detailResp.Data.MainImages {
					if len(img.URLs) > 0 {
						// Find WebP URL first (TikTok often provides multiple formats)
						webpURL := ""
						fallbackURL := ""
						for _, u := range img.URLs {
							if u == "" {
								continue
							}
							if strings.HasSuffix(strings.ToLower(u), ".webp") || strings.Contains(u, "webp") {
								webpURL = u
								break
							}
							if fallbackURL == "" {
								fallbackURL = u
							}
						}
						// Prefer WebP, fallback to first available
						if webpURL != "" {
							imageURLs = append(imageURLs, webpURL)
						} else if fallbackURL != "" {
							imageURLs = append(imageURLs, fallbackURL)
						}
					}
				}

				// Download and save product images locally
				if len(imageURLs) > 0 {
					localPaths := s.downloadAndSaveProductImages(ctx, prod.ID, imageURLs)
					if len(localPaths) > 0 {
						s.updateProductLocalImages(ctx, savedProd.ID, localPaths)
					}
				}
			}
		}
	}

	return count, nil
}

// downloadAndSaveProductImages downloads product images using unified ImageManager
// Returns slice of local paths (thumb size for display). Errors are logged but don't break sync.
func (s *SyncService) downloadAndSaveProductImages(ctx context.Context, productID string, imageURLs []string) []string {
	var localPaths []string

	for i, url := range imageURLs {
		if url == "" {
			continue
		}

		// Use unified ImageManager - handles deduplication and thumbnails
		img, err := s.imgMgr.CacheImage(ctx, s.tenantID, url)
		if err != nil {
			log.Warn().
				Str("service", "tiktok_sync").
				Int("image_index", i).
				Str("product_id", productID).
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

	// Convert to JSON
	pathsJSON, err := json.Marshal(localPaths)
	if err != nil {
		log.Warn().
			Str("service", "tiktok_sync").
			Err(err).
			Msg("Failed to marshal local paths")
		return
	}

	// Update product
	err = s.db.Model(&models.TiktokProduct{}).
		Where("id = ?", productID).
		Update("local_images", pathsJSON).Error
	if err != nil {
		log.Warn().
			Str("service", "tiktok_sync").
			Uint("product_id", productID).
			Err(err).
			Msg("Failed to update local_images")
	}
}
