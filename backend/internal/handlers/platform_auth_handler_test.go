package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestPlatformAuthHandler_GetOAuthURLs_MissingTenant tests GetOAuthURLs without tenant
func TestPlatformAuthHandler_GetOAuthURLs_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewPlatformAuthHandler(nil, "http://localhost:3000", "./data")
	r.GET("/api/platform-auth/urls", handler.GetOAuthURLs)

	req, _ := http.NewRequest("GET", "/api/platform-auth/urls", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestPlatformAuthHandler_DisconnectShopee_MissingTenant tests DisconnectShopee without tenant
func TestPlatformAuthHandler_DisconnectShopee_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewPlatformAuthHandler(nil, "http://localhost:3000", "./data")
	r.GET("/api/platform-auth/shopee/disconnect", handler.DisconnectShopee)

	req, _ := http.NewRequest("GET", "/api/platform-auth/shopee/disconnect", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestPlatformAuthHandler_DisconnectLazada_MissingTenant tests DisconnectLazada without tenant
func TestPlatformAuthHandler_DisconnectLazada_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewPlatformAuthHandler(nil, "http://localhost:3000", "./data")
	r.GET("/api/platform-auth/lazada/disconnect", handler.DisconnectLazada)

	req, _ := http.NewRequest("GET", "/api/platform-auth/lazada/disconnect", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestPlatformAuthHandler_CheckAllConnections_MissingTenant tests CheckAllConnections without tenant
func TestPlatformAuthHandler_CheckAllConnections_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewPlatformAuthHandler(nil, "http://localhost:3000", "./data")
	r.POST("/api/platform-auth/check-all", handler.CheckAllConnections)

	req, _ := http.NewRequest("POST", "/api/platform-auth/check-all", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestConnectionStatus_Structure tests ConnectionStatus struct
func TestConnectionStatus_Structure(t *testing.T) {
	status := ConnectionStatus{
		Platform:    "shopee",
		Connected:   true,
		ShopID:      "12345",
		ShopName:    "Test Shop",
		ExpiresAt:   1704067200,
		ExpiresSoon: false,
		Expired:     false,
	}

	assert.Equal(t, "shopee", status.Platform)
	assert.True(t, status.Connected)
	assert.Equal(t, "12345", status.ShopID)
	assert.Equal(t, "Test Shop", status.ShopName)
}

// TestOAuthURLInfo_Structure tests OAuthURLInfo struct
func TestOAuthURLInfo_Structure(t *testing.T) {
	info := OAuthURLInfo{
		Platform: "lazada",
		AuthURL:  "https://oauth.lazada.com/auth",
		Status:   "ready",
	}

	assert.Equal(t, "lazada", info.Platform)
	assert.Contains(t, info.AuthURL, "lazada")
	assert.Equal(t, "ready", info.Status)
}
