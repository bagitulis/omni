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

func TestSimulationHandler_Simulate(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewSimulationHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/simulation/calculate", strings.NewReader(`{}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Simulate(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		handler := NewSimulationHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/simulation/calculate", strings.NewReader(`not-json`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.Simulate(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("missing required fields returns 400", func(t *testing.T) {
		handler := NewSimulationHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// Missing required fields: target_roas and budget_per_day
		c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/simulation/calculate",
			strings.NewReader(`{"product_id": "prod-1"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.Simulate(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("valid JSON with tenant but no DB returns 500", func(t *testing.T) {
		handler := NewSimulationHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/simulation/calculate",
			strings.NewReader(`{"product_id": "prod-1", "target_roas": 3.0, "budget_per_day": 100.0}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.Simulate(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSimulationHandler_GetProductsFromAds(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewSimulationHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/simulation/products", nil)

		handler.GetProductsFromAds(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewSimulationHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/simulation/products", nil)
		c.Set("tenant_id", "test-tenant")

		handler.GetProductsFromAds(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestSimulationHandler_GetCalendarEvents(t *testing.T) {
	t.Run("no tenant ID required - does not return 401", func(t *testing.T) {
		handler := NewSimulationHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/simulation/calendar", nil)

		// GetCalendarEvents does NOT require tenant auth
		handler.GetCalendarEvents(c)

		// Should not return 401 (no auth check)
		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	})
}

func TestSimulateRequest_StructFields(t *testing.T) {
	t.Run("struct can be instantiated with fields", func(t *testing.T) {
		req := SimulateRequest{
			ProductID:    "prod-123",
			TargetRoas:   3.0,
			BudgetPerDay: 100.0,
			PeriodDays:   30,
		}
		assert.Equal(t, "prod-123", req.ProductID)
		assert.Equal(t, 3.0, req.TargetRoas)
		assert.Equal(t, 100.0, req.BudgetPerDay)
		assert.Equal(t, 30, req.PeriodDays)
	})

	t.Run("struct marshals to snake_case JSON", func(t *testing.T) {
		req := SimulateRequest{
			ProductID:    "prod-1",
			TargetRoas:   2.5,
			BudgetPerDay: 50.0,
			PeriodDays:   7,
		}
		data, err := json.Marshal(req)
		assert.NoError(t, err)

		var m map[string]interface{}
		json.Unmarshal(data, &m)
		assert.Contains(t, m, "product_id")
		assert.Contains(t, m, "target_roas")
		assert.Contains(t, m, "budget_per_day")
		assert.Contains(t, m, "period_days")
	})
}

func TestNewSimulationHandler_Creation(t *testing.T) {
	t.Run("creates handler with base path", func(t *testing.T) {
		handler := NewSimulationHandler("/test/path")
		assert.NotNil(t, handler)
	})

	t.Run("creates handler with empty base path", func(t *testing.T) {
		handler := NewSimulationHandler("")
		assert.NotNil(t, handler)
	})
}
