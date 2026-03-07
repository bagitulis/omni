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

// GetProductList fetches products
func (c *ShopeeAPIClient) GetProductList(ctx context.Context, offset, limit int) ([]map[string]interface{}, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("shopee client not initialized")
	}

	// TODO: Implement product list using Shopee API
	return nil, fmt.Errorf("shopee GetProductList is not yet implemented")
}
