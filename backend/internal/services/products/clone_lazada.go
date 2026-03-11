// Package products provides product cloning - Lazada platform operations
package products

import (
	"context"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	"github.com/rs/zerolog/log"
)

// =============================================================================
// Lazada: Fetch Product from Database
// =============================================================================

// fetchLazadaProductFromDB fetches Lazada product from database
func (s *CloneService) fetchLazadaProductFromDB(ctx context.Context, itemID string) (*ProductData, error) {
	var sku models.LazadaSku
	// Search by item_id, sku_id, or seller_sku (all strings in Lazada model)
	err := s.db.WithContext(ctx).
		Where("item_id = ? OR sku_id = ? OR seller_sku = ?", itemID, itemID, itemID).
		First(&sku).Error
	if err != nil {
		return nil, fmt.Errorf("lazada product not found by id/sku '%s': %w", itemID, err)
	}

	var product models.LazadaProduct
	if sku.ProductID > 0 {
		if err := s.db.WithContext(ctx).First(&product, sku.ProductID).Error; err != nil {
			return nil, fmt.Errorf("lazada product record not found (id=%d): %w", sku.ProductID, err)
		}
	}

	var allSkus []models.LazadaSku
	if sku.ProductID > 0 {
		if err := s.db.WithContext(ctx).Where("product_id = ?", sku.ProductID).Find(&allSkus).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch lazada SKUs for product %d: %w", sku.ProductID, err)
		}
	}

	images := parseImagesFromDB(product.Image)
	variants := buildVariantsFromLazada(allSkus)

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

// =============================================================================
// Lazada: Create Product via API (with auto-category like TikTok)
// =============================================================================

// createLazadaProduct creates a product on Lazada platform
// Flow mirrors clone_tiktok.go: auto-recommend category, migrate images, build XML payload
func (s *CloneService) createLazadaProduct(_ context.Context, data *ProductData) (string, error) {
	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "lazada")
	if err != nil {
		return "", fmt.Errorf("failed to get Lazada credentials: %w", err)
	}

	client := lazadaPkg.NewClient(creds.AppKey, creds.AppSecret, creds.Region)
	client.SetAccessToken(creds.AccessToken)

	// Step 1: Migrate images to Lazada CDN
	migratedImages := s.migrateLazadaImages(client, data.Images)
	if len(migratedImages) == 0 {
		return "", fmt.Errorf("failed to process images - at least one image is required")
	}
	log.Info().Int("count", len(migratedImages)).Msg("Lazada images migrated")

	// Step 2: Get category - use provided or auto-recommend
	categoryID, err := s.getLazadaCategory(client, data)
	if err != nil {
		return "", err
	}

	// Step 3: Get required attributes for this category (optional, for logging)
	s.logLazadaCategoryAttributes(client, categoryID)

	// Step 4: Build SKUs
	skus := s.buildLazadaSKUs(data)

	// Step 5: Build XML payload and create product
	xmlPayload := s.buildLazadaProductXML(data, categoryID, migratedImages, skus)

	// Step 6: Log creation info
	s.logLazadaProductCreation(data, categoryID, migratedImages, skus)

	// Step 7: Call Lazada API
	result, err := client.CreateProductWithPayload(xmlPayload)
	if err != nil {
		return "", fmt.Errorf("lazada API error: %w", err)
	}

	if result.Code != "0" && result.Code != "" {
		return "", fmt.Errorf("lazada error: code=%s", result.Code)
	}

	itemID := result.Data.ItemID.String()
	if itemID != "" {
		log.Info().Str("item_id", itemID).Msg("Lazada product created successfully")
		return itemID, nil
	}

	return "lazada-" + time.Now().Format("20060102150405"), nil
}

// NOTE: Lazada helper methods are defined in clone_lazada_helpers.go
// NOTE: buildVariantsFromLazada is defined in clone_variant_helpers.go
