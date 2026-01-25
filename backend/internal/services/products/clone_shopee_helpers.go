// Package products provides product cloning - Shopee helper methods
package products

import (
	"fmt"
	"strconv"

	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
)

// =============================================================================
// Shopee: Category Helpers
// =============================================================================

// getShopeeCategory gets or recommends category for Shopee
func (s *CloneService) getShopeeCategory(client *shopeePkg.Client, data *ProductData) (int64, error) {
	// If category already provided, use it
	if data.CategoryID != "" {
		var categoryID int64
		if _, err := fmt.Sscanf(data.CategoryID, "%d", &categoryID); err == nil && categoryID > 0 {
			return categoryID, nil
		}
	}

	// Auto-recommend category based on product name
	log.Info().Str("item_name", data.Name).Msg("Getting Shopee category recommendation")
	resp, err := client.GetCategoryRecommend(data.Name, "")
	if err != nil {
		return 0, fmt.Errorf("failed to get category recommendation: %w", err)
	}

	if len(resp.Response.CategoryID) == 0 {
		return 0, fmt.Errorf("no category recommendation found for product: %s", data.Name)
	}

	// Use first recommended category
	categoryID := resp.Response.CategoryID[0]
	log.Info().Int64("category_id", categoryID).Int("total_suggestions", len(resp.Response.CategoryID)).Msg("Got Shopee category recommendation")

	return categoryID, nil
}

// =============================================================================
// Shopee: Attribute Helpers
// =============================================================================

// getShopeeRecommendedAttributes gets recommended attributes for category
func (s *CloneService) getShopeeRecommendedAttributes(client *shopeePkg.Client, categoryID int64, itemName string) []shopeePkg.AttributeInfo {
	resp, err := client.GetRecommendAttribute(categoryID, itemName)
	if err != nil {
		log.Warn().Err(err).Int64("category_id", categoryID).Msg("Failed to get recommended attributes, proceeding without")
		return nil
	}

	var attributes []shopeePkg.AttributeInfo
	for _, attr := range resp.Response.AttributeList {
		// Only include mandatory attributes with values
		if !attr.IsMandatory || len(attr.AttributeValues) == 0 {
			continue
		}

		// Use first recommended value
		attrInfo := shopeePkg.AttributeInfo{
			AttributeID: attr.AttributeID,
			AttributeValueList: []shopeePkg.AttributeValue{
				{ValueID: attr.AttributeValues[0].ValueID},
			},
		}
		attributes = append(attributes, attrInfo)
	}

	log.Info().Int("count", len(attributes)).Int64("category_id", categoryID).Msg("Got Shopee recommended attributes")
	return attributes
}

// =============================================================================
// Shopee: Logistics Helpers
// =============================================================================

// getShopeeLogistics gets enabled logistics for shop
func (s *CloneService) getShopeeLogistics(client *shopeePkg.Client) []shopeePkg.LogisticInfo {
	resp, err := client.GetLogisticChannels()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get logistics, using default")
		return nil
	}

	var logistics []shopeePkg.LogisticInfo
	for _, ch := range resp.Response.LogisticsList {
		if ch.Enabled {
			logistics = append(logistics, shopeePkg.LogisticInfo{
				LogisticID: ch.LogisticID,
				Enabled:    true,
			})
		}
	}

	log.Info().Int("enabled", len(logistics)).Msg("Got Shopee logistics channels")
	return logistics
}

// =============================================================================
// Shopee: Image Upload Helpers
// =============================================================================

// uploadShopeeImages uploads images to Shopee CDN and returns image IDs
func (s *CloneService) uploadShopeeImages(client *shopeePkg.Client, imageURLs []string) []string {
	log.Info().Int("total", len(imageURLs)).Msg("Starting Shopee image upload")
	var imageIDs []string
	for i, url := range imageURLs {
		if i >= 9 { // Shopee allows max 9 images
			break
		}
		log.Debug().Str("url", url).Int("index", i).Msg("Uploading image to Shopee")
		imageID, err := client.UploadImageFromURL(url)
		if err != nil {
			log.Error().Err(err).Str("url", url).Int("index", i).Msg("Failed to upload image to Shopee")
			continue
		}
		log.Info().Str("image_id", imageID).Int("index", i).Msg("Successfully uploaded image")
		imageIDs = append(imageIDs, imageID)
	}
	log.Info().Int("success", len(imageIDs)).Int("total", len(imageURLs)).Msg("Shopee image upload complete")
	return imageIDs
}

// =============================================================================
// Shopee: Product Request Builder
// =============================================================================

// buildShopeeProductRequest builds the Shopee create product request
func (s *CloneService) buildShopeeProductRequest(
	data *ProductData,
	categoryID int64,
	imageIDs []string,
	logistics []shopeePkg.LogisticInfo,
	attributes []shopeePkg.AttributeInfo,
) shopeePkg.CreateProductRequest {
	itemStatus := "NORMAL"
	if data.SaveAsDraft {
		itemStatus = "UNLIST"
	}

	// Ensure minimum price (Shopee requires >= 99)
	price := data.Price
	if price < 99 {
		price = 99
	}

	// Ensure minimum stock
	stock := data.Stock
	if stock < 1 {
		stock = 1
	}

	return shopeePkg.CreateProductRequest{
		ItemName:      data.Name,
		Description:   s.ensureDescription(data.Description, data.Name),
		CategoryID:    categoryID,
		OriginalPrice: price,
		Weight:        0.5, // Default weight 500g
		ItemStatus:    itemStatus,
		Brand: &shopeePkg.BrandInfo{
			BrandID:       0, // 0 = No Brand
			OriginalBrand: "No Brand",
		},
		SellerStock: []shopeePkg.SellerStock{
			{Stock: stock},
		},
		Image: shopeePkg.ImageInfo{
			ImageIDList: imageIDs,
		},
		LogisticInfo: logistics,
		Attributes:   attributes,
	}
}

// logShopeeProductCreation logs product creation details
func (s *CloneService) logShopeeProductCreation(req shopeePkg.CreateProductRequest, stock int) {
	log.Info().
		Str("item_name", req.ItemName).
		Int64("category_id", req.CategoryID).
		Float64("price", req.OriginalPrice).
		Int("stock", stock).
		Int("images", len(req.Image.ImageIDList)).
		Int("logistics", len(req.LogisticInfo)).
		Int("attributes", len(req.Attributes)).
		Msg("Creating Shopee product")
}

// ensureDescription ensures description is not empty (uses shared helper)
func (s *CloneService) ensureDescription(desc, name string) string {
	return EnsureDescription(desc, name, 25)
}

// createShopeeProductWithRequest creates the product using provided request
func (s *CloneService) createShopeeProductWithRequest(client *shopeePkg.Client, req shopeePkg.CreateProductRequest) (string, error) {
	result, err := client.CreateProduct(req)
	if err != nil {
		return "", fmt.Errorf("shopee create product failed: %w", err)
	}

	return strconv.FormatInt(result.Response.ItemID, 10), nil
}
