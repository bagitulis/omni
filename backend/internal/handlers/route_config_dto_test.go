package handlers

import (
	"testing"
	"time"
)

// TestRouteConfigDTOStructure verifies the route config DTO package is importable
func TestRouteConfigDTOStructure(t *testing.T) {
	// Verify RouteConfig struct exists
	now := time.Now()
	config := RouteConfig{
		ID:       "route1",
		TenantID: "tenant123",
		RouteID:  "route_sync_orders",
		Platform: "shopee",
		Category: "sync",
		Name:     "Sync Orders",
		Enabled:  true,
		LastRun:  &now,
	}

	if config.Platform != "shopee" {
		t.Errorf("Expected platform 'shopee', got %s", config.Platform)
	}

	// Verify UpdateConfigRequest struct exists
	updateReq := UpdateConfigRequest{
		Name:     "Updated Sync Orders",
		Enabled:  boolPtr(false),
		Schedule: "0 0 * * *",
	}

	if *updateReq.Enabled != false {
		t.Errorf("Expected enabled false")
	}

	// Verify CreateConfigRequest struct exists
	createReq := CreateConfigRequest{
		RouteID:  "route_sync",
		Platform: "tiktok",
		Category: "order",
		Name:     "Sync TikTok Orders",
		Enabled:  true,
	}

	if createReq.Platform != "tiktok" {
		t.Errorf("Expected platform 'tiktok', got %s", createReq.Platform)
	}

	// Verify RouteExecutionLog struct exists
	log := RouteExecutionLog{
		ID:      "log1",
		RouteID: "route1",
		Status:  "success",
		Message: "Route executed successfully",
	}

	if log.Status != "success" {
		t.Errorf("Expected status 'success', got %s", log.Status)
	}
}

// boolPtr is a helper function to get a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}
