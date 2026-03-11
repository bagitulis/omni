package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInventoryHandler_GetStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/stats", nil)
		// No tenantID set

		handler.GetStats(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "Missing tenant_id")
	})

	t.Run("valid tenant ID with no DB returns error", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/stats", nil)
		c.Set("tenant_id", "test-tenant")

		handler.GetStats(c)

		// Without a valid DB, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestInventoryHandler_GetPlatformStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/platform-status", nil)
		// No tenantID set

		handler.GetPlatformStatus(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "Missing tenant_id")
	})

	t.Run("valid tenant ID returns empty results", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/platform-status", nil)
		c.Set("tenant_id", "test-tenant")

		handler.GetPlatformStatus(c)

		// This endpoint returns a static response without DB access
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, true, resp["success"])
		assert.Equal(t, float64(0), resp["count"])
	})
}
