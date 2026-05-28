package tiktok

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// maxPages is the maximum number of pages to fetch to prevent infinite loops.
// 500 pages * 100 items/page = 50,000 products max.
const maxPages = 500

// fetchAllProducts retrieves all ACTIVATE products from TikTok with retry and pagination.
// Status is hardcoded to "ACTIVATE"; change to "" (empty) for all statuses if deletion detection is needed.
func (s *SyncService) fetchAllProducts(ctx context.Context) ([]tiktokPkg.ProductSearchItem, int, error) {
	zlog := zerolog.Ctx(ctx)
	allProducts := make([]tiktokPkg.ProductSearchItem, 0)
	nextPageToken := ""
	totalCount := 0
	pageCount := 0

	for {
		if err := ctx.Err(); err != nil {
			return nil, 0, fmt.Errorf("context cancelled during product fetch: %w", err)
		}

		pageCount++
		if pageCount > maxPages {
			zlog.Warn().Int("max_pages", maxPages).Int("products_fetched", len(allProducts)).Msg("Reached maxPages limit, stopping pagination")
			break
		}

		resp, err := s.fetchProductPageWithRetry(ctx, nextPageToken)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to fetch product page %d: %w", pageCount, err)
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

	if len(allProducts) < totalCount {
		zlog.Warn().
			Int("total_count", totalCount).
			Int("fetched_count", len(allProducts)).
			Msg("TikTok product fetch incomplete: fetched fewer products than reported total")
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
		// Guard: skip products with empty ID to prevent orphan creation.
		if prod.ID == "" {
			zlog.Warn().Msg("Skipping TikTok product with empty ID")
			continue
		}

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

		detailResp, err := s.client.GetProductDetail(ctx, prod.ID)
		if err != nil {
			zlog.Error().Str("product_id", prod.ID).Err(err).Msg("Failed to get product detail, syncing product without variant/image enrichment")
			detailResp = nil
		} else if detailResp.Code != 0 {
			zlog.Warn().Str("product_id", prod.ID).Int("code", detailResp.Code).Str("message", detailResp.Message).Msg("TikTok API error on product detail, syncing product without variant/image enrichment")
			detailResp = nil
		}

		s.syncProductSkus(ctx, prodRepo, savedProd.ID, prod, detailResp)
		s.updateProductSummary(ctx, tx, savedProd.ID, prod)
		s.syncProductImages(ctx, tx, savedProd.ID, prod.ID, detailResp)

		*count = *count + 1
	}

	return nil
}

func (s *SyncService) syncProductSkus(
	ctx context.Context,
	prodRepo *repositories.TiktokProductRepository,
	productID string,
	prod tiktokPkg.ProductSearchItem,
	detailResp *tiktokPkg.ProductDetailResponse,
) {
	zlog := zerolog.Ctx(ctx)
	variantBySkuID := make(map[string][]tiktokPkg.ProductSalesAttr)

	if detailResp != nil {
		for _, detailSku := range detailResp.Data.Skus {
			variantBySkuID[detailSku.ID] = detailSku.SalesAttributes
		}
	}

	for _, sku := range prod.Skus {
		parsedPrice := parsePrice(sku.Price.SalePrice, sku.Price.OriginalPrice, sku.Price.TaxExclusivePrice)
		parsedQty := sumInventory(sku.Inventory)

		// Debug: log raw TikTok price values to diagnose price=0 bug
		zlog.Debug().
			Str("sku_id", sku.ID).
			Str("seller_sku", sku.SellerSku).
			Str("raw_sale_price", sku.Price.SalePrice).
			Str("raw_original_price", sku.Price.OriginalPrice).
			Float64("parsed_price", parsedPrice).
			Int("parsed_qty", parsedQty).
			Msg("TikTok SKU price debug")

		dbSku := &models.TiktokSku{
			TenantID:  s.tenantID,
			ProductID: productID,
			SkuID:     sku.ID,
			SellerSku: sku.SellerSku,
			Price:     parsedPrice,
			Quantity:  parsedQty,
		}

		if attrs, ok := variantBySkuID[sku.ID]; ok {
			dbSku.VariantName = buildVariantName(attrs)
			dbSku.VariantData = buildVariantData(attrs)
		}

		if err := prodRepo.UpsertSku(ctx, dbSku); err != nil {
			zlog.Warn().Str("sku_id", sku.ID).Err(err).Msg("Failed to upsert TikTok SKU")
		}
	}
}

func (s *SyncService) updateProductSummary(ctx context.Context, tx *gorm.DB, productID string, prod tiktokPkg.ProductSearchItem) {
	// If SearchProducts response has SKU data, use it directly.
	if len(prod.Skus) > 0 {
		totalStock := 0
		for _, sku := range prod.Skus {
			totalStock += sumInventory(sku.Inventory)
		}

		firstSku := prod.Skus[0]
		firstPrice := parsePrice(firstSku.Price.SalePrice, firstSku.Price.OriginalPrice, firstSku.Price.TaxExclusivePrice)

		err := tx.WithContext(ctx).
			Model(&models.TiktokProduct{}).
			Where("id = ?", productID).
			Updates(map[string]interface{}{
				"price":    firstPrice,
				"quantity": totalStock,
			}).Error
		if err != nil {
			zerolog.Ctx(ctx).Warn().Str("product_id", productID).Err(err).Msg("Failed to update TikTok product summary")
		}
		return
	}

	// Fallback: SearchProducts didn't include SKU data, but GetProductDetail
	// may have synced SKUs to DB. Read from tiktok_skus to update summary.
	var dbSkus []models.TiktokSku
	if err := tx.WithContext(ctx).
		Where("product_id = ?", productID).
		Find(&dbSkus).Error; err != nil || len(dbSkus) == 0 {
		return
	}

	totalStock := 0
	var firstPrice float64
	for i, sku := range dbSkus {
		totalStock += sku.Quantity
		if i == 0 {
			firstPrice = sku.Price
		}
	}

	err := tx.WithContext(ctx).
		Model(&models.TiktokProduct{}).
		Where("id = ?", productID).
		Updates(map[string]interface{}{
			"price":    firstPrice,
			"quantity": totalStock,
		}).Error
	if err != nil {
		zerolog.Ctx(ctx).Warn().Str("product_id", productID).Err(err).Msg("Failed to update TikTok product summary from DB")
	}
}

func (s *SyncService) syncProductImages(
	ctx context.Context,
	tx *gorm.DB,
	productID string,
	remoteProductID string,
	detailResp *tiktokPkg.ProductDetailResponse,
) {
	zlog := zerolog.Ctx(ctx)

	if detailResp == nil {
		return
	}

	imageURLs := extractPreferredImageURLs(detailResp.Data.MainImages)
	if len(imageURLs) == 0 {
		return
	}

	err := tx.WithContext(ctx).
		Model(&models.TiktokProduct{}).
		Where("id = ?", productID).
		Update("image", imageURLs[0]).Error
	if err != nil {
		zlog.Warn().Str("product_id", productID).Err(err).Msg("Failed to update TikTok primary image")
	}

	localPaths := s.downloadAndSaveProductImages(ctx, remoteProductID, imageURLs)
	if len(localPaths) > 0 {
		s.updateProductLocalImages(ctx, tx, productID, localPaths)
	}
}

// downloadAndSaveProductImages downloads product images using unified ImageManager.
// Returns slice of local paths (medium size for display). Errors are logged but do not break sync.
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
		localPaths = append(localPaths, paths.Medium)
	}

	return localPaths
}

func (s *SyncService) updateProductLocalImages(ctx context.Context, tx *gorm.DB, productID string, localPaths []string) {
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
		zerolog.Ctx(ctx).Warn().Str("product_id", productID).Err(err).Msg("Failed to update local_images")
	}
}

