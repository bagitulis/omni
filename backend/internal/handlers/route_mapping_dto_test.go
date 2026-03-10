package handlers

import (
	"testing"
)

// TestRouteMappingDTOStructure verifies the route mapping DTO package is importable
func TestRouteMappingDTOStructure(t *testing.T) {
	// Verify RouteInfo struct exists
	route := RouteInfo{
		Method:      "POST",
		Path:        "/api/orders/sync",
		Handler:     "SyncOrders",
		Platform:    "shopee",
		Category:    "order",
		Description: "Synchronize orders from Shopee",
	}

	if route.Method != "POST" {
		t.Errorf("Expected method 'POST', got %s", route.Method)
	}

	// Verify HandlerInfo struct exists
	handler := HandlerInfo{
		Name:    "OrderHandler",
		Package: "handlers",
		Routes:  5,
	}

	if handler.Routes != 5 {
		t.Errorf("Expected 5 routes, got %d", handler.Routes)
	}

	// Verify DuplicateRouteInfo struct exists
	duplicate := DuplicateRouteInfo{
		Path:     "/api/products",
		Methods:  []string{"GET", "POST"},
		Handlers: []string{"ProductHandler", "CreateProductHandler"},
	}

	if len(duplicate.Methods) != 2 {
		t.Errorf("Expected 2 methods, got %d", len(duplicate.Methods))
	}

	// Verify AnalyzeRequest struct exists
	analyzeReq := AnalyzeRequest{
		IncludeStats:    true,
		IncludeMetrics:  true,
		IncludeCoverage: false,
	}

	if analyzeReq.IncludeStats != true {
		t.Errorf("Expected include_stats true")
	}

	// Verify RouteAnalysis struct exists
	analysis := RouteAnalysis{
		TotalRoutes:         42,
		AuthenticatedRoutes: 35,
		PublicRoutes:        7,
	}

	if analysis.TotalRoutes != 42 {
		t.Errorf("Expected total_routes 42, got %d", analysis.TotalRoutes)
	}
}
