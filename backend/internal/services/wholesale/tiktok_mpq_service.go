package wholesale

import (
	"context"
	"fmt"
	"strings"

	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// TiktokProductAPI defines the interface for TikTok API calls needed by MPQ service
// This avoids circular dependency with platform package
type TiktokProductAPI interface {
	// Request makes an HTTP request and returns parsed response
	Request(method, path string, queryParams map[string]string, body interface{}) (map[string]interface{}, error)
}

// TiktokMpqService handles TikTok MPQ operations
// Ported from Node.js: tiktokMpqService.ts
// TikTok does NOT support wholesale tiers, only MPQ + price
type TiktokMpqService struct {
	db         *gorm.DB
	tenantID   string
	tiktokAPI  TiktokProductAPI
	apiVersion string
}

// NewTiktokMpqService creates a new TikTok MPQ service
func NewTiktokMpqService(db *gorm.DB, tenantID string, tiktokAPI TiktokProductAPI) *TiktokMpqService {
	return &TiktokMpqService{
		db:         db,
		tenantID:   tenantID,
		tiktokAPI:  tiktokAPI,
		apiVersion: "202309",
	}
}

// tiktokProductLookup holds lookup result for TikTok product by SKU
type tiktokProductLookup struct {
	ProductID   string // TikTok API product ID (string)
	SkuID       string
	SellerSku   string
	ProductName string
	Price       float64
}

// TiktokMpqResult represents a single TikTok MPQ operation result
type TiktokMpqResult struct {
	ProductID string `json:"product_id"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}

// TiktokBatchMpqResult represents batch TikTok MPQ result
type TiktokBatchMpqResult struct {
	TotalSKUs      int               `json:"total_skus"`
	UniqueProducts int               `json:"unique_products"`
	Processed      int               `json:"processed"`
	Failed         int               `json:"failed"`
	Skipped        []string          `json:"skipped"`
	Success        bool              `json:"success"`
	Results        []TiktokMpqResult `json:"results"`
}

func parseFallbackSku(sku string) (string, string) {
	// Format: tiktok_{productID}_{skuID}
	if !strings.HasPrefix(sku, "tiktok_") {
		return "", ""
	}
	parts := strings.Split(sku, "_")
	if len(parts) < 3 {
		return "", ""
	}
	return parts[1], parts[2]
}

// lookupProductBySku finds TikTok product by seller_sku or fallback pattern
func (s *TiktokMpqService) lookupProductBySku(ctx context.Context, sku string) (*tiktokProductLookup, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not configured")
	}

	var skuRecord models.TiktokSku
	err := s.db.WithContext(ctx).
		Where("seller_sku = ?", sku).
		First(&skuRecord).Error

	if err == gorm.ErrRecordNotFound {
		err = s.db.WithContext(ctx).
			Where("sku_id = ?", sku).
			First(&skuRecord).Error
	}

	// Handle fallback SKU if not found in platform mapping
	if err == gorm.ErrRecordNotFound {
		prodID, skuID := parseFallbackSku(sku)
		if prodID != "" && skuID != "" {
			// Try to find product name from DB even if SKU record is missing
			var product models.TiktokProduct
			var productName string
			if errP := s.db.WithContext(ctx).Where("product_id = ?", prodID).First(&product).Error; errP == nil {
				productName = product.Name
			}
			return &tiktokProductLookup{
				ProductID:   prodID,
				SkuID:       skuID,
				SellerSku:   sku,
				ProductName: productName, // Might be empty, will be resolved via API later
			}, nil
		}
	}

	if err != nil {
		return nil, fmt.Errorf("TikTok SKU not found: %s", sku)
	}

	var product models.TiktokProduct
	err = s.db.WithContext(ctx).
		Where("id = ?", skuRecord.ProductID).
		First(&product).Error
	if err != nil {
		return nil, fmt.Errorf("TikTok product not found for SKU: %s", sku)
	}

	return &tiktokProductLookup{
		ProductID:   product.ProductID,
		SkuID:       skuRecord.SkuID,
		SellerSku:   skuRecord.SellerSku,
		ProductName: product.Name,
		Price:       skuRecord.Price,
	}, nil
}

// updateMpq sets MPQ and updates price for a TikTok product
func (s *TiktokMpqService) updateMpq(productID string, mpq int, skuID string, newPrice float64) *TiktokMpqResult {
	if s.tiktokAPI == nil {
		return &TiktokMpqResult{ProductID: productID, Success: false, Error: "TikTok API client not configured"}
	}

	log.Info().
		Str("tenant_id", s.tenantID).
		Str("product_id", productID).
		Int("mpq", mpq).
		Float64("price", newPrice).
		Msg("Updating TikTok MPQ")

	// Step 1: Get current product data
	productPath := fmt.Sprintf("/product/%s/products/%s", s.apiVersion, productID)
	productData, err := s.tiktokAPI.Request("GET", productPath, nil, nil)
	if err != nil {
		log.Error().Err(err).Str("product_id", productID).Msg("TikTok GET product failed")
		return &TiktokMpqResult{ProductID: productID, Success: false, Error: fmt.Sprintf("get product failed: %v", err)}
	}

	// Extract data from response
	data, ok := productData["data"].(map[string]interface{})
	if !ok {
		return &TiktokMpqResult{ProductID: productID, Success: false, Error: "invalid product data format"}
	}

	// Resolve product title from API if not already known
	productTitle := ""
	if t, ok := data["title"].(string); ok {
		productTitle = t
	}

	// Step 2: Build update payload with MPQ and PUT
	updatePayload := buildTiktokMpqPayload(data, mpq)
	_, err = s.tiktokAPI.Request("PUT", productPath, nil, updatePayload)
	if err != nil {
		log.Error().Err(err).Str("product_id", productID).Msg("TikTok PUT MPQ update failed")
		return &TiktokMpqResult{ProductID: productID, Success: false, Error: fmt.Sprintf("update MPQ failed: %v", err)}
	}

	log.Info().Str("product_id", productID).Int("mpq", mpq).Msg("TikTok MPQ updated successfully")

	// Step 3: Update price if provided
	if newPrice > 0 && skuID != "" {
		if !s.updatePrice(productID, skuID, newPrice) {
			return &TiktokMpqResult{
				ProductID: productID,
				Success:   true,
				Message:   fmt.Sprintf("%s - MPQ=%d set, but price update failed", productTitle, mpq),
			}
		}
	}

	dispName := productTitle
	if dispName == "" {
		dispName = productID
	}

	return &TiktokMpqResult{
		ProductID: productID,
		Success:   true,
		Message:   fmt.Sprintf("%s - MPQ=%d, price=%.0f", dispName, mpq, newPrice),
	}
}

// updatePrice updates price for a TikTok product
func (s *TiktokMpqService) updatePrice(productID, skuID string, price float64) bool {
	pricePath := fmt.Sprintf("/product/%s/products/%s/prices/update", s.apiVersion, productID)
	payload := map[string]interface{}{
		"skus": []map[string]interface{}{
			{
				"id": skuID,
				"price": map[string]interface{}{
					"currency": "IDR",
					"amount":   fmt.Sprintf("%d", int(price)),
				},
				"external_list_prices": []interface{}{},
			},
		},
	}

	log.Info().
		Str("product_id", productID).
		Str("sku_id", skuID).
		Float64("price", price).
		Msg("Updating TikTok price")

	_, err := s.tiktokAPI.Request("POST", pricePath, nil, payload)
	if err != nil {
		log.Error().Err(err).Msg("TikTok price update failed")
		return false
	}

	log.Info().Str("product_id", productID).Float64("price", price).Msg("TikTok price updated")
	return true
}

// buildTiktokMpqPayload builds minimal update payload
func buildTiktokMpqPayload(productData map[string]interface{}, mpq int) map[string]interface{} {
	payload := map[string]interface{}{
		"minimum_order_quantity": mpq,
		"is_cod_allowed":         true,
	}

	for _, field := range []string{
		"title", "description", "main_images", "skus",
		"package_weight", "video", "product_attributes",
	} {
		if v, ok := productData[field]; ok && v != nil {
			payload[field] = v
		}
	}

	if brand, ok := productData["brand"].(map[string]interface{}); ok {
		if id, ok := brand["id"]; ok {
			payload["brand_id"] = id
		}
	}

	if chains, ok := productData["category_chains"].([]interface{}); ok && len(chains) > 0 {
		if last, ok := chains[len(chains)-1].(map[string]interface{}); ok {
			if id, ok := last["id"]; ok {
				payload["category_id"] = id
			}
		}
	}

	return payload
}

// BatchUpdateMpq batch updates MPQ for multiple TikTok SKUs
func (s *TiktokMpqService) BatchUpdateMpq(
	ctx context.Context,
	skuPriceMap map[string]float64,
	mpq int,
) (*TiktokBatchMpqResult, error) {
	result := &TiktokBatchMpqResult{
		TotalSKUs: len(skuPriceMap),
		Results:   make([]TiktokMpqResult, 0),
		Skipped:   make([]string, 0),
	}

	log.Info().
		Str("tenant_id", s.tenantID).
		Int("sku_count", len(skuPriceMap)).
		Int("mpq", mpq).
		Msg("Batch TikTok MPQ update")

	type productInfo struct {
		name  string
		skus  []string
		skuID string
		price float64
	}
	productMap := make(map[string]*productInfo)

	for sku, price := range skuPriceMap {
		lookup, err := s.lookupProductBySku(ctx, sku)
		if err != nil {
			result.Skipped = append(result.Skipped, sku)
			log.Warn().Str("sku", sku).Err(err).Msg("TikTok SKU lookup failed")
			continue
		}

		if existing, ok := productMap[lookup.ProductID]; ok {
			existing.skus = append(existing.skus, sku)
		} else {
			productMap[lookup.ProductID] = &productInfo{
				name:  lookup.ProductName,
				skus:  []string{sku},
				skuID: lookup.SkuID,
				price: price,
			}
		}
	}

	result.UniqueProducts = len(productMap)

	for productID, info := range productMap {
		mpqResult := s.updateMpq(productID, mpq, info.skuID, info.price)
		mpqResult.Message = fmt.Sprintf("%s (SKUs: %v) - %s", info.name, info.skus, mpqResult.Message)
		result.Results = append(result.Results, *mpqResult)

		if mpqResult.Success {
			result.Processed++
		} else {
			result.Failed++
		}
	}

	result.Success = result.Failed == 0
	return result, nil
}
