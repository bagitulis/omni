// Package products provides product cloning - TikTok helper methods
package products

import (
	"fmt"

	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
)

// =============================================================================
// TikTok: Image Upload Helpers
// =============================================================================

// uploadTiktokImages uploads images to TikTok CDN and returns image info and URIs
func (s *CloneService) uploadTiktokImages(client *tiktokPkg.Client, images []string) ([]tiktokPkg.ImageInfo, []string) {
	uploadedImages := make([]tiktokPkg.ImageInfo, 0, len(images))
	imageURIs := make([]string, 0, len(images))

	for i, imgURL := range images {
		if imgURL == "" {
			continue
		}
		log.Info().Int("index", i).Str("url", imgURL[:minInt(50, len(imgURL))]).Msg("Uploading image to TikTok")

		uploadResp, err := client.UploadImage(imgURL, "MAIN_IMAGE")
		if err != nil {
			log.Error().Err(err).Str("url", imgURL).Msg("Failed to upload image to TikTok")
			continue
		}
		if uploadResp.Data.URI != "" {
			uploadedImages = append(uploadedImages, tiktokPkg.ImageInfo{URI: uploadResp.Data.URI})
			imageURIs = append(imageURIs, uploadResp.Data.URI)
			log.Info().Str("uri", uploadResp.Data.URI).Msg("Image uploaded successfully")
		}
	}

	return uploadedImages, imageURIs
}

// =============================================================================
// TikTok: Category Helpers
// =============================================================================

// getTiktokCategory gets recommended category from TikTok API
func (s *CloneService) getTiktokCategory(client *tiktokPkg.Client, data *ProductData, imageURIs []string) (string, error) {
	log.Info().
		Str("source_category", data.CategoryID).
		Str("title", data.Name).
		Int("images", len(imageURIs)).
		Msg("Getting TikTok category recommendation based on product info")

	recommendedID, err := client.GetRecommendedCategoryID(data.Name, data.Description, imageURIs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get recommended category from TikTok")
		return "", fmt.Errorf("category recommendation failed: %w - please ensure product title and description are clear", err)
	}

	log.Info().
		Str("recommended_category_id", recommendedID).
		Str("product_title", data.Name).
		Msg("Got TikTok category from recommendation API")

	return recommendedID, nil
}

// getTiktokRequiredAttributes fetches required attributes for a category
func (s *CloneService) getTiktokRequiredAttributes(client *tiktokPkg.Client, categoryID string) []tiktokPkg.ProductAttribute {
	requiredAttrs, err := client.GetRequiredAttributes(categoryID)
	if err != nil {
		log.Warn().Err(err).Str("category_id", categoryID).Msg("Failed to get required attributes, proceeding without them")
		return []tiktokPkg.ProductAttribute{}
	}
	log.Info().Int("count", len(requiredAttrs)).Str("category_id", categoryID).Msg("Got required attributes for category")
	return requiredAttrs
}

// =============================================================================
// TikTok: SKU Building Helpers
// =============================================================================

// buildTiktokSKUs builds SKU list from product data
func (s *CloneService) buildTiktokSKUs(data *ProductData, warehouseID string) []tiktokPkg.CreateProductSku {
	if len(data.Variants) > 1 {
		return s.buildMultipleTiktokSKUs(data.Variants, warehouseID)
	}
	return []tiktokPkg.CreateProductSku{s.buildSingleTiktokSKU(data, warehouseID)}
}

// buildMultipleTiktokSKUs builds SKUs for products with multiple variants
func (s *CloneService) buildMultipleTiktokSKUs(variants []ProductVariant, warehouseID string) []tiktokPkg.CreateProductSku {
	log.Info().Int("variant_count", len(variants)).Msg("Building TikTok product with multiple variants")
	skus := make([]tiktokPkg.CreateProductSku, 0, len(variants))
	for _, v := range variants {
		sku := tiktokPkg.CreateProductSku{
			SellerSku:       v.SKU,
			Price:           &tiktokPkg.SkuPrice{Amount: fmt.Sprintf("%.0f", v.Price), Currency: "IDR"},
			SalesAttributes: []tiktokPkg.SalesAttr{{Name: "Varian", ValueName: v.Name}},
		}
		if warehouseID != "" {
			sku.Inventory = []tiktokPkg.SkuInventory{{WarehouseID: warehouseID, Quantity: v.Stock}}
		}
		skus = append(skus, sku)
	}
	return skus
}

// buildSingleTiktokSKU builds a single SKU for product without variants
func (s *CloneService) buildSingleTiktokSKU(data *ProductData, warehouseID string) tiktokPkg.CreateProductSku {
	sellerSku, price, stock := "", data.Price, data.Stock
	if len(data.Variants) > 0 {
		v := data.Variants[0]
		sellerSku = v.SKU
		if v.Price > 0 {
			price = v.Price
		}
		if v.Stock > 0 {
			stock = v.Stock
		}
	}
	sku := tiktokPkg.CreateProductSku{
		SellerSku: sellerSku,
		Price:     &tiktokPkg.SkuPrice{Amount: fmt.Sprintf("%.0f", price), Currency: "IDR"},
	}
	if warehouseID != "" {
		sku.Inventory = []tiktokPkg.SkuInventory{{WarehouseID: warehouseID, Quantity: stock}}
	}
	return sku
}

// =============================================================================
// TikTok: Request Building Helpers
// =============================================================================

// buildTiktokCreateRequest builds the TikTok create product request
func (s *CloneService) buildTiktokCreateRequest(
	data *ProductData,
	categoryID string,
	uploadedImages []tiktokPkg.ImageInfo,
	requiredAttrs []tiktokPkg.ProductAttribute,
	skus []tiktokPkg.CreateProductSku,
) tiktokPkg.CreateProductRequest {
	saveMode := "LISTING"
	if data.SaveAsDraft {
		saveMode = "AS_DRAFT"
	}

	return tiktokPkg.CreateProductRequest{
		Title:           data.Name,
		Description:     data.Description,
		CategoryID:      categoryID,
		CategoryVersion: "v2",
		MainImages:      uploadedImages,
		PackageWeight: tiktokPkg.PackageWeight{
			Value: "500",
			Unit:  "GRAM",
		},
		ProductAttributes: requiredAttrs,
		Skus:              skus,
		SaveMode:          saveMode,
	}
}

// logTiktokProductCreation logs product creation details
func (s *CloneService) logTiktokProductCreation(
	data *ProductData,
	categoryID string,
	uploadedImages []tiktokPkg.ImageInfo,
	skus []tiktokPkg.CreateProductSku,
	warehouseID string,
) {
	firstSku := ""
	firstPrice := data.Price
	firstStock := data.Stock

	if len(skus) > 0 {
		firstSku = skus[0].SellerSku
	}
	if len(data.Variants) > 0 {
		firstPrice = data.Variants[0].Price
		firstStock = data.Variants[0].Stock
	}

	log.Info().
		Str("title", data.Name).
		Str("category_id", categoryID).
		Int("images", len(uploadedImages)).
		Str("first_sku", firstSku).
		Int("total_skus", len(skus)).
		Str("first_price", fmt.Sprintf("%.0f", firstPrice)).
		Str("warehouse_id", warehouseID).
		Int("first_stock", firstStock).
		Msg("Creating TikTok product")
}
