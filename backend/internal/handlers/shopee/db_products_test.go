package shopee

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDBProductHandler_GetDBProducts_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewDBProductHandler("/test/path")

	r.GET("/api/shopee/db/products", handler.GetDBProducts)

	req, _ := http.NewRequest("GET", "/api/shopee/db/products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestDBProductHandler_GetDBProducts_WithTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewDBProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/shopee/db/products", handler.GetDBProducts)

	req, _ := http.NewRequest("GET", "/api/shopee/db/products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available in test — expect 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestDBProductHandler_GetDBProducts_WithPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewDBProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/shopee/db/products", handler.GetDBProducts)

	req, _ := http.NewRequest("GET", "/api/shopee/db/products?offset=10&limit=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available in test — expect 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestDBProductHandler_GetMasterProducts_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewDBProductHandler("/test/path")

	r.GET("/api/shopee/db/products/master", handler.GetMasterProducts)

	req, _ := http.NewRequest("GET", "/api/shopee/db/products/master", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestDBProductHandler_GetMasterProducts_WithTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewDBProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/shopee/db/products/master", handler.GetMasterProducts)

	req, _ := http.NewRequest("GET", "/api/shopee/db/products/master", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available in test — expect 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestFlattenedSkuRow_StructFields(t *testing.T) {
	row := FlattenedSkuRow{
		ItemID:    "12345",
		ModelID:   "67890",
		SKU:       "SKU-001",
		ItemName:  "Test Product",
		SKUName:   "Red / Large",
		Image:     "https://example.com/image.jpg",
		Price:     99.99,
		Stock:     10,
		Status:    "NORMAL",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	assert.Equal(t, "12345", row.ItemID)
	assert.Equal(t, "67890", row.ModelID)
	assert.Equal(t, "SKU-001", row.SKU)
	assert.Equal(t, "Test Product", row.ItemName)
	assert.Equal(t, "Red / Large", row.SKUName)
	assert.Equal(t, 99.99, row.Price)
	assert.Equal(t, 10, row.Stock)
	assert.Equal(t, "NORMAL", row.Status)
}

func TestNewDBProductHandler(t *testing.T) {
	handler := NewDBProductHandler("/some/path")
	assert.NotNil(t, handler)
	assert.Equal(t, "/some/path", handler.basePath)
}
