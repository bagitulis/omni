package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestShopeeAnalyticsHandler_GetSettings_MissingTenant tests GetSettings without tenant
func TestShopeeAnalyticsHandler_GetSettings_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewShopeeAnalyticsHandler()
	r.GET("/api/analytics/shopee/settings", handler.GetSettings)

	req, _ := http.NewRequest("GET", "/api/analytics/shopee/settings", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "tenant")
}

// TestShopeeAnalyticsHandler_SaveSettings_MissingTenant tests SaveSettings without tenant
func TestShopeeAnalyticsHandler_SaveSettings_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewShopeeAnalyticsHandler()
	r.POST("/api/analytics/shopee/settings", handler.SaveSettings)

	reqBody := `{"enabled": true}`
	req, _ := http.NewRequest("POST", "/api/analytics/shopee/settings", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestShopeeAnalyticsHandler_GetSyncStatus_MissingTenant tests GetSyncStatus without tenant
func TestShopeeAnalyticsHandler_GetSyncStatus_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewShopeeAnalyticsHandler()
	r.GET("/api/analytics/shopee/sync-status", handler.GetSyncStatus)

	req, _ := http.NewRequest("GET", "/api/analytics/shopee/sync-status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestShopeeAnalyticsHandler_SyncEscrow_MissingTenant tests SyncEscrow without tenant
func TestShopeeAnalyticsHandler_SyncEscrow_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewShopeeAnalyticsHandler()
	r.POST("/api/analytics/shopee/sync", handler.SyncEscrow)

	reqBody := `{"month": 1, "year": 2025}`
	req, _ := http.NewRequest("POST", "/api/analytics/shopee/sync", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestShopeeAnalyticsHandler_DeleteSyncData_MissingTenant tests DeleteSyncData without tenant
func TestShopeeAnalyticsHandler_DeleteSyncData_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewShopeeAnalyticsHandler()
	r.DELETE("/api/analytics/shopee/sync", handler.DeleteSyncData)

	req, _ := http.NewRequest("DELETE", "/api/analytics/shopee/sync", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestShopeeAnalyticsHandler_GetReconciliation_MissingTenant tests GetReconciliation without tenant
func TestShopeeAnalyticsHandler_GetReconciliation_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewShopeeAnalyticsHandler()
	r.GET("/api/analytics/shopee/reconciliation", handler.GetReconciliation)

	req, _ := http.NewRequest("GET", "/api/analytics/shopee/reconciliation", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestShopeeAnalyticsHandler_GetShippingFeeAnalysis_MissingTenant tests GetShippingFeeAnalysis without tenant
func TestShopeeAnalyticsHandler_GetShippingFeeAnalysis_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewShopeeAnalyticsHandler()
	r.GET("/api/analytics/shopee/shipping-fee", handler.GetShippingFeeAnalysis)

	req, _ := http.NewRequest("GET", "/api/analytics/shopee/shipping-fee", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestNewShopeeAnalyticsHandler tests handler creation
func TestNewShopeeAnalyticsHandler(t *testing.T) {
	handler := NewShopeeAnalyticsHandler()
	assert.NotNil(t, handler)
}
