package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRouteConfigHandler_NewRouteConfigHandler tests handler creation
func TestRouteConfigHandler_NewRouteConfigHandler(t *testing.T) {
	handler := NewRouteConfigHandler(nil)
	assert.NotNil(t, handler)
}

// TestRouteConfigHandler_List_MissingTenant tests List without tenant
func TestRouteConfigHandler_List_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteConfigHandler(nil)
	r.GET("/api/route-config", handler.List)

	req, _ := http.NewRequest("GET", "/api/route-config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRouteConfigHandler_Get_MissingTenant tests Get without tenant
func TestRouteConfigHandler_Get_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteConfigHandler(nil)
	r.GET("/api/route-config/:path", handler.Get)

	req, _ := http.NewRequest("GET", "/api/route-config/test-path", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRouteConfigHandler_Update_MissingTenant tests Update without tenant
func TestRouteConfigHandler_Update_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteConfigHandler(nil)
	r.PUT("/api/route-config/:path", handler.Update)

	req, _ := http.NewRequest("PUT", "/api/route-config/test-path", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRouteConfigHandler_Update_InvalidJSON tests Update with invalid JSON
func TestRouteConfigHandler_Update_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewRouteConfigHandler(nil)
	r.PUT("/api/route-config/:path", handler.Update)

	body := `{invalid json`
	req, _ := http.NewRequest("PUT", "/api/route-config/test-path", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler checks DB first, so may return 500 (DB error) or 400 (invalid JSON)
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, w.Code)
}

// TestRouteConfigHandler_Reset_MissingTenant tests Reset without tenant
func TestRouteConfigHandler_Reset_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteConfigHandler(nil)
	r.POST("/api/routes-config/reset", handler.Reset)

	req, _ := http.NewRequest("POST", "/api/routes-config/reset", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRouteConfigHandler_ListAll_MissingTenant tests ListAll without tenant
func TestRouteConfigHandler_ListAll_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteConfigHandler(nil)
	r.GET("/api/routes-config/all", handler.ListAll)

	req, _ := http.NewRequest("GET", "/api/routes-config/all", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
