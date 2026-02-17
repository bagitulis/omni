package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

	tmpDir := t.TempDir()
	createFrontendAPIMock(t, tmpDir, `
import api from "./client"
export async function run() {
  await api.get("/orders")
}
`)

	handler := NewRouteHandler(tmpDir)
	r.GET("/api/orders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})
	handler.SetEngine(r)
	r.GET("/api/routes/analyze", handler.AnalyzeRoutes)

	req, _ := http.NewRequest("GET", "/api/routes/analyze", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			TotalRoutes       int            `json:"total_routes"`
			ByMethod          map[string]int `json:"by_method"`
			ByTag             map[string]int `json:"by_tag"`
			Conflicts         []interface{}  `json:"conflicts"`
			UnprotectedRoutes []string       `json:"unprotected_routes"`
		} `json:"data"`
	}

	requireDecodeJSON(t, w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	assert.GreaterOrEqual(t, response.Data.TotalRoutes, 1)
	assert.NotNil(t, response.Data.ByMethod)
	assert.NotNil(t, response.Data.ByTag)
	assert.NotNil(t, response.Data.Conflicts)
	assert.NotNil(t, response.Data.UnprotectedRoutes)
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
	assert.Contains(t, w.Body.String(), "files_scanned")
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

// TestRouteHandler_GetRouteCoverage tests GetRouteCoverage
func TestRouteHandler_GetRouteCoverage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpDir := t.TempDir()
	createFrontendAPIMock(t, tmpDir, `
import api from "./client"
export async function run() {
  await api.get("/orders")
  await api.delete("/orders")
}
`)

	handler := NewRouteHandler(tmpDir)
	r.GET("/api/orders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})
	handler.SetEngine(r)
	r.GET("/api/routes/mapping", handler.GetRouteCoverage)

	req, _ := http.NewRequest("GET", "/api/routes/mapping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			TotalRoutes    int    `json:"total_routes"`
			ConnectionRate string `json:"connection_rate"`
			Categories     struct {
				FrontendOnly []struct {
					Status string `json:"status"`
				} `json:"frontend_only"`
			} `json:"categories"`
		} `json:"data"`
	}

	requireDecodeJSON(t, w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	assert.GreaterOrEqual(t, response.Data.TotalRoutes, 1)
	assert.NotEmpty(t, response.Data.ConnectionRate)
	assert.NotEmpty(t, response.Data.Categories.FrontendOnly)
	assert.Equal(t, "method_mismatch", response.Data.Categories.FrontendOnly[0].Status)
}

// TestRouteHandler_GetRouteCoverageStatistics tests GetRouteCoverageStatistics
func TestRouteHandler_GetRouteCoverageStatistics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpDir := t.TempDir()
	createFrontendAPIMock(t, tmpDir, `
import api from "./client"
export async function run() {
  await api.get("/orders")
  await api.delete("/orders")
}
`)

	handler := NewRouteHandler(tmpDir)
	r.GET("/api/orders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})
	handler.SetEngine(r)
	r.GET("/api/routes/mapping/stats", handler.GetRouteCoverageStatistics)

	req, _ := http.NewRequest("GET", "/api/routes/mapping/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Statistics        map[string]interface{} `json:"statistics"`
			TotalCalledRoutes int                    `json:"total_called_routes"`
			TotalDisconnected int                    `json:"total_disconnected"`
		} `json:"data"`
	}

	requireDecodeJSON(t, w.Body.Bytes(), &response)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data.Statistics)
	assert.Equal(t, 2, response.Data.TotalCalledRoutes)
	assert.Equal(t, 1, response.Data.TotalDisconnected)
	assert.Equal(t, float64(1), response.Data.Statistics["method_mismatch_total"])
}

func requireDecodeJSON(t *testing.T, data []byte, target interface{}) {
	t.Helper()
	err := json.Unmarshal(data, target)
	assert.NoError(t, err)
}

func createFrontendAPIMock(t *testing.T, basePath, content string) {
	t.Helper()
	apiDir := filepath.Join(basePath, "frontend", "src", "api")
	err := os.MkdirAll(apiDir, 0o755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(apiDir, "route_mapping_mock.ts"), []byte(content), 0o644)
	assert.NoError(t, err)
}
