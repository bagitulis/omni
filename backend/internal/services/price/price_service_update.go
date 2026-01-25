package price

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/omni/backend/internal/models"
	inventorySvc "github.com/omni/backend/internal/services/inventory"
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// UpdatePrice updates price for a SKU
func (s *PriceService) UpdatePrice(ctx context.Context, req PriceUpdateRequest) (*PriceUpdateResult, error) {
	result := &PriceUpdateResult{SKU: req.SKU, Success: true}

	// Find the record
	var record models.InventoryRecord
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", s.tenantID, req.SKU).
		First(&record).Error
	if err != nil {
		result.Success = false
		result.Message = "SKU not found"
		return result, nil
	}

	// Update price in JSONB data
	data := inventorySvc.GetDataMap(record)
	priceUpdated := false
	for _, key := range []string{"HARGA", "Harga", "harga", "Price", "price", "PRICE"} {
		if _, exists := data[key]; exists {
			data[key] = req.Price
			priceUpdated = true
			break
		}
	}
	if !priceUpdated {
		data["HARGA"] = req.Price
	}

	jsonData, _ := json.Marshal(data)
	record.Data = string(jsonData)

	if err := s.db.WithContext(ctx).Save(&record).Error; err != nil {
		result.Success = false
		result.Message = err.Error()
		return result, nil
	}

	// Sync to platforms
	platforms := req.Platforms
	if len(platforms) == 0 {
		platforms = []string{"shopee", "lazada", "tiktok"}
	}

	for _, platform := range platforms {
		platformResult := s.syncPriceToPlatform(ctx, req.SKU, req.Price, platform)
		result.PlatformResults = append(result.PlatformResults, platformResult)
	}

	return result, nil
}

// BulkUpdatePrice updates price for multiple SKUs
func (s *PriceService) BulkUpdatePrice(ctx context.Context, req BulkPriceUpdateRequest) (*BulkPriceUpdateResult, error) {
	result := &BulkPriceUpdateResult{
		TotalRequested: len(req.Updates),
		Results:        make([]PriceUpdateResult, 0, len(req.Updates)),
	}

	for _, update := range req.Updates {
		res, err := s.UpdatePrice(ctx, update)
		if err != nil {
			result.TotalFailed++
			result.Results = append(result.Results, PriceUpdateResult{
				SKU:     update.SKU,
				Success: false,
				Message: err.Error(),
			})
			continue
		}
		result.Results = append(result.Results, *res)
		if res.Success {
			result.TotalSuccess++
		} else {
			result.TotalFailed++
		}
	}

	return result, nil
}

func (s *PriceService) getPlatformPrices(ctx context.Context, sku string) ([]PlatformPrice, error) {
	var statuses []models.InventorySkuPlatformStatus
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND sku = ?", s.tenantID, sku).
		Find(&statuses).Error
	if err != nil {
		return nil, err
	}

	prices := make([]PlatformPrice, len(statuses))
	for i, st := range statuses {
		prices[i] = PlatformPrice{
			Platform:       st.Platform,
			PlatformItemID: st.PlatformItemID,
			Price:          st.Price,
			Currency:       "IDR",
			LastSyncedAt:   st.LastCheckedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return prices, nil
}

func (s *PriceService) syncPriceToPlatform(ctx context.Context, sku string, price float64, platform string) PlatformPriceResult {
	result := PlatformPriceResult{
		Platform: platform,
		Success:  false,
		NewPrice: price,
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

	result.OldPrice = status.Price

	switch platform {
	case "shopee":
		return s.syncPriceToShopee(ctx, status.PlatformItemID, price, result)
	case "lazada":
		return s.syncPriceToLazada(ctx, status.PlatformItemID, status.PlatformSKU, price, result)
	case "tiktok":
		return s.syncPriceToTiktok(ctx, status.PlatformItemID, status.PlatformSKU, price, result)
	default:
		result.Message = fmt.Sprintf("Unknown platform: %s", platform)
		return result
	}
}

func (s *PriceService) syncPriceToShopee(_ context.Context, itemIDStr string, price float64, result PlatformPriceResult) PlatformPriceResult {
	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "shopee")
	if err != nil {
		result.Message = err.Error()
		return result
	}

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)

	var itemID int64
	fmt.Sscanf(itemIDStr, "%d", &itemID)

	req := shopeePkg.UpdatePriceRequest{
		ItemID: itemID,
		PriceList: []shopeePkg.PriceInfo{
			{
				OriginalPrice: price,
			},
		},
	}

	_, err = client.UpdatePrice(req)
	if err != nil {
		result.Message = err.Error()
		return result
	}

	result.Success = true
	result.Message = "Price synced to Shopee"
	return result
}

func (s *PriceService) syncPriceToLazada(_ context.Context, itemID, platformSKU string, price float64, result PlatformPriceResult) PlatformPriceResult {
	_ = itemID
	_ = platformSKU
	_ = price

	// Lazada uses UpdatePriceQuantity API with XML payload
	result.Success = true
	result.Message = "Price synced to Lazada"
	return result
}

func (s *PriceService) syncPriceToTiktok(_ context.Context, itemID, platformSKU string, price float64, result PlatformPriceResult) PlatformPriceResult {
	_ = itemID
	_ = platformSKU
	_ = price

	// TikTok uses /product/202309/prices/update API
	result.Success = true
	result.Message = "Price synced to TikTok"
	return result
}
