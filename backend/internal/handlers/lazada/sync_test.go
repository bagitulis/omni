package lazada

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSyncHandler_New(t *testing.T) {
	h := NewSyncHandler("/test/path")
	assert.NotNil(t, h)
	assert.Equal(t, "/test/path", h.basePath)
	assert.Nil(t, h.cacheService)
}

func TestSyncHandlerWithCache_New(t *testing.T) {
	h := NewSyncHandlerWithCache("/test/path", nil)
	assert.NotNil(t, h)
	assert.Equal(t, "/test/path", h.basePath)
	assert.Nil(t, h.cacheService)
}

// ---- invalidateAnalyticsCache (nil safe) ----

func TestSyncHandler_InvalidateAnalyticsCache_NilService(t *testing.T) {
	h := NewSyncHandler("/test/path")
	// Should not panic when cacheService is nil
	assert.NotPanics(t, func() {
		h.invalidateAnalyticsCache("test-tenant", "sync_orders")
	})
}

// ---- SyncOrders ----

func TestSyncHandler_SyncOrders_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSyncHandler("/test/path")
	r.POST("/api/lazada/sync/orders", h.SyncOrders)

	req, _ := http.NewRequest("POST", "/api/lazada/sync/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestSyncHandler_SyncOrders_WithTenantID_NoDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSyncHandler("/test/path")
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/lazada/sync/orders", h.SyncOrders)

	req, _ := http.NewRequest("POST", "/api/lazada/sync/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available → 500 (database connection failed)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

// ---- SyncProducts ----

func TestSyncHandler_SyncProducts_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSyncHandler("/test/path")
	r.POST("/api/lazada/sync/products", h.SyncProducts)

	req, _ := http.NewRequest("POST", "/api/lazada/sync/products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestSyncHandler_SyncProducts_WithTenantID_NoDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSyncHandler("/test/path")
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/lazada/sync/products", h.SyncProducts)

	req, _ := http.NewRequest("POST", "/api/lazada/sync/products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available → 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}
