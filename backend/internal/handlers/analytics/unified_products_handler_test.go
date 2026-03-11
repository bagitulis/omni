package analytics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUnifiedHandler_GetClassifiedProducts(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewUnifiedHandler("", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/unified/products/classified", nil)

		handler.GetClassifiedProducts(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewUnifiedHandler("", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/unified/products/classified", nil)
		c.Set("tenant_id", "test-tenant")

		handler.GetClassifiedProducts(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUnifiedHandler_GetTopProducts(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewUnifiedHandler("", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/unified/products/top", nil)

		handler.GetTopProducts(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewUnifiedHandler("", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/unified/products/top", nil)
		c.Set("tenant_id", "test-tenant")

		handler.GetTopProducts(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUnifiedProductsHandler_ClassifyProduct(t *testing.T) {
	handler := NewUnifiedHandler("", nil)

	tests := []struct {
		name           string
		roas           float64
		expectedAction string
	}{
		{"roas >= 5 returns SCALE_UP", 5.0, "SCALE_UP"},
		{"roas >= 5 high returns SCALE_UP", 10.0, "SCALE_UP"},
		{"roas >= 2 returns MAINTAIN", 2.0, "MAINTAIN"},
		{"roas 3.5 returns MAINTAIN", 3.5, "MAINTAIN"},
		{"roas >= 1 returns REDUCE", 1.0, "REDUCE"},
		{"roas 1.5 returns REDUCE", 1.5, "REDUCE"},
		{"roas < 1 returns STOP", 0.5, "STOP"},
		{"roas 0 returns STOP", 0.0, "STOP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, action := handler.classifyProduct(
				"prod-1", "Test Product",
				100.0, tt.roas*100.0, tt.roas,
				10, 0.05,
			)
			assert.Equal(t, tt.expectedAction, action)
			assert.Equal(t, tt.expectedAction, item["action"])
		})
	}
}
