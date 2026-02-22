package stock

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// getPlatformStock retrieves platform stock info for a SKU
func (s *StockService) getPlatformStock(ctx context.Context, sku string) ([]models.PlatformStockInfo, error) {
	var statuses []models.InventorySkuPlatformStatus
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND sku = ?", s.tenantID, sku).
		Find(&statuses).Error
	if err != nil {
		return nil, err
	}

	platforms := make([]models.PlatformStockInfo, len(statuses))
	for i, st := range statuses {
		platforms[i] = models.PlatformStockInfo{
			Platform:       st.Platform,
			PlatformItemID: st.PlatformItemID,
			PlatformSKU:    st.PlatformSKU,
			Stock:          st.Stock,
			Status:         st.Status,
			LastSyncedAt:   st.LastCheckedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return platforms, nil
}

// syncToPlatform syncs stock to a specific platform via SDK
func (s *StockService) syncToPlatform(ctx context.Context, sku string, quantity int, platform string) models.PlatformUpdateResult {
	result := models.PlatformUpdateResult{
		Platform: platform,
		Success:  false,
	}

	if s.credService == nil {
		result.Message = "Credential service not initialized"
		return result
	}

	var status models.InventorySkuPlatformStatus
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND sku = ? AND platform = ?", s.tenantID, sku, platform).
		First(&status).Error
	if err != nil {
		result.Message = fmt.Sprintf("SKU not linked to %s", platform)
		return result
	}

	switch platform {
	case "shopee":
		return s.syncToShopee(status.PlatformItemID, quantity)
	case "lazada":
		return s.syncToLazada(status.PlatformItemID, status.PlatformSKU, quantity)
	case "tiktok":
		return s.syncToTiktok(status.PlatformItemID, status.PlatformSKU, quantity)
	default:
		result.Message = fmt.Sprintf("Unknown platform: %s", platform)
		return result
	}
}

func (s *StockService) syncToShopee(itemIDStr string, quantity int) models.PlatformUpdateResult {
	result := models.PlatformUpdateResult{Platform: "shopee", Success: false}

	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "shopee")
	if err != nil {
		result.Message = err.Error()
		return result
	}

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)

	var itemID int64
	fmt.Sscanf(itemIDStr, "%d", &itemID)

	resp, err := client.UpdateStock(shopeePkg.UpdateStockRequest{
		ItemID: itemID,
		StockList: []shopeePkg.StockListItem{
			{ModelID: 0, SellerStock: []shopeePkg.SellerStock{{Stock: quantity}}},
		},
	})
	if err != nil {
		result.Message = err.Error()
		return result
	}
	if resp.Error != "" {
		result.Message = fmt.Sprintf("Shopee API: %s - %s", resp.Error, resp.Message)
		return result
	}

	result.Success = true
	result.Message = "Stock synced to Shopee"
	return result
}

func (s *StockService) syncToLazada(itemID, platformSKU string, quantity int) models.PlatformUpdateResult {
	result := models.PlatformUpdateResult{Platform: "lazada", Success: false}

	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "lazada")
	if err != nil {
		result.Message = err.Error()
		return result
	}

	client := lazadaPkg.NewClient(creds.AppKey, creds.AppSecret, creds.Region)
	client.SetAccessToken(creds.AccessToken)

	resp, err := client.UpdatePriceQuantity(lazadaPkg.UpdatePriceQuantityRequest{
		ItemID:    itemID,
		SkuID:     platformSKU,
		SellerSku: platformSKU,
		Quantity:  quantity,
	})
	if err != nil {
		result.Message = err.Error()
		return result
	}
	if resp.Code != "0" {
		result.Message = fmt.Sprintf("Lazada API: %s - %s", resp.Code, resp.Message)
		return result
	}

	result.Success = true
	result.Message = "Stock synced to Lazada"
	return result
}

func (s *StockService) syncToTiktok(productID, skuID string, quantity int) models.PlatformUpdateResult {
	result := models.PlatformUpdateResult{Platform: "tiktok", Success: false}

	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "tiktok")
	if err != nil {
		result.Message = err.Error()
		return result
	}

	client := tiktokPkg.NewClient(creds.AppKey, creds.AppSecret)
	client.SetCredentials(creds.AccessToken, creds.ShopCipher)

	resp, err := client.UpdateInventory(tiktokPkg.UpdateInventoryRequest{
		ProductID: productID,
		Skus: []tiktokPkg.InventorySkuInfo{
			{ID: skuID, Inventory: []tiktokPkg.InventoryQuantity{{Quantity: quantity}}},
		},
	})
	if err != nil {
		result.Message = err.Error()
		return result
	}
	if resp.Code != 0 {
		result.Message = fmt.Sprintf("TikTok API: %d - %s", resp.Code, resp.Message)
		return result
	}

	result.Success = true
	result.Message = "Stock synced to TikTok"
	return result
}
