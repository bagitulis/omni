package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDevLoginInfo_DevMode(t *testing.T) {
	// Ensure we're in dev mode
	os.Setenv("GO_ENV", "development")
	defer os.Unsetenv("GO_ENV")

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a minimal auth handler (nil services - we only test the env check)
	handler := &AuthHandler{}
	router.GET("/auth/dev-info", handler.DevLoginInfo)

	req, _ := http.NewRequest("GET", "/auth/dev-info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.True(t, response["dev_mode"].(bool))
	assert.Equal(t, "tester", response["username"])

	// Check tenants are returned
	tenants := response["tenants"].([]interface{})
	assert.Len(t, tenants, 2)
}

func TestDevLoginInfo_ProductionMode(t *testing.T) {
	// Set production mode
	os.Setenv("GO_ENV", "production")
	defer os.Unsetenv("GO_ENV")

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &AuthHandler{}
	router.GET("/auth/dev-info", handler.DevLoginInfo)

	req, _ := http.NewRequest("GET", "/auth/dev-info", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 404 in production
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Equal(t, "Not found", response["error"])
}

func TestDevLogin_ProductionMode(t *testing.T) {
	// Set production mode
	os.Setenv("GO_ENV", "production")
	defer os.Unsetenv("GO_ENV")

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &AuthHandler{}
	router.POST("/auth/dev-login", handler.DevLogin)

	body := `{"tenant_id": "yumna_bertigamart"}`
	req, _ := http.NewRequest("POST", "/auth/dev-login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 404 in production - endpoint hidden
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDevLogin_InvalidTenant(t *testing.T) {
	// Dev mode
	os.Setenv("GO_ENV", "development")
	defer os.Unsetenv("GO_ENV")

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &AuthHandler{}
	router.POST("/auth/dev-login", handler.DevLogin)

	body := `{"tenant_id": "invalid_tenant"}`
	req, _ := http.NewRequest("POST", "/auth/dev-login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 400 for invalid tenant
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["error"], "Invalid tenant_id")
}

func TestDevLogin_MissingTenantID(t *testing.T) {
	// Dev mode
	os.Setenv("GO_ENV", "development")
	defer os.Unsetenv("GO_ENV")

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &AuthHandler{}
	router.POST("/auth/dev-login", handler.DevLogin)

	body := `{}`
	req, _ := http.NewRequest("POST", "/auth/dev-login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 400 for missing tenant_id
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
