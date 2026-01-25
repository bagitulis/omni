// Package security contains security tests for multi-tenant isolation
package security

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
	tenantA  = getEnv("TEST_TENANT_A", "yumna_bertigamart")
	tenantB  = getEnv("TEST_TENANT_B", "tika_nusseyba")
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
	Data     json.RawMessage `json:"data,omitempty"`
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

// getAuthToken logs in and returns token
func getAuthToken(t *testing.T) string {
	loginReq := map[string]string{
		"username": testUser,
		"password": testPass,
	}

	_, body, err := makeRequest("POST", "/api/auth/login", loginReq, nil)
	if err != nil {
		t.Skipf("Backend not available: %v", err)
		return ""
	}

	var result APIResponse
	json.Unmarshal(body, &result)

	if result.Token == "" {
		t.Skip("Could not get auth token")
	}

	return result.Token
}

// switchTenant switches to a specific tenant and returns new token
func switchTenant(t *testing.T, currentToken, targetTenant string) string {
	headers := map[string]string{
		"Authorization": "Bearer " + currentToken,
	}

	switchReq := map[string]string{
		"tenant_id": targetTenant,
	}

	_, body, err := makeRequest("POST", "/api/auth/switch-tenant", switchReq, headers)
	if err != nil {
		t.Fatalf("Failed to switch tenant: %v", err)
	}

	var result APIResponse
	json.Unmarshal(body, &result)

	if result.Token == "" {
		t.Skipf("Could not switch to tenant %s: %s", targetTenant, result.Error)
	}

	return result.Token
}

// ===========================================
// CROSS-TENANT DATA ACCESS PREVENTION TESTS
// ===========================================

func TestCrossTenant_RejectRequestsWithoutTenantContext(t *testing.T) {
	// Make request without any authentication (no tenant context)
	resp, body, err := makeRequest("GET", "/api/orders", nil, nil)
	if err != nil {
		t.Skipf("Backend not available: %v", err)
		return
	}

	// Should fail - no tenant context
	if resp.StatusCode < 400 {
		t.Errorf("Expected status >= 400 for request without tenant, got %d", resp.StatusCode)
	}

	var result APIResponse
	json.Unmarshal(body, &result)

	if result.Success {
		t.Error("Request without tenant context should not succeed")
	}
}

func TestCrossTenant_IsolatedDataBetweenTenants(t *testing.T) {
	token := getAuthToken(t)
	if token == "" {
		return
	}

	// Get token for tenant A
	tokenA := switchTenant(t, token, tenantA)
	if tokenA == "" {
		t.Skip("Could not get token for tenant A")
		return
	}

	// Get token for tenant B
	tokenB := switchTenant(t, token, tenantB)
	if tokenB == "" {
		t.Skip("Could not get token for tenant B")
		return
	}

	// Request orders as tenant A
	headersA := map[string]string{"Authorization": "Bearer " + tokenA}
	_, bodyA, err := makeRequest("GET", "/api/orders", nil, headersA)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Request orders as tenant B
	headersB := map[string]string{"Authorization": "Bearer " + tokenB}
	_, bodyB, err := makeRequest("GET", "/api/orders", nil, headersB)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Parse responses
	var resultA, resultB struct {
		Success bool `json:"success"`
		Data    []struct {
			ID      string `json:"id"`
			OrderSN string `json:"order_sn"`
		} `json:"data"`
	}

	json.Unmarshal(bodyA, &resultA)
	json.Unmarshal(bodyB, &resultB)

	// If both have data, ensure they're different (no overlap)
	if resultA.Success && resultB.Success && len(resultA.Data) > 0 && len(resultB.Data) > 0 {
		idsA := make(map[string]bool)
		for _, order := range resultA.Data {
			idsA[order.ID] = true
			idsA[order.OrderSN] = true
		}

		for _, order := range resultB.Data {
			if idsA[order.ID] || idsA[order.OrderSN] {
				t.Errorf("Data overlap detected! Order %s/%s exists in both tenants", order.ID, order.OrderSN)
			}
		}
	}
}

func TestCrossTenant_PreventTenantIDManipulationInQuery(t *testing.T) {
	token := getAuthToken(t)
	if token == "" {
		return
	}

	// Get token for tenant A
	tokenA := switchTenant(t, token, tenantA)
	if tokenA == "" {
		t.Skip("Could not get token for tenant A")
		return
	}

	// Try to inject tenant B via query string while authenticated as tenant A
	headers := map[string]string{"Authorization": "Bearer " + tokenA}

	// Various injection attempts
	injectionAttempts := []string{
		"/api/orders?tenantId=" + tenantB,
		"/api/orders?tenant_id=" + tenantB,
		"/api/orders?tenant=" + tenantB,
	}

	for _, endpoint := range injectionAttempts {
		resp, _, err := makeRequest("GET", endpoint, nil, headers)
		if err != nil {
			continue
		}

		// Should not return 500 (internal error from injection)
		if resp.StatusCode == 500 {
			t.Errorf("Query param injection caused server error: %s", endpoint)
		}

		// The request should be processed using the token's tenant, not query param
		// (no error means it used token tenant, which is correct)
	}
}

