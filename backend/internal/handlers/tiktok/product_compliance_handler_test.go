package tiktok

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestComplianceHandler_Constructor(t *testing.T) {
	handler := NewComplianceHandler("/data/path")
	assert.NotNil(t, handler)
	assert.Equal(t, "/data/path", handler.basePath)
}

func TestComplianceHandler_GetCompliance_Returns501(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewComplianceHandler("/test/path")
	r.GET("/api/tiktok/products/:productId/compliance", handler.GetCompliance)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/PROD123/compliance", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestComplianceHandler_UpdateCompliance_Returns501(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewComplianceHandler("/test/path")
	r.PUT("/api/tiktok/products/:productId/compliance", handler.UpdateCompliance)

	req, _ := http.NewRequest("PUT", "/api/tiktok/products/PROD123/compliance", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestComplianceHandler_PublishGlobal_Returns501(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewComplianceHandler("/test/path")
	r.POST("/api/tiktok/products/publish-global", handler.PublishGlobal)

	req, _ := http.NewRequest("POST", "/api/tiktok/products/publish-global", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestComplianceHandler_GetGlobalProducts_Returns501(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := NewComplianceHandler("/test/path")
	r.GET("/api/tiktok/products/global-products", handler.GetGlobalProducts)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/global-products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}
