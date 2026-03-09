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
	assert.NotEmpty(t, resp["go_version"])

	// Check services map
	services, ok := resp["services"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "connected", services["database"])
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
	expectedFields := []string{"success", "status", "timestamp", "uptime", "version", "go_version", "services"}
	for _, field := range expectedFields {
		_, exists := resp[field]
		assert.True(t, exists, "Field %s should exist", field)
	}
}