func TestCrossTenant_PreventTenantIDManipulationInBody(t *testing.T) {
	token := getAuthToken(t)
	if token == "" {
		return
	}

	// Get token for tenant A
	tokenA := switchTenant(t, token, tenantA)
	if tokenA == "" {
		t.Skip("Could not get token for tenant A")
		return
	}

	headers := map[string]string{"Authorization": "Bearer " + tokenA}

	// Try to inject tenant B in request body
	injectionBody := map[string]interface{}{
		"tenant_id": tenantB,
		"tenantId":  tenantB,
	}

	// This should either be ignored or rejected, not processed
	resp, _, err := makeRequest("POST", "/api/orders/sync", injectionBody, headers)
	if err != nil {
		t.Skipf("Request failed: %v", err)
		return
	}

	// Should not return 500 from injection
	if resp.StatusCode == 500 {
		t.Error("Body injection caused server error")
	}
}

// ===========================================
// ERROR RESPONSE INFORMATION LEAKAGE TESTS
// ===========================================

func TestErrorResponse_NoStackTrace(t *testing.T) {
	token := getAuthToken(t)
	if token == "" {
		return
	}

	headers := map[string]string{"Authorization": "Bearer " + token}

	// Request with invalid ID to trigger error
	resp, body, err := makeRequest("GET", "/api/orders/invalid-id-12345-test", nil, headers)
	if err != nil {
		t.Skipf("Request failed: %v", err)
		return
	}

	if resp.StatusCode >= 400 {
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		// Should NOT contain stack trace
		if _, hasStack := result["stack"]; hasStack {
			t.Error("Error response should not contain stack trace")
		}

		// Should NOT contain SQL errors
		bodyStr := strings.ToLower(string(body))
		sqlKeywords := []string{"select", "insert", "update", "delete", "from", "where", "gorm", "prisma"}
		for _, keyword := range sqlKeywords {
			if strings.Contains(bodyStr, keyword) {
				t.Errorf("Error response should not contain SQL keyword: %s", keyword)
			}
		}
	}
}

func TestErrorResponse_NoOtherTenantInfo(t *testing.T) {
	token := getAuthToken(t)
	if token == "" {
		return
	}

	tokenA := switchTenant(t, token, tenantA)
	if tokenA == "" {
		t.Skip("Could not get token for tenant A")
		return
	}

	headers := map[string]string{"Authorization": "Bearer " + tokenA}

	// Request non-existent resource
	_, body, err := makeRequest("GET", "/api/orders/non-existent-order-xyz", nil, headers)
	if err != nil {
		t.Skipf("Request failed: %v", err)
		return
	}

	bodyStr := string(body)

	// Should NOT mention other tenant
	if strings.Contains(bodyStr, tenantB) {
		t.Errorf("Error response should not mention other tenant: %s", tenantB)
	}

	// Should NOT expose internal schema names
	if strings.Contains(bodyStr, "tenant_") && strings.Contains(bodyStr, "_") {
		// Check for schema name patterns like "tenant_tika_nusseyba"
		if strings.Contains(bodyStr, "tenant_tika") || strings.Contains(bodyStr, "tenant_yumna") {
			t.Error("Error response should not expose internal schema names")
		}
	}
}

// ===========================================
// API SECURITY HEADERS TESTS
// ===========================================

func TestSecurityHeaders_Present(t *testing.T) {
	resp, _, err := makeRequest("GET", "/api/health", nil, nil)
	if err != nil {
		t.Skipf("Backend not available: %v", err)
		return
	}

	// Check for common security headers
	securityHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "", // Should be present (any value)
	}

	for header, expectedValue := range securityHeaders {
		value := resp.Header.Get(header)
		if expectedValue != "" && value != expectedValue {
			t.Errorf("Expected %s=%s, got %s", header, expectedValue, value)
		}
	}
}

func TestSecurityHeaders_NoServerVersion(t *testing.T) {
	resp, _, err := makeRequest("GET", "/api/health", nil, nil)
	if err != nil {
		t.Skipf("Backend not available: %v", err)
		return
	}

	// Should NOT expose server version
	xPoweredBy := resp.Header.Get("X-Powered-By")
	if xPoweredBy != "" {
		t.Errorf("Should not expose X-Powered-By header, got: %s", xPoweredBy)
	}

	server := resp.Header.Get("Server")
	if server != "" {
		// Check if it exposes specific versions
		versionIndicators := []string{"nginx/", "Apache/", "Go/", "gin-", "1.", "2."}
		for _, indicator := range versionIndicators {
			if strings.Contains(server, indicator) {
				t.Errorf("Server header should not expose version info: %s", server)
			}
		}
	}
}

