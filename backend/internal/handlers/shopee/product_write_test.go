package shopee

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestProductHandler_CreateProduct_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.POST("/api/shopee/products", handler.CreateProduct)

	body := `{"name": "Test Product", "original_price": 10.0, "weight": 0.5, "category_id": 1}`
	req, _ := http.NewRequest("POST", "/api/shopee/products", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductHandler_CreateProduct_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/products", handler.CreateProduct)

	body := `{invalid json}`
	req, _ := http.NewRequest("POST", "/api/shopee/products", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductHandler_CreateProduct_ValidRequestNoCreds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/products", handler.CreateProduct)

	body := `{"name": "Test Product", "original_price": 10.0, "weight": 0.5, "category_id": 1, "stock": 1}`
	req, _ := http.NewRequest("POST", "/api/shopee/products", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Shopee credentials not available in test environment
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductHandler_UpdateProduct_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.PUT("/api/shopee/products/:itemId", handler.UpdateProduct)

	body := `{"name": "Updated Product"}`
	req, _ := http.NewRequest("PUT", "/api/shopee/products/12345", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductHandler_UpdateProduct_InvalidItemID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.PUT("/api/shopee/products/:itemId", handler.UpdateProduct)

	body := `{"name": "Updated Product"}`
	req, _ := http.NewRequest("PUT", "/api/shopee/products/not-a-number", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductHandler_UpdateProduct_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.PUT("/api/shopee/products/:itemId", handler.UpdateProduct)

	body := `{invalid json}`
	req, _ := http.NewRequest("PUT", "/api/shopee/products/12345", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductHandler_DeleteProduct_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.DELETE("/api/shopee/products/:itemId", handler.DeleteProduct)

	req, _ := http.NewRequest("DELETE", "/api/shopee/products/12345", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductHandler_DeleteProduct_InvalidItemID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.DELETE("/api/shopee/products/:itemId", handler.DeleteProduct)

	req, _ := http.NewRequest("DELETE", "/api/shopee/products/not-a-number", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestNewProductHandler(t *testing.T) {
	handler := NewProductHandler("/some/path")
	assert.NotNil(t, handler)
	assert.Equal(t, "/some/path", handler.basePath)
}
