// Package lighthouse provides Lighthouse performance testing for the OMNI frontend
package lighthouse

// GetAllRoutes returns all routes to test - matching frontend/src/App.tsx static routes
// Parameterized routes are excluded (require data seeding) - see TESTING_RULES.md Skipped Routes section
func GetAllRoutes() []Route {
	return []Route{
		// Core
		{Name: "Dashboard", Path: "/"},
		{Name: "Order Manager", Path: "/order-manager"},
		{Name: "Route Mapping", Path: "/route-mapping"},
		{Name: "Inventory", Path: "/inventory"},
		{Name: "Settings", Path: "/settings"},

		// Products
		{Name: "Products", Path: "/products"},
		{Name: "Add Product", Path: "/products/add"},
		{Name: "Product Sync History", Path: "/products/sync-history"},
		// SKIP: parameterized route - requires valid product ID: /products/:id/edit
		// SKIP: requires file upload fixture: /products/import
		// SKIP: parameterized route - requires valid platform slug: /order-manager/:platform

		// Analytics
		{Name: "Analytics Hub", Path: "/analytics"},
		{Name: "Shopee Ads Analytics", Path: "/analytics/shopee-ads"},
		{Name: "TikTok Ads Analytics", Path: "/analytics/tiktok-ads"},
		{Name: "ML Dashboard", Path: "/analytics/ml"},
		{Name: "Budget Simulator", Path: "/analytics/budget-simulator"},
		{Name: "Product Classification", Path: "/analytics/product-classification"},
		{Name: "AI Reports", Path: "/analytics/ai-reports"},

		// Reports
		{Name: "Shopee Report", Path: "/report/shopee"},
		{Name: "TikTok Report", Path: "/report/tiktok"},

		// Script Monitor
		{Name: "Script Monitor", Path: "/script-monitor"},
	}
}

// GetQuickTestRoutes returns a subset of routes for quick testing (3 routes)
func GetQuickTestRoutes() []Route {
	return []Route{
		{Name: "Dashboard", Path: "/"},
		{Name: "Order Manager", Path: "/order-manager"},
		{Name: "Analytics Hub", Path: "/analytics"},
	}
}

// GetRoutesByCategory returns routes grouped by category
func GetRoutesByCategory(category string) []Route {
	all := GetAllRoutes()
	var filtered []Route

	categoryPrefixes := map[string][]string{
		"core":      {"/", "/inventory", "/route-mapping", "/settings"},
		"product":   {"/products"},
		"order":     {"/order-manager"},
		"script":    {"/script-monitor"},
		"report":    {"/report"},
		"analytics": {"/analytics"},
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
