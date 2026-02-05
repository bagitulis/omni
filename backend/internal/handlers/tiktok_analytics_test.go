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

// TestTiktokAnalyticsHandler_GetSettings_MissingTenant tests GetSettings without tenant
func TestTiktokAnalyticsHandler_GetSettings_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewTiktokAnalyticsHandler()
	r.GET("/api/analytics/tiktok/settings", handler.GetSettings)

	req, _ := http.NewRequest("GET", "/api/analytics/tiktok/settings", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "tenant")
}

// TestTiktokAnalyticsHandler_SaveSettings_MissingTenant tests SaveSettings without tenant
func TestTiktokAnalyticsHandler_SaveSettings_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewTiktokAnalyticsHandler()
	r.POST("/api/analytics/tiktok/settings", handler.SaveSettings)

	reqBody := `{"enabled": true}`
	req, _ := http.NewRequest("POST", "/api/analytics/tiktok/settings", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestTiktokAnalyticsHandler_GetSyncStatus_MissingTenant tests GetSyncStatus without tenant
func TestTiktokAnalyticsHandler_GetSyncStatus_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewTiktokAnalyticsHandler()
	r.GET("/api/analytics/tiktok/sync-status", handler.GetSyncStatus)

	req, _ := http.NewRequest("GET", "/api/analytics/tiktok/sync-status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestTiktokAnalyticsHandler_SyncEscrow_MissingTenant tests SyncEscrow without tenant
func TestTiktokAnalyticsHandler_SyncEscrow_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewTiktokAnalyticsHandler()
	r.POST("/api/analytics/tiktok/sync", handler.SyncEscrow)

	reqBody := `{"month": 1, "year": 2025}`
	req, _ := http.NewRequest("POST", "/api/analytics/tiktok/sync", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestTiktokAnalyticsHandler_DeleteSyncData_MissingTenant tests DeleteSyncData without tenant
func TestTiktokAnalyticsHandler_DeleteSyncData_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewTiktokAnalyticsHandler()
	r.DELETE("/api/analytics/tiktok/sync", handler.DeleteSyncData)

	req, _ := http.NewRequest("DELETE", "/api/analytics/tiktok/sync", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestTiktokAnalyticsHandler_GetReconciliation_MissingTenant tests GetReconciliation without tenant
func TestTiktokAnalyticsHandler_GetReconciliation_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewTiktokAnalyticsHandler()
	r.GET("/api/analytics/tiktok/reconciliation", handler.GetReconciliation)

	req, _ := http.NewRequest("GET", "/api/analytics/tiktok/reconciliation", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestTiktokAnalyticsHandler_GetShippingFeeAnalysis_MissingTenant tests GetShippingFeeAnalysis without tenant
func TestTiktokAnalyticsHandler_GetShippingFeeAnalysis_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewTiktokAnalyticsHandler()
	r.GET("/api/analytics/tiktok/shipping-fee", handler.GetShippingFeeAnalysis)

	req, _ := http.NewRequest("GET", "/api/analytics/tiktok/shipping-fee", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestNewTiktokAnalyticsHandler tests handler creation
func TestNewTiktokAnalyticsHandler(t *testing.T) {
	handler := NewTiktokAnalyticsHandler()
	assert.NotNil(t, handler)
}
