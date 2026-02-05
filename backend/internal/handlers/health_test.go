package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestHealthCheck_Success tests the health check endpoint returns correct status
func TestHealthCheck_Success(t *testing.T) {
	r := gin.New()
	r.GET("/api/health", HealthCheck)

	req, _ := http.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "healthy", resp["status"])
	assert.NotEmpty(t, resp["timestamp"])
	assert.NotEmpty(t, resp["uptime"])
	assert.Equal(t, "1.0.0", resp["version"])
	assert.NotEmpty(t, resp["goVersion"])

	// Check services map
	services, ok := resp["services"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "connected", services["database"])
	assert.Equal(t, "connected", services["cache"])
}

// TestStatusCheck_NoTenant tests status check without tenant ID
func TestStatusCheck_NoTenant(t *testing.T) {
	r := gin.New()
	r.GET("/api/status", StatusCheck)

	req, _ := http.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "connected", resp["connection_status"])
	assert.NotEmpty(t, resp["timestamp"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, data["message"], "Backend is running")
}

// TestStatusCheck_WithTenantHeader tests status check with tenant ID header
func TestStatusCheck_WithTenantHeader(t *testing.T) {
	r := gin.New()
	r.GET("/api/status", StatusCheck)

	req, _ := http.NewRequest("GET", "/api/status", nil)
	req.Header.Set("x-tenant-id", "test-tenant-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "connected", resp["connection_status"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, data["shopee"], "Shopee Token Status")
	assert.Contains(t, data["lazada"], "Lazada Token Status")
	assert.Contains(t, data["tiktok"], "TikTok Token Status")
}

// TestStatusCheck_WithTenantContext tests status check with tenant ID from context
func TestStatusCheck_WithTenantContext(t *testing.T) {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "context-tenant-456")
		c.Next()
	})
	r.GET("/api/status", StatusCheck)

	req, _ := http.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "connected", resp["connection_status"])

	// With tenant context, should return platform statuses
	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, data["shopee"], "Shopee Token Status")
}

// TestHealthCheck_ResponseFormat tests that health check response format is correct
func TestHealthCheck_ResponseFormat(t *testing.T) {
	r := gin.New()
	r.GET("/api/health", HealthCheck)

	req, _ := http.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// All expected fields should be present
	expectedFields := []string{"success", "status", "timestamp", "uptime", "version", "goVersion", "services"}
	for _, field := range expectedFields {
		_, exists := resp[field]
		assert.True(t, exists, "Field %s should exist", field)
	}
}
