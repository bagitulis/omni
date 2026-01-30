// Package lighthouse provides Lighthouse performance testing for the OMNI frontend
package lighthouse

// GetAllRoutes returns all routes to test - matching frontend/src/router/routes.ts
func GetAllRoutes() []Route {
	return []Route{
		// Core
		{Name: "Dashboard", Path: "/"},
		{Name: "Inventory", Path: "/inventory"},
		{Name: "Route Mapping", Path: "/route-mapping"},

		// Product Manager
		{Name: "Product Manager - Shopee", Path: "/product-manager/shopee"},
		{Name: "Product Manager - Lazada", Path: "/product-manager/lazada"},
		{Name: "Product Manager - Tiktok", Path: "/product-manager/tiktok"},

		// Order Manager
		{Name: "Order Manager", Path: "/order-manager"},

		// Script Monitor
		{Name: "Script Monitor - Current", Path: "/script-monitor/current"},
		{Name: "Script Monitor - Queue", Path: "/script-monitor/queue"},
		{Name: "Script Monitor - History", Path: "/script-monitor/history"},
		{Name: "Script Monitor - Auto Functions", Path: "/script-monitor/auto-functions"},

		// Report
		{Name: "Report - Shopee", Path: "/report/shopee"},
		{Name: "Report - Tiktok", Path: "/report/tiktok"},

		// Analytics
		{Name: "Analytics - Hub", Path: "/analytics/hub"},
		{Name: "Analytics - Simulator", Path: "/analytics/simulator"},
		{Name: "Analytics - Classification", Path: "/analytics/classification"},
		{Name: "Analytics - ML Dashboard", Path: "/analytics/ml"},
		{Name: "Analytics - Tiktok Ads", Path: "/analytics/tiktok-ads"},
		{Name: "Analytics - Shopee Ads", Path: "/analytics/shopee-ads"},
		{Name: "Analytics - AI Reports", Path: "/analytics/ai-reports"},

		// Settings
		{Name: "Settings - Google Sheets", Path: "/settings/google-sheets"},
		{Name: "Settings - Webhook", Path: "/settings/webhook"},
	}
}

// GetQuickTestRoutes returns a subset of routes for quick testing
func GetQuickTestRoutes() []Route {
	return []Route{
		{Name: "Dashboard", Path: "/"},
		{Name: "Order Manager", Path: "/order-manager"},
		{Name: "Analytics - Hub", Path: "/analytics/hub"},
	}
}

// GetRoutesByCategory returns routes grouped by category
func GetRoutesByCategory(category string) []Route {
	all := GetAllRoutes()
	var filtered []Route

	categoryPrefixes := map[string][]string{
		"core":      {"/", "/inventory", "/route-mapping"},
		"product":   {"/product-manager"},
		"order":     {"/order-manager"},
		"script":    {"/script-monitor"},
		"report":    {"/report"},
		"analytics": {"/analytics"},
		"settings":  {"/settings"},
	}

	prefixes, ok := categoryPrefixes[category]
	if !ok {
		return all
	}

	for _, route := range all {
		for _, prefix := range prefixes {
			if route.Path == prefix || (len(prefix) > 1 && len(route.Path) >= len(prefix) && route.Path[:len(prefix)] == prefix) {
				filtered = append(filtered, route)
				break
			}
		}
	}

	return filtered
}
