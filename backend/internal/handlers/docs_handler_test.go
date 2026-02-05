package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestGetAPIInfo_Success tests the API info endpoint
func TestGetAPIInfo_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	docsHandler := NewDocsHandler()
	r.GET("/api/docs", docsHandler.GetAPIInfo)

	req, _ := http.NewRequest("GET", "/api/docs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)

	// Check required fields
	assert.Equal(t, "Omni Backend API", data["name"])
	assert.Equal(t, "1.0.0", data["version"])
	assert.Contains(t, data["description"], "Multi-platform")
	assert.NotEmpty(t, data["docs"])
}

// TestGetAPIInfo_ContainsEndpoints tests that API info includes endpoint listing
func TestGetAPIInfo_ContainsEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	docsHandler := NewDocsHandler()
	r.GET("/api/docs", docsHandler.GetAPIInfo)

	req, _ := http.NewRequest("GET", "/api/docs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	endpoints, ok := data["endpoints"].(map[string]interface{})
	assert.True(t, ok)

	// Check that platform endpoints are listed
	assert.NotEmpty(t, endpoints["health"])
	assert.NotEmpty(t, endpoints["auth"])
	assert.NotEmpty(t, endpoints["shopee"])
	assert.NotEmpty(t, endpoints["lazada"])
	assert.NotEmpty(t, endpoints["tiktok"])
}

// TestGetAPIInfo_ContainsAuthInfo tests that API info includes authentication details
func TestGetAPIInfo_ContainsAuthInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	docsHandler := NewDocsHandler()
	r.GET("/api/docs", docsHandler.GetAPIInfo)

	req, _ := http.NewRequest("GET", "/api/docs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	auth, ok := data["authentication"].(map[string]interface{})
	assert.True(t, ok)

	assert.Contains(t, auth["type"], "Bearer")
	assert.Contains(t, auth["header"], "Authorization")
}

// TestGetAPIInfo_ContainsHeaders tests that API info includes required headers
func TestGetAPIInfo_ContainsHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	docsHandler := NewDocsHandler()
	r.GET("/api/docs", docsHandler.GetAPIInfo)

	req, _ := http.NewRequest("GET", "/api/docs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	headers, ok := data["headers"].(map[string]interface{})
	assert.True(t, ok)

	assert.NotEmpty(t, headers["x-tenant-id"])
	assert.NotEmpty(t, headers["x-csrf-token"])
	assert.NotEmpty(t, headers["x-request-id"])
}

// TestGetSwaggerJSON_Success tests the Swagger JSON endpoint
func TestGetSwaggerJSON_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	docsHandler := NewDocsHandler()
	r.GET("/api/docs/swagger", docsHandler.GetSwaggerJSON)

	req, _ := http.NewRequest("GET", "/api/docs/swagger", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	// Check OpenAPI specification fields
	assert.Equal(t, "3.0.3", resp["openapi"])
	assert.NotNil(t, resp["info"])
	assert.NotNil(t, resp["paths"])
	assert.NotNil(t, resp["components"])
}

// TestGetSwaggerJSON_ContainsInfo tests Swagger JSON contains API info
func TestGetSwaggerJSON_ContainsInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	docsHandler := NewDocsHandler()
	r.GET("/api/docs/swagger", docsHandler.GetSwaggerJSON)

	req, _ := http.NewRequest("GET", "/api/docs/swagger", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	info := resp["info"].(map[string]interface{})
	assert.Equal(t, "Omni Backend API", info["title"])
	assert.Equal(t, "1.0.0", info["version"])
	assert.Contains(t, info["description"], "Multi-platform")
}

// TestGetSwaggerJSON_ContainsPaths tests Swagger JSON contains API paths
func TestGetSwaggerJSON_ContainsPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	docsHandler := NewDocsHandler()
	r.GET("/api/docs/swagger", docsHandler.GetSwaggerJSON)

	req, _ := http.NewRequest("GET", "/api/docs/swagger", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	paths := resp["paths"].(map[string]interface{})
	assert.NotNil(t, paths["/health"])
	assert.NotNil(t, paths["/auth/login"])
	assert.NotNil(t, paths["/shopee/orders"])
}

// TestGetSwaggerJSON_ContainsSecuritySchemes tests Swagger JSON contains security info
func TestGetSwaggerJSON_ContainsSecuritySchemes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	docsHandler := NewDocsHandler()
	r.GET("/api/docs/swagger", docsHandler.GetSwaggerJSON)

	req, _ := http.NewRequest("GET", "/api/docs/swagger", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	components := resp["components"].(map[string]interface{})
	securitySchemes := components["securitySchemes"].(map[string]interface{})
	bearerAuth := securitySchemes["bearerAuth"].(map[string]interface{})

	assert.Equal(t, "http", bearerAuth["type"])
	assert.Equal(t, "bearer", bearerAuth["scheme"])
	assert.Equal(t, "JWT", bearerAuth["bearerFormat"])
}

// TestDocsHandler_NewDocsHandler tests handler creation
func TestDocsHandler_NewDocsHandler(t *testing.T) {
	handler := NewDocsHandler()
	assert.NotNil(t, handler)
}
