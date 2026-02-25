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

func TestProductExtendedHandler_GetCategories_WithTenantID(t *testing.T) {
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

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
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

func TestProductExtendedHandler_GetAttributes_WithTenantID(t *testing.T) {
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

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
}

// ---- Category / Attribute / Option JSON tags ----

func TestCategory_JSONFields(t *testing.T) {
	cat := Category{
		ID:       100,
		Name:     "Electronics",
		ParentID: 0,
		Level:    1,
		IsLeaf:   false,
		Children: nil,
	}
	data, err := json.Marshal(cat)
	assert.NoError(t, err)

	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, float64(100), out["id"])
	assert.Equal(t, "Electronics", out["name"])
	assert.Equal(t, float64(1), out["level"])
	assert.Equal(t, false, out["is_leaf"])
}

func TestAttribute_JSONFields(t *testing.T) {
	attr := Attribute{
		Name:          "brand",
		Label:         "Brand",
		InputType:     "singleSelect",
		IsMandatory:   true,
		IsSaleProp:    false,
		AttributeType: "normal",
	}
	data, err := json.Marshal(attr)
	assert.NoError(t, err)

	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, "brand", out["name"])
	assert.Equal(t, "Brand", out["label"])
	assert.Equal(t, "singleSelect", out["input_type"])
	assert.Equal(t, true, out["is_mandatory"])
	assert.Equal(t, false, out["is_sale_prop"])
	assert.Equal(t, "normal", out["attribute_type"])
}

func TestOption_JSONFields(t *testing.T) {
	opt := Option{Name: "Black", Value: "Black"}
	data, err := json.Marshal(opt)
	assert.NoError(t, err)

	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, "Black", out["name"])
	assert.Equal(t, "Black", out["value"])
}
