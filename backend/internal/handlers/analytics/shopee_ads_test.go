package analytics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAdsHandler_GetData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewAdsHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/shopee-ads/data", nil)

		handler.GetData(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant ID with no DB returns error", func(t *testing.T) {
		handler := NewAdsHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/shopee-ads/data?limit=50&offset=0", nil)
		c.Set("tenantID", "test-tenant")

		handler.GetData(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNewAdsHandler(t *testing.T) {
	t.Run("creates handler with base path", func(t *testing.T) {
		handler := NewAdsHandler("/test/path")
		assert.NotNil(t, handler)
		assert.Equal(t, "/test/path", handler.basePath)
	})
}
