package models

import (
	"testing"
)

func TestGetTableName(t *testing.T) {
	tests := []struct {
		pascalCase string
		expected   string
	}{
		// User
		{"User", "users"},

		// Platform Config
		{"PlatformConfig", "platform_configs"},

		// Shopee
		{"ShopeeOrder", "shopee_orders"},
		{"ShopeeOrderItem", "shopee_order_items"},
		{"ShopeeProduct", "shopee_products"},
		{"ShopeeSku", "shopee_skus"},
		{"ShopeeEscrowSync", "shopee_escrow_sync"},
		{"ShopeeEscrowOrder", "shopee_escrow_orders"},
		{"ShopeeEscrowItem", "shopee_escrow_items"},

		// Lazada
		{"LazadaOrder", "lazada_orders"},
		{"LazadaOrderItem", "lazada_order_items"},
		{"LazadaProduct", "lazada_products"},
		{"LazadaSku", "lazada_skus"},

		// Tiktok
		{"TiktokOrder", "tiktok_orders"},
		{"TiktokOrderItem", "tiktok_order_items"},
		{"TiktokProduct", "tiktok_products"},
		{"TiktokSku", "tiktok_skus"},
		{"TiktokEscrowSync", "tiktok_escrow_sync"},

		// Inventory
		{"InventorySettings", "inventory_settings"},
		{"InventoryRecord", "inventory_records"},
		{"InventorySyncHistory", "inventory_sync_history"},

		// Settings
		{"GoogleSheetsSettings", "google_sheets_settings"},
		{"FilterPreference", "filter_preferences"},
		{"Spreadsheet", "spreadsheets"},
		{"RouteConfig", "route_configs"},
		{"WholesaleSettings", "wholesale_settings"},

		// Unified Products
		{"Product", "products"},
		{"ProductSKU", "product_skus"},

		// OAuth
		{"OAuthState", "oauth_states"},
		{"OAuthLog", "oauth_logs"},

		// Webhooks
		{"WebhookLog", "webhook_logs"},
		{"WebhookOrderEvent", "webhook_order_events"},

		// Jobs
		{"Job", "jobs"},
		{"JobHistory", "job_history"},
		{"AutoFunctionsConfig", "auto_functions_config"},

		// Global Config
		{"GlobalConfig", "global_config"},
	}

	for _, tt := range tests {
		t.Run(tt.pascalCase, func(t *testing.T) {
			got := GetTableName(tt.pascalCase)
			if got != tt.expected {
				t.Errorf("GetTableName(%s) = %v, want %v", tt.pascalCase, got, tt.expected)
			}
		})
	}
}

func TestGetTableName_Fallback(t *testing.T) {
	// Test fallback to snake_case conversion for unmapped names
	tests := []struct {
		pascalCase string
		expected   string
	}{
		{"UnknownModel", "unknown_model"},
		{"SomeNewTable", "some_new_table"},
		{"ABC", "a_b_c"},
		{"Simple", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.pascalCase, func(t *testing.T) {
			got := GetTableName(tt.pascalCase)
			if got != tt.expected {
				t.Errorf("GetTableName(%s) fallback = %v, want %v", tt.pascalCase, got, tt.expected)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PascalCase", "pascal_case"},
		{"camelCase", "camel_case"},
		{"Simple", "simple"},
		{"ABC", "a_b_c"},
		{"XMLParser", "x_m_l_parser"},
		{"getHTTPResponse", "get_h_t_t_p_response"},
		{"", ""},
		{"lowercase", "lowercase"},
		{"A", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toSnakeCase(tt.input)
			if got != tt.expected {
				t.Errorf("toSnakeCase(%s) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// Test that all models have consistent TableName implementations
func TestModelTableNames(t *testing.T) {
	// Verify specific models return correct table names
	user := User{}
	if got := user.TableName(); got != "users" {
		t.Errorf("User.TableName() = %v, want 'users'", got)
	}

	shopeeOrder := ShopeeOrder{}
	if got := shopeeOrder.TableName(); got != "shopee_orders" {
		t.Errorf("ShopeeOrder.TableName() = %v, want 'shopee_orders'", got)
	}

	lazadaOrder := LazadaOrder{}
	if got := lazadaOrder.TableName(); got != "lazada_orders" {
		t.Errorf("LazadaOrder.TableName() = %v, want 'lazada_orders'", got)
	}

	tiktokOrder := TiktokOrder{}
	if got := tiktokOrder.TableName(); got != "tiktok_orders" {
		t.Errorf("TiktokOrder.TableName() = %v, want 'tiktok_orders'", got)
	}

	invRecord := InventoryRecord{}
	if got := invRecord.TableName(); got != "inventory_records" {
		t.Errorf("InventoryRecord.TableName() = %v, want 'inventory_records'", got)
	}

	platformCfg := PlatformConfig{}
	if got := platformCfg.TableName(); got != "platform_configs" {
		t.Errorf("PlatformConfig.TableName() = %v, want 'platform_configs'", got)
	}
}

// Test table name mapping completeness
func TestTableNameMappingExists(t *testing.T) {
	requiredMappings := []string{
		"User",
		"ShopeeOrder",
		"ShopeeOrderItem",
		"ShopeeProduct",
		"LazadaOrder",
		"LazadaOrderItem",
		"LazadaProduct",
		"TiktokOrder",
		"TiktokOrderItem",
		"TiktokProduct",
		"InventoryRecord",
		"PlatformConfig",
		"Job",
		"WebhookLog",
	}

	for _, name := range requiredMappings {
		t.Run(name, func(t *testing.T) {
			tableName := GetTableName(name)
			// Should not be empty
			if tableName == "" {
				t.Errorf("GetTableName(%s) returned empty string", name)
			}
			// Should contain underscore (snake_case) or be lowercase
			if tableName[0] >= 'A' && tableName[0] <= 'Z' {
				t.Errorf("GetTableName(%s) = %v, should be snake_case", name, tableName)
			}
		})
	}
}
