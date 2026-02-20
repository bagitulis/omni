package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func (s *SyncService) fetchAllProducts(ctx context.Context) ([]tiktokPkg.ProductSearchItem, int, error) {
	allProducts := make([]tiktokPkg.ProductSearchItem, 0)
	nextPageToken := ""
	totalCount := 0

	for {
		resp, err := s.client.SearchProductsV202502("", 100, nextPageToken)
		if err != nil {
			return nil, 0, err
		}

		allProducts = append(allProducts, resp.Data.Products...)
		if resp.Data.TotalCount > totalCount {
			totalCount = resp.Data.TotalCount
		}

		if resp.Data.NextPageToken == "" || resp.Data.NextPageToken == nextPageToken {
			break
		}

		nextPageToken = resp.Data.NextPageToken
	}

	return allProducts, totalCount, nil
}

func (s *SyncService) syncProductsWithDB(
	ctx context.Context,
	tx *gorm.DB,
	prodRepo *repositories.TiktokProductRepository,
	products []tiktokPkg.ProductSearchItem,
	count *int,
) error {
	zlog := zerolog.Ctx(ctx)

	for _, prod := range products {
		dbProd := &models.TiktokProduct{
			TenantID:    s.tenantID,
			ProductID:   prod.ID,
			Name:        prod.Title,
			Description: prod.Description,
			Status:      prod.Status,
		}

		if err := prodRepo.Upsert(ctx, dbProd); err != nil {
			zlog.Warn().Str("product_id", prod.ID).Err(err).Msg("Failed to upsert TikTok product")
			continue
		}

		savedProd, err := prodRepo.FindByProductID(ctx, prod.ID)
		if err != nil {
			zlog.Warn().Str("product_id", prod.ID).Err(err).Msg("Failed to load saved TikTok product")
			continue
		}

		if savedProd == nil {
			zlog.Warn().Str("product_id", prod.ID).Msg("Saved TikTok product not found after upsert")
			continue
		}

		s.syncProductSkus(ctx, prodRepo, savedProd.ID, prod)
		s.updateProductSummary(ctx, tx, savedProd.ID, prod)
		s.syncProductImages(ctx, tx, savedProd.ID, prod.ID)

		*count = *count + 1
	}

	return nil
}

func (s *SyncService) syncProductSkus(
	ctx context.Context,
	prodRepo *repositories.TiktokProductRepository,
	productID uint,
	prod tiktokPkg.ProductSearchItem,
) {
	zlog := zerolog.Ctx(ctx)

	for _, sku := range prod.Skus {
		dbSku := &models.TiktokSku{
			TenantID:  s.tenantID,
			ProductID: productID,
			SkuID:     sku.ID,
			SellerSku: sku.SellerSku,
			Price:     parsePrice(sku.Price.SalePrice, sku.Price.OriginalPrice),
			Quantity:  sumInventory(sku.Inventory),
		}

		if err := prodRepo.UpsertSku(ctx, dbSku); err != nil {
			zlog.Warn().Str("sku_id", sku.ID).Err(err).Msg("Failed to upsert TikTok SKU")
		}
	}
}

func (s *SyncService) updateProductSummary(ctx context.Context, tx *gorm.DB, productID uint, prod tiktokPkg.ProductSearchItem) {
	if len(prod.Skus) == 0 {
		return
	}

	totalStock := 0
	for _, sku := range prod.Skus {
		totalStock += sumInventory(sku.Inventory)
	}

	firstSku := prod.Skus[0]
	firstPrice := parsePrice(firstSku.Price.SalePrice, firstSku.Price.OriginalPrice)

	err := tx.WithContext(ctx).
		Model(&models.TiktokProduct{}).
		Where("id = ?", productID).
		Updates(map[string]interface{}{
			"price":    firstPrice,
			"quantity": totalStock,
		}).Error
	if err != nil {
		zerolog.Ctx(ctx).Warn().Uint("product_id", productID).Err(err).Msg("Failed to update TikTok product summary")
	}
}

