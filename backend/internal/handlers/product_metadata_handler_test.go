package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestProductMetadataHandler_GetShopeeCategories(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing_tenant_id", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.GET("/api/products/create/shopee/categories", handler.GetShopeeCategories)

		req, _ := http.NewRequest("GET", "/api/products/create/shopee/categories", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid_tenant_returns_categories", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.Use(func(c *gin.Context) {
			c.Set("tenantID", "test-tenant")
			c.Next()
		})
		r.GET("/api/products/create/shopee/categories", handler.GetShopeeCategories)

		req, _ := http.NewRequest("GET", "/api/products/create/shopee/categories", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
		data := resp["data"].(map[string]interface{})
		assert.NotNil(t, data["categories"])
	})
}

func TestProductMetadataHandler_GetShopeeAttributes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing_tenant_id", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.GET("/api/products/create/shopee/attributes/:catId", handler.GetShopeeAttributes)

		req, _ := http.NewRequest("GET", "/api/products/create/shopee/attributes/123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid_tenant_returns_attributes", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.Use(func(c *gin.Context) {
			c.Set("tenantID", "test-tenant")
			c.Next()
		})
		r.GET("/api/products/create/shopee/attributes/:catId", handler.GetShopeeAttributes)

		req, _ := http.NewRequest("GET", "/api/products/create/shopee/attributes/123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
	})
}

func TestProductMetadataHandler_GetShopeeBrands(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing_tenant_id", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.GET("/api/products/create/shopee/brands/:catId", handler.GetShopeeBrands)

		req, _ := http.NewRequest("GET", "/api/products/create/shopee/brands/123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid_tenant_returns_brands", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.Use(func(c *gin.Context) {
			c.Set("tenantID", "test-tenant")
			c.Next()
		})
		r.GET("/api/products/create/shopee/brands/:catId", handler.GetShopeeBrands)

		req, _ := http.NewRequest("GET", "/api/products/create/shopee/brands/123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
	})
}

func TestProductMetadataHandler_GetShopeeLogistics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing_tenant_id", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.GET("/api/products/create/shopee/logistics", handler.GetShopeeLogistics)

		req, _ := http.NewRequest("GET", "/api/products/create/shopee/logistics", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid_tenant_returns_logistics", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.Use(func(c *gin.Context) {
			c.Set("tenantID", "test-tenant")
			c.Next()
		})
		r.GET("/api/products/create/shopee/logistics", handler.GetShopeeLogistics)

		req, _ := http.NewRequest("GET", "/api/products/create/shopee/logistics", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestProductMetadataHandler_GetLazadaCategories(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid_tenant_returns_categories", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.Use(func(c *gin.Context) {
			c.Set("tenantID", "test-tenant")
			c.Next()
		})
		r.GET("/api/products/create/lazada/categories", handler.GetLazadaCategories)

		req, _ := http.NewRequest("GET", "/api/products/create/lazada/categories", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
	})
}

func TestProductMetadataHandler_GetTiktokCategories(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid_tenant_returns_categories", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.Use(func(c *gin.Context) {
			c.Set("tenantID", "test-tenant")
			c.Next()
		})
		r.GET("/api/products/create/tiktok/categories", handler.GetTiktokCategories)

		req, _ := http.NewRequest("GET", "/api/products/create/tiktok/categories", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestProductMetadataHandler_UploadImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing_tenant_id", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.POST("/api/products/create/upload-image", handler.UploadImage)

		body, _ := json.Marshal(map[string]interface{}{
			"platform":  "shopee",
			"image_url": "https://example.com/image.jpg",
		})
		req, _ := http.NewRequest("POST", "/api/products/create/upload-image", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid_request_returns_success", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.Use(func(c *gin.Context) {
			c.Set("tenantID", "test-tenant")
			c.Next()
		})
		r.POST("/api/products/create/upload-image", handler.UploadImage)

		body, _ := json.Marshal(map[string]interface{}{
			"platform":  "shopee",
			"image_url": "https://example.com/image.jpg",
		})
		req, _ := http.NewRequest("POST", "/api/products/create/upload-image", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
	})

	t.Run("missing_platform_returns_error", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.Use(func(c *gin.Context) {
			c.Set("tenantID", "test-tenant")
			c.Next()
		})
		r.POST("/api/products/create/upload-image", handler.UploadImage)

		body, _ := json.Marshal(map[string]interface{}{
			"image_url": "https://example.com/image.jpg",
		})
		req, _ := http.NewRequest("POST", "/api/products/create/upload-image", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestProductMetadataHandler_ValidateProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing_tenant_id", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.POST("/api/products/create/validate", handler.ValidateProduct)

		body, _ := json.Marshal(map[string]interface{}{
			"platform": "shopee",
			"product":  map[string]interface{}{"title": "Test"},
		})
		req, _ := http.NewRequest("POST", "/api/products/create/validate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid_product", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.Use(func(c *gin.Context) {
			c.Set("tenantID", "test-tenant")
			c.Next()
		})
		r.POST("/api/products/create/validate", handler.ValidateProduct)

		body, _ := json.Marshal(map[string]interface{}{
			"platform": "shopee",
			"product":  map[string]interface{}{"title": "Test Product"},
		})
		req, _ := http.NewRequest("POST", "/api/products/create/validate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
	})

	t.Run("missing_title_returns_validation_errors", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.Use(func(c *gin.Context) {
			c.Set("tenantID", "test-tenant")
			c.Next()
		})
		r.POST("/api/products/create/validate", handler.ValidateProduct)

		body, _ := json.Marshal(map[string]interface{}{
			"platform": "shopee",
			"product":  map[string]interface{}{"price": 10000},
		})
		req, _ := http.NewRequest("POST", "/api/products/create/validate", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		data := resp["data"].(map[string]interface{})
		assert.False(t, data["valid"].(bool))
	})
}

func TestProductMetadataHandler_GetTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing_tenant_id", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.GET("/api/products/create/templates", handler.GetTemplates)

		req, _ := http.NewRequest("GET", "/api/products/create/templates", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns_all_templates", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.Use(func(c *gin.Context) {
			c.Set("tenantID", "test-tenant")
			c.Next()
		})
		r.GET("/api/products/create/templates", handler.GetTemplates)

		req, _ := http.NewRequest("GET", "/api/products/create/templates", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp["success"].(bool))
	})

	t.Run("filters_by_platform", func(t *testing.T) {
		r := gin.New()
		handler := NewProductMetadataHandler("")

		r.Use(func(c *gin.Context) {
			c.Set("tenantID", "test-tenant")
			c.Next()
		})
		r.GET("/api/products/create/templates", handler.GetTemplates)

		req, _ := http.NewRequest("GET", "/api/products/create/templates?platform=shopee", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
