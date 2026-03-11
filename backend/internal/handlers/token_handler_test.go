package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestTokenHandler_NewTokenHandler tests handler creation
func TestTokenHandler_NewTokenHandler(t *testing.T) {
	handler := NewTokenHandler(nil)
	assert.NotNil(t, handler)
}

// TestTokenHandler_GetTokenStatus_MissingTenant tests GetTokenStatus without tenant
func TestTokenHandler_GetTokenStatus_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewTokenHandler(nil)
	r.GET("/api/tokens/:platform/status", handler.GetTokenStatus)

	req, _ := http.NewRequest("GET", "/api/tokens/shopee/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestTokenHandler_GetTokenStatus_InvalidPlatform tests GetTokenStatus with invalid platform
func TestTokenHandler_GetTokenStatus_InvalidPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewTokenHandler(nil)
	r.GET("/api/tokens/:platform/status", handler.GetTokenStatus)

	req, _ := http.NewRequest("GET", "/api/tokens/invalid/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTokenHandler_GetAllTokenStatus_MissingTenant tests GetAllTokenStatus without tenant
func TestTokenHandler_GetAllTokenStatus_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewTokenHandler(nil)
	r.GET("/api/tokens/status", handler.GetAllTokenStatus)

	req, _ := http.NewRequest("GET", "/api/tokens/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestTokenHandler_RefreshToken_MissingTenant tests RefreshToken without tenant
func TestTokenHandler_RefreshToken_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewTokenHandler(nil)
	r.POST("/api/tokens/:platform/refresh", handler.RefreshToken)

	req, _ := http.NewRequest("POST", "/api/tokens/shopee/refresh", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestTokenHandler_RefreshToken_MissingPlatform tests RefreshToken without platform param
func TestTokenHandler_RefreshToken_EmptyPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewTokenHandler(nil)
	// Route with empty platform
	r.POST("/api/tokens//refresh", handler.RefreshToken)

	req, _ := http.NewRequest("POST", "/api/tokens//refresh", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Empty platform should fail
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusNotFound}, w.Code)
}

// TestTokenHandler_GetPlatformTokenStatus_MissingTenant tests GetPlatformTokenStatus without tenant
func TestTokenHandler_GetPlatformTokenStatus_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewTokenHandler(nil)
	r.GET("/api/platform/:platform/token-status", handler.GetPlatformTokenStatus)

	req, _ := http.NewRequest("GET", "/api/platform/shopee/token-status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestTokenHandler_GetPlatformTokenStatus_InvalidPlatform tests GetPlatformTokenStatus with invalid platform
func TestTokenHandler_GetPlatformTokenStatus_InvalidPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewTokenHandler(nil)
	r.GET("/api/platform/:platform/token-status", handler.GetPlatformTokenStatus)

	req, _ := http.NewRequest("GET", "/api/platform/invalid/token-status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTokenHandler_RefreshAllTokens_MissingTenant tests RefreshAllTokens without tenant
func TestTokenHandler_RefreshAllTokens_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewTokenHandler(nil)
	r.POST("/api/tokens/refresh-all", handler.RefreshAllTokens)

	req, _ := http.NewRequest("POST", "/api/tokens/refresh-all", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestTokenHandler_GetStatusWithTokens_NoTenant tests GetStatusWithTokens endpoint
func TestTokenHandler_GetStatusWithTokens_NoTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewTokenHandler(nil)
	r.GET("/api/status", handler.GetStatusWithTokens)

	req, _ := http.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Without tenant should still return 200 with generic status
	assert.Equal(t, http.StatusOK, w.Code)
}
