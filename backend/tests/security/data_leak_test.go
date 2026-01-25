// Package security contains security tests for data leak prevention
package security

import (
	"encoding/json"
	"strings"
	"testing"
)

// ===========================================
// DATA LEAK PREVENTION TESTS
// ===========================================

// TestDataLeak_PasswordNotExposed ensures password is never exposed in responses
func TestDataLeak_PasswordNotExposed(t *testing.T) {
	loginReq := map[string]string{
		"username": testUser,
		"password": testPass,
	}

	_, body, err := makeRequest("POST", "/api/auth/login", loginReq, nil)
	if err != nil {
		t.Skipf("Backend not available: %v", err)
		return
	}

	bodyStr := string(body)

	// Should NOT contain the actual password
	if strings.Contains(bodyStr, testPass) {
		t.Error("Response should never contain the user's password")
	}

	// Should NOT contain password hash patterns
	hashPatterns := []string{
		"$2b$",    // bcrypt prefix
		"$2a$",    // bcrypt prefix
		"$2y$",    // bcrypt prefix
		"$argon2", // argon2 prefix
	}

	for _, pattern := range hashPatterns {
		if strings.Contains(bodyStr, pattern) {
			t.Errorf("Response should not contain password hash pattern: %s", pattern)
		}
	}

	// User object should not have password field
	var result struct {
		User struct {
			Password string `json:"password"`
		} `json:"user"`
	}

	json.Unmarshal(body, &result)

	if result.User.Password != "" {
		t.Error("User response should not include password field")
	}
}

