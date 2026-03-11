package inventory

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"strconv"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"gorm.io/gorm"
)

// PlatformStockResult represents result of stock update for a platform
type PlatformStockResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	ItemID  string `json:"item_id,omitempty"`
	ModelID string `json:"model_id,omitempty"`
	SkuID   string `json:"sku_id,omitempty"`
}

// StockUpdateOrchestratorResult represents the full stock update result
type StockUpdateOrchestratorResult struct {
	SKU       string                          `json:"sku"`
	Success   bool                            `json:"success"`
	Platforms map[string]*PlatformStockResult `json:"platforms"`
	Errors    []string                        `json:"errors,omitempty"`
}

// StockUpdateOrchestrator coordinates stock updates across all platforms
type StockUpdateOrchestrator struct {
	db          *gorm.DB
	tenantID    string
	credService *services.CredentialService
	idFetcher   *ProductIdFetcher
}

// NewStockUpdateOrchestrator creates a new orchestrator
func NewStockUpdateOrchestrator(db *gorm.DB, tenantID string, credService *services.CredentialService) *StockUpdateOrchestrator {
	return &StockUpdateOrchestrator{
		db:          db,
		tenantID:    tenantID,
		credService: credService,
		idFetcher:   NewProductIdFetcher(db, tenantID),
	}
}

// UpdateStock updates stock for a SKU across specified platforms
func (o *StockUpdateOrchestrator) UpdateStock(ctx context.Context, sku string, stock int, platforms []string) (*StockUpdateOrchestratorResult, error) {
	result := &StockUpdateOrchestratorResult{
		SKU:       sku,
		Success:   false,
		Platforms: make(map[string]*PlatformStockResult),
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
				platformResult := o.updateShopeeStock(ctx, platformIds.Shopee, stock)
				result.Platforms["shopee"] = platformResult
			} else {
				result.Platforms["shopee"] = &PlatformStockResult{
					Success: false,
					Error:   "SKU not found in Shopee",
				}
			}

		case "lazada":
			if platformIds.Lazada != nil {
				platformResult := o.updateLazadaStock(ctx, platformIds.Lazada, stock)
				result.Platforms["lazada"] = platformResult
			} else {
				result.Platforms["lazada"] = &PlatformStockResult{
					Success: false,
					Error:   "SKU not found in Lazada",
				}
			}

		case "tiktok":
			if platformIds.Tiktok != nil {
				platformResult := o.updateTiktokStock(ctx, platformIds.Tiktok, stock)
				result.Platforms["tiktok"] = platformResult
			} else {
				result.Platforms["tiktok"] = &PlatformStockResult{
					Success: false,
					Error:   "SKU not found in TikTok",
				}
			}
		default:
			result.Platforms[platform] = &PlatformStockResult{
				Success: false,
				Error:   fmt.Sprintf("unsupported platform: %s", platform),
			}
		}
	}

	result.Success, result.Errors = summarizeStockResults(result.Platforms)

	return result, nil
}

func summarizeStockResults(platformResults map[string]*PlatformStockResult) (bool, []string) {
	errors := make([]string, 0)
	successCount := 0

	for _, platformResult := range platformResults {
		if platformResult == nil {
			continue
		}

		if platformResult.Success {
			successCount++
			continue
		}

		if platformResult.Error != "" && !isStockNotFoundError(platformResult.Error) {
			errors = append(errors, platformResult.Error)
		}
	}

	return successCount > 0, errors
}

func isStockNotFoundError(errMsg string) bool {
	return errMsg == "SKU not found in Shopee" ||
		errMsg == "SKU not found in Lazada" ||
		errMsg == "SKU not found in TikTok"
}

// updateShopeeStock updates stock on Shopee platform
func (o *StockUpdateOrchestrator) updateShopeeStock(_ context.Context, ids *ShopeeProductIds, stock int) *PlatformStockResult {
	result := &PlatformStockResult{
		Success: false,
		ItemID:  ids.ItemID,
		ModelID: ids.ModelID,
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

	// Shopee UpdateStock API format: { item_id, stock_list: [{ model_id, seller_stock: [{ stock }] }] }
	req := shopeePkg.UpdateStockRequest{
		ItemID: itemID,
		StockList: []shopeePkg.StockListItem{
			{
				ModelID: modelID,
				SellerStock: []shopeePkg.SellerStock{
					{Stock: stock},
				},
			},
		},
	}

	resp, err := client.UpdateStock(req)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	if resp.Error != "" {
		result.Error = rawShopeeAPIError(resp.Error, resp.Message)
		return result
	}

	result.Success = true
	log.Info().Msgf("[StockOrchestrator] ✅ Shopee stock updated: item_id=%s, model_id=%s, stock=%d",
		ids.ItemID, ids.ModelID, stock)
	return result
}

// updateLazadaStock updates stock on Lazada platform
func (o *StockUpdateOrchestrator) updateLazadaStock(_ context.Context, ids *LazadaProductIds, stock int) *PlatformStockResult {
	result := &PlatformStockResult{
		Success: false,
		ItemID:  ids.ItemID,
		SkuID:   ids.SkuID,
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

	req := lazadaPkg.UpdatePriceQuantityRequest{
		ItemID:    ids.ItemID,
		SkuID:     ids.SkuID,
		SellerSku: ids.SellerSku,
		Quantity:  stock,
	}

	resp, err := client.UpdatePriceQuantity(req)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	if resp.Code != "0" {
		result.Error = rawLazadaAPIError(resp.Code, resp.Message)
		return result
	}

	result.Success = true
	log.Info().Msgf("[StockOrchestrator] ✅ Lazada stock updated: item_id=%s, sku_id=%s, stock=%d",
		ids.ItemID, ids.SkuID, stock)
	return result
}

// updateTiktokStock updates stock on TikTok platform
func (o *StockUpdateOrchestrator) updateTiktokStock(_ context.Context, ids *TiktokProductIds, stock int) *PlatformStockResult {
	result := &PlatformStockResult{
		Success: false,
		ItemID:  ids.ProductID,
		SkuID:   ids.SkuID,
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

	// TikTok inventory update format (no warehouse_id needed for simple update)
	req := tiktokPkg.UpdateInventoryRequest{
		ProductID: ids.ProductID,
		Skus: []tiktokPkg.InventorySkuInfo{
			{
				ID: ids.SkuID,
				Inventory: []tiktokPkg.InventoryQuantity{
					{Quantity: stock},
				},
			},
		},
	}

	resp, err := client.UpdateInventory(req)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	if resp.Code != 0 {
		result.Error = rawTiktokAPIError(resp.Code, resp.Message)
		return result
	}

	result.Success = true
	log.Info().Msgf("[StockOrchestrator] ✅ TikTok stock updated: product_id=%s, sku_id=%s, stock=%d",
		ids.ProductID, ids.SkuID, stock)
	return result
}

// GetPlatformConfig returns platform config from database
func (o *StockUpdateOrchestrator) GetPlatformConfig(ctx context.Context, platform string) (*models.PlatformConfig, error) {
	var config models.PlatformConfig
	err := o.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", o.tenantID, platform).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}
