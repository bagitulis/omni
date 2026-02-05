package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRouteHandler_NewRouteHandler tests handler creation
func TestRouteHandler_NewRouteHandler(t *testing.T) {
	handler := NewRouteHandler("")
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.mappingService)
	assert.NotNil(t, handler.scannerService)
}

// TestRouteHandler_GetAllRoutes tests GetAllRoutes
func TestRouteHandler_GetAllRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteHandler("")
	r.GET("/api/routes", handler.GetAllRoutes)

	req, _ := http.NewRequest("GET", "/api/routes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "total")
}

// TestRouteHandler_GetRoutesByTag tests GetRoutesByTag
func TestRouteHandler_GetRoutesByTag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteHandler("")
	r.GET("/api/routes/tag/:tag", handler.GetRoutesByTag)

	req, _ := http.NewRequest("GET", "/api/routes/tag/orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "tag")
	assert.Contains(t, w.Body.String(), "orders")
}

// TestRouteHandler_AnalyzeRoutes tests AnalyzeRoutes
func TestRouteHandler_AnalyzeRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteHandler("")
	r.GET("/api/routes/analyze", handler.AnalyzeRoutes)

	req, _ := http.NewRequest("GET", "/api/routes/analyze", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

// TestRouteHandler_ScanRoutes tests ScanRoutes
func TestRouteHandler_ScanRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteHandler("")
	r.GET("/api/routes/scan", handler.ScanRoutes)

	req, _ := http.NewRequest("GET", "/api/routes/scan", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "filesScanned")
}

// TestRouteHandler_ScanMiddleware tests ScanMiddleware
func TestRouteHandler_ScanMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteHandler("")
	r.GET("/api/routes/middleware", handler.ScanMiddleware)

	req, _ := http.NewRequest("GET", "/api/routes/middleware", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

// TestRouteHandler_FormatAsTable tests FormatAsTable
func TestRouteHandler_FormatAsTable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRouteHandler("")
	r.GET("/api/routes/table", handler.FormatAsTable)

	req, _ := http.NewRequest("GET", "/api/routes/table", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Returns plain text markdown table
}
