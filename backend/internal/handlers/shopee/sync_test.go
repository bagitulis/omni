package shopee

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSyncHandler_SyncOrders_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewSyncHandler("/tmp/test-sync")
	r.POST("/sync/orders", handler.SyncOrders)

	req, _ := http.NewRequest("POST", "/sync/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestSyncHandler_SyncOrders_WithTenantID_NoDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	handler := NewSyncHandler("/tmp/nonexistent-path-sync")
	r.POST("/sync/orders", handler.SyncOrders)

	req, _ := http.NewRequest("POST", "/sync/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should fail at DB level since no real DB exists at that path
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestSyncHandler_SyncOrders_DaysQueryParam_Valid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	handler := NewSyncHandler("/tmp/nonexistent-path-sync")
	r.POST("/sync/orders", handler.SyncOrders)

	// days=15 is valid (1-30 range)
	req, _ := http.NewRequest("POST", "/sync/orders?days=15", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Will fail at DB but should not return 400 (days param is valid)
	assert.NotEqual(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestSyncHandler_SyncOrders_DaysQueryParam_OutOfRange(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	handler := NewSyncHandler("/tmp/nonexistent-path-sync")
	r.POST("/sync/orders", handler.SyncOrders)

	// days=999 is out of range — handler clamps it to 7 and proceeds to DB
	req, _ := http.NewRequest("POST", "/sync/orders?days=999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should reach DB check (not short-circuit at days validation)
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestSyncHandler_SyncProducts_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewSyncHandler("/tmp/test-sync")
	r.POST("/sync/products", handler.SyncProducts)

	req, _ := http.NewRequest("POST", "/sync/products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestSyncHandler_SyncProducts_WithTenantID_NoDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	handler := NewSyncHandler("/tmp/nonexistent-path-sync")
	r.POST("/sync/products", handler.SyncProducts)

	req, _ := http.NewRequest("POST", "/sync/products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should fail at DB level
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestNewSyncHandlerWithCache_NotNil(t *testing.T) {
	handler := NewSyncHandlerWithCache("/tmp/test-path", nil)
	assert.NotNil(t, handler)
	assert.Equal(t, "/tmp/test-path", handler.basePath)
}

func TestNewSyncHandler_NotNil(t *testing.T) {
	handler := NewSyncHandler("/tmp/test-path")
	assert.NotNil(t, handler)
	assert.Equal(t, "/tmp/test-path", handler.basePath)
}
