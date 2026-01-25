// Package inventory provides price update orchestration across platforms
package inventory

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/omni/backend/internal/services"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"gorm.io/gorm"
)

// PlatformPriceResult represents result of price update for a platform
type PlatformPriceResult struct {
	Success  bool    `json:"success"`
	Error    string  `json:"error,omitempty"`
	ItemID   string  `json:"item_id,omitempty"`
	ModelID  string  `json:"model_id,omitempty"`
	SkuID    string  `json:"sku_id,omitempty"`
	OldPrice float64 `json:"old_price,omitempty"`
	NewPrice float64 `json:"new_price,omitempty"`
}

// PriceUpdateOrchestratorResult represents the full price update result
type PriceUpdateOrchestratorResult struct {
	SKU       string                          `json:"sku"`
	Success   bool                            `json:"success"`
	Platforms map[string]*PlatformPriceResult `json:"platforms"`
	Errors    []string                        `json:"errors,omitempty"`
}

// PriceUpdateOrchestrator coordinates price updates across all platforms
type PriceUpdateOrchestrator struct {
	db          *gorm.DB
	tenantID    string
	credService *services.CredentialService
	idFetcher   *ProductIdFetcher
}

// NewPriceUpdateOrchestrator creates a new orchestrator
func NewPriceUpdateOrchestrator(db *gorm.DB, tenantID string, credService *services.CredentialService) *PriceUpdateOrchestrator {
	return &PriceUpdateOrchestrator{
		db:          db,
		tenantID:    tenantID,
		credService: credService,
		idFetcher:   NewProductIdFetcher(db, tenantID),
	}
}

// UpdatePrice updates price for a SKU across specified platforms
func (o *PriceUpdateOrchestrator) UpdatePrice(ctx context.Context, sku string, price float64, platforms []string) (*PriceUpdateOrchestratorResult, error) {
	result := &PriceUpdateOrchestratorResult{
		SKU:       sku,
		Success:   false,
		Platforms: make(map[string]*PlatformPriceResult),
		Errors:    []string{},
	}

	// Default to all platforms if none specified
	if len(platforms) == 0 {
		platforms = []string{"shopee", "lazada", "tiktok"}
	}

	// Fetch platform IDs
	platformIds, err := o.idFetcher.FetchBySku(ctx, sku)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch platform IDs: %w", err)
	}

	// Update each platform
	for _, platform := range platforms {
		switch platform {
		case "shopee":
			if platformIds.Shopee != nil {
				platformResult := o.updateShopeePrice(ctx, platformIds.Shopee, price)
				result.Platforms["shopee"] = platformResult
			} else {
				result.Platforms["shopee"] = &PlatformPriceResult{
					Success: false,
					Error:   "SKU not found in Shopee",
				}
			}

		case "lazada":
			if platformIds.Lazada != nil {
				platformResult := o.updateLazadaPrice(ctx, platformIds.Lazada, price)
				result.Platforms["lazada"] = platformResult
			} else {
				result.Platforms["lazada"] = &PlatformPriceResult{
					Success: false,
					Error:   "SKU not found in Lazada",
				}
			}

		case "tiktok":
			if platformIds.Tiktok != nil {
				platformResult := o.updateTiktokPrice(ctx, platformIds.Tiktok, price)
				result.Platforms["tiktok"] = platformResult
			} else {
				result.Platforms["tiktok"] = &PlatformPriceResult{
					Success: false,
					Error:   "SKU not found in TikTok",
				}
			}
		}
	}

	// Determine overall success
	successCount := 0
	for _, p := range result.Platforms {
		if p.Success {
			successCount++
		} else if p.Error != "" && !isNotFoundError(p.Error) {
			result.Errors = append(result.Errors, p.Error)
		}
	}

	// Success if at least one platform updated OR no platforms found (skip)
	result.Success = successCount > 0 || len(result.Errors) == 0

	return result, nil
}

// isNotFoundError checks if error is a "not found" error
func isNotFoundError(errMsg string) bool {
	return errMsg == "SKU not found in Shopee" ||
		errMsg == "SKU not found in Lazada" ||
		errMsg == "SKU not found in TikTok"
}

