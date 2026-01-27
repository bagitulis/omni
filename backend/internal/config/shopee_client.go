package config

import (
	"encoding/json"
	"fmt"

	"github.com/omni/backend/pkg/shopee"
)

// GetShopeeClient returns a configured Shopee client for the given tenant
// For now, returns a test client with default credentials
// TODO: Load from platform_configs table per tenant
func GetShopeeClient(tenantID, basePath string) (*shopee.Client, error) {
	if tenantID == "" {
		return nil, ErrMissingTenantID
	}

	// For now, create a test client with default credentials
	// In production, this should load from platform_configs table
	partnerID := int64(100000) // Test partner ID
	partnerKey := "test_key"   // Test partner key
	isProduction := false

	client := shopee.NewClient(partnerID, partnerKey, isProduction)

	// Set shop credentials (should be loaded from database)
	shopID := int64(123456)
	accessToken := "test_token"
	client.SetShopCredentials(shopID, accessToken)

	return client, nil
}

// Helper function to parse JSON from platform_config
func parseConfigValue(value string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil, fmt.Errorf("failed to parse config value: %w", err)
	}
	return result, nil
}
