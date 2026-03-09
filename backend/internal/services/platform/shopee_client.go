package platform

import (
	"context"
	"fmt"
	"strconv"

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

// GetClient returns the underlying Shopee client for advanced operations
func (c *ShopeeAPIClient) GetClient() *shopeePkg.Client {
	return c.client
}

// GetProductList fetches products using Shopee API v2
// API: GET /api/v2/product/get_item_list
// Docs: https://open.shopee.com/documents/v2/v2.product.get_item_list
func (c *ShopeeAPIClient) GetProductList(_ context.Context, offset, limit int) ([]map[string]interface{}, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("shopee client not initialized")
	}

	rawResp, err := c.client.GetItemList(offset, limit, "NORMAL")
	if err != nil {
		return nil, err
	}

	// Parse response: { error, message, response: { item: [...], total_count, has_next_page, next_offset } }
	if errStr, ok := rawResp["error"].(string); ok && errStr != "" {
		msg, _ := rawResp["message"].(string)
		return nil, fmt.Errorf("shopee API error: %s - %s", errStr, msg)
	}

	response, ok := rawResp["response"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("shopee GetProductList: unexpected response shape (missing 'response' object)")
	}

	items, ok := response["item"].([]interface{})
	if !ok {
		return []map[string]interface{}{}, nil // No items = empty result (valid)
	}

	productMaps := make([]map[string]interface{}, 0, len(items))
	for _, p := range items {
		if pm, ok := p.(map[string]interface{}); ok {
			productMaps = append(productMaps, pm)
		}
	}

	return productMaps, nil
}
