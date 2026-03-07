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

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandler_GetShopeeSettings(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/shopee/settings", nil)

		handler.GetShopeeSettings(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant ID with no DB returns error", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/shopee/settings", nil)
		c.Set("tenantID", "test-tenant")

		handler.GetShopeeSettings(c)

		// Without valid DB config, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_GetShopeeSyncStatus(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/shopee/sync-status", nil)

		handler.GetShopeeSyncStatus(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestHandler_GetTiktokSettings(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/tiktok/settings", nil)

		handler.GetTiktokSettings(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestHandler_GetTiktokSyncStatus(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/tiktok/sync-status", nil)

		handler.GetTiktokSyncStatus(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestHandler_SyncShopeeEscrow(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/shopee/sync", strings.NewReader(`{"month": 1, "year": 2024}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.SyncShopeeEscrow(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/shopee/sync", strings.NewReader(`invalid`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.SyncShopeeEscrow(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid month returns 400", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/shopee/sync", strings.NewReader(`{"month": 13, "year": 2024}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.SyncShopeeEscrow(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "Invalid month")
	})

	t.Run("invalid year returns 400", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/shopee/sync", strings.NewReader(`{"month": 6, "year": 2019}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.SyncShopeeEscrow(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "Invalid year")
	})

	t.Run("valid request returns 501 not implemented", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/shopee/sync", strings.NewReader(`{"month": 6, "year": 2024, "force_resync": true}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.SyncShopeeEscrow(c)

		assert.Equal(t, http.StatusNotImplemented, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestHandler_SyncTiktokEscrow(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/tiktok/sync", strings.NewReader(`{"month": 1, "year": 2024}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.SyncTiktokEscrow(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid request returns 501 not implemented", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/analytics/tiktok/sync", strings.NewReader(`{"month": 6, "year": 2024}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.SyncTiktokEscrow(c)

		assert.Equal(t, http.StatusNotImplemented, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestNewHandler(t *testing.T) {
	t.Run("creates handler with base path", func(t *testing.T) {
		handler := NewHandler("/some/path")
		assert.NotNil(t, handler)
		assert.Equal(t, "/some/path", handler.basePath)
	})
}
