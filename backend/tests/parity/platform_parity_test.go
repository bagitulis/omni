// Package parity provides tests comparing Go backend responses with Node.js backend
package parity

import (
	"net/http"
	"testing"
	"time"
)

// TestAuthVerifyEndpoint tests /api/auth/verify parity
func TestAuthVerifyEndpoint(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	headers := map[string]string{
		"Authorization": "Bearer " + cfg.AuthToken,
	}

	nodeResp, nodeBody, nodeTime, err := makeRequest(client, "GET", cfg.NodeJSURL+"/api/auth/verify", headers)
	if err != nil {
		t.Skipf("Node.js server not available: %v", err)
	}

	goResp, goBody, goTime, err := makeRequest(client, "GET", cfg.GoURL+"/api/auth/verify", headers)
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

// TestAuthTenantsEndpoint tests /api/auth/tenants parity
func TestAuthTenantsEndpoint(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	headers := map[string]string{
		"Authorization": "Bearer " + cfg.AuthToken,
	}

	nodeResp, nodeBody, nodeTime, err := makeRequest(client, "GET", cfg.NodeJSURL+"/api/auth/tenants", headers)
	if err != nil {
		t.Skipf("Node.js server not available: %v", err)
	}

	goResp, goBody, goTime, err := makeRequest(client, "GET", cfg.GoURL+"/api/auth/tenants", headers)
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

// TestLazadaOrdersEndpoint tests /api/lazada/orders parity
func TestLazadaOrdersEndpoint(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	headers := getAuthHeaders(cfg)

	nodeResp, nodeBody, nodeTime, err := makeRequest(client, "GET", cfg.NodeJSURL+"/api/lazada/orders?page=1&pageSize=10", headers)
	if err != nil {
		t.Fatalf("Node.js request failed: %v", err)
	}

	goResp, goBody, goTime, err := makeRequest(client, "GET", cfg.GoURL+"/api/lazada/orders?page=1&pageSize=10", headers)
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

// TestLazadaProductsEndpoint tests /api/lazada/products parity
func TestLazadaProductsEndpoint(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	headers := getAuthHeaders(cfg)

	nodeResp, nodeBody, nodeTime, err := makeRequest(client, "GET", cfg.NodeJSURL+"/api/lazada/products?page=1&pageSize=10", headers)
	if err != nil {
		t.Fatalf("Node.js request failed: %v", err)
	}

	goResp, goBody, goTime, err := makeRequest(client, "GET", cfg.GoURL+"/api/lazada/products?page=1&pageSize=10", headers)
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

// TestTiktokOrdersEndpoint tests /api/tiktok/orders parity
func TestTiktokOrdersEndpoint(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	headers := getAuthHeaders(cfg)

	nodeResp, nodeBody, nodeTime, err := makeRequest(client, "GET", cfg.NodeJSURL+"/api/tiktok/orders?page=1&pageSize=10", headers)
	if err != nil {
		t.Fatalf("Node.js request failed: %v", err)
	}

	goResp, goBody, goTime, err := makeRequest(client, "GET", cfg.GoURL+"/api/tiktok/orders?page=1&pageSize=10", headers)
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

// TestTiktokProductsEndpoint tests /api/tiktok/products parity
func TestTiktokProductsEndpoint(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	headers := getAuthHeaders(cfg)

	nodeResp, nodeBody, nodeTime, err := makeRequest(client, "GET", cfg.NodeJSURL+"/api/tiktok/products?page=1&pageSize=10", headers)
	if err != nil {
		t.Fatalf("Node.js request failed: %v", err)
	}

	goResp, goBody, goTime, err := makeRequest(client, "GET", cfg.GoURL+"/api/tiktok/products?page=1&pageSize=10", headers)
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

// TestWebhookLogsEndpoint tests /api/webhooks/logs parity
func TestWebhookLogsEndpoint(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	headers := getAuthHeaders(cfg)

	nodeResp, nodeBody, nodeTime, err := makeRequest(client, "GET", cfg.NodeJSURL+"/api/webhooks/logs?page=1&pageSize=10", headers)
	if err != nil {
		t.Fatalf("Node.js request failed: %v", err)
	}

	goResp, goBody, goTime, err := makeRequest(client, "GET", cfg.GoURL+"/api/webhooks/logs?page=1&pageSize=10", headers)
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

// TestWebhookStatsEndpoint tests /api/webhooks/stats parity
func TestWebhookStatsEndpoint(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	headers := getAuthHeaders(cfg)

	nodeResp, nodeBody, nodeTime, err := makeRequest(client, "GET", cfg.NodeJSURL+"/api/webhooks/stats", headers)
	if err != nil {
		t.Fatalf("Node.js request failed: %v", err)
	}

	goResp, goBody, goTime, err := makeRequest(client, "GET", cfg.GoURL+"/api/webhooks/stats", headers)
	if err != nil {
		t.Fatalf("Go request failed: %v", err)
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

// TestWholesaleSettingsEndpoint tests /api/wholesale/settings parity
func TestWholesaleSettingsEndpoint(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	headers := getAuthHeaders(cfg)

	nodeResp, nodeBody, nodeTime, err := makeRequest(client, "GET", cfg.NodeJSURL+"/api/wholesale/settings", headers)
	if err != nil {
		t.Fatalf("Node.js request failed: %v", err)
	}

	goResp, goBody, goTime, err := makeRequest(client, "GET", cfg.GoURL+"/api/wholesale/settings", headers)
	if err != nil {
		t.Fatalf("Go request failed: %v", err)
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
