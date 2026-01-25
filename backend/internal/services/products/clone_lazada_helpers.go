// Package products provides product cloning - Lazada helper methods
package products

import (
	"fmt"
	"strings"

	lazadaPkg "github.com/omni/backend/pkg/lazada"
	"github.com/rs/zerolog/log"
)

// LazadaSkuPayload represents SKU data for XML payload
type LazadaSkuPayload struct {
	SellerSku string
	Name      string // Variant name (e.g., "Lemon", "Cofee")
	Price     float64
	Quantity  int
	Images    []string
}

// =============================================================================
// Lazada: Image Migration Helpers
// =============================================================================

// migrateLazadaImages migrates images to Lazada CDN
func (s *CloneService) migrateLazadaImages(client *lazadaPkg.Client, images []string) []string {
	migratedImages := make([]string, 0, len(images))

	for i, imgURL := range images {
		if imgURL == "" {
			continue
		}
		log.Info().Int("index", i).Str("url", imgURL[:minInt(50, len(imgURL))]).Msg("Migrating image to Lazada CDN")

		migratedURL, err := client.MigrateImage(imgURL)
		if err != nil {
			log.Error().Err(err).Str("url", imgURL).Msg("Failed to migrate image, using original")
			migratedImages = append(migratedImages, imgURL)
			continue
		}
		migratedImages = append(migratedImages, migratedURL)
		log.Info().Str("migrated_url", migratedURL[:minInt(50, len(migratedURL))]).Msg("Image migrated successfully")
	}

	return migratedImages
}

// =============================================================================
// Lazada: Category Helpers
// =============================================================================

// getLazadaCategory gets category - uses provided or auto-recommends
func (s *CloneService) getLazadaCategory(client *lazadaPkg.Client, data *ProductData) (int64, error) {
	// If category ID already provided, use it
	if data.CategoryID != "" {
		var categoryID int64
		_, err := fmt.Sscanf(data.CategoryID, "%d", &categoryID)
		if err == nil && categoryID > 0 {
			log.Info().Int64("category_id", categoryID).Msg("Using provided category ID")
			return categoryID, nil
		}
	}

	// Auto-recommend category based on product name
	log.Info().Str("product_name", data.Name).Msg("Getting Lazada category recommendation")

	categoryID, err := client.GetRecommendedCategoryID(data.Name)
	if err != nil {
		return 0, fmt.Errorf("category recommendation failed: %w - please ensure product title is clear", err)
	}

	log.Info().
		Int64("recommended_category_id", categoryID).
		Str("product_title", data.Name).
		Msg("Got Lazada category from recommendation API")

	return categoryID, nil
}

// logLazadaCategoryAttributes logs required attributes for category (informational)
func (s *CloneService) logLazadaCategoryAttributes(client *lazadaPkg.Client, categoryID int64) {
	attrs, err := client.GetRequiredAttributes(categoryID)
	if err != nil {
		log.Warn().Err(err).Int64("category_id", categoryID).Msg("Failed to get category attributes")
		return
	}
	log.Info().Int("count", len(attrs)).Int64("category_id", categoryID).Msg("Got required attributes for category")
}

// =============================================================================
// Lazada: SKU Building Helpers
// =============================================================================

