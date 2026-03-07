// Package products provides product cloning between platforms
package products

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/services/inventory"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CloneService handles product cloning between platforms
type CloneService struct {
	db          *gorm.DB
	tenantID    string
	dbPath      string
	credService *services.CredentialService
	httpClient  *http.Client
	idFetcher   *inventory.ProductIdFetcher
}

// NewCloneService creates a new clone service
func NewCloneService(db *gorm.DB, tenantID string) *CloneService {
	return &CloneService{
		db:         db,
		tenantID:   tenantID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		idFetcher:  inventory.NewProductIdFetcher(db, tenantID),
	}
}

// NewCloneServiceWithCreds creates a clone service with credential support
func NewCloneServiceWithCreds(db *gorm.DB, tenantID, dbPath string) *CloneService {
	return &CloneService{
		db:          db,
		tenantID:    tenantID,
		dbPath:      dbPath,
		credService: services.NewCredentialService(dbPath),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		idFetcher:   inventory.NewProductIdFetcher(db, tenantID),
	}
}

// Clone clones a product from one platform to another
func (s *CloneService) Clone(ctx context.Context, req CloneRequest) (*CloneResult, error) {
	result := &CloneResult{
		ID:             generateCloneID(),
		SourcePlatform: req.SourcePlatform,
		TargetPlatform: req.TargetPlatform,
		SourceItemID:   req.SourceItemID,
		Status:         "in_progress",
		Progress:       0,
		StartedAt:      time.Now(),
	}

	// 1. Fetch product data from source platform (database)
	result.Progress = 10
	productData, err := s.fetchProductData(ctx, req.SourcePlatform, req.SourceItemID)
	if err != nil {
		return s.failResult(result, "Failed to fetch product: %v", err)
	}

	// 2. Override price/stock from inventory if enabled (default: true)
	result.Progress = 20
	if req.SKU != "" {
		invData, err := s.applyInventoryData(ctx, req.SKU, productData, req.UseInventory)
		if err != nil {
			// Log warning but don't fail - use platform data as fallback
			log.Warn().
				Err(err).
				Str("service", "clone").
				Str("sku", req.SKU).
				Msg("Failed to get inventory data, using platform data as fallback")
		} else if invData != nil && invData.Found {
			log.Info().
				Str("service", "clone").
				Str("sku", req.SKU).
				Float64("price", invData.Price).
				Int("stock", invData.Stock).
				Msg("Applied inventory data for clone")
		}
	}

	// 3. Map category if needed
	result.Progress = 30
	targetCategoryID := req.CategoryID
	if targetCategoryID == "" {
		targetCategoryID, err = s.mapCategory(ctx, req.SourcePlatform, req.TargetPlatform, productData.CategoryID)
		if err != nil {
			return s.failResult(result, "Failed to map category: %v", err)
		}
	}
	productData.CategoryID = targetCategoryID
	productData.SaveAsDraft = req.SaveAsDraft // Pass saveAsDraft option to createProduct

	// 4. Download and re-upload images
	result.Progress = 50
	imageURLs, err := s.processImages(ctx, req.TargetPlatform, productData.Images)
	if err != nil {
		return s.failResult(result, "Failed to process images: %v", err)
	}
	productData.Images = imageURLs

	// 5. Update price if explicitly requested (overrides inventory)
	if req.UpdatePrice && req.NewPrice > 0 {
		productData.Price = req.NewPrice
	}

	// Apply price adjustment logic (percentage, fixed amount, etc.)
	if req.PriceAdjustmentType != "" {
		switch req.PriceAdjustmentType {
		case "fixed":
			if req.PriceAdjustmentValue > 0 {
				productData.Price = req.PriceAdjustmentValue
			}
		case "percentage":
			// Increase/Decrease by percentage. e.g. 10 means +10%, -10 means -10%
			productData.Price = productData.Price * (1 + req.PriceAdjustmentValue/100)
		case "amount":
			// Increase/Decrease by fixed amount. e.g. 1000 means +1000
			productData.Price = productData.Price + req.PriceAdjustmentValue
		}
		// Ensure price is not negative
		if productData.Price < 0 {
			productData.Price = 0
		}
	}

	// 6. Create product on target platform
	result.Progress = 80
	targetItemID, err := s.createProduct(ctx, req.TargetPlatform, productData)
	if err != nil {
		return s.failResult(result, "Failed to create product: %v", err)
	}

	// 7. Trigger product sync to update database with new product from API
	result.Progress = 90
	syncResult := s.triggerProductSync(ctx, req.TargetPlatform)

	return s.successResultWithSync(result, targetItemID, syncResult)
}

