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

// TestGetDashboardSummary_MissingTenant tests dashboard summary without tenant
func TestGetDashboardSummary_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAnalyticsHandler(nil)

	r.GET("/api/analytics/dashboard", handler.GetDashboardSummary)

	req, _ := http.NewRequest("GET", "/api/analytics/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should fail because no tenant ID
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "tenant")
}

// TestGetOrderAnalytics_MissingTenant tests order analytics without tenant
func TestGetOrderAnalytics_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAnalyticsHandler(nil)

	r.GET("/api/analytics/orders", handler.GetOrderAnalytics)

	req, _ := http.NewRequest("GET", "/api/analytics/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

// TestGetRevenueAnalytics_MissingTenant tests revenue analytics without tenant
func TestGetRevenueAnalytics_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAnalyticsHandler(nil)

	r.GET("/api/analytics/revenue", handler.GetRevenueAnalytics)

	req, _ := http.NewRequest("GET", "/api/analytics/revenue", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetAnalyticsSettings_MissingTenant tests getting settings without tenant
func TestGetAnalyticsSettings_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAnalyticsHandler(nil)

	r.GET("/api/analytics/settings", handler.GetAnalyticsSettings)

	req, _ := http.NewRequest("GET", "/api/analytics/settings", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUpdateAnalyticsSettings_MissingTenant tests updating settings without tenant
func TestUpdateAnalyticsSettings_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAnalyticsHandler(nil)

	r.PUT("/api/analytics/settings", handler.UpdateAnalyticsSettings)

	reqBody := `{"platform": "shopee", "price_column": "price"}`
	req, _ := http.NewRequest("PUT", "/api/analytics/settings", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// NOTE: TestUpdateAnalyticsSettings_InvalidBody requires a DB mock to test properly.
// The handler validates tenant and creates a service (which needs DB) BEFORE
// validating the request body. This test is skipped until a proper mock DB
// can be injected.

// TestGetEscrowSyncStatus_MissingTenant tests escrow sync status without tenant
func TestGetEscrowSyncStatus_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAnalyticsHandler(nil)

	r.GET("/api/analytics/escrow-sync", handler.GetEscrowSyncStatus)

	req, _ := http.NewRequest("GET", "/api/analytics/escrow-sync", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAnalyticsHandler_ParseDateRange tests date range parsing
func TestAnalyticsHandler_ParseDateRange(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAnalyticsHandler(nil)
	r := gin.New()

	r.GET("/test", func(c *gin.Context) {
		start, end := handler.parseDateRange(c)
		c.JSON(http.StatusOK, gin.H{
			"start": start.Format("2006-01-02"),
			"end":   end.Format("2006-01-02"),
		})
	})

	// Test with custom dates
	req, _ := http.NewRequest("GET", "/test?startDate=2025-01-01&endDate=2025-01-31", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "2025-01-01", resp["start"])
	assert.Equal(t, "2025-01-31", resp["end"])
}

// TestAnalyticsHandler_ParseDateRange_Default tests default date range
func TestAnalyticsHandler_ParseDateRange_Default(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAnalyticsHandler(nil)
	r := gin.New()

	r.GET("/test", func(c *gin.Context) {
		start, end := handler.parseDateRange(c)
		c.JSON(http.StatusOK, gin.H{
			"has_start": !start.IsZero(),
			"has_end":   !end.IsZero(),
		})
	})

	// No date params - should use defaults
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["has_start"])
	assert.Equal(t, true, resp["has_end"])
}
