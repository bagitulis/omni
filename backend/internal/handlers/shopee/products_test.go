package shopee

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestProductHandler_GetProducts_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.GET("/api/shopee/products", handler.GetProducts)

	req, _ := http.NewRequest("GET", "/api/shopee/products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductHandler_GetProducts_WithTenantID_NoDB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/shopee/products", handler.GetProducts)

	req, _ := http.NewRequest("GET", "/api/shopee/products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available in test environment
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductHandler_GetProducts_WithQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/shopee/products", handler.GetProducts)

	req, _ := http.NewRequest("GET", "/api/shopee/products?itemStatus=NORMAL&offset=0&limit=20", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available in test environment
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductHandler_GetProductByID_InvalidItemID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/shopee/products/:itemId", handler.GetProductByID)

	req, _ := http.NewRequest("GET", "/api/shopee/products/not-a-number", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductHandler_GetProductByID_ValidID_NoDB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/shopee/products/:itemId", handler.GetProductByID)

	req, _ := http.NewRequest("GET", "/api/shopee/products/12345", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available — expect 500 or 404
	assert.True(t, w.Code == http.StatusInternalServerError || w.Code == http.StatusNotFound)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}
