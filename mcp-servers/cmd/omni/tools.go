package main

import "mcp-servers/pkg/mcp"

func buildToolDefinitions() []mcp.Tool {
	return []mcp.Tool{
		// Menu
		{
			Name:        "omni_menu",
			Description: "TRIGGER: 'MCP Omni' or 'Omni Menu' - Shows available omni management tools",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		// Token Management
		{
			Name:        "get_token_status",
			Description: "TRIGGER: 'Omni token status' - Get status of all platform tokens (connected/expired/missing)",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "get_platform_token_status",
			Description: "TRIGGER: 'Omni token status [platform]' - Get token status for a specific platform (shopee/lazada/tiktok)",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{"platform": {Type: "string", Description: "Platform name: shopee, lazada, or tiktok"}},
				Required:   []string{"platform"},
			},
		},
		{
			Name:        "refresh_token",
			Description: "TRIGGER: 'Omni refresh [platform]' - Manually refresh a specific platform token",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{"platform": {Type: "string", Description: "Platform: shopee, lazada, or tiktok"}},
				Required:   []string{"platform"},
			},
		},
		{
			Name:        "refresh_all_tokens",
			Description: "TRIGGER: 'Omni refresh all' - Refresh all platform tokens",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		// Order Management
		{
			Name:        "get_today_orders",
			Description: "TRIGGER: 'Omni orders today' - Get today's orders from all platforms",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "get_unpaid_orders",
			Description: "TRIGGER: 'Omni unpaid' - Get orders awaiting payment",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "get_unprocess_orders",
			Description: "TRIGGER: 'Omni unprocess' - Get orders not yet processed",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "sync_all_orders",
			Description: "TRIGGER: 'Omni sync all' - Trigger full order sync across all platforms",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "get_order_detail",
			Description: "TRIGGER: 'Omni order [order_sn]' - Get detailed info for a specific order",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{"order_sn": {Type: "string", Description: "Order serial number"}},
				Required:   []string{"order_sn"},
			},
		},
		// Credential Management
		{
			Name:        "disconnect_store",
			Description: "TRIGGER: 'Omni disconnect [platform] [store]' - Disconnect a store from a platform",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"platform":         {Type: "string", Description: "Platform: shopee, lazada, or tiktok"},
					"store_identifier": {Type: "string", Description: "Store identifier to disconnect"},
				},
				Required: []string{"platform", "store_identifier"},
			},
		},
		{
			Name:        "refresh_store_credential",
			Description: "TRIGGER: 'Omni refresh credential [platform] [store]' - Refresh a store's OAuth credential",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"platform":         {Type: "string", Description: "Platform: shopee, lazada, or tiktok"},
					"store_identifier": {Type: "string", Description: "Store identifier to refresh credentials for"},
				},
				Required: []string{"platform", "store_identifier"},
			},
		},
		// Auto-Function Management
		{
			Name:        "get_auto_functions",
			Description: "TRIGGER: 'Omni autofunc' - List all auto-functions with enabled status and schedules",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "toggle_auto_function",
			Description: "TRIGGER: 'Omni autofunc enable/disable [name]' - Enable or disable an auto-function",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"name":    {Type: "string", Description: "Auto-function name"},
					"enabled": {Type: "boolean", Description: "true to enable, false to disable"},
				},
				Required: []string{"name", "enabled"},
			},
		},
		{
			Name:        "run_auto_function",
			Description: "TRIGGER: 'Omni autofunc run [name]' - Manually trigger an auto-function",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{"name": {Type: "string", Description: "Auto-function name"}},
				Required:   []string{"name"},
			},
		},
		{
			Name:        "get_auto_function_history",
			Description: "TRIGGER: 'Omni autofunc history' - Get execution history of auto-functions",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		// Analytics & Monitoring
		{
			Name:        "get_system_health",
			Description: "TRIGGER: 'Omni health' - Get system health status (services, uptime, resource usage)",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "get_analytics_dashboard",
			Description: "TRIGGER: 'Omni dashboard' - Get analytics dashboard data",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "get_job_stats",
			Description: "TRIGGER: 'Omni jobs' - Get background job statistics",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		// Config & Settings
		{
			Name:        "get_settings",
			Description: "TRIGGER: 'Omni settings' - Get general settings for current tenant",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "update_settings",
			Description: "TRIGGER: 'Omni settings update' - Update general settings (pass key-value pairs)",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{"settings": {Type: "object", Description: "Settings key-value pairs to update"}},
				Required:   []string{"settings"},
			},
		},
		{
			Name:        "get_inventory_config",
			Description: "TRIGGER: 'Omni inventory config' - Get inventory management configuration",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		// Tenant & Product Management
		{
			Name:        "list_tenants",
			Description: "TRIGGER: 'Omni tenants' - List all tenants (multi-tenant overview)",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "sync_products",
			Description: "TRIGGER: 'Omni sync products' - Trigger product sync across all platforms",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		// Config Update Tools
		{
			Name:        "update_inventory_settings",
			Description: "TRIGGER: 'Omni update inventory settings' - Update inventory settings (spreadsheet_id, sheet_name, price columns, ratios)",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"settings": {Type: "object", Description: "Settings to update: spreadsheet_id, sheet_name, header_row, data_start_row, key_column, auto_sync, sync_interval_seconds, price_column, price_column_shopee, price_column_tiktok, price_column_lazada, shopee_ratio, tiktok_ratio, total_column, auto_column"},
				},
				Required: []string{"settings"},
			},
		},
		{
			Name:        "get_analytics_settings",
			Description: "TRIGGER: 'Omni analytics settings' - Get analytics settings (formula, price column, platform)",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "update_analytics_settings",
			Description: "TRIGGER: 'Omni update analytics' - Update analytics settings (formula_deduction, formula_multiplier, price_column, platform)",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"settings": {Type: "object", Description: "Settings: platform, price_column, sku_column, formula_deduction, formula_multiplier"},
				},
				Required: []string{"settings"},
			},
		},
		{
			Name:        "update_auto_function",
			Description: "TRIGGER: 'Omni autofunc update [name]' - Update auto-function config (interval_minutes, start_time, end_time)",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"name":     {Type: "string", Description: "Auto-function name"},
					"settings": {Type: "object", Description: "Config to update: interval_minutes, start_time, end_time"},
				},
				Required: []string{"name", "settings"},
			},
		},
		{
			Name:        "get_google_sheets_settings",
			Description: "TRIGGER: 'Omni sheets settings' - Get Google Sheets detailed settings",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "update_google_sheets_settings",
			Description: "TRIGGER: 'Omni update sheets' - Update Google Sheets settings (spreadsheet IDs, sheet names)",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"settings": {Type: "object", Description: "Settings: spreadsheet_id, selected_sheet, inventory_spreadsheet_id, inventory_sheet_name, wallet_spreadsheet_id, shipping_spreadsheet_id, order_spreadsheet_id, manual_mode"},
				},
				Required: []string{"settings"},
			},
		},
		{
			Name:        "update_inventory_config",
			Description: "TRIGGER: 'Omni update inventory config' - Update inventory management config",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"settings": {Type: "object", Description: "Inventory config to update"},
				},
				Required: []string{"settings"},
			},
		},
	}
}
