package tiktok

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Constructor tests ---

func TestNewSyncHandler(t *testing.T) {
	h := NewSyncHandler("/some/base/path")
	assert.NotNil(t, h)
	assert.Equal(t, "/some/base/path", h.basePath)
	assert.Nil(t, h.cacheService)
}

func TestNewSyncHandlerWithCache(t *testing.T) {
	h := NewSyncHandlerWithCache("/some/base/path", nil)
	assert.NotNil(t, h)
	assert.Equal(t, "/some/base/path", h.basePath)
	assert.Nil(t, h.cacheService)
}

func TestNewSyncHandlerWithCache_NilCacheDoesNotPanic(t *testing.T) {
	// Constructing with nil cache should not panic
	assert.NotPanics(t, func() {
		h := NewSyncHandlerWithCache("/test", nil)
		_ = h
	})
}

// --- SyncOrders handler tests ---

func TestSyncHandler_SyncOrders_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSyncHandler("/test/path")
	r.POST("/sync/orders", h.SyncOrders)

	req, _ := http.NewRequest(http.MethodPost, "/sync/orders", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.NotEmpty(t, resp["error"])
}

func TestSyncHandler_SyncOrders_WithTenant_ReturnsServerError(t *testing.T) {
	// In test environment there is no real DB.
	// config.GetTenantDB will fail → handler returns 500 (DB connection failed).
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	h := NewSyncHandler("/test/path")
	r.POST("/sync/orders", h.SyncOrders)

	req, _ := http.NewRequest(http.MethodPost, "/sync/orders", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Must NOT be 401 (tenant was provided)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	// DB unavailable in test env → 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.NotEmpty(t, resp["error"])
}

func TestSyncHandler_SyncOrders_ResponseIsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSyncHandler("/test/path")
	r.POST("/sync/orders", h.SyncOrders)

	req, _ := http.NewRequest(http.MethodPost, "/sync/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasSuccess := resp["success"]
	assert.True(t, hasSuccess, "response must have a 'success' field")
}

// --- SyncProducts handler tests ---

func TestSyncHandler_SyncProducts_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSyncHandler("/test/path")
	r.POST("/sync/products", h.SyncProducts)

	req, _ := http.NewRequest(http.MethodPost, "/sync/products", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.NotEmpty(t, resp["error"])
}

func TestSyncHandler_SyncProducts_WithTenant_ReturnsServerError(t *testing.T) {
	// config.GetTenantDB will fail in test env → 500
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	h := NewSyncHandler("/test/path")
	r.POST("/sync/products", h.SyncProducts)

	req, _ := http.NewRequest(http.MethodPost, "/sync/products", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.NotEmpty(t, resp["error"])
}

func TestSyncHandler_SyncProducts_ResponseIsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSyncHandler("/test/path")
	r.POST("/sync/products", h.SyncProducts)

	req, _ := http.NewRequest(http.MethodPost, "/sync/products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasSuccess := resp["success"]
	assert.True(t, hasSuccess, "response must have a 'success' field")
}

// --- invalidateAnalyticsCache with nil cache (no panic) ---

func TestSyncHandler_InvalidateAnalyticsCache_NilCache(t *testing.T) {
	h := NewSyncHandler("/test/path")
	// Should not panic when cacheService is nil
	assert.NotPanics(t, func() {
		h.invalidateAnalyticsCache("test-tenant", "sync_orders")
	})
}
