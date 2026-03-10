package lazada

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestProductExtendedHandler_New(t *testing.T) {
	h := NewProductExtendedHandler("/test/path")
	assert.NotNil(t, h)
	assert.Equal(t, "/test/path", h.basePath)
}

// ---- GetProductsFromDB ----

func TestProductExtendedHandler_GetProductsFromDB_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewProductExtendedHandler("/test/path")
	r.GET("/api/lazada/products/db", h.GetProductsFromDB)

	req, _ := http.NewRequest("GET", "/api/lazada/products/db", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestProductExtendedHandler_GetProductsFromDB_WithTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewProductExtendedHandler("/test/path")
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.GET("/api/lazada/products/db", h.GetProductsFromDB)

	req, _ := http.NewRequest("GET", "/api/lazada/products/db?page=1&pageSize=20", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

// ---- GetProductFromDB ----

func TestProductExtendedHandler_GetProductFromDB_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewProductExtendedHandler("/test/path")
	r.GET("/api/lazada/products/db/:itemId", h.GetProductFromDB)

	req, _ := http.NewRequest("GET", "/api/lazada/products/db/ITEM123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestProductExtendedHandler_GetProductFromDB_WithTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewProductExtendedHandler("/test/path")
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.GET("/api/lazada/products/db/:itemId", h.GetProductFromDB)

	req, _ := http.NewRequest("GET", "/api/lazada/products/db/ITEM123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available → 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

// ---- GetCategories ----

func TestProductExtendedHandler_GetCategories_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewProductExtendedHandler("/test/path")
	r.GET("/api/lazada/products/categories", h.GetCategories)

	req, _ := http.NewRequest("GET", "/api/lazada/products/categories", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestProductExtendedHandler_GetCategories_WithTenantID_FailsWithoutDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewProductExtendedHandler("/test/path")
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.GET("/api/lazada/products/categories", h.GetCategories)

	req, _ := http.NewRequest("GET", "/api/lazada/products/categories", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should fail since no DB/credentials available
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

// ---- GetAttributes ----

func TestProductExtendedHandler_GetAttributes_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewProductExtendedHandler("/test/path")
	r.GET("/api/lazada/products/attributes/:categoryId", h.GetAttributes)

	req, _ := http.NewRequest("GET", "/api/lazada/products/attributes/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestProductExtendedHandler_GetAttributes_InvalidCategoryID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewProductExtendedHandler("/test/path")
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.GET("/api/lazada/products/attributes/:categoryId", h.GetAttributes)

	req, _ := http.NewRequest("GET", "/api/lazada/products/attributes/not-a-number", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestProductExtendedHandler_GetAttributes_WithTenantID_FailsWithoutDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewProductExtendedHandler("/test/path")
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.GET("/api/lazada/products/attributes/:categoryId", h.GetAttributes)

	req, _ := http.NewRequest("GET", "/api/lazada/products/attributes/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should fail since no DB/credentials available
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}
