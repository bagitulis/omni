package master_product

import (
	"context"
	"fmt"
	"strconv"

	"github.com/omni/backend/internal/models"
)

// ImportFromShopeeStaging imports Shopee staging products into master products.
// All per-product errors are accumulated in result.Errors; only fatal top-level
// DB failures return a non-nil error.
func (s *StagingImportService) ImportFromShopeeStaging(
	ctx context.Context,
	tenantID string,
) (*StagingImportResult, error) {
	var result StagingImportResult

	// 1. Fetch all ShopeeProducts for this tenant.
	var products []models.ShopeeProduct
	if err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("shopee staging import: fetch products: %w", err)
	}

	for _, p := range products {
		// 2a. Find or create a master product by product name.
		masterProduct, created, err := s.findOrCreateMasterProduct(ctx, tenantID, p.Name)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Errorf("shopee product %d: find/create master product: %w", p.ItemID, err).Error())
			continue
		}

		if created {
			result.ProductsCreated++
		} else {
			result.ProductsMatched++
		}

		// 2b. Fetch all SKUs for this product.
		var skus []models.ShopeeSku
		if err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND item_id = ?", tenantID, p.ItemID).
			Find(&skus).Error; err != nil {
			result.Errors = append(result.Errors,
				fmt.Errorf("shopee product %d: fetch skus: %w", p.ItemID, err).Error())
			continue
		}

		if len(skus) == 0 {
			// 2c. No SKUs found — create a single default SKU.
			s.processShopeeDefaultSku(ctx, tenantID, p, masterProduct, &result)
			continue
		}

		// 2d. Process each SKU.
		for _, sku := range skus {
			s.processShoepeeSku(ctx, tenantID, p, sku, masterProduct, &result)
		}
	}

	return &result, nil
}

// processShoepeeSku creates a MasterProductSku and a platform link for one ShopeeSku.
func (s *StagingImportService) processShoepeeSku(
	ctx context.Context,
	tenantID string,
	p models.ShopeeProduct,
	sku models.ShopeeSku,
	masterProduct *models.MasterProduct,
	result *StagingImportResult,
) {
	sellerSku := sku.SellerSku
	if sellerSku == "" {
		sellerSku = fmt.Sprintf("shopee_%d", p.ItemID)
	}

	masterSku := models.MasterProductSku{
		TenantID:        tenantID,
		MasterProductID: masterProduct.ID,
		SellerSku:       sellerSku,
		VariantName:     sku.VariantName,
		Price:           sku.Price,
		Stock:           sku.Quantity,
	}

	if err := s.repo.CreateSku(ctx, &masterSku); err != nil {
		result.Errors = append(result.Errors,
			fmt.Errorf("shopee product %d sku: %w", p.ItemID, err).Error())
		return
	}
	result.SkusCreated++

	platformSkuID := ""
	if sku.ModelID != nil {
		platformSkuID = strconv.FormatInt(*sku.ModelID, 10)
	}

	platformItemID := strconv.FormatInt(p.ItemID, 10)
	link := &models.MasterProductPlatformLink{
		MasterProductID:   masterProduct.ID,
		MasterSkuID:       &masterSku.ID,
		Platform:          "shopee",
		PlatformProductID: platformItemID,
		PlatformItemID:    platformItemID,
		PlatformSkuID:     platformSkuID,
		SyncStatus:        models.SyncStatusSynced,
	}

	if err := s.repo.UpsertPlatformLink(ctx, link); err != nil {
		result.Errors = append(result.Errors,
			fmt.Errorf("shopee product %d: upsert platform link: %w", p.ItemID, err).Error())
		return
	}
	result.LinksCreated++
}

// processShopeeDefaultSku creates a single default MasterProductSku and platform link
// when no ShopeeSku rows exist for the product.
func (s *StagingImportService) processShopeeDefaultSku(
	ctx context.Context,
	tenantID string,
	p models.ShopeeProduct,
	masterProduct *models.MasterProduct,
	result *StagingImportResult,
) {
	sellerSku := fmt.Sprintf("shopee_%d", p.ItemID)

	masterSku := models.MasterProductSku{
		TenantID:        tenantID,
		MasterProductID: masterProduct.ID,
		SellerSku:       sellerSku,
		VariantName:     "",
		Price:           p.Price,
		Stock:           p.Quantity,
	}

	if err := s.repo.CreateSku(ctx, &masterSku); err != nil {
		result.Errors = append(result.Errors,
			fmt.Errorf("shopee product %d sku: %w", p.ItemID, err).Error())
		return
	}
	result.SkusCreated++

	platformItemID := strconv.FormatInt(p.ItemID, 10)
	link := &models.MasterProductPlatformLink{
		MasterProductID:   masterProduct.ID,
		MasterSkuID:       &masterSku.ID,
		Platform:          "shopee",
		PlatformProductID: platformItemID,
		PlatformItemID:    platformItemID,
		PlatformSkuID:     "",
		SyncStatus:        models.SyncStatusSynced,
	}

	if err := s.repo.UpsertPlatformLink(ctx, link); err != nil {
		result.Errors = append(result.Errors,
			fmt.Errorf("shopee product %d: upsert platform link: %w", p.ItemID, err).Error())
		return
	}
	result.LinksCreated++
}
