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

// TestReportHandler_GetShopeeAdsReport_MissingTenant tests GetShopeeAdsReport without tenant
func TestReportHandler_GetShopeeAdsReport_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewReportHandler(nil)
	r.GET("/api/reports/shopee/ads", handler.GetShopeeAdsReport)

	req, _ := http.NewRequest("GET", "/api/reports/shopee/ads", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "tenant")
}

// TestReportHandler_GetLatestShopeeAdsReport_MissingTenant tests GetLatestShopeeAdsReport without tenant
func TestReportHandler_GetLatestShopeeAdsReport_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewReportHandler(nil)
	r.GET("/api/reports/shopee/ads/latest", handler.GetLatestShopeeAdsReport)

	req, _ := http.NewRequest("GET", "/api/reports/shopee/ads/latest", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestReportHandler_GetShopeeAdsReportByDate_MissingTenant tests GetShopeeAdsReportByDate without tenant
func TestReportHandler_GetShopeeAdsReportByDate_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewReportHandler(nil)
	r.GET("/api/reports/shopee/ads/:date", handler.GetShopeeAdsReportByDate)

	req, _ := http.NewRequest("GET", "/api/reports/shopee/ads/2025-01-15", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestReportHandler_GetShopeeAdsReportByDate_MissingDate tests GetShopeeAdsReportByDate without date
func TestReportHandler_GetShopeeAdsReportByDate_MissingDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewReportHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/reports/shopee/ads/", nil)
	c.Params = gin.Params{{Key: "date", Value: ""}}
	c.Set("tenantID", "test-tenant")

	handler.GetShopeeAdsReportByDate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "Date required")
}

// TestReportHandler_GetTiktokAdsReport_MissingTenant tests GetTiktokAdsReport without tenant
func TestReportHandler_GetTiktokAdsReport_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewReportHandler(nil)
	r.GET("/api/reports/tiktok/ads", handler.GetTiktokAdsReport)

	req, _ := http.NewRequest("GET", "/api/reports/tiktok/ads", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestReportHandler_GetLatestTiktokAdsReport_MissingTenant tests GetLatestTiktokAdsReport without tenant
func TestReportHandler_GetLatestTiktokAdsReport_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewReportHandler(nil)
	r.GET("/api/reports/tiktok/ads/latest", handler.GetLatestTiktokAdsReport)

	req, _ := http.NewRequest("GET", "/api/reports/tiktok/ads/latest", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestNewReportHandler tests handler creation
func TestNewReportHandler(t *testing.T) {
	handler := NewReportHandler(nil)
	assert.NotNil(t, handler)
}
