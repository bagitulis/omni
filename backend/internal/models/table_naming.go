package models

import (
	"strings"
	"unicode"
)

// tableNameMapping maps PascalCase model names to snake_case PostgreSQL table names
var tableNameMapping = map[string]string{
	// User
	"User": "users",

	// Platform Config
	"PlatformConfig": "platform_configs",

	// Analytics
	"AnalyticsSettings": "analytics_settings",

	// Shopee
	"ShopeeOrder":       "shopee_orders",
	"ShopeeOrderItem":   "shopee_order_items",
	"ShopeeProduct":     "shopee_products",
	"ShopeeSku":         "shopee_skus",
	"ShopeeEscrowSync":  "shopee_escrow_sync",
	"ShopeeEscrowOrder": "shopee_escrow_orders",
	"ShopeeEscrowItem":  "shopee_escrow_items",

	// Lazada
	"LazadaOrder":     "lazada_orders",
	"LazadaOrderItem": "lazada_order_items",
	"LazadaProduct":   "lazada_products",
	"LazadaSku":       "lazada_skus",

	// Tiktok
	"TiktokOrder":       "tiktok_orders",
	"TiktokOrderItem":   "tiktok_order_items",
	"TiktokProduct":     "tiktok_products",
	"TiktokSku":         "tiktok_skus",
	"TiktokEscrowSync":  "tiktok_escrow_sync",
	"TiktokEscrowOrder": "tiktok_escrow_orders",
	"TiktokEscrowItem":  "tiktok_escrow_items",

	// Inventory
	"InventorySettings":          "inventory_settings",
	"InventoryRecord":            "inventory_records",
	"InventorySyncHistory":       "inventory_sync_history",
	"InventorySkuPlatformStatus": "inventory_sku_platform_status",
	"SheetSnapshot":              "sheet_snapshots",
	"MarketplaceSyncHistory":     "marketplace_sync_history",

	// Settings
	"GoogleSheetsSettings": "google_sheets_settings",
	"FilterPreference":     "filter_preferences",
	"Spreadsheet":          "spreadsheets",
	"RouteConfig":          "route_configs",
	"WholesaleSettings":    "wholesale_settings",

	// Unified Products
	"Product":    "products",
	"ProductSKU": "product_skus",

	// Master Products
	"MasterProduct":             "master_products",
	"MasterProductSku":          "master_product_skus",
	"MasterProductPlatformLink": "master_product_platform_links",

	// OAuth
	"OAuthState": "oauth_states",
	"OAuthLog":   "oauth_logs",

	// Webhooks
	"WebhookLog":            "webhook_logs",
	"WebhookOrderEvent":     "webhook_order_events",
	"WebhookProductEvent":   "webhook_product_events",
	"WebhookReturnEvent":    "webhook_return_events",
	"WebhookMarketingEvent": "webhook_marketing_events",
	"WebhookShopeeEvent":    "webhook_shopee_events",
	"WebhookWebchatEvent":   "webhook_webchat_events",
	"WebhookFBSEvent":       "webhook_fbs_events",

	// Jobs
	"Job":                  "jobs",
	"JobHistory":           "job_history",
	"AutoFunctionsConfig":  "auto_functions_config",
	"AutoFunctionsHistory": "auto_functions_history",

	// Order Today
	"OrderTodayItem": "order_today_items",
	"LockedOrder":    "locked_orders",

	// Audit
	"AuditLog": "audit_logs",

	// Ads
	"ShopeeAdsUploadBatch":    "shopee_ads_upload_batches",
	"ShopeeAdsProductData":    "shopee_ads_product_data",
	"TiktokAdsUploadBatch":    "tiktok_ads_upload_batches",
	"TiktokAdsCreativeData":   "tiktok_ads_creative_data",
	"TiktokAdsProductSummary": "tiktok_ads_product_summaries",
	"TiktokAdsMLPrediction":   "tiktok_ads_ml_predictions",

	// ML Reports
	"MLReport": "ml_reports",
	"MLJob":    "ml_jobs",

	// Global Config (system schema)
	"GlobalConfig": "global_config",

	// Image Gallery
	"Image": "images",

	// Security/Auth
	"RefreshSession": "refresh_sessions",
}

// GetTableName returns the PostgreSQL table name (snake_case)
func GetTableName(pascalCaseName string) string {
	if snakeName, exists := tableNameMapping[pascalCaseName]; exists {
		return snakeName
	}
	// Fallback: convert PascalCase to snake_case
	return toSnakeCase(pascalCaseName)
}

// toSnakeCase converts PascalCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
