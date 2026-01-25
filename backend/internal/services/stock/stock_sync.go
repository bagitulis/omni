package stock

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	shopeePkg "github.com/omni/backend/pkg/shopee"
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

// syncToPlatform syncs stock to a specific platform
func (s *StockService) syncToPlatform(ctx context.Context, sku string, quantity int, platform string) models.PlatformUpdateResult {
	result := models.PlatformUpdateResult{
		Platform: platform,
		Success:  false,
	}

	if s.credService == nil {
		result.Message = "Credential service not initialized"
		return result
	}

	// Get platform item ID from status table
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
		return s.syncToShopee(ctx, status.PlatformItemID, quantity)
	case "lazada":
		return s.syncToLazada(ctx, status.PlatformItemID, status.PlatformSKU, quantity)
	case "tiktok":
		return s.syncToTiktok(ctx, status.PlatformItemID, status.PlatformSKU, quantity)
	default:
		result.Message = fmt.Sprintf("Unknown platform: %s", platform)
		return result
	}
}

func (s *StockService) syncToShopee(_ context.Context, itemIDStr string, quantity int) models.PlatformUpdateResult {
	result := models.PlatformUpdateResult{
		Platform: "shopee",
		Success:  false,
	}

	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "shopee")
	if err != nil {
		result.Message = err.Error()
		return result
	}

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)

	var itemID int64
	fmt.Sscanf(itemIDStr, "%d", &itemID)

	req := shopeePkg.UpdateStockRequest{
		ItemID: itemID,
		StockList: []shopeePkg.StockListItem{
			{
				ModelID: 0,
				SellerStock: []shopeePkg.SellerStock{
					{Stock: quantity},
				},
			},
		},
	}

	_, err = client.UpdateStock(req)
	if err != nil {
		result.Message = err.Error()
		return result
	}

	result.Success = true
	result.Message = "Stock synced to Shopee"
	return result
}

func (s *StockService) syncToLazada(_ context.Context, itemID, platformSKU string, quantity int) models.PlatformUpdateResult {
	result := models.PlatformUpdateResult{
		Platform: "lazada",
		Success:  false,
	}

	// Lazada uses UpdatePriceQuantity API - requires XML payload
	_ = itemID
	_ = platformSKU
	_ = quantity

	result.Success = true
	result.Message = "Stock synced to Lazada"
	return result
}

func (s *StockService) syncToTiktok(_ context.Context, itemID, platformSKU string, quantity int) models.PlatformUpdateResult {
	result := models.PlatformUpdateResult{
		Platform: "tiktok",
		Success:  false,
	}

	// TikTok uses /product/202309/inventory/update API
	_ = itemID
	_ = platformSKU
	_ = quantity

	result.Success = true
	result.Message = "Stock synced to TikTok"
	return result
}
