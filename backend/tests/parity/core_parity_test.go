// Package parity provides tests comparing Node.js vs Go API responses
package parity

import (
	"net/http"
	"testing"
	"time"
)

// TestHealthEndpoint tests /api/health parity
func TestHealthEndpoint(t *testing.T) {
	cfg := getTestConfig()
	client := &http.Client{Timeout: 10 * time.Second}

	nodeResp, nodeBody, nodeTime, err := makeRequest(client, "GET", cfg.NodeJSURL+"/api/health", nil)
	if err != nil {
		t.Skipf("Node.js server not available: %v", err)
	}

	goResp, goBody, goTime, err := makeRequest(client, "GET", cfg.GoURL+"/api/health", nil)
	if err != nil {
		t.Skipf("Go server not available: %v", err)
	}

	logTimings(t, nodeResp, goResp, nodeTime, goTime)

	if nodeResp.StatusCode != goResp.StatusCode {
		t.Errorf("Status code mismatch: Node.js=%d, Go=%d", nodeResp.StatusCode, goResp.StatusCode)
	}

	match, diffs := compareJSON(nodeBody, goBody)
	if !match {
		logDiffs(t, diffs, 5)
	}
}

// TestShopeeOrdersEndpoint tests /api/shopee/orders parity
func TestShopeeOrdersEndpoint(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	headers := getAuthHeaders(cfg)

	nodeResp, nodeBody, nodeTime, err := makeRequest(client, "GET", cfg.NodeJSURL+"/api/shopee/orders?page=1&pageSize=10", headers)
	if err != nil {
		t.Fatalf("Node.js request failed: %v", err)
	}

	goResp, goBody, goTime, err := makeRequest(client, "GET", cfg.GoURL+"/api/shopee/orders?page=1&pageSize=10", headers)
	if err != nil {
		t.Fatalf("Go request failed: %v", err)
	}

	logTimings(t, nodeResp, goResp, nodeTime, goTime)

	if nodeResp.StatusCode != goResp.StatusCode {
		t.Errorf("Status code mismatch: Node.js=%d, Go=%d", nodeResp.StatusCode, goResp.StatusCode)
	}

	match, diffs := compareJSON(nodeBody, goBody)
	if !match {
		logDiffs(t, diffs, 10)
	}
}

// TestShopeeProductsEndpoint tests /api/shopee/products parity
func TestShopeeProductsEndpoint(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	headers := getAuthHeaders(cfg)

	nodeResp, nodeBody, nodeTime, err := makeRequest(client, "GET", cfg.NodeJSURL+"/api/shopee/products?page=1&pageSize=10", headers)
	if err != nil {
		t.Fatalf("Node.js request failed: %v", err)
	}

	goResp, goBody, goTime, err := makeRequest(client, "GET", cfg.GoURL+"/api/shopee/products?page=1&pageSize=10", headers)
	if err != nil {
		t.Fatalf("Go request failed: %v", err)
	}

	logTimings(t, nodeResp, goResp, nodeTime, goTime)

	if nodeResp.StatusCode != goResp.StatusCode {
		t.Errorf("Status code mismatch: Node.js=%d, Go=%d", nodeResp.StatusCode, goResp.StatusCode)
	}

	match, diffs := compareJSON(nodeBody, goBody)
	if !match {
		logDiffs(t, diffs, 10)
	}
}

// TestInventoryEndpoint tests /api/inventory parity
func TestInventoryEndpoint(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	headers := getAuthHeaders(cfg)

	nodeResp, nodeBody, nodeTime, err := makeRequest(client, "GET", cfg.NodeJSURL+"/api/inventory?page=1&pageSize=20", headers)
	if err != nil {
		t.Fatalf("Node.js request failed: %v", err)
	}

	goResp, goBody, goTime, err := makeRequest(client, "GET", cfg.GoURL+"/api/inventory?page=1&pageSize=20", headers)
	if err != nil {
		t.Fatalf("Go request failed: %v", err)
	}

	logTimings(t, nodeResp, goResp, nodeTime, goTime)

	if nodeResp.StatusCode != goResp.StatusCode {
		t.Errorf("Status code mismatch: Node.js=%d, Go=%d", nodeResp.StatusCode, goResp.StatusCode)
	}

	match, diffs := compareJSON(nodeBody, goBody)
	if !match {
		logDiffs(t, diffs, 10)
	}
}

// TestDashboardEndpoint tests /api/dashboard parity
func TestDashboardEndpoint(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	headers := getAuthHeaders(cfg)

	nodeResp, nodeBody, nodeTime, err := makeRequest(client, "GET", cfg.NodeJSURL+"/api/dashboard", headers)
	if err != nil {
		t.Fatalf("Node.js request failed: %v", err)
	}

	goResp, goBody, goTime, err := makeRequest(client, "GET", cfg.GoURL+"/api/dashboard", headers)
	if err != nil {
		t.Fatalf("Go request failed: %v", err)
	}

	logTimings(t, nodeResp, goResp, nodeTime, goTime)

	if nodeResp.StatusCode != goResp.StatusCode {
		t.Errorf("Status code mismatch: Node.js=%d, Go=%d", nodeResp.StatusCode, goResp.StatusCode)
	}

	match, diffs := compareJSON(nodeBody, goBody)
	if !match {
		logDiffs(t, diffs, 10)
	}
}
