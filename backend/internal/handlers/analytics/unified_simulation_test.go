package analytics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ─── classifyProduct (pure function) ───────────────────────────────────────

func TestUnifiedHandler_ClassifyProduct(t *testing.T) {
	handler := NewUnifiedHandler("", nil)

	tests := []struct {
		name           string
		roas           float64
		expectedAction string
	}{
		{"SCALE_UP when roas >= 5", 5.0, "SCALE_UP"},
		{"SCALE_UP when roas > 5", 10.0, "SCALE_UP"},
		{"MAINTAIN when roas >= 2 and < 5", 2.0, "MAINTAIN"},
		{"MAINTAIN when roas = 4.9", 4.9, "MAINTAIN"},
		{"REDUCE when roas >= 1 and < 2", 1.0, "REDUCE"},
		{"REDUCE when roas = 1.5", 1.5, "REDUCE"},
		{"STOP when roas < 1", 0.5, "STOP"},
		{"STOP when roas = 0", 0.0, "STOP"},
		{"STOP when roas negative", -1.0, "STOP"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			item, action := handler.classifyProduct(
				"prod-123", "Test Product",
				100.0, 200.0, tc.roas, 10, 0.05,
			)
			assert.Equal(t, tc.expectedAction, action)
			assert.Equal(t, tc.expectedAction, item["action"])
			assert.Equal(t, "prod-123", item["product_id"])
			assert.Equal(t, "Test Product", item["product_name"])
			assert.NotEmpty(t, item["action_label"])
			assert.NotEmpty(t, item["recommendation"])
		})
	}
}

// ─── UnifiedHandler 401 tests ───────────────────────────────────────────────

func TestUnifiedHandler_GetUnifiedSummary_MissingTenant(t *testing.T) {
	handler := NewUnifiedHandler("", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/unified/summary", nil)

	handler.GetUnifiedSummary(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, false, resp["success"])
}

func TestUnifiedHandler_GetUnifiedKPI_MissingTenant(t *testing.T) {
	handler := NewUnifiedHandler("", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/unified/kpi", nil)

	handler.GetUnifiedKPI(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, false, resp["success"])
}

func TestUnifiedHandler_RefreshCache_MissingTenant(t *testing.T) {
	handler := NewUnifiedHandler("", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/cache/refresh", nil)

	handler.RefreshCache(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUnifiedHandler_GetCacheStatus_MissingTenant(t *testing.T) {
	handler := NewUnifiedHandler("", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/cache/status", nil)

	handler.GetCacheStatus(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUnifiedHandler_GetClassifiedProducts_MissingTenant(t *testing.T) {
	handler := NewUnifiedHandler("", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/products/classified", nil)

	handler.GetClassifiedProducts(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUnifiedHandler_GetTopProducts_MissingTenant(t *testing.T) {
	handler := NewUnifiedHandler("", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/products/top", nil)

	handler.GetTopProducts(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ─── SimulationHandler tests ────────────────────────────────────────────────

func TestSimulationHandler_Simulate_MissingTenant(t *testing.T) {
	handler := NewSimulationHandler("")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/simulation/calculate",
		strings.NewReader(`{"product_id":"p1","target_roas":2.0,"budget_per_day":100.0}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Simulate(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSimulationHandler_Simulate_InvalidJSON(t *testing.T) {
	handler := NewSimulationHandler("")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/simulation/calculate",
		strings.NewReader(`invalid`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("tenant_id", "test-tenant")

	handler.Simulate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSimulationHandler_GetProductsFromAds_MissingTenant(t *testing.T) {
	handler := NewSimulationHandler("")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/products/from-ads", nil)

	handler.GetProductsFromAds(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSimulationHandler_GetCalendarEvents_Default(t *testing.T) {
	handler := NewSimulationHandler("")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/intelligence/calendar", nil)

	handler.GetCalendarEvents(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(30), data["days_ahead"])
}

func TestSimulationHandler_GetCalendarEvents_CustomDays(t *testing.T) {
	handler := NewSimulationHandler("")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/intelligence/calendar?days=14", nil)
	// Set query param
	r := gin.New()
	r.GET("/api/analytics/intelligence/calendar", handler.GetCalendarEvents)
	req, _ := http.NewRequest(http.MethodGet, "/api/analytics/intelligence/calendar?days=14", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)

	assert.Equal(t, http.StatusOK, w2.Code)
	var resp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["success"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(14), data["days_ahead"])
}

// ─── TiktokAdsHandler 401 tests ─────────────────────────────────────────────

func TestTiktokAdsHandler_GetData_MissingTenant(t *testing.T) {
	handler := NewTiktokAdsHandler("")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/tiktok-ads/data", nil)

	handler.GetData(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTiktokAdsHandler_GetUploads_MissingTenant(t *testing.T) {
	handler := NewTiktokAdsHandler("")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/tiktok-ads/uploads", nil)

	handler.GetUploads(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNewTiktokAdsHandler(t *testing.T) {
	handler := NewTiktokAdsHandler("/test")
	assert.NotNil(t, handler)
	assert.Equal(t, "/test", handler.basePath)
}

func TestNewSimulationHandler(t *testing.T) {
	handler := NewSimulationHandler("/test")
	assert.NotNil(t, handler)
	assert.Equal(t, "/test", handler.basePath)
}
