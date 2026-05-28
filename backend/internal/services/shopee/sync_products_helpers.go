package shopee

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func (s *ProductSyncService) syncProductsByItemIDs(ctx context.Context, allItemIDs []int64) (int, error) {
	zlog := zerolog.Ctx(ctx)
	if len(allItemIDs) == 0 {
		zlog.Warn().Msg("Shopee API returned 0 products - skipping sync to prevent data loss")
		return 0, nil
	}

	count := 0
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.clearShopeeProductCacheWithDB(ctx, tx); err != nil {
			return err
		}

		txProdRepo := repositories.NewShopeeProductRepository(tx)
		txSkuRepo := repositories.NewShopeeSkuRepository(tx)
		return s.syncItemDetailBatches(ctx, tx, txProdRepo, txSkuRepo, allItemIDs, &count)
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *ProductSyncService) syncItemDetailBatches(
	ctx context.Context,
	tx *gorm.DB,
	prodRepo *repositories.ShopeeProductRepository,
	skuRepo *repositories.ShopeeSkuRepository,
	allItemIDs []int64,
	count *int,
) error {
	zlog := zerolog.Ctx(ctx)
	batchSize := 50 // API limit.

	log.Info().Ints64("item_ids", allItemIDs).Int("total_items", len(allItemIDs)).Msg("[Shopee SyncProductsByIDs] Starting sync")

	var failedItemIDs []int64

	for i := 0; i < len(allItemIDs); i += batchSize {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled during batch sync: %w", err)
		}

		end := i + batchSize
		if end > len(allItemIDs) {
			end = len(allItemIDs)
		}
		batchIDs := allItemIDs[i:end]

		log.Info().Ints64("batch_ids", batchIDs).Int("batch_size", len(batchIDs)).Msg("[Shopee SyncProductsByIDs] Fetching product details")

		detailResp, err := s.client.GetProductDetailWithImages(ctx, batchIDs)
		if err != nil {
			log.Error().Err(err).Ints64("batch_ids", batchIDs).Msg("[Shopee SyncProductsByIDs] API call FAILED")
			zlog.Warn().Err(err).Int("batch_size", len(batchIDs)).Int("batch_start", i).Msg("Shopee product detail batch failed, continuing")
			failedItemIDs = append(failedItemIDs, batchIDs...)
			continue
		}

		log.Info().Int("item_count", len(detailResp.Response.ItemList)).Msg("[Shopee SyncProductsByIDs] API response received")
		if len(detailResp.Response.ItemList) == 0 {
			log.Warn().Str("api_error", detailResp.Error).Str("api_message", detailResp.Message).Msg("[Shopee SyncProductsByIDs] Empty item_list in response")
		}

		for _, prod := range detailResp.Response.ItemList {
			log.Info().Int64("item_id", prod.ItemID).Str("name", prod.ItemName).Int("stock", prod.StockInfoV2.SummaryInfo.TotalAvailableStock).Msg("[Shopee SyncProductsByIDs] Processing product")
			if err := s.upsertProductWithDependencies(ctx, tx, prodRepo, skuRepo, prod); err != nil {
				log.Error().Err(err).Int64("item_id", prod.ItemID).Msg("[Shopee SyncProductsByIDs] Upsert FAILED")
				zlog.Warn().Err(err).Int64("item_id", prod.ItemID).Msg("Failed to persist Shopee product")
				continue
			}
			log.Info().Int64("item_id", prod.ItemID).Msg("[Shopee SyncProductsByIDs] ✅ Persisted")
			*count = *count + 1
		}
	}

	if len(failedItemIDs) > 0 {
		zlog.Warn().
			Int("failed_items", len(failedItemIDs)).
			Int("total_items", len(allItemIDs)).
			Msg("Shopee product sync completed with batch failures")
	}

	return nil
}

