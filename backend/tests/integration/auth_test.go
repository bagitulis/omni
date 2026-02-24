// Package integration contains integration tests for the API
package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// Test configuration
var (
	baseURL  = getEnv("TEST_BASE_URL", "http://localhost:3000")
	testUser = getEnv("TEST_USERNAME", "tester")
	testPass = getEnv("TEST_PASSWORD", "tester@123")
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// APIResponse represents a generic API response
type APIResponse struct {
	Success  bool            `json:"success"`
	Message  string          `json:"message,omitempty"`
	Error    string          `json:"error,omitempty"`
	Token    string          `json:"token,omitempty"`
	TenantID string          `json:"tenant_id,omitempty"`
	User     json.RawMessage `json:"user,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
	Tenants  json.RawMessage `json:"tenants,omitempty"`
}

// TenantInfo represents tenant information
type TenantInfo struct {
	ID       string `json:"id"`
	ShopName string `json:"shop_name"`
}

// httpClient with timeout
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// makeRequest helper function
func makeRequest(method, endpoint string, body interface{}, headers map[string]string) (*http.Response, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, baseURL+endpoint, reqBody)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, err
	}

	return resp, respBody, nil
}

// ===========================================
// AUTH API INTEGRATION TESTS
// ===========================================

func TestHealthEndpoint(t *testing.T) {
	resp, body, err := makeRequest("GET", "/api/health", nil, nil)
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if status, ok := result["status"].(string); !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", result["status"])
	}
}

func TestLogin_Success(t *testing.T) {
	loginReq := map[string]string{
		"username": testUser,
		"password": testPass,
	}

	resp, body, err := makeRequest("POST", "/api/auth/login", loginReq, nil)
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	var result APIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Error: %s", resp.StatusCode, result.Error)
		return
	}

	if !result.Success {
		t.Errorf("Expected success=true, got false. Error: %s", result.Error)
	}

	if result.Token == "" {
		t.Error("Expected non-empty token")
	}

	if result.TenantID == "" {
		t.Error("Expected non-empty tenant_id")
	}

	// Verify snake_case in response (per AGENTS.MD)
	if !strings.Contains(string(body), "tenant_id") {
		t.Error("Response should use snake_case for tenant_id")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	loginReq := map[string]string{
		"username": "nonexistent",
		"password": "wrongpassword",
	}

	resp, body, err := makeRequest("POST", "/api/auth/login", loginReq, nil)
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	var result APIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}

	if result.Success {
		t.Error("Expected success=false for invalid credentials")
	}

	if result.Error == "" {
		t.Error("Expected error message for invalid credentials")
	}
}

func TestLogin_MissingFields(t *testing.T) {
	// Test missing password
	resp, body, err := makeRequest("POST", "/api/auth/login", map[string]string{
		"username": "test",
	}, nil)
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	var result APIResponse
	json.Unmarshal(body, &result)

	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 400 or 401 for missing password, got %d", resp.StatusCode)
	}

	// Test empty body
	resp2, _, _ := makeRequest("POST", "/api/auth/login", map[string]string{}, nil)
	if resp2 != nil && resp2.StatusCode == http.StatusOK {
		t.Error("Expected failure for empty body")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	loginReq := map[string]string{
		"username": testUser,
		"password": "wrongpassword123",
	}

	resp, body, err := makeRequest("POST", "/api/auth/login", loginReq, nil)
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	var result APIResponse
	json.Unmarshal(body, &result)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for wrong password, got %d", resp.StatusCode)
	}

	if result.Success {
		t.Error("Expected success=false for wrong password")
	}
}

func TestGetTenants_Authenticated(t *testing.T) {
	// First login to get token
	loginReq := map[string]string{
		"username": testUser,
		"password": testPass,
	}

	_, loginBody, err := makeRequest("POST", "/api/auth/login", loginReq, nil)
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	var loginResult APIResponse
	json.Unmarshal(loginBody, &loginResult)

	if loginResult.Token == "" {
		t.Skip("Could not get token for authenticated test")
		return
	}

	// Now get tenants with token
	headers := map[string]string{
		"Authorization": "Bearer " + loginResult.Token,
	}

	resp, body, err := makeRequest("GET", "/api/auth/tenants", nil, headers)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	var result APIResponse
	json.Unmarshal(body, &result)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Error: %s", resp.StatusCode, result.Error)
		return
	}

	if !result.Success {
		t.Errorf("Expected success=true, got false. Error: %s", result.Error)
	}

	// Parse tenants
	var tenants []TenantInfo
	if err := json.Unmarshal(result.Tenants, &tenants); err != nil {
		t.Errorf("Failed to parse tenants: %v", err)
		return
	}

	if len(tenants) == 0 {
		t.Error("Expected at least one tenant")
	}

	// Verify snake_case in tenant response
	if !strings.Contains(string(body), "shop_name") {
		t.Error("Tenant response should use snake_case for shop_name")
	}
}

func TestSwitchTenant_DeveloperOnly(t *testing.T) {
	// First login to get token
	loginReq := map[string]string{
		"username": testUser,
		"password": testPass,
	}

	_, loginBody, err := makeRequest("POST", "/api/auth/login", loginReq, nil)
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	var loginResult APIResponse
	json.Unmarshal(loginBody, &loginResult)

	if loginResult.Token == "" {
		t.Skip("Could not get token for authenticated test")
		return
	}

	// Get available tenants first
	headers := map[string]string{
		"Authorization": "Bearer " + loginResult.Token,
	}

	_, tenantsBody, _ := makeRequest("GET", "/api/auth/tenants", nil, headers)
	var tenantsResult APIResponse
	json.Unmarshal(tenantsBody, &tenantsResult)

	var tenants []TenantInfo
	json.Unmarshal(tenantsResult.Tenants, &tenants)

	if len(tenants) < 1 {
		t.Skip("Need at least 1 tenant for switch test")
		return
	}

	// Try to switch to a different tenant
	targetTenant := tenants[0].ID
	if loginResult.TenantID == targetTenant && len(tenants) > 1 {
		targetTenant = tenants[1].ID
	}

	switchReq := map[string]string{
		"tenant_id": targetTenant,
	}

	resp, body, err := makeRequest("POST", "/api/auth/switch-tenant", switchReq, headers)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	var result APIResponse
	json.Unmarshal(body, &result)

	// Developer should be able to switch
	if resp.StatusCode == http.StatusForbidden {
		t.Log("User is not a developer - switch tenant correctly rejected")
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for developer switch, got %d. Error: %s", resp.StatusCode, result.Error)
		return
	}

	if result.Token == "" {
		t.Error("Expected new token after tenant switch")
	}

	if result.TenantID != targetTenant {
		t.Errorf("Expected tenant_id=%s, got %s", targetTenant, result.TenantID)
	}
}

func TestVerifyToken_Valid(t *testing.T) {
	// First login to get token
	loginReq := map[string]string{
		"username": testUser,
		"password": testPass,
	}

	_, loginBody, err := makeRequest("POST", "/api/auth/login", loginReq, nil)
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	var loginResult APIResponse
	json.Unmarshal(loginBody, &loginResult)

	if loginResult.Token == "" {
		t.Skip("Could not get token")
		return
	}

	// Verify the token
	headers := map[string]string{
		"Authorization": "Bearer " + loginResult.Token,
	}

	resp, body, err := makeRequest("GET", "/api/auth/verify", nil, headers)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if valid, ok := result["valid"].(bool); !ok || !valid {
		t.Error("Expected valid=true for valid token")
	}
}

func TestVerifyToken_Invalid(t *testing.T) {
	headers := map[string]string{
		"Authorization": "Bearer invalid.token.here",
	}

	resp, body, err := makeRequest("GET", "/api/auth/verify", nil, headers)
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for invalid token, got %d", resp.StatusCode)
	}

	if valid, ok := result["valid"].(bool); ok && valid {
		t.Error("Expected valid=false for invalid token")
	}
}

func TestVerifyToken_NoToken(t *testing.T) {
	resp, _, err := makeRequest("GET", "/api/auth/verify", nil, nil)
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for missing token, got %d", resp.StatusCode)
	}
}

// ===========================================
// REQUEST VALIDATION TESTS
// ===========================================

func TestLogin_InvalidJSON(t *testing.T) {
	req, _ := http.NewRequest("POST", baseURL+"/api/auth/login", strings.NewReader("invalid-json"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", resp.StatusCode)
	}
}

func TestLogin_ContentTypeJSON(t *testing.T) {
	loginReq := map[string]string{
		"username": testUser,
		"password": testPass,
	}

	resp, _, err := makeRequest("POST", "/api/auth/login", loginReq, nil)
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

// ===========================================
// RESPONSE FORMAT TESTS (AGENTS.MD Compliance)
// ===========================================

func TestResponseFormat_SnakeCase(t *testing.T) {
	loginReq := map[string]string{
		"username": testUser,
		"password": testPass,
	}

	_, body, err := makeRequest("POST", "/api/auth/login", loginReq, nil)
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	// Check for snake_case keys
	bodyStr := string(body)

	// Should have snake_case
	snakeCaseFields := []string{"tenant_id", "success"}
	for _, field := range snakeCaseFields {
		if !strings.Contains(bodyStr, `"`+field+`"`) {
			t.Errorf("Response should contain snake_case field: %s", field)
		}
	}

	// Should NOT have camelCase for these fields
	camelCaseFields := []string{"tenantId"}
	for _, field := range camelCaseFields {
		if strings.Contains(bodyStr, `"`+field+`"`) {
			t.Errorf("Response should NOT contain camelCase field: %s (use snake_case)", field)
		}
	}
}

func TestResponseFormat_SuccessField(t *testing.T) {
	// Test success response
	loginReq := map[string]string{
		"username": testUser,
		"password": testPass,
	}

	_, body, err := makeRequest("POST", "/api/auth/login", loginReq, nil)
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// All responses should have "success" field
	if _, ok := result["success"]; !ok {
		t.Error("Response must include 'success' field")
	}
}