// updateShopeePrice updates price on Shopee platform
func (o *PriceUpdateOrchestrator) updateShopeePrice(_ context.Context, ids *ShopeeProductIds, price float64) *PlatformPriceResult {
	result := &PlatformPriceResult{
		Success:  false,
		ItemID:   ids.ItemID,
		ModelID:  ids.ModelID,
		NewPrice: price,
	}

	if o.credService == nil {
		result.Error = "Credential service not initialized"
		return result
	}

	creds, err := o.credService.GetPlatformCredentials(o.tenantID, "shopee")
	if err != nil {
		result.Error = fmt.Sprintf("Failed to get Shopee credentials: %s", err.Error())
		return result
	}

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)

	itemID, _ := strconv.ParseInt(ids.ItemID, 10, 64)
	var modelID int64
	if ids.ModelID != "" {
		modelID, _ = strconv.ParseInt(ids.ModelID, 10, 64)
	}

	// Shopee UpdatePrice API format
	req := shopeePkg.UpdatePriceRequest{
		ItemID: itemID,
		PriceList: []shopeePkg.PriceInfo{
			{
				ModelID:       modelID,
				OriginalPrice: price,
			},
		},
	}

	resp, err := client.UpdatePrice(req)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	if resp.Error != "" {
		result.Error = fmt.Sprintf("%s: %s", resp.Error, resp.Message)
		return result
	}

	result.Success = true
	log.Printf("[PriceOrchestrator] ✅ Shopee price updated: item_id=%s, model_id=%s, price=%.2f",
		ids.ItemID, ids.ModelID, price)
	return result
}

// updateLazadaPrice updates price on Lazada platform
func (o *PriceUpdateOrchestrator) updateLazadaPrice(_ context.Context, ids *LazadaProductIds, price float64) *PlatformPriceResult {
	result := &PlatformPriceResult{
		Success:  false,
		ItemID:   ids.ItemID,
		SkuID:    ids.SkuID,
		NewPrice: price,
	}

	if o.credService == nil {
		result.Error = "Credential service not initialized"
		return result
	}

	creds, err := o.credService.GetPlatformCredentials(o.tenantID, "lazada")
	if err != nil {
		result.Error = fmt.Sprintf("Failed to get Lazada credentials: %s", err.Error())
		return result
	}

	client := lazadaPkg.NewClient(creds.AppKey, creds.AppSecret, creds.Region)
	client.SetAccessToken(creds.AccessToken)

	// Lazada UpdatePrice via UpdatePriceQuantity API with price field
	req := lazadaPkg.UpdatePriceRequest{
		ItemID:    ids.ItemID,
		SkuID:     ids.SkuID,
		SellerSku: ids.SellerSku,
		Price:     price,
	}

	resp, err := client.UpdatePrice(req)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	if resp.Code != "0" {
		result.Error = fmt.Sprintf("code=%s: %s", resp.Code, resp.Message)
		return result
	}

	result.Success = true
	log.Printf("[PriceOrchestrator] ✅ Lazada price updated: item_id=%s, sku_id=%s, price=%.2f",
		ids.ItemID, ids.SkuID, price)
	return result
}

// updateTiktokPrice updates price on TikTok platform
func (o *PriceUpdateOrchestrator) updateTiktokPrice(_ context.Context, ids *TiktokProductIds, price float64) *PlatformPriceResult {
	result := &PlatformPriceResult{
		Success:  false,
		ItemID:   ids.ProductID,
		SkuID:    ids.SkuID,
		NewPrice: price,
	}

	if o.credService == nil {
		result.Error = "Credential service not initialized"
		return result
	}

	creds, err := o.credService.GetPlatformCredentials(o.tenantID, "tiktok")
	if err != nil {
		result.Error = fmt.Sprintf("Failed to get TikTok credentials: %s", err.Error())
		return result
	}

	client := tiktokPkg.NewClient(creds.AppKey, creds.AppSecret)
	client.SetCredentials(creds.AccessToken, creds.ShopCipher)

	// TikTok price is in cents, convert
	priceStr := fmt.Sprintf("%.0f", price*100)

	req := tiktokPkg.UpdateProductRequest{
		ProductID: ids.ProductID,
		Skus: []tiktokPkg.UpdateProductSku{
			{
				ID:            ids.SkuID,
				OriginalPrice: priceStr,
			},
		},
	}

	resp, err := client.UpdateProduct(req)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	if resp.Code != 0 {
		result.Error = fmt.Sprintf("code=%d: %s", resp.Code, resp.Message)
		return result
	}

	result.Success = true
	log.Printf("[PriceOrchestrator] ✅ TikTok price updated: product_id=%s, sku_id=%s, price=%.2f",
		ids.ProductID, ids.SkuID, price)
	return result
}