// applyInventoryData fetches inventory and applies price/stock to productData
func (s *CloneService) applyInventoryData(ctx context.Context, sku string, productData *ProductData, useInventory bool) (*InventoryData, error) {
	// Default to using inventory data
	if !useInventory {
		return nil, nil
	}

	fetcher := NewInventoryFetcher(s.db, s.tenantID)
	invData, err := fetcher.GetInventoryBySellerSKU(ctx, sku)
	if err != nil {
		return nil, err
	}

	if invData.Found {
		// Override price and stock from inventory
		if invData.Price > 0 {
			productData.Price = invData.Price
		}
		if invData.Stock >= 0 {
			productData.Stock = invData.Stock
		}

		// Also update variants if present
		for i := range productData.Variants {
			if productData.Variants[i].SKU == sku {
				if invData.Price > 0 {
					productData.Variants[i].Price = invData.Price
				}
				if invData.Stock >= 0 {
					productData.Variants[i].Stock = invData.Stock
				}
			}
		}
	}

	return invData, nil
}

// BatchClone clones multiple products
func (s *CloneService) BatchClone(ctx context.Context, req BatchCloneRequest) (*BatchCloneResult, error) {
	result := &BatchCloneResult{
		BatchID:        generateCloneID(),
		TotalRequested: len(req.SourceItemIDs),
		Status:         "in_progress",
		Results:        make([]CloneResult, 0, len(req.SourceItemIDs)),
	}

	for _, itemID := range req.SourceItemIDs {
		cloneReq := CloneRequest{
			SourcePlatform:       req.SourcePlatform,
			TargetPlatform:       req.TargetPlatform,
			SourceItemID:         itemID,
			CategoryID:           req.CategoryID,
			PriceAdjustmentType:  req.PriceAdjustmentType,
			PriceAdjustmentValue: req.PriceAdjustmentValue,
			SaveAsDraft:          req.SaveAsDraft,
			UseInventory:         true,
		}

		cloneResult, err := s.Clone(ctx, cloneReq)
		if err != nil {
			result.TotalFailed++
			result.Results = append(result.Results, CloneResult{
				SourcePlatform: req.SourcePlatform,
				SourceItemID:   itemID,
				Status:         "failed",
				Message:        err.Error(),
			})
			continue
		}

		result.Results = append(result.Results, *cloneResult)
		if cloneResult.Status == "completed" {
			result.TotalSuccess++
		} else {
			result.TotalFailed++
		}
	}

	result.Status = s.determineBatchStatus(result.TotalFailed, result.TotalSuccess)
	return result, nil
}

// GetCloneStatus gets the status of a clone operation
func (s *CloneService) GetCloneStatus(_ context.Context, cloneID string) (*CloneResult, error) {
	return &CloneResult{
		ID:     cloneID,
		Status: "completed",
	}, nil
}

// GetProductData fetches product data from a platform by SKU
func (s *CloneService) GetProductData(ctx context.Context, platform, sku string) (*ProductData, error) {
	return s.fetchProductData(ctx, platform, sku)
}

// GetAvailableTargets returns available clone target platforms for a SKU
// Uses ProductIdFetcher to find platform IDs across different tables
func (s *CloneService) GetAvailableTargets(ctx context.Context, sku string) (*CloneTargetsResult, error) {
	result := &CloneTargetsResult{
		SKU: sku,
		Status: PlatformStatus{
			Shopee: false,
			Lazada: false,
			TikTok: false,
		},
		Sources: []string{},
		Targets: []string{},
	}

	// Use ProductIdFetcher to find platform IDs (more robust)
	platformIds, err := s.idFetcher.FetchBySku(ctx, sku)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch platform IDs: %w", err)
	}

	// Check which platforms have this SKU
	if platformIds.Shopee != nil {
		result.Status.Shopee = true
		result.Sources = append(result.Sources, "shopee")
	} else {
		result.Targets = append(result.Targets, "shopee")
	}

	if platformIds.Lazada != nil {
		result.Status.Lazada = true
		result.Sources = append(result.Sources, "lazada")
	} else {
		result.Targets = append(result.Targets, "lazada")
	}

	if platformIds.Tiktok != nil {
		result.Status.TikTok = true
		result.Sources = append(result.Sources, "tiktok")
	} else {
		result.Targets = append(result.Targets, "tiktok")
	}

	return result, nil
}

// checkSkuOnPlatform checks if SKU exists on a platform
// Uses ProductIdFetcher for more robust detection
func (s *CloneService) checkSkuOnPlatform(ctx context.Context, platform, sku string) bool {
	platformIds, err := s.idFetcher.FetchBySku(ctx, sku)
	if err != nil {
		return false
	}

	switch platform {
	case "shopee":
		return platformIds.Shopee != nil
	case "lazada":
		return platformIds.Lazada != nil
	case "tiktok":
		return platformIds.Tiktok != nil
	}

	return false
}
