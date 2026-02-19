// Package master_product provides import service for Master Product
package master_product

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ImportFromLazadaStaging imports all Lazada staging products into Master Product for a tenant.
// It iterates LazadaProduct rows, finds-or-creates a MasterProduct per product, then maps
// every LazadaSku into a MasterProductSku with a corresponding MasterProductPlatformLink.
func (s *StagingImportService) ImportFromLazadaStaging(ctx context.Context, tenantID string) (*StagingImportResult, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	result := &StagingImportResult{}
	imageAggregator := NewImageAggregator(s.db)

	// 1. Fetch all Lazada products for the tenant.
	var products []models.LazadaProduct
	if err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("fetch lazada products: %w", err)
	}

	for _, p := range products {
		// 2a. Fetch all SKUs for this product (join via ItemID string).
		var skus []models.LazadaSku
		if err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND item_id = ?", tenantID, p.ItemID).
			Find(&skus).Error; err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("lazada product %s fetch skus: %w", p.ItemID, err).Error())
			continue
		}

		// 2a. Find or create master product by name.
		masterProduct, created, err := s.resolveLazadaMasterProduct(ctx, tenantID, p.Name, p.ItemID, skus)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("lazada product %s master product: %w", p.ItemID, err).Error())
			result.ProductsSkipped++
			continue
		}

		// 2b. Track created vs matched.
		if created {
			result.ProductsCreated++
		} else {
			result.ProductsMatched++
		}

		if err := s.normalizeLazadaMasterProductTitle(ctx, masterProduct, p.Name); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("lazada product %s normalize title: %w", p.ItemID, err).Error())
		}

		// 2e. No SKUs — create one default SKU with an empty variant name.
		if len(skus) == 0 {
			sellerSku := fmt.Sprintf("lazada_%s", p.ItemID)

			defaultSku := &models.MasterProductSku{
				TenantID:        tenantID,
				MasterProductID: masterProduct.ID,
				SellerSku:       sellerSku,
				VariantName:     "",
			}

			upsertedSku, created, err := s.upsertMasterSku(ctx, defaultSku)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("lazada product %s default sku: %w", p.ItemID, err).Error())
				continue
			}
			if created {
				result.SkusCreated++
			}

			link := &models.MasterProductPlatformLink{
				MasterProductID:   masterProduct.ID,
				MasterSkuID:       &upsertedSku.ID,
				Platform:          "lazada",
				PlatformProductID: p.ItemID,
				PlatformItemID:    p.ItemID,
				PlatformSkuID:     "",
				SyncStatus:        models.SyncStatusSynced,
			}
			if err := s.repo.UpsertPlatformLink(ctx, link); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("lazada product %s default sku link: %w", p.ItemID, err).Error())
				continue
			}
			result.LinksCreated++

			if err := imageAggregator.AggregateImagesForProduct(ctx, masterProduct.ID); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("lazada product %s aggregate images: %w", p.ItemID, err).Error())
			}
			continue
		}

		// 2d. Process each SKU.
		for _, sku := range skus {
			// ii. Seller SKU: prefer SellerSku, fallback to "lazada_{ItemID}".
			sellerSku := sku.SellerSku
			if sellerSku == "" {
				sellerSku = fmt.Sprintf("lazada_%s", p.ItemID)
			}

			// i. Variant name: prefer VariantName. Use Name only when it's not identical to seller_sku.
			variantName := resolveLazadaVariantName(sku, sellerSku)

			// iii. Create master SKU.
			masterSku := &models.MasterProductSku{
				TenantID:        tenantID,
				MasterProductID: masterProduct.ID,
				SellerSku:       sellerSku,
				VariantName:     variantName,
				Price:           sku.Price,
				Stock:           sku.Quantity,
			}

			// iv. Persist SKU; on error append and continue.
			upsertedSku, created, err := s.upsertMasterSku(ctx, masterSku)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("lazada product %s sku: %w", p.ItemID, err).Error())
				continue
			}

			if err := s.normalizeLegacyLazadaSkuVariant(ctx, upsertedSku, variantName); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("lazada product %s normalize sku variant: %w", p.ItemID, err).Error())
			}
			if created {
				result.SkusCreated++
			}

			// vi. Upsert platform link.
			link := &models.MasterProductPlatformLink{
				MasterProductID:   masterProduct.ID,
				MasterSkuID:       &upsertedSku.ID,
				Platform:          "lazada",
				PlatformProductID: p.ItemID,
				PlatformItemID:    p.ItemID,
				PlatformSkuID:     sku.SkuID,
				SyncStatus:        models.SyncStatusSynced,
			}
			// vii. On error append and continue.
			if err := s.repo.UpsertPlatformLink(ctx, link); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("lazada product %s sku link: %w", p.ItemID, err).Error())
				continue
			}
			// viii.
			result.LinksCreated++
		}

		if err := imageAggregator.AggregateImagesForProduct(ctx, masterProduct.ID); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("lazada product %s aggregate images: %w", p.ItemID, err).Error())
		}
	}

	return result, nil
}

