package master_product

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
)

// PreviewFromShopee previews what will be imported from Shopee without creating
func (s *ImportService) PreviewFromShopee(ctx context.Context, tenantID string, shopeeItemID int64) (*PreviewResult, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	// Check if already imported
	existingLinks, _ := s.repo.FindPlatformLinksByItemID(ctx, tenantID, "shopee", strconv.FormatInt(shopeeItemID, 10))
	alreadyExists := len(existingLinks) > 0

	// Get Shopee client
	client, err := s.getShopeeClient(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Fetch product details
	productResp, err := client.GetProductDetailWithImages([]int64{shopeeItemID})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shopee product: %w", err)
	}

	if len(productResp.Response.ItemList) == 0 {
		return nil, ErrShopeeProductNotFound
	}

	product := productResp.Response.ItemList[0]

	// Fetch models/SKUs
	modelResp, err := client.GetModelList(shopeeItemID)
	if err != nil {
		log.Warn().Err(err).Int64("item_id", shopeeItemID).Msg("Failed to fetch models, product may have no variants")
	}

	result := &PreviewResult{
		Title:         truncateString(product.ItemName, models.MasterProductMaxTitleLength),
		Description:   truncateString(product.Description, models.MasterProductMaxDescriptionLength),
		Images:        limitImages(product.Image.ImageURLList, models.MasterProductMaxImages),
		SKUs:          []PreviewSku{},
		AlreadyExists: alreadyExists,
	}

	// Process models
	if modelResp != nil && len(modelResp.Response.Model) > 0 {
		for _, model := range modelResp.Response.Model {
			sku := PreviewSku{
				SellerSku:   model.ModelSKU,
				VariantName: buildVariantName(modelResp.Response.TierVariation, model.TierIndex),
				Stock:       model.StockInfoV2.SummaryInfo.TotalAvailableStock,
				Valid:       true,
			}

			// Get price
			if len(model.PriceInfo) > 0 {
				sku.Price = model.PriceInfo[0].CurrentPrice
			}

			// Validate seller_sku
			if model.ModelSKU == "" {
				sku.Valid = false
				sku.SkipReason = "missing seller_sku"
				result.SkippedSKUs++
			} else {
				result.ValidSKUs++
			}

			result.SKUs = append(result.SKUs, sku)
			result.TotalSKUs++
		}
	}

	return result, nil
}

