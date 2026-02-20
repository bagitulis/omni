package master_product

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
)

// ImportFromTiktokStaging imports TikTok staging products into master products.
// All per-product errors are accumulated in result.Errors; only fatal top-level
// DB failures return a non-nil error.
func (s *StagingImportService) ImportFromTiktokStaging(
	ctx context.Context,
	tenantID string,
) (*StagingImportResult, error) {
	var result StagingImportResult

	// 1. Fetch all TiktokProducts for this tenant.
	var products []models.TiktokProduct
	if err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("tiktok staging import: fetch products: %w", err)
	}

	for _, p := range products {
		// 2a. Tier 1: Check if already linked by platform product ID.
		masterProduct, err := s.findMasterProductByPlatformItemID(ctx, tenantID, "tiktok", p.ProductID)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Errorf("tiktok product %s: find by platform item_id: %w", p.ProductID, err).Error())
			continue
		}

		created := false
		if masterProduct == nil {
			// 2b. Tier 2: Try matching by normalized product name.
			masterProduct, created, err = s.findOrCreateMasterProduct(ctx, tenantID, p.Name)
			if err != nil {
				result.Errors = append(result.Errors,
					fmt.Errorf("tiktok product %s: find/create master product: %w", p.ProductID, err).Error())
				continue
			}
		}

		// 2c. Tier 3: If still creating a new product, check by seller SKU first.
		if created {
			var tiktokSkus []models.TiktokSku
			if skuErr := s.db.WithContext(ctx).
				Where("tenant_id = ? AND product_id = ?", tenantID, p.ID).
				Find(&tiktokSkus).Error; skuErr == nil && len(tiktokSkus) > 0 {

				sellerSkuList := make([]string, 0, len(tiktokSkus))
				for _, sk := range tiktokSkus {
					if sk.SellerSku != "" {
						sellerSkuList = append(sellerSkuList, sk.SellerSku)
					}
				}

				if matchedBySku, skuErr := s.findMasterProductBySellerSkus(ctx, tenantID, sellerSkuList); skuErr == nil && matchedBySku != nil {
					_ = s.repo.Delete(ctx, masterProduct.ID)
					masterProduct = matchedBySku
					created = false
					result.ProductsCreated--
				}
			}
		}

		if created {
			result.ProductsCreated++
		} else {
			result.ProductsMatched++
		}

		// 2b. Fetch all SKUs for this product using the internal uint PK (p.ID), not p.ProductID.
		var skus []models.TiktokSku
		if err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND product_id = ?", tenantID, p.ID).
			Find(&skus).Error; err != nil {
			result.Errors = append(result.Errors,
				fmt.Errorf("tiktok product %s: fetch skus: %w", p.ProductID, err).Error())
			continue
		}

		if len(skus) == 0 {
			// 2c. No SKUs found — create a single default SKU.
			s.processTiktokDefaultSku(ctx, tenantID, p, masterProduct, &result)
		} else {
			// 2d. Process each SKU.
			for _, sku := range skus {
				s.processTiktokSku(ctx, tenantID, p, sku, masterProduct, &result)
			}
		}

		// 2e. Aggregate images from platform products into master product
		aggregator := NewImageAggregator(s.db)
		if err := aggregator.AggregateImagesForProduct(ctx, masterProduct.ID); err != nil {
			result.Errors = append(result.Errors,
				fmt.Errorf("tiktok product %s: aggregate images: %w", p.ProductID, err).Error())
		}
	}

	return &result, nil
}

// processTiktokSku creates a MasterProductSku and a platform link for one TiktokSku.
func (s *StagingImportService) processTiktokSku(
	ctx context.Context,
	tenantID string,
	p models.TiktokProduct,
	sku models.TiktokSku,
	masterProduct *models.MasterProduct,
	result *StagingImportResult,
) {
	sellerSku := sku.SellerSku
	if sellerSku == "" {
		sellerSku = fmt.Sprintf("tiktok_%s", p.ProductID)
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
			fmt.Errorf("tiktok product %s upsert sku: %w", p.ProductID, err).Error())
		return
	}
	if created {
		result.SkusCreated++
	}

	link := &models.MasterProductPlatformLink{
		MasterProductID:   masterProduct.ID,
		MasterSkuID:       &upsertedSku.ID,
		Platform:          "tiktok",
		PlatformProductID: p.ProductID,
		PlatformItemID:    p.ProductID,
		PlatformSkuID:     sku.SkuID,
		SyncStatus:        models.SyncStatusSynced,
	}

	if err := s.repo.UpsertPlatformLink(ctx, link); err != nil {
		result.Errors = append(result.Errors,
			fmt.Errorf("tiktok product %s: upsert platform link: %w", p.ProductID, err).Error())
		return
	}
	result.LinksCreated++
}

// processTiktokDefaultSku creates a single default MasterProductSku and platform link
// when no TiktokSku rows exist for the product.
func (s *StagingImportService) processTiktokDefaultSku(
	ctx context.Context,
	tenantID string,
	p models.TiktokProduct,
	masterProduct *models.MasterProduct,
	result *StagingImportResult,
) {
	sellerSku := fmt.Sprintf("tiktok_%s", p.ProductID)

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
			fmt.Errorf("tiktok product %s upsert sku: %w", p.ProductID, err).Error())
		return
	}
	if created {
		result.SkusCreated++
	}

	link := &models.MasterProductPlatformLink{
		MasterProductID:   masterProduct.ID,
		MasterSkuID:       &upsertedSku.ID,
		Platform:          "tiktok",
		PlatformProductID: p.ProductID,
		PlatformItemID:    p.ProductID,
		PlatformSkuID:     "",
		SyncStatus:        models.SyncStatusSynced,
	}

	if err := s.repo.UpsertPlatformLink(ctx, link); err != nil {
		result.Errors = append(result.Errors,
			fmt.Errorf("tiktok product %s: upsert platform link: %w", p.ProductID, err).Error())
		return
	}
	result.LinksCreated++
}