func (s *ProductSyncService) upsertProductWithDependencies(
	ctx context.Context,
	tx *gorm.DB,
	prodRepo *repositories.ShopeeProductRepository,
	skuRepo *repositories.ShopeeSkuRepository,
	prod shopeePkg.ProductDetailWithImages,
) error {
	price := float64(0)
	if len(prod.PriceInfo) > 0 {
		price = prod.PriceInfo[0].CurrentPrice
		if price == 0 {
			price = prod.PriceInfo[0].OriginalPrice
		}
	}

	stock := prod.StockInfoV2.SummaryInfo.TotalAvailableStock
	imageURLs := prod.Image.ImageURLList
	primaryImage := ""
	if len(imageURLs) > 0 {
		primaryImage = imageURLs[0]
	}

	dbProd := &models.ShopeeProduct{
		TenantID:    s.tenantID,
		ItemID:      prod.ItemID,
		Name:        prod.ItemName,
		Description: prod.Description,
		Image:       primaryImage,
		Price:       price,
		Quantity:    stock,
	}

	if err := prodRepo.Upsert(ctx, dbProd); err != nil {
		return err
	}

	savedProd, err := prodRepo.FindByItemID(ctx, prod.ItemID)
	if err != nil {
		return err
	}

	if err := s.syncProductSKUs(ctx, skuRepo, savedProd, prod.ItemID, prod.ItemSKU); err != nil {
		return err
	}

	if len(imageURLs) == 0 {
		return nil
	}

	localPaths := s.downloadAndSaveProductImages(ctx, prod.ItemID, imageURLs)
	if len(localPaths) == 0 {
		return nil
	}

	return s.updateProductLocalImages(ctx, tx, savedProd.ID, localPaths)
}

// syncProductSKUs fetches and saves SKUs (models) for a product.
// itemSKU is the parent-level SKU from get_item_base_info, used when no models exist.
func (s *ProductSyncService) syncProductSKUs(
	ctx context.Context,
	skuRepo *repositories.ShopeeSkuRepository,
	product *models.ShopeeProduct,
	itemID int64,
	itemSKU string,
) error {
	modelResp, err := s.client.GetModelList(ctx, itemID)
	if err != nil {
		return err
	}

	tierVariations := modelResp.Response.TierVariation
	modelList := modelResp.Response.Model

	if len(modelList) == 0 {
		sku := &models.ShopeeSku{
			TenantID:    s.tenantID,
			ProductID:   product.ID,
			ItemID:      itemID,
			SellerSku:   itemSKU,
			Price:       product.Price,
			Quantity:    product.Quantity,
			VariantName: "",
		}
		return skuRepo.Upsert(ctx, sku)
	}

	for _, m := range modelList {
		variantName := buildVariantName(m.TierIndex, tierVariations)

		price := float64(0)
		if len(m.PriceInfo) > 0 {
			price = m.PriceInfo[0].CurrentPrice
			if price == 0 {
				price = m.PriceInfo[0].OriginalPrice
			}
		}

		quantity := m.StockInfoV2.SummaryInfo.TotalAvailableStock
		modelID := m.ModelID

		log.Info().
			Int64("item_id", itemID).
			Int64("model_id", modelID).
			Str("model_sku", m.ModelSKU).
			Str("variant_name", variantName).
			Int("stock", quantity).
			Float64("price", price).
			Msg("[Shopee SyncProductsByIDs] Model/SKU stock")

		sellerSku := m.ModelSKU
		if sellerSku == "" {
			sellerSku = itemSKU // Fallback: use parent item_sku
		}

		sku := &models.ShopeeSku{
			TenantID:    s.tenantID,
			ProductID:   product.ID,
			ItemID:      itemID,
			ModelID:     &modelID,
			SellerSku:   sellerSku,
			Price:       price,
			Quantity:    quantity,
			VariantName: variantName,
		}

		if err := skuRepo.Upsert(ctx, sku); err != nil {
			return err
		}
	}

	return nil
}

// downloadAndSaveProductImages downloads product images using unified ImageManager.
// Returns local paths (medium size). Errors are logged but do not fail sync.
func (s *ProductSyncService) downloadAndSaveProductImages(ctx context.Context, itemID int64, imageURLs []string) []string {
	zlog := zerolog.Ctx(ctx)
	localPaths := make([]string, 0, len(imageURLs))

	for i, url := range imageURLs {
		if url == "" {
			continue
		}

		img, err := s.imgMgr.CacheImage(ctx, s.tenantID, url)
		if err != nil {
			zlog.Warn().Err(err).Int("image_index", i).Int64("item_id", itemID).Msg("Failed to cache image")
			continue
		}

		paths := s.imgMgr.GetPaths(img)
		localPaths = append(localPaths, paths.Medium)
	}

	return localPaths
}

// updateProductLocalImages updates a product row with local image paths.
func (s *ProductSyncService) updateProductLocalImages(ctx context.Context, db *gorm.DB, productID string, localPaths []string) error {
	if len(localPaths) == 0 {
		return nil
	}

	pathsJSON, err := json.Marshal(localPaths)
	if err != nil {
		return err
	}

	return db.WithContext(ctx).
		Model(&models.ShopeeProduct{}).
		Where("id = ?", productID).
		Update("local_images", pathsJSON).Error
}
