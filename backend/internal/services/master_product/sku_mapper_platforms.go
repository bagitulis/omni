// Package master_product provides SKU mapping service
package master_product

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
)

// ================== Platform-specific SKU lookup ==================

func (m *SkuMapper) findShopeeSku(ctx context.Context, sellerSku string) AutoMapResult {
	normalizedSku := normalizeSkuForLookup(sellerSku)

	result := AutoMapResult{
		Platform: models.PlatformShopee,
		Found:    false,
	}

	var skus []models.ShopeeSku
	err := m.db.WithContext(ctx).
		Where("tenant_id = ? AND LOWER(seller_sku) = LOWER(?)", m.tenantID, normalizedSku).
		Find(&skus).Error

	if err != nil || len(skus) == 0 {
		result.Message = "SKU not found in Shopee"
		return result
	}

	if len(skus) > 1 {
		result.Ambiguous = true
		result.Message = fmt.Sprintf("Found %d products with this SKU - manual selection required", len(skus))
		return result
	}

	sku := skus[0]
	result.Found = true
	result.PlatformItemID = fmt.Sprintf("%d", sku.ItemID)
	if sku.ModelID != nil {
		result.PlatformSkuID = fmt.Sprintf("%d", *sku.ModelID)
	}
	result.Message = "SKU found in Shopee"

	return result
}

func (m *SkuMapper) findTiktokSku(ctx context.Context, sellerSku string) AutoMapResult {
	normalizedSku := normalizeSkuForLookup(sellerSku)

	result := AutoMapResult{
		Platform: models.PlatformTiktok,
		Found:    false,
	}

	var skus []models.TiktokSku
	err := m.db.WithContext(ctx).
		Where("tenant_id = ? AND LOWER(seller_sku) = LOWER(?)", m.tenantID, normalizedSku).
		Find(&skus).Error

	if err != nil || len(skus) == 0 {
		result.Message = "SKU not found in TikTok"
		return result
	}

	if len(skus) > 1 {
		result.Ambiguous = true
		result.Message = fmt.Sprintf("Found %d products with this SKU - manual selection required", len(skus))
		return result
	}

	sku := skus[0]
	result.Found = true
	result.PlatformItemID = fmt.Sprintf("%d", sku.ProductID)
	result.PlatformSkuID = sku.SkuID
	result.Message = "SKU found in TikTok"

	return result
}

func (m *SkuMapper) findLazadaSku(ctx context.Context, sellerSku string) AutoMapResult {
	normalizedSku := normalizeSkuForLookup(sellerSku)

	result := AutoMapResult{
		Platform: models.PlatformLazada,
		Found:    false,
	}

	var skus []models.LazadaSku
	err := m.db.WithContext(ctx).
		Where("tenant_id = ? AND (LOWER(seller_sku) = LOWER(?) OR LOWER(shop_sku) = LOWER(?))", m.tenantID, normalizedSku, normalizedSku).
		Find(&skus).Error

	if err != nil || len(skus) == 0 {
		result.Message = "SKU not found in Lazada"
		return result
	}

	if len(skus) > 1 {
		result.Ambiguous = true
		result.Message = fmt.Sprintf("Found %d products with this SKU - manual selection required", len(skus))
		return result
	}

	sku := skus[0]
	result.Found = true
	result.PlatformItemID = sku.ItemID
	result.PlatformSkuID = sku.SkuID
	result.Message = "SKU found in Lazada"

	return result
}
