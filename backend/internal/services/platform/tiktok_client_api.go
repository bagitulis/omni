package platform

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// GetOrderList fetches orders by status using POST /order/202309/orders/search
func (c *TiktokAPIClient) GetOrderList(ctx context.Context, status string, days int) ([]map[string]interface{}, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("tiktok client not initialized")
	}

	// IMPORTANT: Check token expiry and refresh if needed BEFORE API call
	if err := c.config.EnsureValidToken(ctx); err != nil {
		expiry := c.config.GetTokenExpiry()
		if expiry > 0 && time.Now().UnixMilli() >= expiry {
			// Hard-expired: token is definitely expired and refresh failed — abort
			return nil, fmt.Errorf("tiktok: token expired and refresh failed: %w", err)
		}
		// Buffer/unknown expiry (expiry==0 or within buffer) — warn and continue
		log.Warn().Err(err).Str("platform", "tiktok").Msg("Token validation warning, continuing with existing token")
	}

	// Reload config from DB to get latest token (after potential refresh)
	if err := c.config.LoadConfig(ctx); err != nil {
		return nil, fmt.Errorf("reload config: %w", err)
	}

	// Calculate time range (TikTok uses Unix timestamp in seconds)
	endTime := time.Now().Unix()
	startTime := endTime - int64(days*24*60*60)

	allOrders := []map[string]interface{}{}
	pageToken := ""

	for {
		queryParams := map[string]string{
			"page_size":   "100",
			"shop_cipher": c.config.ShopCipher,
		}
		if pageToken != "" {
			queryParams["page_token"] = pageToken
		}

		// TikTok requires POST with body for order search
		body := map[string]interface{}{
			"create_time_ge": startTime,
			"create_time_lt": endTime,
			"order_status":   status,
		}

		result, err := c.request("POST", "/order/202309/orders/search", queryParams, body)
		if err != nil {
			return allOrders, err
		}

		// Parse response
		data, ok := result["data"].(map[string]interface{})
		if !ok {
			break
		}

		orders, ok := data["orders"].([]interface{})
		if !ok || len(orders) == 0 {
			break
		}

		// Convert orders to map
		for _, o := range orders {
			if orderMap, ok := o.(map[string]interface{}); ok {
				// Map 'id' to 'order_sn' for consistency
				if id, ok := orderMap["id"].(string); ok {
					orderMap["order_sn"] = id
				}

				// For AWAITING_COLLECTION (processed) orders, extract tracking info
				// TikTok search API includes this at order level or in packages array
				if status == "AWAITING_COLLECTION" {
					// Try to get tracking from packages array first
					if packages, ok := orderMap["packages"].([]interface{}); ok && len(packages) > 0 {
						if pkg, ok := packages[0].(map[string]interface{}); ok {
							if trackingNo, ok := pkg["tracking_number"].(string); ok && trackingNo != "" {
								orderMap["tracking_number"] = trackingNo
							}
							if carrier, ok := pkg["shipping_provider_name"].(string); ok && carrier != "" {
								orderMap["shipping_provider_name"] = carrier
							}
						}
					}
					// Fallback: check order-level tracking fields
					if _, hasTracking := orderMap["tracking_number"]; !hasTracking {
						if trackingNo, ok := orderMap["tracking_no"].(string); ok {
							orderMap["tracking_number"] = trackingNo
						}
					}
				}

				allOrders = append(allOrders, orderMap)
			}
		}

		// Check for next page
		if nextToken, ok := data["next_page_token"].(string); ok && nextToken != "" {
			pageToken = nextToken
		} else {
			break
		}

		// Rate limiting
		time.Sleep(200 * time.Millisecond)
	}

	return allOrders, nil
}

// GetOrderDetails fetches order details using GET /order/202309/orders
func (c *TiktokAPIClient) GetOrderDetails(ctx context.Context, orderIDs []string) ([]map[string]interface{}, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("tiktok client not initialized")
	}

	// IMPORTANT: Check token expiry and refresh if needed BEFORE API call
	if err := c.config.EnsureValidToken(ctx); err != nil {
		expiry := c.config.GetTokenExpiry()
		if expiry > 0 && time.Now().UnixMilli() >= expiry {
			// Hard-expired: token is definitely expired and refresh failed — abort
			return nil, fmt.Errorf("tiktok: token expired and refresh failed: %w", err)
		}
		// Buffer/unknown expiry (expiry==0 or within buffer) — warn and continue
		log.Warn().Err(err).Str("platform", "tiktok").Msg("Token validation warning, continuing with existing token")
	}

	// Reload config from DB to get latest token (after potential refresh)
	if err := c.config.LoadConfig(ctx); err != nil {
		return nil, fmt.Errorf("reload config: %w", err)
	}

	// TikTok allows up to 50 order IDs per request
	allOrders := []map[string]interface{}{}
	batchSize := 50

	for i := 0; i < len(orderIDs); i += batchSize {
		end := i + batchSize
		if end > len(orderIDs) {
			end = len(orderIDs)
		}
		batch := orderIDs[i:end]

		queryParams := map[string]string{
			"ids":         strings.Join(batch, ","),
			"shop_cipher": c.config.ShopCipher,
		}

		result, err := c.request("GET", "/order/202309/orders", queryParams, nil)
		if err != nil {
			continue
		}

		data, ok := result["data"].(map[string]interface{})
		if !ok {
			continue
		}

		orders, ok := data["orders"].([]interface{})
		if !ok {
			continue
		}

		for _, o := range orders {
			if orderMap, ok := o.(map[string]interface{}); ok {
				allOrders = append(allOrders, orderMap)
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	return allOrders, nil
}

// GetProductList fetches products using POST /product/202309/products/search
func (c *TiktokAPIClient) GetProductList(ctx context.Context, offset, limit int) ([]map[string]interface{}, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("tiktok client not initialized")
	}

	queryParams := map[string]string{
		"page_size":   fmt.Sprintf("%d", limit),
		"shop_cipher": c.config.ShopCipher,
	}

	body := map[string]interface{}{}

	result, err := c.request("POST", "/product/202309/products/search", queryParams, body)
	if err != nil {
		return nil, err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return []map[string]interface{}{}, nil
	}

	products, ok := data["products"].([]interface{})
	if !ok {
		return []map[string]interface{}{}, nil
	}

	productMaps := make([]map[string]interface{}, 0, len(products))
	for _, p := range products {
		if pm, ok := p.(map[string]interface{}); ok {
			productMaps = append(productMaps, pm)
		}
	}

	return productMaps, nil
}