func (s *StagingImportService) resolveLazadaMasterProduct(
	ctx context.Context,
	tenantID string,
	productName string,
	itemID string,
	skus []models.LazadaSku,
) (*models.MasterProduct, bool, error) {
	matchedByItemID, err := s.findMasterProductByLazadaItemID(ctx, tenantID, itemID)
	if err != nil {
		return nil, false, err
	}
	if matchedByItemID != nil {
		return matchedByItemID, false, nil
	}

	trimmedName := strings.TrimSpace(productName)
	if trimmedName != "" {
		return s.findOrCreateMasterProduct(ctx, tenantID, trimmedName)
	}

	matchedBySku, err := s.findMasterProductByLazadaSellerSkus(ctx, tenantID, skus)
	if err != nil {
		return nil, false, err
	}
	if matchedBySku != nil {
		return matchedBySku, false, nil
	}

	fallbackTitle := fmt.Sprintf("Lazada Item %s", strings.TrimSpace(itemID))
	return s.findOrCreateMasterProduct(ctx, tenantID, fallbackTitle)
}

func (s *StagingImportService) findMasterProductByLazadaItemID(
	ctx context.Context,
	tenantID string,
	itemID string,
) (*models.MasterProduct, error) {
	trimmedItemID := strings.TrimSpace(itemID)
	if trimmedItemID == "" {
		return nil, nil
	}

	links, err := s.repo.FindPlatformLinksByItemID(ctx, tenantID, "lazada", trimmedItemID)
	if err != nil {
		return nil, fmt.Errorf("find lazada platform links by item_id %s: %w", trimmedItemID, err)
	}
	if len(links) == 0 {
		return nil, nil
	}

	product, err := s.repo.FindByTenantAndID(ctx, tenantID, links[0].MasterProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find master product %d for lazada item_id %s: %w", links[0].MasterProductID, trimmedItemID, err)
	}

	return product, nil
}

func (s *StagingImportService) findMasterProductByLazadaSellerSkus(
	ctx context.Context,
	tenantID string,
	skus []models.LazadaSku,
) (*models.MasterProduct, error) {
	for _, sku := range skus {
		sellerSku := strings.TrimSpace(sku.SellerSku)
		if sellerSku == "" {
			continue
		}

		existingSku, err := s.repo.FindBySku(ctx, tenantID, sellerSku)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, fmt.Errorf("find master sku by seller_sku %s: %w", sellerSku, err)
		}

		product, err := s.repo.FindByTenantAndID(ctx, tenantID, existingSku.MasterProductID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, fmt.Errorf("find master product %d for seller_sku %s: %w", existingSku.MasterProductID, sellerSku, err)
		}

		return product, nil
	}

	return nil, nil
}

func resolveLazadaVariantName(sku models.LazadaSku, sellerSku string) string {
	variantName := strings.TrimSpace(sku.VariantName)
	if variantName != "" {
		return variantName
	}

	nameValue := strings.TrimSpace(sku.Name)
	trimmedSellerSku := strings.TrimSpace(sellerSku)
	if nameValue != "" && !strings.EqualFold(nameValue, trimmedSellerSku) {
		return nameValue
	}

	return ""
}

func (s *StagingImportService) normalizeLazadaMasterProductTitle(
	ctx context.Context,
	masterProduct *models.MasterProduct,
	productName string,
) error {
	if masterProduct == nil {
		return nil
	}

	trimmedName := strings.TrimSpace(productName)
	if trimmedName == "" {
		return nil
	}

	if strings.TrimSpace(masterProduct.Title) != "" {
		return nil
	}

	masterProduct.Title = trimmedName
	return s.repo.Update(ctx, masterProduct)
}

func (s *StagingImportService) normalizeLegacyLazadaSkuVariant(
	ctx context.Context,
	sku *models.MasterProductSku,
	incomingVariantName string,
) error {
	if sku == nil {
		return nil
	}

	if strings.TrimSpace(incomingVariantName) != "" {
		return nil
	}

	trimmedSellerSku := strings.TrimSpace(sku.SellerSku)
	trimmedVariantName := strings.TrimSpace(sku.VariantName)
	if trimmedSellerSku == "" || trimmedVariantName == "" {
		return nil
	}

	if !strings.EqualFold(trimmedVariantName, trimmedSellerSku) {
		return nil
	}

	sku.VariantName = ""
	return s.repo.UpdateSku(ctx, sku)
}
