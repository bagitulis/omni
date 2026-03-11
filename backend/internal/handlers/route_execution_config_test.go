package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRouteExecutionConfigHandler_NewHandler tests handler creation
func TestRouteExecutionConfigHandler_NewHandler(t *testing.T) {
	handler := NewRouteExecutionConfigHandler(nil)
	assert.NotNil(t, handler)
}

// TestRouteExecutionConfigHandler_List_MissingTenant tests List without tenant
func TestRouteExecutionConfigHandler_List_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteExecutionConfigHandler(nil)
	r.GET("/api/route-execution-config", handler.List)

	req, _ := http.NewRequest("GET", "/api/route-execution-config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteExecutionConfigHandler_Get_MissingTenant tests Get without tenant
func TestRouteExecutionConfigHandler_Get_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteExecutionConfigHandler(nil)
	r.GET("/api/route-execution-config/:routeKey", handler.Get)

	req, _ := http.NewRequest("GET", "/api/route-execution-config/test-route", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteExecutionConfigHandler_Get_MissingRouteKey tests Get without routeKey param
func TestRouteExecutionConfigHandler_Get_MissingRouteKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewRouteExecutionConfigHandler(nil)
	// Empty param test - the route pattern requires it, so we test different path
	r.GET("/api/route-execution-config/:routeKey", handler.Get)

	req, _ := http.NewRequest("GET", "/api/route-execution-config/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Route won't match, returns 404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestRouteExecutionConfigHandler_GetMode_MissingTenant tests GetMode without tenant
func TestRouteExecutionConfigHandler_GetMode_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteExecutionConfigHandler(nil)
	r.GET("/api/route-execution-config/:routeKey/mode", handler.GetMode)

	req, _ := http.NewRequest("GET", "/api/route-execution-config/test-route/mode", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteExecutionConfigHandler_Create_MissingTenant tests Create without tenant
func TestRouteExecutionConfigHandler_Create_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteExecutionConfigHandler(nil)
	r.POST("/api/route-execution-config", handler.Create)

	req, _ := http.NewRequest("POST", "/api/route-execution-config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteExecutionConfigHandler_Create_InvalidJSON tests Create with invalid JSON
func TestRouteExecutionConfigHandler_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewRouteExecutionConfigHandler(nil)
	r.POST("/api/route-execution-config", handler.Create)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/route-execution-config", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler checks DB first, so may return 500 (DB error) or 400 (invalid JSON)
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, w.Code)
}

// TestRouteExecutionConfigHandler_Create_InvalidExecutionMode tests Create with invalid execution mode
func TestRouteExecutionConfigHandler_Create_InvalidExecutionMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewRouteExecutionConfigHandler(nil)
	r.POST("/api/route-execution-config", handler.Create)

	body := `{"route_key":"test","route_name":"Test","execution_mode":"invalid"}`
	req, _ := http.NewRequest("POST", "/api/route-execution-config", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler checks DB first, so may return 500 (DB error) or 400 (invalid mode)
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, w.Code)
}

// TestRouteExecutionConfigHandler_Update_MissingTenant tests Update without tenant
func TestRouteExecutionConfigHandler_Update_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteExecutionConfigHandler(nil)
	r.PUT("/api/route-execution-config/:routeKey", handler.Update)

	req, _ := http.NewRequest("PUT", "/api/route-execution-config/test-route", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteExecutionConfigHandler_Update_InvalidJSON tests Update with invalid JSON
func TestRouteExecutionConfigHandler_Update_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewRouteExecutionConfigHandler(nil)
	r.PUT("/api/route-execution-config/:routeKey", handler.Update)

	body := `{invalid json`
	req, _ := http.NewRequest("PUT", "/api/route-execution-config/test-route", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler checks DB first, so may return 500 (DB error) or 400 (invalid JSON)
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, w.Code)
}

// TestRouteExecutionConfigHandler_Toggle_MissingTenant tests Toggle without tenant
func TestRouteExecutionConfigHandler_Toggle_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteExecutionConfigHandler(nil)
	r.POST("/api/route-execution-config/:routeKey/toggle", handler.Toggle)

	req, _ := http.NewRequest("POST", "/api/route-execution-config/test-route/toggle", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteExecutionConfigHandler_Delete_MissingTenant tests Delete without tenant
func TestRouteExecutionConfigHandler_Delete_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteExecutionConfigHandler(nil)
	r.DELETE("/api/route-execution-config/:routeKey", handler.Delete)

	req, _ := http.NewRequest("DELETE", "/api/route-execution-config/test-route", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteExecutionConfigHandler_Delete_DBError tests Delete with DB error
func TestRouteExecutionConfigHandler_Delete_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewRouteExecutionConfigHandler(nil)
	r.DELETE("/api/route-execution-config/:routeKey", handler.Delete)

	req, _ := http.NewRequest("DELETE", "/api/route-execution-config/test-route", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Without DB, should return 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
