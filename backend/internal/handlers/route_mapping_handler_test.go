package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRouteMappingHandler_NewHandler tests handler creation
func TestRouteMappingHandler_NewHandler(t *testing.T) {
	handler := NewRouteMappingHandler(nil)
	assert.NotNil(t, handler)
}

// TestRouteMappingHandler_GetRoutes_MissingTenant tests GetRoutes without tenant
func TestRouteMappingHandler_GetRoutes_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteMappingHandler(nil)
	r.GET("/api/route-mapping/routes", handler.GetRoutes)

	req, _ := http.NewRequest("GET", "/api/route-mapping/routes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteMappingHandler_GetRoutes_WithTenant tests GetRoutes with tenant and nil engine
func TestRouteMappingHandler_GetRoutes_WithTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteMappingHandler(nil)
	r.GET("/api/route-mapping/routes", handler.GetRoutes)

	req, _ := http.NewRequest("GET", "/api/route-mapping/routes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "total")
}

// TestRouteMappingHandler_GetRoutesByPlatform_MissingTenant tests GetRoutesByPlatform without tenant
func TestRouteMappingHandler_GetRoutesByPlatform_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteMappingHandler(nil)
	r.GET("/api/route-mapping/routes/:platform", handler.GetRoutesByPlatform)

	req, _ := http.NewRequest("GET", "/api/route-mapping/routes/shopee", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteMappingHandler_GetRoutesByPlatform_WithTenant tests GetRoutesByPlatform with tenant
func TestRouteMappingHandler_GetRoutesByPlatform_WithTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteMappingHandler(nil)
	r.GET("/api/route-mapping/routes/:platform", handler.GetRoutesByPlatform)

	req, _ := http.NewRequest("GET", "/api/route-mapping/routes/shopee", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "platform")
	assert.Contains(t, w.Body.String(), "shopee")
}

// TestRouteMappingHandler_GetHandlers_MissingTenant tests GetHandlers without tenant
func TestRouteMappingHandler_GetHandlers_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteMappingHandler(nil)
	r.GET("/api/route-mapping/handlers", handler.GetHandlers)

	req, _ := http.NewRequest("GET", "/api/route-mapping/handlers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteMappingHandler_GetHandlers_WithTenant tests GetHandlers with tenant
func TestRouteMappingHandler_GetHandlers_WithTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteMappingHandler(nil)
	r.GET("/api/route-mapping/handlers", handler.GetHandlers)

	req, _ := http.NewRequest("GET", "/api/route-mapping/handlers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "handlers")
}

// TestRouteMappingHandler_GetServices_MissingTenant tests GetServices without tenant
func TestRouteMappingHandler_GetServices_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteMappingHandler(nil)
	r.GET("/api/route-mapping/services", handler.GetServices)

	req, _ := http.NewRequest("GET", "/api/route-mapping/services", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteMappingHandler_GetServices_WithTenant tests GetServices with tenant
func TestRouteMappingHandler_GetServices_WithTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteMappingHandler(nil)
	r.GET("/api/route-mapping/services", handler.GetServices)

	req, _ := http.NewRequest("GET", "/api/route-mapping/services", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "services")
}

// TestRouteMappingHandler_GetMiddleware_MissingTenant tests GetMiddleware without tenant
func TestRouteMappingHandler_GetMiddleware_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteMappingHandler(nil)
	r.GET("/api/route-mapping/middleware", handler.GetMiddleware)

	req, _ := http.NewRequest("GET", "/api/route-mapping/middleware", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteMappingHandler_GetMiddleware_WithTenant tests GetMiddleware with tenant
func TestRouteMappingHandler_GetMiddleware_WithTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteMappingHandler(nil)
	r.GET("/api/route-mapping/middleware", handler.GetMiddleware)

	req, _ := http.NewRequest("GET", "/api/route-mapping/middleware", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "middlewares")
}

// TestRouteMappingHandler_GetUnused_MissingTenant tests GetUnused without tenant
func TestRouteMappingHandler_GetUnused_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteMappingHandler(nil)
	r.GET("/api/route-mapping/unused", handler.GetUnused)

	req, _ := http.NewRequest("GET", "/api/route-mapping/unused", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteMappingHandler_GetUnused_WithTenant tests GetUnused with tenant
func TestRouteMappingHandler_GetUnused_WithTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteMappingHandler(nil)
	r.GET("/api/route-mapping/unused", handler.GetUnused)

	req, _ := http.NewRequest("GET", "/api/route-mapping/unused", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

// TestRouteMappingHandler_GetDuplicates_MissingTenant tests GetDuplicates without tenant
func TestRouteMappingHandler_GetDuplicates_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteMappingHandler(nil)
	r.GET("/api/route-mapping/duplicates", handler.GetDuplicates)

	req, _ := http.NewRequest("GET", "/api/route-mapping/duplicates", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteMappingHandler_GetDuplicates_WithTenant tests GetDuplicates with tenant
func TestRouteMappingHandler_GetDuplicates_WithTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteMappingHandler(nil)
	r.GET("/api/route-mapping/duplicates", handler.GetDuplicates)

	req, _ := http.NewRequest("GET", "/api/route-mapping/duplicates", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "duplicates")
}

// TestRouteMappingHandler_AnalyzeRoutes_MissingTenant tests AnalyzeRoutes without tenant
func TestRouteMappingHandler_AnalyzeRoutes_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteMappingHandler(nil)
	r.POST("/api/route-mapping/analyze", handler.AnalyzeRoutes)

	req, _ := http.NewRequest("POST", "/api/route-mapping/analyze", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRouteMappingHandler_AnalyzeRoutes_WithTenant tests AnalyzeRoutes with tenant
func TestRouteMappingHandler_AnalyzeRoutes_WithTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewRouteMappingHandler(nil)
	r.POST("/api/route-mapping/analyze", handler.AnalyzeRoutes)

	// Empty body is fine as request struct fields are optional
	req, _ := http.NewRequest("POST", "/api/route-mapping/analyze", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Response format includes "success":true
	assert.Contains(t, w.Body.String(), `"success":true`)
}