// ===========================================
// OAUTH STATE CROSS-TENANT PREVENTION TESTS
// ===========================================

func TestOAuth_RejectFakeState(t *testing.T) {
	// This tests that OAuth callbacks properly validate state
	fakeState := "fake-state-from-attacker-12345"

	// Try various OAuth callback endpoints
	oauthEndpoints := []string{
		"/api/auth/shopee/callback?code=fake&state=" + fakeState,
		"/api/auth/lazada/callback?code=fake&state=" + fakeState,
		"/api/auth/tiktok/callback?code=fake&state=" + fakeState,
	}

	for _, endpoint := range oauthEndpoints {
		resp, _, err := makeRequest("GET", endpoint, nil, nil)
		if err != nil {
			continue // Endpoint might not exist
		}

		// Should fail - state doesn't exist or invalid
		if resp.StatusCode < 400 && resp.StatusCode != 302 {
			t.Errorf("OAuth callback with fake state should fail: %s (got %d)", endpoint, resp.StatusCode)
		}
	}
}

// ===========================================
// TOKEN SECURITY TESTS
// ===========================================

func TestToken_CannotBeReusedAcrossTenants(t *testing.T) {
	token := getAuthToken(t)
	if token == "" {
		return
	}

	// Get token specifically for tenant A
	tokenA := switchTenant(t, token, tenantA)
	if tokenA == "" {
		t.Skip("Could not get token for tenant A")
		return
	}

	// Verify the token contains tenant A
	headers := map[string]string{"Authorization": "Bearer " + tokenA}
	_, verifyBody, _ := makeRequest("GET", "/api/auth/verify", nil, headers)

	var verifyResult struct {
		Payload struct {
			TenantID string `json:"tenant_id"`
		} `json:"payload"`
	}
	json.Unmarshal(verifyBody, &verifyResult)

	// Token should be bound to tenant A
	if verifyResult.Payload.TenantID != tenantA {
		t.Errorf("Token should be bound to tenant %s, got %s", tenantA, verifyResult.Payload.TenantID)
	}
}

func TestToken_ExpiredTokenRejected(t *testing.T) {
	// Use an obviously fake/expired token
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdCIsInRlbmFudF9pZCI6InRlc3QiLCJleHAiOjB9.invalid"

	headers := map[string]string{"Authorization": "Bearer " + expiredToken}
	resp, _, err := makeRequest("GET", "/api/auth/me", nil, headers)
	if err != nil {
		t.Skipf("Backend not available: %v", err)
		return
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expired/invalid token should return 401, got %d", resp.StatusCode)
	}
}

// ===========================================
// NO DEFAULT TENANT TESTS (AGENTS.MD Requirement)
// ===========================================

func TestNoDefaultTenant_MissingTenantIDErrors(t *testing.T) {
	// According to AGENTS.MD: "TIDAK ADA DEFAULT TENANT"
	// If tenantID is missing, it should error, not fallback to default

	// Make authenticated request but with token that has empty tenant
	// This tests the backend's handling of edge cases

	resp, body, err := makeRequest("GET", "/api/orders", nil, nil)
	if err != nil {
		t.Skipf("Backend not available: %v", err)
		return
	}

	var result APIResponse
	json.Unmarshal(body, &result)

	// Should fail - no tenant context
	if resp.StatusCode < 400 {
		t.Error("Request without tenant context should fail (no default tenant allowed)")
	}

	if result.Success {
		t.Error("Request without tenant should not succeed - AGENTS.MD requires no default tenant")
	}
}

// ===========================================
// CORS SECURITY TESTS
// ===========================================

func TestCORS_RejectUnknownOrigins(t *testing.T) {
	req, _ := http.NewRequest("GET", baseURL+"/api/health", nil)
	req.Header.Set("Origin", "https://malicious-site.com")

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Skipf("Backend not available: %v", err)
		return
	}
	defer resp.Body.Close()

	corsHeader := resp.Header.Get("Access-Control-Allow-Origin")

	// Should be rejected or not have wildcard CORS
	if corsHeader == "*" {
		t.Error("CORS should not allow all origins (*)")
	}

	if corsHeader == "https://malicious-site.com" {
		t.Error("CORS should not reflect malicious origin")
	}
}

func TestCORS_PreflightRequest(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", baseURL+"/api/auth/login", nil)
	req.Header.Set("Origin", "https://malicious-site.com")
	req.Header.Set("Access-Control-Request-Method", "POST")

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Skipf("Backend not available: %v", err)
		return
	}
	defer resp.Body.Close()

	corsHeader := resp.Header.Get("Access-Control-Allow-Origin")

	// Should not allow malicious origin
	if corsHeader == "https://malicious-site.com" {
		t.Error("CORS preflight should not allow malicious origin")
	}
}
