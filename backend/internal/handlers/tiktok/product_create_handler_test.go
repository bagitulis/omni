package tiktok

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestProductCreateHandler_Constructor(t *testing.T) {
	handler := NewProductCreateHandler("/data/path")
	assert.NotNil(t, handler)
	assert.Equal(t, "/data/path", handler.basePath)
}

func TestProductCreateHandler_SaveDraft_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.POST("/api/tiktok/products/draft", handler.SaveDraft)

	body := `{"title": "Test Product"}`
	req, _ := http.NewRequest("POST", "/api/tiktok/products/draft", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductCreateHandler_SaveDraft_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/tiktok/products/draft", handler.SaveDraft)

	body := `{invalid json}`
	req, _ := http.NewRequest("POST", "/api/tiktok/products/draft", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductCreateHandler_SaveDraft_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/tiktok/products/draft", handler.SaveDraft)

	body := `{"title": "My Draft Product", "description": "A test product"}`
	req, _ := http.NewRequest("POST", "/api/tiktok/products/draft", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, data["draftId"])
}

func TestProductCreateHandler_SaveDraft_WithExistingID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/tiktok/products/draft", handler.SaveDraft)

	body := `{"id": "existing-draft-123", "title": "Existing Draft"}`
	req, _ := http.NewRequest("POST", "/api/tiktok/products/draft", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "existing-draft-123", data["draftId"])
}

func TestProductCreateHandler_PublishDraft_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.POST("/api/tiktok/products/publish/:draftId", handler.PublishDraft)

	req, _ := http.NewRequest("POST", "/api/tiktok/products/publish/draft-001", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductCreateHandler_PublishDraft_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/tiktok/products/publish/:draftId", handler.PublishDraft)

	req, _ := http.NewRequest("POST", "/api/tiktok/products/publish/draft-001", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "draft-001", data["draftId"])
}

func TestProductCreateHandler_GetProductsFromDB_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.GET("/api/tiktok/products/db", handler.GetProductsFromDB)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/db", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductCreateHandler_GetProductsFromDB_WithTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/tiktok/products/db", handler.GetProductsFromDB)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/db", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available in test environment
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestProductCreateHandler_GetProductFromDB_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.GET("/api/tiktok/products/db/:productId", handler.GetProductFromDB)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/db/PROD123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductCreateHandler_GetProductFromDB_WithTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/tiktok/products/db/:productId", handler.GetProductFromDB)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/db/PROD123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available in test environment
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestProductCreateHandler_SearchProducts_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.GET("/api/tiktok/products/search", handler.SearchProducts)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/search?q=test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductCreateHandler_SearchProducts_WithTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/tiktok/products/search", handler.SearchProducts)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/search?q=test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available in test environment
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestProductCreateHandler_GetCategories_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.GET("/api/tiktok/products/categories", handler.GetCategories)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/categories", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductCreateHandler_GetCategories_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/tiktok/products/categories", handler.GetCategories)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/categories", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	categories, ok := data["categories"].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, categories)
}

func TestProductCreateHandler_GetCategories_WithParentID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/tiktok/products/categories", handler.GetCategories)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/categories?parentId=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
}

func TestProductCreateHandler_GetAttributes_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.GET("/api/tiktok/products/categories/:categoryId/attributes", handler.GetAttributes)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/categories/cat-001/attributes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductCreateHandler_GetAttributes_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/tiktok/products/categories/:categoryId/attributes", handler.GetAttributes)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/categories/cat-001/attributes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	_, hasAttributes := data["attributes"]
	assert.True(t, hasAttributes)
}

func TestProductCreateHandler_GetRules_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.GET("/api/tiktok/products/categories/:categoryId/rules", handler.GetRules)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/categories/cat-001/rules", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductCreateHandler_GetRules_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/tiktok/products/categories/:categoryId/rules", handler.GetRules)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/categories/cat-001/rules", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
}

func TestProductCreateHandler_GetBrands_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.GET("/api/tiktok/products/brands", handler.GetBrands)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/brands", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductCreateHandler_GetBrands_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/tiktok/products/brands", handler.GetBrands)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/brands", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	brands, ok := data["brands"].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, brands)
}

func TestProductCreateHandler_GetDeliveryOptions_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.GET("/api/tiktok/products/delivery-options", handler.GetDeliveryOptions)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/delivery-options", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductCreateHandler_GetDeliveryOptions_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/tiktok/products/delivery-options", handler.GetDeliveryOptions)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/delivery-options", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	options, ok := data["options"].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, options)
}

func TestProductCreateHandler_GetWarehouses_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.GET("/api/tiktok/products/warehouses", handler.GetWarehouses)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/warehouses", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductCreateHandler_GetWarehouses_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductCreateHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/tiktok/products/warehouses", handler.GetWarehouses)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/warehouses", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	warehouses, ok := data["warehouses"].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, warehouses)
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	time.Sleep(2 * time.Millisecond) // ensure UnixNano differs
	id2 := generateID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	// IDs should be unique (generated from UnixNano)
	assert.NotEqual(t, id1, id2)
}

func TestGetMockCategories_NoParent(t *testing.T) {
	categories := getMockCategories("")
	assert.NotEmpty(t, categories)
	for _, cat := range categories {
		assert.Equal(t, 1, cat.Level)
		assert.False(t, cat.IsLeaf)
	}
}

func TestGetMockCategories_WithParent(t *testing.T) {
	categories := getMockCategories("parent-1")
	assert.NotEmpty(t, categories)
	for _, cat := range categories {
		assert.Equal(t, "parent-1", cat.ParentID)
		assert.Equal(t, 2, cat.Level)
		assert.True(t, cat.IsLeaf)
	}
}

func TestGetMockAttributes(t *testing.T) {
	attributes := getMockAttributes()
	assert.NotEmpty(t, attributes)
	// brand attribute should be required
	var brandAttr *Attribute
	for i, a := range attributes {
		if a.ID == "brand" {
			brandAttr = &attributes[i]
			break
		}
	}
	assert.NotNil(t, brandAttr)
	assert.True(t, brandAttr.Required)
}