// buildLazadaSKUs builds SKU list from product data
func (s *CloneService) buildLazadaSKUs(data *ProductData) []LazadaSkuPayload {
	var skus []LazadaSkuPayload

	if len(data.Variants) > 1 {
		// Multiple variants - log for debugging
		log.Info().Int("variant_count", len(data.Variants)).Msg("Building Lazada product with multiple variants")

		for i, variant := range data.Variants {
			sku := LazadaSkuPayload{
				SellerSku: variant.SKU,
				Name:      variant.Name,
				Price:     variant.Price,
				Quantity:  variant.Stock,
			}
			if sku.Price <= 0 {
				sku.Price = data.Price
			}
			if sku.Quantity <= 0 {
				sku.Quantity = data.Stock
			}
			skus = append(skus, sku)
			log.Info().
				Int("index", i).
				Str("sku", variant.SKU).
				Str("name", variant.Name).
				Float64("price", sku.Price).
				Int("stock", sku.Quantity).
				Msg("Added Lazada variant SKU")
		}
	} else if len(data.Variants) == 1 {
		// Single variant
		sku := LazadaSkuPayload{
			SellerSku: data.Variants[0].SKU,
			Name:      data.Variants[0].Name,
			Price:     data.Variants[0].Price,
			Quantity:  data.Variants[0].Stock,
		}
		if sku.Price <= 0 {
			sku.Price = data.Price
		}
		if sku.Quantity <= 0 {
			sku.Quantity = data.Stock
		}
		skus = append(skus, sku)
	} else {
		// No variants - single SKU
		skus = append(skus, LazadaSkuPayload{
			SellerSku: "",
			Price:     data.Price,
			Quantity:  data.Stock,
		})
	}

	return skus
}

// =============================================================================
// Lazada: XML Payload Building Helpers
// =============================================================================

// buildLazadaProductXML builds the XML payload for Lazada product creation
// Per Lazada docs: /product/create expects XML payload
func (s *CloneService) buildLazadaProductXML(data *ProductData, categoryID int64, images []string, skus []LazadaSkuPayload) string {
	// Build images XML
	imagesXML := ""
	for _, img := range images {
		imagesXML += fmt.Sprintf("<Image>%s</Image>", escapeXML(img))
	}

	// Determine if we have multiple variants (need color/size attribute)
	hasMultipleVariants := len(skus) > 1

	// Build SKUs XML
	skusXML := ""
	for _, sku := range skus {
		skuXML := fmt.Sprintf(`<Sku>
        <SellerSku>%s</SellerSku>
        <quantity>%d</quantity>
        <price>%.2f</price>
        <package_weight>0.5</package_weight>
        <package_height>5</package_height>
        <package_width>20</package_width>
        <package_length>30</package_length>
        <Images>%s</Images>`, escapeXML(sku.SellerSku), sku.Quantity, sku.Price, imagesXML)

		// Add variant name as color_family if multiple variants
		if hasMultipleVariants && sku.Name != "" {
			skuXML += fmt.Sprintf(`
        <color_family>%s</color_family>`, escapeXML(sku.Name))
		}

		skuXML += `
      </Sku>`
		skusXML += skuXML
	}

	// Build full XML payload
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Request>
  <Product>
    <PrimaryCategory>%d</PrimaryCategory>
    <Attributes>
      <name>%s</name>
      <description><![CDATA[%s]]></description>
      <brand>No Brand</brand>
      <warranty_type>No Warranty</warranty_type>
    </Attributes>
    <Skus>
      %s
    </Skus>
  </Product>
</Request>`, categoryID, escapeXML(data.Name), data.Description, skusXML)
}

// escapeXML escapes special XML characters using strings.Replace
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// =============================================================================
// Lazada: Logging Helpers
// =============================================================================

// logLazadaProductCreation logs product creation details
func (s *CloneService) logLazadaProductCreation(data *ProductData, categoryID int64, images []string, skus []LazadaSkuPayload) {
	firstSku := ""
	firstPrice := data.Price
	firstStock := data.Stock

	if len(skus) > 0 {
		firstSku = skus[0].SellerSku
		if skus[0].Price > 0 {
			firstPrice = skus[0].Price
		}
		if skus[0].Quantity > 0 {
			firstStock = skus[0].Quantity
		}
	}

	log.Info().
		Str("title", data.Name).
		Int64("category_id", categoryID).
		Int("images", len(images)).
		Str("first_sku", firstSku).
		Int("total_skus", len(skus)).
		Float64("first_price", firstPrice).
		Int("first_stock", firstStock).
		Msg("Creating Lazada product")
}
