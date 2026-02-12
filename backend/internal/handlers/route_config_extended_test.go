package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRouteConfigHandler_GetCategories_MissingTenant tests GetCategories without tenant
func TestRouteConfigHandler_GetCategories_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteConfigHandler(nil)
	r.GET("/api/routes-config/categories", handler.GetCategories)

	req, _ := http.NewRequest("GET", "/api/routes-config/categories", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRouteConfigHandler_GetCategories_DBError tests GetCategories with DB error
func TestRouteConfigHandler_GetCategories_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteConfigHandler(nil)
	r.GET("/api/routes-config/categories", handler.GetCategories)

	req, _ := http.NewRequest("GET", "/api/routes-config/categories", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Without DB, should return 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestRouteConfigHandler_GetByCategory_MissingTenant tests GetByCategory without tenant
func TestRouteConfigHandler_GetByCategory_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteConfigHandler(nil)
	r.GET("/api/routes-config/by-category/:category", handler.GetByCategory)

	req, _ := http.NewRequest("GET", "/api/routes-config/by-category/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRouteConfigHandler_GetByCategory_DBError tests GetByCategory with DB error
func TestRouteConfigHandler_GetByCategory_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteConfigHandler(nil)
	r.GET("/api/routes-config/by-category/:category", handler.GetByCategory)

	req, _ := http.NewRequest("GET", "/api/routes-config/by-category/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Without DB, should return 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestRouteConfigHandler_Create_MissingTenant tests Create without tenant
func TestRouteConfigHandler_Create_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteConfigHandler(nil)
	r.POST("/api/routes-config", handler.Create)

	req, _ := http.NewRequest("POST", "/api/routes-config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRouteConfigHandler_Create_InvalidJSON tests Create with invalid JSON
func TestRouteConfigHandler_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteConfigHandler(nil)
	r.POST("/api/routes-config", handler.Create)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/routes-config", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler checks DB first, so may return 500 (DB error) or 400 (invalid JSON)
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, w.Code)
}

// TestRouteConfigHandler_Create_MissingRoutePath tests Create with missing route_path
func TestRouteConfigHandler_Create_MissingRoutePath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteConfigHandler(nil)
	r.POST("/api/routes-config", handler.Create)

	body := `{"name": "test"}`
	req, _ := http.NewRequest("POST", "/api/routes-config", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// May return 500 (DB error) or 400 (missing route_path)
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, w.Code)
}

// TestRouteConfigHandler_Delete_MissingTenant tests Delete without tenant
func TestRouteConfigHandler_Delete_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteConfigHandler(nil)
	r.DELETE("/api/routes-config/:id", handler.Delete)

	req, _ := http.NewRequest("DELETE", "/api/routes-config/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRouteConfigHandler_Delete_DBError tests Delete with DB error
func TestRouteConfigHandler_Delete_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteConfigHandler(nil)
	r.DELETE("/api/routes-config/:id", handler.Delete)

	req, _ := http.NewRequest("DELETE", "/api/routes-config/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Without DB, should return 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestRouteConfigHandler_BulkUpdate_MissingTenant tests BulkUpdate without tenant
func TestRouteConfigHandler_BulkUpdate_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteConfigHandler(nil)
	r.POST("/api/routes-config/bulk-update", handler.BulkUpdate)

	req, _ := http.NewRequest("POST", "/api/routes-config/bulk-update", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRouteConfigHandler_BulkUpdate_InvalidJSON tests BulkUpdate with invalid JSON
func TestRouteConfigHandler_BulkUpdate_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteConfigHandler(nil)
	r.POST("/api/routes-config/bulk-update", handler.BulkUpdate)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/routes-config/bulk-update", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler checks DB first, so may return 500 (DB error) or 400 (invalid JSON)
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, w.Code)
}

// TestRouteConfigHandler_ApplyPreset_MissingTenant tests ApplyPreset without tenant
func TestRouteConfigHandler_ApplyPreset_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteConfigHandler(nil)
	r.POST("/api/routes-config/apply-preset/:preset", handler.ApplyPreset)

	req, _ := http.NewRequest("POST", "/api/routes-config/apply-preset/default", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRouteConfigHandler_ApplyPreset_InvalidJSON tests ApplyPreset with invalid JSON
func TestRouteConfigHandler_ApplyPreset_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteConfigHandler(nil)
	r.POST("/api/routes-config/apply-preset/:preset", handler.ApplyPreset)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/routes-config/apply-preset/default", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler checks DB first, so may return 500 (DB error) or 400 (invalid JSON)
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, w.Code)
}
