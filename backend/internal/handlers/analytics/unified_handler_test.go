package analytics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUnifiedHandler_GetUnifiedSummary(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewUnifiedHandler("", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/unified/summary", nil)

		handler.GetUnifiedSummary(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewUnifiedHandler("", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/unified/summary", nil)
		c.Set("tenant_id", "test-tenant")

		handler.GetUnifiedSummary(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUnifiedHandler_GetUnifiedKPI(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewUnifiedHandler("", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/unified/kpi", nil)

		handler.GetUnifiedKPI(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewUnifiedHandler("", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/unified/kpi", nil)
		c.Set("tenant_id", "test-tenant")

		handler.GetUnifiedKPI(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUnifiedHandler_RefreshCache(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewUnifiedHandler("", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/unified/cache/refresh", nil)

		handler.RefreshCache(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestUnifiedHandler_GetCacheStatus(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewUnifiedHandler("", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/unified/cache/status", nil)

		handler.GetCacheStatus(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestNewUnifiedHandler(t *testing.T) {
	t.Run("creates handler with nil cache", func(t *testing.T) {
		handler := NewUnifiedHandler("", nil)
		assert.NotNil(t, handler)
	})

	t.Run("creates handler with base path", func(t *testing.T) {
		handler := NewUnifiedHandler("/data/path", nil)
		assert.NotNil(t, handler)
		assert.Equal(t, "/data/path", handler.basePath)
	})
}
