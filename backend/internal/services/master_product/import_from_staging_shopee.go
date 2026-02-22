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
	result := NewStagingImportResult()

	// 1. Fetch all ShopeeProducts for this tenant.
	var products []models.ShopeeProduct
	if err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("shopee staging import: fetch products: %w", err)
	}

	for _, p := range products {
		platformItemID := strconv.FormatInt(p.ItemID, 10)

		// 2a. Tier 1: Check if already linked by platform item ID.
		masterProduct, err := s.findMasterProductByPlatformItemID(ctx, tenantID, "shopee", platformItemID)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Errorf("shopee product %d: find by platform item_id: %w", p.ItemID, err).Error())
			continue
		}

		created := false
		if masterProduct == nil {
			// 2b. Tier 2: Try matching by normalized product name.
			masterProduct, created, err = s.findOrCreateMasterProduct(ctx, tenantID, p.Name)
			if err != nil {
				result.Errors = append(result.Errors,
					fmt.Errorf("shopee product %d: find/create master product: %w", p.ItemID, err).Error())
				continue
			}
		}

		// 2c. Tier 3: If still creating a new product, check by seller SKU first.
		if created {
			// Fetch SKUs to check for seller_sku matches
			var shopeeSkus []models.ShopeeSku
			if skuErr := s.db.WithContext(ctx).
				Where("tenant_id = ? AND item_id = ?", tenantID, p.ItemID).
				Find(&shopeeSkus).Error; skuErr == nil && len(shopeeSkus) > 0 {

				sellerSkuList := make([]string, 0, len(shopeeSkus))
				for _, sk := range shopeeSkus {
					if sk.SellerSku != "" {
						sellerSkuList = append(sellerSkuList, sk.SellerSku)
					}
				}

				if matchedBySku, skuErr := s.findMasterProductBySellerSkus(ctx, tenantID, sellerSkuList); skuErr == nil && matchedBySku != nil {
					// Delete the just-created empty master product and use the SKU match instead
					_ = s.repo.Delete(ctx, masterProduct.ID)
					masterProduct = matchedBySku
					created = false
				}
			}
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
		} else {
			// 2d. Process each SKU.
			for _, sku := range skus {
				s.processShoepeeSku(ctx, tenantID, p, sku, masterProduct, &result)
			}
		}

		// 2e. Aggregate images from platform products into master product
		aggregator := NewImageAggregator(s.db)
		if err := aggregator.AggregateImagesForProduct(ctx, masterProduct.ID); err != nil {
			result.Errors = append(result.Errors,
				fmt.Errorf("shopee product %d: aggregate images: %w", p.ItemID, err).Error())
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

	masterSku := &models.MasterProductSku{
		TenantID:        tenantID,
		MasterProductID: masterProduct.ID,
		SellerSku:       sellerSku,
		VariantName:     sku.VariantName,
		Price:           sku.Price,
		Stock:           sku.Quantity,
	}

	upsertedSku, created, err := s.upsertMasterSku(ctx, masterSku)
	if err != nil {
		result.Errors = append(result.Errors,
			fmt.Errorf("shopee product %d upsert sku: %w", p.ItemID, err).Error())
		return
	}
	if created {
		result.SkusCreated++
	}

	platformSkuID := ""
	if sku.ModelID != nil {
		platformSkuID = strconv.FormatInt(*sku.ModelID, 10)
	}

	platformItemID := strconv.FormatInt(p.ItemID, 10)
	link := &models.MasterProductPlatformLink{
		MasterProductID:   masterProduct.ID,
		MasterSkuID:       &upsertedSku.ID,
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

	masterSku := &models.MasterProductSku{
		TenantID:        tenantID,
		MasterProductID: masterProduct.ID,
		SellerSku:       sellerSku,
		VariantName:     "",
		Price:           p.Price,
		Stock:           p.Quantity,
	}

	upsertedSku, created, err := s.upsertMasterSku(ctx, masterSku)
	if err != nil {
		result.Errors = append(result.Errors,
			fmt.Errorf("shopee product %d upsert sku: %w", p.ItemID, err).Error())
		return
	}
	if created {
		result.SkusCreated++
	}

	platformItemID := strconv.FormatInt(p.ItemID, 10)
	link := &models.MasterProductPlatformLink{
		MasterProductID:   masterProduct.ID,
		MasterSkuID:       &upsertedSku.ID,
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
