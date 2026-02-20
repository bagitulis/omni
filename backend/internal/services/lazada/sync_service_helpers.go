package lazada

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func (s *SyncService) persistProductWithSkus(
	ctx context.Context,
	tx *gorm.DB,
	prodRepo *repositories.LazadaProductRepository,
	apiProduct lazadaPkg.Product,
) error {
	itemID := apiProduct.ItemID.String()
	dbProduct := s.buildLazadaProductModel(apiProduct)

	if err := prodRepo.Upsert(ctx, dbProduct); err != nil {
		return fmt.Errorf("upsert lazada product item_id=%s: %w", itemID, err)
	}

	if len(apiProduct.Images) > 0 {
		localPaths := s.downloadAndSaveProductImages(ctx, itemID, apiProduct.Images)
		s.updateProductLocalImagesWithDB(ctx, tx, dbProduct.ID, localPaths)
	}

	for _, apiSku := range apiProduct.Skus {
		dbSku := s.buildLazadaSkuModel(apiProduct.ItemID.String(), dbProduct.ID, apiSku)
		if err := prodRepo.UpsertSku(ctx, dbSku); err != nil {
			return fmt.Errorf("upsert lazada sku sku_id=%s: %w", dbSku.SkuID, err)
		}
	}

	return nil
}

func (s *SyncService) buildLazadaProductModel(apiProduct lazadaPkg.Product) *models.LazadaProduct {
	productName := apiProduct.Name
	if apiProduct.Attributes.Name != "" {
		productName = apiProduct.Attributes.Name
	}

	description := apiProduct.Description
	if apiProduct.Attributes.Description != "" {
		description = apiProduct.Attributes.Description
	}

	brand := apiProduct.Brand
	if apiProduct.Attributes.Brand != "" {
		brand = apiProduct.Attributes.Brand
	}

	imageURL := ""
	if len(apiProduct.Images) > 0 {
		imageURL = apiProduct.Images[0]
	}

	return &models.LazadaProduct{
		TenantID:    s.tenantID,
		ItemID:      apiProduct.ItemID.String(),
		Name:        productName,
		Description: description,
		Brand:       brand,
		Price:       apiProduct.Price,
		Status:      apiProduct.Status,
		Image:       imageURL,
	}
}

func (s *SyncService) buildLazadaSkuModel(itemID string, productID uint, apiSku lazadaPkg.ProductSku) *models.LazadaSku {
	skuName := apiSku.SellerSku
	if skuName == "" {
		skuName = apiSku.ShopSku
	}

	variantName := apiSku.Variation
	if variantName == "" && apiSku.Pilihan != "" {
		variantName = apiSku.Pilihan
	}

	return &models.LazadaSku{
		TenantID:     s.tenantID,
		ItemID:       itemID,
		ProductID:    productID,
		SkuID:        apiSku.SkuID.String(),
		ShopSku:      apiSku.ShopSku,
		SellerSku:    apiSku.SellerSku,
		Name:         skuName,
		VariantName:  variantName,
		Price:        apiSku.Price,
		SpecialPrice: apiSku.SpecialPrice,
		Quantity:     apiSku.Quantity,
		Available:    apiSku.Available,
	}
}

func (s *SyncService) buildProductRows(apiProduct lazadaPkg.Product) []map[string]interface{} {
	itemID := apiProduct.ItemID.String()
	productName := apiProduct.Name
	if apiProduct.Attributes.Name != "" {
		productName = apiProduct.Attributes.Name
	}

	rows := make([]map[string]interface{}, 0, len(apiProduct.Skus)+1)
	for _, apiSku := range apiProduct.Skus {
		skuName := apiSku.SellerSku
		if skuName == "" {
			skuName = apiSku.ShopSku
		}

		variantName := apiSku.Variation
		if variantName == "" && apiSku.Pilihan != "" {
			variantName = apiSku.Pilihan
		}

		rows = append(rows, map[string]interface{}{
			"item_id":      itemID,
			"sku_id":       apiSku.SkuID.String(),
			"sku_name":     skuName,
			"product_name": productName,
			"variant_name": variantName,
			"price":        apiSku.Price,
			"quantity":     apiSku.Quantity,
			"status":       apiProduct.Status,
		})
	}

	if len(apiProduct.Skus) == 0 {
		rows = append(rows, map[string]interface{}{
			"item_id":      itemID,
			"sku_id":       "",
			"sku_name":     productName,
			"product_name": productName,
			"variant_name": "",
			"price":        apiProduct.Price,
			"quantity":     0,
			"status":       apiProduct.Status,
		})
	}

	return rows
}

// downloadAndSaveProductImages downloads product images using unified ImageManager.
// Returns slice of local paths (medium size for display). Errors are logged but don't break sync.
func (s *SyncService) downloadAndSaveProductImages(ctx context.Context, itemID string, imageURLs []string) []string {
	localPaths := make([]string, 0, len(imageURLs))
	zlog := zerolog.Ctx(ctx)

	for imageIndex, imageURL := range imageURLs {
		if imageURL == "" {
			continue
		}

		img, err := s.imgMgr.CacheImage(ctx, s.tenantID, imageURL)
		if err != nil {
			zlog.Warn().
				Str("service", "lazada_sync").
				Int("image_index", imageIndex).
				Str("item_id", itemID).
				Err(err).
				Msg("Failed to cache image")
			continue
		}

		paths := s.imgMgr.GetPaths(img)
		localPaths = append(localPaths, paths.Medium)
	}

	return localPaths
}

// updateProductLocalImages updates the product with local image paths.
func (s *SyncService) updateProductLocalImages(ctx context.Context, productID uint, localPaths []string) {
	s.updateProductLocalImagesWithDB(ctx, s.db, productID, localPaths)
}

func (s *SyncService) updateProductLocalImagesWithDB(ctx context.Context, db *gorm.DB, productID uint, localPaths []string) {
	if len(localPaths) == 0 || productID == 0 {
		return
	}

	zlog := zerolog.Ctx(ctx)

	pathsJSON, err := json.Marshal(localPaths)
	if err != nil {
		zlog.Warn().
			Str("service", "lazada_sync").
			Err(err).
			Msg("Failed to marshal local paths")
		return
	}

	err = db.WithContext(ctx).
		Model(&models.LazadaProduct{}).
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
