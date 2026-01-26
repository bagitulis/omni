package platform

import (
	"context"
	"fmt"
	"time"

	"github.com/omni/backend/internal/utils"
)

// GetOrderList fetches orders by status
// Lazada uses GET /orders/get with created_after, created_before, status, offset, limit
func (c *LazadaAPIClient) GetOrderList(ctx context.Context, status string, days int) ([]map[string]interface{}, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("lazada client not initialized")
	}

	// IMPORTANT: Check token expiry and refresh if needed BEFORE API call
	if err := c.config.EnsureValidToken(ctx); err != nil {
		fmt.Printf("[WARN] Lazada token validation warning: %v\n", err)
	}

	// Reload config from DB to get latest token (after potential refresh)
	if err := c.config.LoadConfig(ctx); err != nil {
		return nil, fmt.Errorf("reload config: %w", err)
	}

	// Calculate time range (Lazada uses ISO 8601 format)
	endDate := utils.NowWIB()
	startDate := endDate.Add(-time.Duration(days) * 24 * time.Hour)

	// Format: YYYY-MM-DDTHH:MM:SS+07:00 (Indonesia WIB timezone)
	createdAfter := startDate.Format("2006-01-02T15:04:05+07:00")
	createdBefore := endDate.Format("2006-01-02T15:04:05+07:00")

	allOrders := []map[string]interface{}{}
	offset := 0
	limit := 100

	for {
		params := map[string]string{
			"created_after":  createdAfter,
			"created_before": createdBefore,
			"status":         status,
			"offset":         fmt.Sprintf("%d", offset),
			"limit":          fmt.Sprintf("%d", limit),
			"sort_by":        "created_at",
			"sort_direction": "DESC",
		}

		result, err := c.request("/orders/get", params)
		if err != nil {
			return allOrders, err
		}

		// Parse response data
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
				// Map order_id to order_sn for consistency
				// IMPORTANT: order_id is a large int64, may come as float64 from JSON
				// Must convert properly to avoid scientific notation
				if orderID, ok := orderMap["order_id"]; ok {
					switch v := orderID.(type) {
					case float64:
						// Convert float64 to int64 string without scientific notation
						orderMap["order_sn"] = fmt.Sprintf("%.0f", v)
					case int64:
						orderMap["order_sn"] = fmt.Sprintf("%d", v)
					case string:
						orderMap["order_sn"] = v
					default:
						orderMap["order_sn"] = fmt.Sprintf("%v", v)
					}
				}
				orderMap["order_status"] = status
				allOrders = append(allOrders, orderMap)
			}
		}

		offset += limit

		// Stop if we got fewer results than limit
		if len(orders) < limit {
			break
		}

		// Rate limiting
		time.Sleep(200 * time.Millisecond)
	}

	return allOrders, nil
}

// GetOrderDetails fetches order details using /order/items/get
func (c *LazadaAPIClient) GetOrderDetails(ctx context.Context, orderIDs []string) ([]map[string]interface{}, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("lazada client not initialized")
	}

	// IMPORTANT: Check token expiry and refresh if needed BEFORE API call
	if err := c.config.EnsureValidToken(ctx); err != nil {
		fmt.Printf("[WARN] Lazada token validation warning: %v\n", err)
	}

	// Reload config from DB to get latest token (after potential refresh)
	if err := c.config.LoadConfig(ctx); err != nil {
		return nil, fmt.Errorf("reload config: %w", err)
	}

	allItems := []map[string]interface{}{}

	for _, orderID := range orderIDs {
		params := map[string]string{
			"order_id": orderID,
		}

		result, err := c.request("/order/items/get", params)
		if err != nil {
			// Skip failed orders, continue with others
			continue
		}

		// Parse response
		data, ok := result["data"].([]interface{})
		if !ok {
			continue
		}

		for _, item := range data {
			if itemMap, ok := item.(map[string]interface{}); ok {
				itemMap["order_id"] = orderID
				allItems = append(allItems, itemMap)
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	return allItems, nil
}

// GetProductList fetches products
func (c *LazadaAPIClient) GetProductList(ctx context.Context, offset, limit int) ([]map[string]interface{}, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("lazada client not initialized")
	}

	params := map[string]string{
		"offset": fmt.Sprintf("%d", offset),
		"limit":  fmt.Sprintf("%d", limit),
	}

	result, err := c.request("/products/get", params)
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
