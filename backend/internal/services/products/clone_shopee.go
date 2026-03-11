// Package products provides product cloning - Shopee platform operations
package products

import (
	"context"
	"fmt"
	"strconv"

	"github.com/omni/backend/internal/models"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
)

// =============================================================================
// Shopee: Fetch Product from Database
// =============================================================================

// fetchShopeeProductFromDB fetches Shopee product from database, enriched with API data if needed
func (s *CloneService) fetchShopeeProductFromDB(ctx context.Context, itemID string) (*ProductData, error) {
	// Find SKU first by item_id (if numeric) or seller_sku (if string)
	var sku models.ShopeeSku
	var err error

	// Try to parse as numeric item_id first
	if numericID, parseErr := strconv.ParseInt(itemID, 10, 64); parseErr == nil {
		// It's a numeric ID, search by item_id
		err = s.db.WithContext(ctx).
			Where("item_id = ?", numericID).
			First(&sku).Error
	}

	// If not found or not numeric, search by seller_sku
	if err != nil || sku.ID == 0 {
		err = s.db.WithContext(ctx).
			Where("seller_sku = ?", itemID).
			First(&sku).Error
	}

	if err != nil {
		return nil, fmt.Errorf("shopee product not found: %w", err)
	}

	// Get product
	var product models.ShopeeProduct
	if sku.ProductID > 0 {
		if err := s.db.WithContext(ctx).First(&product, sku.ProductID).Error; err != nil {
			return nil, fmt.Errorf("shopee product record not found (id=%d): %w", sku.ProductID, err)
		}
	}

	// Get all SKUs for this product
	var allSkus []models.ShopeeSku
	if sku.ProductID > 0 {
		if err := s.db.WithContext(ctx).Where("product_id = ?", sku.ProductID).Find(&allSkus).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch shopee SKUs for product %d: %w", sku.ProductID, err)
		}
	}

	// Parse images from product DB
	images := parseImagesFromDB(product.Image)
	categoryID := ""

	// If images empty, try to fetch from Shopee API
	if len(images) == 0 && sku.ItemID > 0 {
		log.Info().Int64("item_id", sku.ItemID).Msg("Fetching product data from Shopee API (DB images empty)")
		apiImages, apiCategoryID := s.fetchShopeeProductFromAPI(sku.ItemID)
		if len(apiImages) > 0 {
			images = apiImages
			log.Info().Int("count", len(images)).Msg("Got images from Shopee API")
		}
		if apiCategoryID != "" {
			categoryID = apiCategoryID
		}
	}

	// Build variants
	variants := buildVariantsFromShopee(allSkus)

	return &ProductData{
		ItemID:      itemID,
		Name:        product.Name,
		Description: product.Description,
		Price:       sku.Price,
		Stock:       sku.Quantity,
		CategoryID:  categoryID,
		Images:      images,
		Variants:    variants,
	}, nil
}

// fetchShopeeProductFromAPI fetches product details from Shopee API
func (s *CloneService) fetchShopeeProductFromAPI(itemID int64) (images []string, categoryID string) {
	if s.credService == nil {
		log.Error().Msg("CredService is nil, cannot fetch from Shopee API")
		return nil, ""
	}

	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "shopee")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Shopee credentials for API fetch")
		return nil, ""
	}

	log.Info().
		Int64("partner_id", creds.PartnerID).
		Int64("shop_id", creds.ShopID).
		Str("token_prefix", creds.AccessToken[:20]).
		Bool("is_production", creds.IsProduction).
		Msg("Using Shopee credentials")

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)

	log.Info().Int64("item_id", itemID).Msg("Calling Shopee GetProductDetailWithImages API")
	resp, err := client.GetProductDetailWithImages([]int64{itemID})
	if err != nil {
		log.Error().Err(err).Int64("item_id", itemID).Msg("Shopee API call failed")
		return nil, ""
	}

	log.Info().
		Str("error", resp.Error).
		Str("message", resp.Message).
		Int("item_count", len(resp.Response.ItemList)).
		Msg("Shopee API response")

	if len(resp.Response.ItemList) > 0 {
		item := resp.Response.ItemList[0]
		categoryID = strconv.FormatInt(item.CategoryID, 10)

		log.Info().
			Str("category_id", categoryID).
			Int("image_url_count", len(item.Image.ImageURLList)).
			Int("image_id_count", len(item.Image.ImageIDList)).
			Msg("Got product data from Shopee")

		if len(item.Image.ImageURLList) > 0 {
			return item.Image.ImageURLList, categoryID
		}
	}

	log.Warn().Msg("No images found from Shopee API")
	return nil, categoryID
}

// =============================================================================
// Shopee: Create Product via API
// =============================================================================

// createShopeeProduct creates a product on Shopee platform
func (s *CloneService) createShopeeProduct(_ context.Context, data *ProductData) (string, error) {
	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "shopee")
	if err != nil {
		return "", err
	}

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)

	// Step 1: Upload images to Shopee CDN
	imageIDs := s.uploadShopeeImages(client, data.Images)
	if len(imageIDs) == 0 {
		return "", fmt.Errorf("failed to upload any images to Shopee")
	}
	log.Info().Int("count", len(imageIDs)).Msg("Uploaded images to Shopee CDN")

	// Step 2: Get category (from data or auto-recommend)
	categoryID, err := s.getShopeeCategory(client, data)
	if err != nil {
		return "", err
	}
	log.Info().Int64("category_id", categoryID).Msg("Using Shopee category")

	// Step 3: Get recommended attributes
	attributes := s.getShopeeRecommendedAttributes(client, categoryID, data.Name)

	// Step 4: Get logistics
	logistics := s.getShopeeLogistics(client)

	// Step 5: Build and submit product request
	req := s.buildShopeeProductRequest(data, categoryID, imageIDs, logistics, attributes)

	// Log creation info
	s.logShopeeProductCreation(req, data.Stock)

	// Create product
	return s.createShopeeProductWithRequest(client, req)
}

// NOTE: Shopee helper methods are defined in clone_shopee_helpers.go
// NOTE: buildVariantsFromShopee is defined in clone_variant_helpers.go
