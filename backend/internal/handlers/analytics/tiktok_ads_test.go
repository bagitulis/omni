package analytics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTiktokAdsHandler_GetData(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewTiktokAdsHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/tiktok/ads", nil)

		handler.GetData(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewTiktokAdsHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/tiktok/ads", nil)
		c.Set("tenantID", "test-tenant")

		handler.GetData(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestTiktokAdsHandler_GetUploads(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewTiktokAdsHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/tiktok/ads/uploads", nil)

		handler.GetUploads(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewTiktokAdsHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/tiktok/ads/uploads", nil)
		c.Set("tenantID", "test-tenant")

		handler.GetUploads(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestTiktokAdsHandler_Upload(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewTiktokAdsHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/tiktok/ads/upload", nil)

		handler.Upload(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant returns 501 (not implemented)", func(t *testing.T) {
		handler := NewTiktokAdsHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/tiktok/ads/upload", nil)
		c.Set("tenantID", "test-tenant")

		handler.Upload(c)

		// Upload is not yet implemented - returns 501
		assert.Equal(t, http.StatusNotImplemented, w.Code)
	})
}

func TestNewTiktokAdsHandler_Creation(t *testing.T) {
	t.Run("creates handler with base path", func(t *testing.T) {
		handler := NewTiktokAdsHandler("/some/path")
		assert.NotNil(t, handler)
		assert.Equal(t, "/some/path", handler.basePath)
	})

	t.Run("creates handler with empty base path", func(t *testing.T) {
		handler := NewTiktokAdsHandler("")
		assert.NotNil(t, handler)
	})
}