func (s *SyncService) syncProductImages(ctx context.Context, tx *gorm.DB, productID uint, remoteProductID string) {
	zlog := zerolog.Ctx(ctx)

	detailResp, err := s.client.GetProductDetail(remoteProductID)
	if err != nil {
		zlog.Warn().Str("product_id", remoteProductID).Err(err).Msg("Failed to get product detail for images, product saved without images")
		return
	}

	imageURLs := extractPreferredImageURLs(detailResp.Data.MainImages)
	if len(imageURLs) == 0 {
		return
	}

	err = tx.WithContext(ctx).
		Model(&models.TiktokProduct{}).
		Where("id = ?", productID).
		Update("image", imageURLs[0]).Error
	if err != nil {
		zlog.Warn().Uint("product_id", productID).Err(err).Msg("Failed to update TikTok primary image")
	}

	localPaths := s.downloadAndSaveProductImages(ctx, remoteProductID, imageURLs)
	if len(localPaths) > 0 {
		s.updateProductLocalImages(ctx, tx, productID, localPaths)
	}
}

// downloadAndSaveProductImages downloads product images using unified ImageManager.
// Returns slice of local paths (thumb size for display). Errors are logged but do not break sync.
func (s *SyncService) downloadAndSaveProductImages(ctx context.Context, productID string, imageURLs []string) []string {
	zlog := zerolog.Ctx(ctx)
	localPaths := make([]string, 0, len(imageURLs))

	for i, url := range imageURLs {
		if url == "" {
			continue
		}

		img, err := s.imgMgr.CacheImage(ctx, s.tenantID, url)
		if err != nil {
			zlog.Warn().Int("image_index", i).Str("product_id", productID).Err(err).Msg("Failed to cache image")
			continue
		}

		paths := s.imgMgr.GetPaths(img)
		localPaths = append(localPaths, paths.Thumb)
	}

	return localPaths
}

func (s *SyncService) updateProductLocalImages(ctx context.Context, tx *gorm.DB, productID uint, localPaths []string) {
	if len(localPaths) == 0 {
		return
	}

	pathsJSON, err := json.Marshal(localPaths)
	if err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Msg("Failed to marshal local paths")
		return
	}

	err = tx.WithContext(ctx).
		Model(&models.TiktokProduct{}).
		Where("id = ?", productID).
		Update("local_images", pathsJSON).Error
	if err != nil {
		zerolog.Ctx(ctx).Warn().Uint("product_id", productID).Err(err).Msg("Failed to update local_images")
	}
}

func parsePrice(salePrice string, originalPrice string) float64 {
	price := 0.0
	if salePrice != "" {
		_, _ = fmt.Sscanf(salePrice, "%f", &price)
		return price
	}

	if originalPrice != "" {
		_, _ = fmt.Sscanf(originalPrice, "%f", &price)
	}

	return price
}

func sumInventory(inventory []struct {
	WarehouseID string `json:"warehouse_id"`
	Quantity    int    `json:"quantity"`
}) int {
	totalQty := 0
	for _, inv := range inventory {
		totalQty += inv.Quantity
	}

	return totalQty
}

func extractPreferredImageURLs(images []tiktokPkg.ProductImage) []string {
	imageURLs := make([]string, 0, len(images))

	for _, img := range images {
		if len(img.URLs) == 0 {
			continue
		}

		webpURL := ""
		fallbackURL := ""
		for _, u := range img.URLs {
			if u == "" {
				continue
			}

			if strings.HasSuffix(strings.ToLower(u), ".webp") || strings.Contains(strings.ToLower(u), "webp") {
				webpURL = u
				break
			}

			if fallbackURL == "" {
				fallbackURL = u
			}
		}

		if webpURL != "" {
			imageURLs = append(imageURLs, webpURL)
			continue
		}

		if fallbackURL != "" {
			imageURLs = append(imageURLs, fallbackURL)
		}
	}

	return imageURLs
}
