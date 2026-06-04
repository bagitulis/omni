package main

import "fmt"

var omniClient *OmniClient

// handleToolCall routes tool calls to the appropriate handler function
func handleToolCall(name string, args map[string]interface{}) (interface{}, error) {
	switch name {
	case "omni_menu":
		return getMenu(), nil
	case "get_token_status":
		return handleGetTokenStatus()
	case "get_platform_token_status":
		return handleGetPlatformTokenStatus(args)
	case "refresh_token":
		return handleRefreshToken(args)
	case "refresh_all_tokens":
		return handleRefreshAllTokens()
	case "get_today_orders":
		return handleGetTodayOrders()
	case "get_unpaid_orders":
		return handleGetUnpaidOrders()
	case "get_unprocess_orders":
		return handleGetUnprocessOrders()
	case "sync_all_orders":
		return handleSyncAllOrders()
	case "get_order_detail":
		return handleGetOrderDetail(args)
	case "disconnect_store":
		return handleDisconnectStore(args)
	case "refresh_store_credential":
		return handleRefreshStoreCredential(args)
	case "get_auto_functions":
		return handleGetAutoFunctions()
	case "toggle_auto_function":
		return handleToggleAutoFunction(args)
	case "run_auto_function":
		return handleRunAutoFunction(args)
	case "get_auto_function_history":
		return handleGetAutoFunctionHistory()
	case "get_system_health":
		return handleGetSystemHealth()
	case "get_analytics_dashboard":
		return handleGetAnalyticsDashboard()
	case "get_job_stats":
		return handleGetJobStats()
	case "get_settings":
		return handleGetSettings()
	case "update_settings":
		return handleUpdateSettings(args)
	case "get_inventory_config":
		return handleGetInventoryConfig()
	case "list_tenants":
		return handleListTenants()
	case "sync_products":
		return handleSyncProducts()
	case "update_inventory_settings":
		return handleUpdateInventorySettings(args)
	case "get_analytics_settings":
		return handleGetAnalyticsSettings()
	case "update_analytics_settings":
		return handleUpdateAnalyticsSettings(args)
	case "update_auto_function":
		return handleUpdateAutoFunction(args)
	case "get_google_sheets_settings":
		return handleGetGoogleSheetsSettings()
	case "update_google_sheets_settings":
		return handleUpdateGoogleSheetsSettings(args)
	case "update_inventory_config":
		return handleUpdateInventoryConfig(args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func getMenu() map[string]interface{} {
	return map[string]interface{}{
		"title":       "Omni Management MCP Server",
		"description": "Manage omni e-commerce platform via Hermes AI agent",
		"shortcuts":   []string{"MCP Omni", "Omni"},
		"categories": map[string][]map[string]string{
			"Token Management": {
				{"trigger": "Omni token status", "label": "All tokens"},
				{"trigger": "Omni token status [platform]", "label": "Platform token"},
				{"trigger": "Omni refresh [platform]", "label": "Refresh token"},
				{"trigger": "Omni refresh all", "label": "Refresh all"},
			},
			"Order Management": {
				{"trigger": "Omni orders today", "label": "Today orders"},
				{"trigger": "Omni unpaid", "label": "Unpaid"},
				{"trigger": "Omni unprocess", "label": "Unprocess"},
				{"trigger": "Omni sync all", "label": "Sync all"},
				{"trigger": "Omni order [SN]", "label": "Order detail"},
			},
			"Credentials": {
				{"trigger": "Omni disconnect [platform] [store]", "label": "Disconnect store"},
				{"trigger": "Omni refresh credential [platform] [store]", "label": "Refresh credential"},
			},
			"Auto-Functions": {
				{"trigger": "Omni autofunc", "label": "List functions"},
				{"trigger": "Omni autofunc enable/disable [name]", "label": "Toggle"},
				{"trigger": "Omni autofunc run [name]", "label": "Run now"},
				{"trigger": "Omni autofunc history", "label": "History"},
			},
			"Analytics": {
				{"trigger": "Omni health", "label": "System health"},
				{"trigger": "Omni dashboard", "label": "Dashboard"},
				{"trigger": "Omni jobs", "label": "Job stats"},
			},
			"Settings": {
				{"trigger": "Omni settings", "label": "Get settings"},
				{"trigger": "Omni settings update", "label": "Update settings"},
				{"trigger": "Omni inventory config", "label": "Inventory config"},
			},
		},
	}
}
