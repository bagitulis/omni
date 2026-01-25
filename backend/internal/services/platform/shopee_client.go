package platform

import (
	"context"
	"fmt"
	"strconv"
	"time"

	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// ShopeeAPIClient implements APIClient for Shopee
type ShopeeAPIClient struct {
	config      *ShopeeConfigManager
	client      *shopeePkg.Client
	initialized bool
}

// NewShopeeAPIClient creates a Shopee API client
func NewShopeeAPIClient(config *ShopeeConfigManager) *ShopeeAPIClient {
	c := &ShopeeAPIClient{
		config: config,
	}

	// Initialize client if config is available
	if config.IsConfigured() {
		partnerID, _ := strconv.ParseInt(config.PartnerID, 10, 64)
		shopID, _ := strconv.ParseInt(config.ShopID, 10, 64)
		
		c.client = shopeePkg.NewClient(partnerID, config.PartnerKey, true)
		c.client.SetShopCredentials(shopID, config.GetAccessToken())
		c.initialized = true
	}

	return c
}

// IsInitialized returns whether client is ready
func (c *ShopeeAPIClient) IsInitialized() bool {
	return c.initialized && c.config.IsConfigured()
}

// GetOrderList fetches orders by status
func (c *ShopeeAPIClient) GetOrderList(ctx context.Context, status string, days int) ([]map[string]interface{}, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("shopee client not initialized")
	}

	// IMPORTANT: Check token expiry and refresh if needed BEFORE API call
	if err := c.config.EnsureValidToken(ctx); err != nil {
		// Log warning but continue - might still work if token is close to expiry
		fmt.Printf("[WARN] Token validation warning: %v\n", err)
	}

	// Reload config from DB to get latest token (after potential refresh)
	if err := c.config.LoadConfig(ctx); err != nil {
		return nil, fmt.Errorf("reload config: %w", err)
	}
	
	shopID, _ := strconv.ParseInt(c.config.ShopID, 10, 64)
	c.client.SetShopCredentials(shopID, c.config.GetAccessToken())

	// Calculate time range
	timeTo := time.Now().Unix()
	timeFrom := time.Now().AddDate(0, 0, -days).Unix()

	// Map status to Shopee time_range_field
	timeRangeField := "create_time"
	if status == "SHIPPED" || status == "COMPLETED" {
		timeRangeField = "update_time"
	}

	// Call the actual client method with status filter
	response, err := c.client.GetOrderList(timeFrom, timeTo, timeRangeField, status)
	if err != nil {
		return nil, fmt.Errorf("get order list: %w", err)
	}

	if response == nil || response.Response.OrderList == nil {
		return []map[string]interface{}{}, nil
	}

	// Convert to generic format
	result := make([]map[string]interface{}, 0, len(response.Response.OrderList))
	for _, order := range response.Response.OrderList {
		result = append(result, map[string]interface{}{
			"order_sn":   order.OrderSN,
			"platform":   "shopee",
			"status":     status,
		})
	}

	return result, nil
}

// GetOrderDetails fetches order details
func (c *ShopeeAPIClient) GetOrderDetails(ctx context.Context, orderIDs []string) ([]map[string]interface{}, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("shopee client not initialized")
	}

	// IMPORTANT: Check token expiry and refresh if needed BEFORE API call
	if err := c.config.EnsureValidToken(ctx); err != nil {
		// Log warning but continue - might still work if token is close to expiry
		fmt.Printf("[WARN] Token validation warning: %v\n", err)
	}

	// Reload config from DB to get latest token (after potential refresh)
	if err := c.config.LoadConfig(ctx); err != nil {
		return nil, fmt.Errorf("reload config: %w", err)
	}

	shopID, _ := strconv.ParseInt(c.config.ShopID, 10, 64)
	c.client.SetShopCredentials(shopID, c.config.GetAccessToken())

	// Call the actual client method (GetOrderDetail, singular)
	response, err := c.client.GetOrderDetail(orderIDs)
	if err != nil {
		return nil, fmt.Errorf("get order details: %w", err)
	}

	if response == nil || response.Response.OrderList == nil {
		return []map[string]interface{}{}, nil
	}

	// Convert to generic format - includes items
	result := make([]map[string]interface{}, 0, len(response.Response.OrderList))
	for _, order := range response.Response.OrderList {
		// Convert item_list to generic format
		items := make([]interface{}, 0, len(order.ItemList))
		for _, item := range order.ItemList {
			items = append(items, map[string]interface{}{
				"item_id":    item.ItemID,
				"model_id":   item.ModelID,
				"item_name":  item.ItemName,
				"model_name": item.ModelName,
				"item_sku":   item.ItemSKU,
				"model_sku":  item.ModelSKU,
				"quantity":   item.ModelQuantityPurchased,
				"price":      item.ModelOriginalPrice,
			})
		}

		result = append(result, map[string]interface{}{
			"order_sn":       order.OrderSN,
			"status":         order.OrderStatus,
			"total_amount":   order.TotalAmount,
			"buyer_username": order.BuyerUsername,
			"create_time":    order.CreateTime,
			"platform":       "shopee",
			"items":          items,
		})
	}

	return result, nil
}

// GetProductList fetches products
func (c *ShopeeAPIClient) GetProductList(ctx context.Context, offset, limit int) ([]map[string]interface{}, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("shopee client not initialized")
	}

	// TODO: Implement product list
	return []map[string]interface{}{}, nil
}
