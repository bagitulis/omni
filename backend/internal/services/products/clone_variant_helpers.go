// Package products provides variant building helpers for product cloning
package products

import (
	"github.com/omni/backend/internal/models"
)

// =============================================================================
// Variant Building Functions - DRY pattern for all platforms
// =============================================================================

// buildVariantsFromShopee converts Shopee SKUs to ProductVariant slice
func buildVariantsFromShopee(allSkus []models.ShopeeSku) []ProductVariant {
	return buildVariants(allSkus, func(i int) (string, string, float64, int) {
		sku := allSkus[i]
		return sku.SellerSku, sku.VariantName, sku.Price, sku.Quantity
	})
}

// buildVariantsFromLazada converts Lazada SKUs to ProductVariant slice
func buildVariantsFromLazada(allSkus []models.LazadaSku) []ProductVariant {
	return buildVariants(allSkus, func(i int) (string, string, float64, int) {
		sku := allSkus[i]
		sellerSku := sku.SellerSku
		if sellerSku == "" {
			sellerSku = sku.SkuID
		}
		return sellerSku, sku.VariantName, sku.Price, sku.Quantity
	})
}

// buildVariantsFromTiktok converts TikTok SKUs to ProductVariant slice
func buildVariantsFromTiktok(allSkus []models.TiktokSku) []ProductVariant {
	return buildVariants(allSkus, func(i int) (string, string, float64, int) {
		sku := allSkus[i]
		sellerSku := sku.SellerSku
		if sellerSku == "" {
			sellerSku = sku.SkuID
		}
		return sellerSku, sku.VariantName, sku.Price, sku.Quantity
	})
}

// buildVariants is a generic variant builder using a field extractor function
func buildVariants[T any](skus []T, extract func(i int) (sku, name string, price float64, stock int)) []ProductVariant {
	variants := make([]ProductVariant, 0, len(skus))
	for i := range skus {
		sku, name, price, stock := extract(i)
		variants = append(variants, ProductVariant{
			SKU:   sku,
			Name:  name,
			Price: price,
			Stock: stock,
		})
	}
	return variants
}

// =============================================================================
// Description Helper - Shared by all platforms
// =============================================================================

// EnsureDescription ensures description is not empty (Shopee requires min 25 chars)
func EnsureDescription(desc, name string, minLength int) string {
	if desc != "" && len(desc) >= minLength {
		return desc
	}
	// Default description template
	return name + " - Produk berkualitas tinggi dengan harga terjangkau. Segera order sekarang!"
}
