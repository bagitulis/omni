package analytics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAdsHandler_GetDashboard(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewAdsHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/shopee/ads/dashboard", nil)

		handler.GetDashboard(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant ID with no DB returns 500", func(t *testing.T) {
		handler := NewAdsHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/shopee/ads/dashboard", nil)
		c.Set("tenant_id", "test-tenant")

		handler.GetDashboard(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestNewAdsHandler_ShopeeAdsDashboard(t *testing.T) {
	t.Run("creates handler with base path", func(t *testing.T) {
		handler := NewAdsHandler("/some/path")
		assert.NotNil(t, handler)
		assert.Equal(t, "/some/path", handler.basePath)
	})

	t.Run("creates handler with empty base path", func(t *testing.T) {
		handler := NewAdsHandler("")
		assert.NotNil(t, handler)
		assert.Equal(t, "", handler.basePath)
	})
}
