// Package products provides product cloning - TikTok platform operations
package products

import (
	"context"
	"fmt"
	"strconv"

	"github.com/omni/backend/internal/models"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
)

// =============================================================================
// TikTok: Fetch Product from Database
// =============================================================================

// fetchTiktokProductFromDB fetches TikTok product from database
// Falls back to API if images are empty in database
func (s *CloneService) fetchTiktokProductFromDB(ctx context.Context, itemID string) (*ProductData, error) {
	var sku models.TiktokSku
	var err error

	// Try to parse as numeric product_id first
	if numericID, parseErr := strconv.ParseUint(itemID, 10, 64); parseErr == nil {
		err = s.db.WithContext(ctx).
			Where("product_id = ?", numericID).
			First(&sku).Error
	}

	// If not found, search by sku_id or seller_sku (strings)
	if err != nil || sku.ID == "" {
		err = s.db.WithContext(ctx).
			Where("sku_id = ? OR seller_sku = ?", itemID, itemID).
			First(&sku).Error
	}

	if err != nil {
		return nil, fmt.Errorf("tiktok product not found by id/sku '%s': %w", itemID, err)
	}

	var product models.TiktokProduct
	if sku.ProductID != "" {
		if err := s.db.WithContext(ctx).Where("id = ?", sku.ProductID).First(&product).Error; err != nil {
			return nil, fmt.Errorf("tiktok product record not found (id=%s): %w", sku.ProductID, err)
		}
	}

	var allSkus []models.TiktokSku
	if sku.ProductID != "" {
		if err := s.db.WithContext(ctx).Where("product_id = ?", sku.ProductID).Find(&allSkus).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch tiktok SKUs for product %s: %w", sku.ProductID, err)
		}
	}

	images := parseImagesFromDB(product.Image)
	variants := buildVariantsFromTiktok(allSkus)

	// Fallback: If images empty in DB, fetch from TikTok API
	if len(images) == 0 && product.ProductID != "" {
		log.Info().
			Str("product_id", product.ProductID).
			Msg("No images in DB, fetching from TikTok API")

		apiImages := s.fetchTiktokImagesFromAPI(ctx, product.ProductID)
		if len(apiImages) > 0 {
			images = apiImages
		}
	}

	return &ProductData{
		ItemID:      itemID,
		Name:        product.Name,
		Description: product.Description,
		Price:       sku.Price,
		Stock:       sku.Quantity,
		CategoryID:  "",
		Images:      images,
		Variants:    variants,
	}, nil
}

// fetchTiktokImagesFromAPI fetches product images from TikTok API
func (s *CloneService) fetchTiktokImagesFromAPI(ctx context.Context, productID string) []string {
	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "tiktok")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get TikTok credentials for API fetch")
		return nil
	}

	client := tiktokPkg.NewClient(creds.AppKey, creds.AppSecret)
	client.SetCredentials(creds.AccessToken, creds.ShopCipher)

	resp, err := client.GetProductDetail(ctx, productID)
	if err != nil {
		log.Error().Err(err).Str("product_id", productID).Msg("Failed to fetch TikTok product detail")
		return nil
	}

	if resp.Code != 0 || len(resp.Data.MainImages) == 0 {
		log.Warn().Int("code", resp.Code).Msg("TikTok API returned no images")
		return nil
	}

	var images []string
	for _, img := range resp.Data.MainImages {
		// Prefer URLs array, fallback to URI
		if len(img.URLs) > 0 {
			images = append(images, img.URLs[0])
		} else if img.URI != "" {
			images = append(images, img.URI)
		}
	}

	log.Info().Int("count", len(images)).Msg("Fetched images from TikTok API")
	return images
}

// =============================================================================
// TikTok: Create Product via API
// =============================================================================

// createTiktokProduct creates a product on TikTok platform
func (s *CloneService) createTiktokProduct(ctx context.Context, data *ProductData) (string, error) {
	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "tiktok")
	if err != nil {
		return "", fmt.Errorf("failed to get TikTok credentials: %w", err)
	}

	client := tiktokPkg.NewClient(creds.AppKey, creds.AppSecret)
	client.SetCredentials(creds.AccessToken, creds.ShopCipher)

	// Step 1: Get warehouse ID
	warehouseID, err := client.GetDefaultWarehouseID(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get warehouse ID, proceeding without it")
	}
	log.Info().Str("warehouse_id", warehouseID).Msg("Using warehouse for TikTok product")

	// Step 2: Upload images to TikTok CDN
	uploadedImages, imageURIs := s.uploadTiktokImages(client, data.Images)
	if len(uploadedImages) == 0 {
		return "", fmt.Errorf("failed to upload any images to TikTok - at least one image is required")
	}

	// Step 3: Get recommended category from TikTok (uses image URIs for better recommendation)
	categoryID, err := s.getTiktokCategory(ctx, client, data, imageURIs)
	if err != nil {
		return "", err
	}

	// Step 4: Get required attributes for this category
	requiredAttrs := s.getTiktokRequiredAttributes(ctx, client, categoryID)

	// Step 5: Build SKUs with variants
	skus := s.buildTiktokSKUs(data, warehouseID)

	// Step 6: Create product request
	req := s.buildTiktokCreateRequest(data, categoryID, uploadedImages, requiredAttrs, skus)

	// Log creation info
	s.logTiktokProductCreation(data, categoryID, uploadedImages, skus, warehouseID)

	// Step 7: Call TikTok API
	result, err := client.CreateProduct(ctx, req)
	if err != nil {
		return "", fmt.Errorf("TikTok CreateProduct: %w", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("TikTok CreateProduct rejected (code=%d): %s", result.Code, result.Message)
	}

	if result.Data.ProductID != "" {
		log.Info().Str("product_id", result.Data.ProductID).Msg("TikTok product created successfully")
		return result.Data.ProductID, nil
	}

	return "", fmt.Errorf("no product_id returned from TikTok")
}

// NOTE: TikTok helper methods are defined in clone_tiktok_helpers.go
// NOTE: buildVariantsFromTiktok is defined in clone_variant_helpers.go