// ImportFromShopee imports a product from Shopee to Master Product
func (s *ImportService) ImportFromShopee(ctx context.Context, tenantID string, shopeeItemID int64) (*ImportResult, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	log.Info().
		Str("tenant_id", tenantID).
		Int64("shopee_item_id", shopeeItemID).
		Msg("Starting Shopee import")

	// Check if already imported
	existingLinks, _ := s.repo.FindPlatformLinksByItemID(ctx, tenantID, "shopee", strconv.FormatInt(shopeeItemID, 10))
	if len(existingLinks) > 0 {
		return nil, ErrProductAlreadyExists
	}

	// Get Shopee client
	client, err := s.getShopeeClient(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Fetch product details
	productResp, err := client.GetProductDetailWithImages([]int64{shopeeItemID})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shopee product: %w", err)
	}

	if len(productResp.Response.ItemList) == 0 {
		return nil, ErrShopeeProductNotFound
	}

	product := productResp.Response.ItemList[0]

	// Fetch models/SKUs
	modelResp, err := client.GetModelList(shopeeItemID)
	if err != nil {
		log.Warn().Err(err).Int64("item_id", shopeeItemID).Msg("Failed to fetch models")
	}

	// Build images as JSONArray (list of URLs)
	imgList := limitImages(product.Image.ImageURLList, models.MasterProductMaxImages)
	images := make(models.JSONArray, len(imgList))
	for i, imgURL := range imgList {
		images[i] = imgURL
	}

	// Create master product
	masterProduct := &models.MasterProduct{
		TenantID:    tenantID,
		Title:       truncateString(product.ItemName, models.MasterProductMaxTitleLength),
		Description: truncateString(product.Description, models.MasterProductMaxDescriptionLength),
		Images:      images,
		Status:      models.MasterProductStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, masterProduct); err != nil {
		return nil, fmt.Errorf("failed to create master product: %w", err)
	}

	result := &ImportResult{
		MasterProduct: masterProduct,
	}

	// Create product-level platform link
	productLink := &models.MasterProductPlatformLink{
		MasterProductID:   masterProduct.ID,
		Platform:          models.PlatformShopee,
		PlatformItemID:    strconv.FormatInt(shopeeItemID, 10),
		PlatformProductID: strconv.FormatInt(shopeeItemID, 10),
		SyncStatus:        models.SyncStatusSynced,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	now := time.Now()
	productLink.LastSyncedAt = &now

	if err := s.repo.UpsertPlatformLink(ctx, productLink); err != nil {
		log.Error().Err(err).Msg("Failed to upsert product platform link")
	} else {
		result.PlatformLinks++
	}

	// Process models/SKUs
	if modelResp != nil && len(modelResp.Response.Model) > 0 {
		for _, model := range modelResp.Response.Model {
			// Skip SKUs without seller_sku
			if model.ModelSKU == "" {
				log.Warn().
					Int64("item_id", shopeeItemID).
					Int64("model_id", model.ModelID).
					Msg("Skipping SKU without seller_sku")
				result.SkusSkipped++
				continue
			}

			// Get price
			var price float64
			if len(model.PriceInfo) > 0 {
				price = model.PriceInfo[0].CurrentPrice
			}

			// Build variant data
			variantData := make(models.JSONMap)
			if modelResp.Response.TierVariation != nil {
				for i, tierIdx := range model.TierIndex {
					if i < len(modelResp.Response.TierVariation) {
						tier := modelResp.Response.TierVariation[i]
						if tierIdx < len(tier.OptionList) {
							variantData[tier.Name] = tier.OptionList[tierIdx].Option
						}
					}
				}
			}

			// Create SKU
			sku := &models.MasterProductSku{
				TenantID:        tenantID,
				MasterProductID: masterProduct.ID,
				SellerSku:       model.ModelSKU,
				VariantName:     buildVariantName(modelResp.Response.TierVariation, model.TierIndex),
				VariantData:     variantData,
				Price:           price,
				Stock:           model.StockInfoV2.SummaryInfo.TotalAvailableStock,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			if err := s.repo.CreateSku(ctx, sku); err != nil {
				log.Error().Err(err).Str("seller_sku", model.ModelSKU).Msg("Failed to create SKU")
				result.SkusSkipped++
				continue
			}

			result.SkusImported++

			// Create SKU-level platform link
			skuLink := &models.MasterProductPlatformLink{
				MasterProductID: masterProduct.ID,
				MasterSkuID:     &sku.ID,
				Platform:        models.PlatformShopee,
				PlatformItemID:  strconv.FormatInt(shopeeItemID, 10),
				PlatformSkuID:   strconv.FormatInt(model.ModelID, 10),
				SyncStatus:      models.SyncStatusSynced,
				LastSyncedAt:    &now,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			if err := s.repo.UpsertPlatformLink(ctx, skuLink); err != nil {
				log.Error().Err(err).Msg("Failed to upsert SKU platform link")
			} else {
				result.PlatformLinks++
			}
		}
	}

	// Create default SKU for single-variant products (no models)
	if result.SkusImported == 0 {
		log.Info().
			Str("tenant_id", tenantID).
			Int64("shopee_item_id", shopeeItemID).
			Msg("Creating default SKU for product without variants")

		// Get price from product's price_info if available
		var price float64
		if len(product.PriceInfo) > 0 {
			price = product.PriceInfo[0].CurrentPrice
		}

		// Get stock from product's stock_info_v2 if available
		stock := product.StockInfoV2.SummaryInfo.TotalAvailableStock

		// Create default SKU with item_id as seller_sku
		defaultSku := &models.MasterProductSku{
			TenantID:        tenantID,
			MasterProductID: masterProduct.ID,
			SellerSku:       strconv.FormatInt(shopeeItemID, 10),
			VariantName:     "",
			VariantData:     make(models.JSONMap),
			Price:           price,
			Stock:           stock,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if err := s.repo.CreateSku(ctx, defaultSku); err != nil {
			log.Error().Err(err).Msg("Failed to create default SKU")
		} else {
			result.SkusImported++

			// Create platform link for default SKU
			defaultSkuLink := &models.MasterProductPlatformLink{
				MasterProductID: masterProduct.ID,
				MasterSkuID:     &defaultSku.ID,
				Platform:        models.PlatformShopee,
				PlatformItemID:  strconv.FormatInt(shopeeItemID, 10),
				SyncStatus:      models.SyncStatusSynced,
				LastSyncedAt:    &now,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			if err := s.repo.CreatePlatformLink(ctx, defaultSkuLink); err != nil {
				log.Error().Err(err).Msg("Failed to create default SKU platform link")
			} else {
				result.PlatformLinks++
			}
		}
	}

	// Fetch complete product with relations
	result.MasterProduct, _ = s.repo.FindByID(ctx, masterProduct.ID)

	log.Info().
		Str("tenant_id", tenantID).
		Uint("master_product_id", masterProduct.ID).
		Int("skus_imported", result.SkusImported).
		Int("skus_skipped", result.SkusSkipped).
		Msg("Shopee import completed")

	return result, nil
}

// getShopeeClient creates a Shopee API client for the tenant
func (s *ImportService) getShopeeClient(ctx context.Context, tenantID string) (*shopee.Client, error) {
	// Get tenant database
	tenantDB, err := config.GetTenantDB(tenantID, s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	// Get system database for global credentials
	systemDB, err := config.GetSystemDB(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get system database: %w", err)
	}

	// Get tenant credentials
	credRepo := repositories.NewPlatformCredentialsRepository(tenantDB)
	tenantCreds, err := credRepo.GetShopeeCredentials(ctx)
	if err != nil {
		return nil, ErrShopeeNotConfigured
	}

	if tenantCreds.ShopIDInt == 0 || tenantCreds.AccessToken == "" {
		return nil, ErrShopeeNotConfigured
	}

	// Get global credentials
	configRepo := repositories.NewGlobalConfigRepository(systemDB)
	globalCreds, err := configRepo.GetShopeeCredentials(ctx)
	if err != nil || globalCreds.PartnerID == 0 {
		return nil, ErrShopeeNotConfigured
	}

	// Create client
	client := shopee.NewClient(globalCreds.PartnerID, globalCreds.PartnerKey, true)
	client.SetShopCredentials(tenantCreds.ShopIDInt, tenantCreds.AccessToken)

	return client, nil
}