// TestDataLeak_SensitiveFieldsNotExposed ensures sensitive fields are not exposed
func TestDataLeak_SensitiveFieldsNotExposed(t *testing.T) {
	loginReq := map[string]string{
		"username": testUser,
		"password": testPass,
	}

	_, body, err := makeRequest("POST", "/api/auth/login", loginReq, nil)
	if err != nil {
		t.Skipf("Backend not available: %v", err)
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// Sensitive fields that should NOT be exposed
	sensitiveFields := []string{
		"password",
		"passwordHash",
		"password_hash",
		"secret",
		"secretKey",
		"secret_key",
		"refreshToken", // Only access token should be exposed
		"apiKey",
		"api_key",
		"privateKey",
		"private_key",
	}

	bodyStr := strings.ToLower(string(body))

	for _, field := range sensitiveFields {
		if strings.Contains(bodyStr, `"`+strings.ToLower(field)+`"`) {
			t.Errorf("Response should not expose sensitive field: %s", field)
		}
	}
}

// TestDataLeak_InternalIDsNotExposed ensures internal DB IDs are properly handled
func TestDataLeak_InternalIDsNotExposed(t *testing.T) {
	token := getAuthToken(t)
	if token == "" {
		return
	}

	headers := map[string]string{"Authorization": "Bearer " + token}

	// Check user info endpoint
	_, body, err := makeRequest("GET", "/api/auth/me", nil, headers)
	if err != nil {
		t.Skipf("Request failed: %v", err)
		return
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// Internal implementation fields that should NOT be exposed
	internalFields := []string{
		"_id",
		"__v",
		"internalId",
		"internal_id",
		"dbId",
		"db_id",
		"prismaId",
		"gormId",
	}

	bodyStr := strings.ToLower(string(body))

	for _, field := range internalFields {
		if strings.Contains(bodyStr, `"`+strings.ToLower(field)+`"`) {
			t.Errorf("Response should not expose internal field: %s", field)
		}
	}
}

// TestDataLeak_TokenStructure ensures tokens don't expose sensitive info
func TestDataLeak_TokenStructure(t *testing.T) {
	loginReq := map[string]string{
		"username": testUser,
		"password": testPass,
	}

	_, body, err := makeRequest("POST", "/api/auth/login", loginReq, nil)
	if err != nil {
		t.Skipf("Backend not available: %v", err)
		return
	}

	var result APIResponse
	json.Unmarshal(body, &result)

	if result.Token == "" {
		t.Skip("No token in response")
		return
	}

	// JWT tokens should be 3 parts separated by dots
	parts := strings.Split(result.Token, ".")
	if len(parts) != 3 {
		t.Error("Token should be a valid JWT (3 parts)")
		return
	}

	// Token should not be too long (indicates too much data)
	if len(result.Token) > 1000 {
		t.Error("Token is suspiciously long - may contain too much sensitive data")
	}
}

// TestDataLeak_ErrorMessagesNotVerbose ensures error messages don't leak info
func TestDataLeak_ErrorMessagesNotVerbose(t *testing.T) {
	// Try login with wrong password
	loginReq := map[string]string{
		"username": testUser,
		"password": "wrongpassword123",
	}

	_, body, err := makeRequest("POST", "/api/auth/login", loginReq, nil)
	if err != nil {
		t.Skipf("Backend not available: %v", err)
		return
	}

	var result APIResponse
	json.Unmarshal(body, &result)

	if result.Success {
		t.Skip("Login succeeded unexpectedly")
		return
	}

	bodyStr := strings.ToLower(result.Error)

	// Error should NOT specify which part was wrong (security best practice)
	// Should say "invalid credentials" not "password is wrong" or "user not found"
	tooSpecificErrors := []string{
		"password is wrong",
		"password incorrect",
		"user not found",
		"username not found",
		"no user with",
		"user does not exist",
	}

	for _, specific := range tooSpecificErrors {
		if strings.Contains(bodyStr, specific) {
			t.Errorf("Error message is too specific (security risk): %s", result.Error)
		}
	}
}

// TestDataLeak_DatabaseErrorsNotExposed ensures DB errors are not exposed
func TestDataLeak_DatabaseErrorsNotExposed(t *testing.T) {
	token := getAuthToken(t)
	if token == "" {
		return
	}

	headers := map[string]string{"Authorization": "Bearer " + token}

	// Try various endpoints with bad data to trigger potential DB errors
	badRequests := []struct {
		method   string
		endpoint string
		body     interface{}
	}{
		{"GET", "/api/orders/' OR '1'='1", nil}, // SQL injection attempt
		{"GET", "/api/orders/; DROP TABLE users;--", nil},
		{"GET", "/api/products/../../../../etc/passwd", nil}, // Path traversal
		{"POST", "/api/orders/sync", map[string]interface{}{"shop_id": "'; DROP TABLE orders;--"}},
	}

	for _, req := range badRequests {
		_, body, err := makeRequest(req.method, req.endpoint, req.body, headers)
		if err != nil {
			continue
		}

		bodyStr := strings.ToLower(string(body))

		// Should NOT expose database errors
		dbErrorPatterns := []string{
			"syntax error",
			"sql error",
			"query error",
			"database error",
			"connection refused",
			"pq:",              // PostgreSQL error prefix
			"gorm:",            // GORM error
			"prisma:",          // Prisma error
			"record not found", // Should be generic "not found"
			"duplicate key",
			"foreign key",
			"constraint violation",
			"table",
			"column",
			"schema",
		}

		for _, pattern := range dbErrorPatterns {
			if strings.Contains(bodyStr, pattern) {
				t.Errorf("Response should not expose database error pattern '%s' for %s", pattern, req.endpoint)
			}
		}
	}
}

// TestDataLeak_DebugInfoNotExposed ensures debug info is not in production responses
func TestDataLeak_DebugInfoNotExposed(t *testing.T) {
	token := getAuthToken(t)
	if token == "" {
		return
	}

	headers := map[string]string{"Authorization": "Bearer " + token}

	endpoints := []string{
		"/api/health",
		"/api/auth/me",
		"/api/orders",
		"/api/products",
	}

	for _, endpoint := range endpoints {
		_, body, err := makeRequest("GET", endpoint, nil, headers)
		if err != nil {
			continue
		}

		var result map[string]interface{}
		json.Unmarshal(body, &result)

		// Debug fields that should NOT be present
		debugFields := []string{
			"debug",
			"trace",
			"stack",
			"stackTrace",
			"stack_trace",
			"query",
			"sql",
			"rawQuery",
			"raw_query",
			"executionTime",
			"execution_time",
			"memoryUsage",
			"memory_usage",
		}

		for _, field := range debugFields {
			if _, exists := result[field]; exists {
				t.Errorf("Response should not contain debug field '%s' in %s", field, endpoint)
			}
		}
	}
}

// TestDataLeak_ListEndpointsPaginated ensures list endpoints don't dump all data
func TestDataLeak_ListEndpointsPaginated(t *testing.T) {
	token := getAuthToken(t)
	if token == "" {
		return
	}

	headers := map[string]string{"Authorization": "Bearer " + token}

	listEndpoints := []string{
		"/api/orders",
		"/api/products",
	}

	for _, endpoint := range listEndpoints {
		_, body, err := makeRequest("GET", endpoint, nil, headers)
		if err != nil {
			continue
		}

		var result struct {
			Success bool          `json:"success"`
			Data    []interface{} `json:"data"`
			Meta    struct {
				Total int `json:"total"`
				Page  int `json:"page"`
				Limit int `json:"limit"`
			} `json:"meta"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			continue
		}

		if !result.Success {
			continue
		}

		// Should have reasonable pagination (not dump thousands of records)
		if len(result.Data) > 100 {
			t.Errorf("Endpoint %s returned too many records (%d) - should be paginated", endpoint, len(result.Data))
		}

		// Should include pagination meta
		if result.Meta.Limit == 0 && len(result.Data) > 20 {
			t.Errorf("Endpoint %s should include pagination metadata", endpoint)
		}
	}
}

// TestDataLeak_NoEnvironmentVariables ensures env vars are not exposed
func TestDataLeak_NoEnvironmentVariables(t *testing.T) {
	token := getAuthToken(t)
	if token == "" {
		return
	}

	headers := map[string]string{"Authorization": "Bearer " + token}

	// Check various endpoints
	endpoints := []string{
		"/api/health",
		"/api/config", // If exists
		"/api/settings",
	}

	envVarPatterns := []string{
		"DB_PASSWORD",
		"JWT_SECRET",
		"API_KEY",
		"SECRET_KEY",
		"PRIVATE_KEY",
		"AWS_SECRET",
		"DATABASE_URL",
		"REDIS_URL",
		"ENCRYPTION_KEY",
	}

	for _, endpoint := range endpoints {
		_, body, err := makeRequest("GET", endpoint, nil, headers)
		if err != nil {
			continue
		}

		bodyStr := strings.ToUpper(string(body))

		for _, pattern := range envVarPatterns {
			if strings.Contains(bodyStr, pattern) {
				t.Errorf("Response should not expose environment variable pattern '%s' in %s", pattern, endpoint)
			}
		}
	}
}

// TestDataLeak_CrossTenantDataNotLeaked ensures tenant A cannot see tenant B data
func TestDataLeak_CrossTenantDataNotLeaked(t *testing.T) {
	token := getAuthToken(t)
	if token == "" {
		return
	}

	// Switch to tenant A
	tokenA := switchTenant(t, token, tenantA)
	if tokenA == "" {
		t.Skip("Could not switch to tenant A")
		return
	}

	headers := map[string]string{"Authorization": "Bearer " + tokenA}

	// Get data from tenant A
	_, body, err := makeRequest("GET", "/api/orders", nil, headers)
	if err != nil {
		t.Skipf("Request failed: %v", err)
		return
	}

	bodyStr := string(body)

	// Response should NOT contain tenant B's identifier
	if strings.Contains(bodyStr, tenantB) {
		t.Errorf("Tenant A's response should not contain tenant B's data or identifiers")
	}

	// Check for any leaked schema references
	if strings.Contains(bodyStr, "tenant_"+tenantB) {
		t.Error("Response should not expose other tenant's schema name")
	}
}
