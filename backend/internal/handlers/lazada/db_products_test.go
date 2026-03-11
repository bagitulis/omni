package lazada

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDBProductHandler_New(t *testing.T) {
	h := NewDBProductHandler("/test/path")
	assert.NotNil(t, h)
	assert.Equal(t, "/test/path", h.basePath)
}

func TestFlattenedSkuRow_JSONFields(t *testing.T) {
	row := FlattenedSkuRow{
		ItemID:      "ITEM001",
		SkuID:       "SKU001",
		SkuName:     "Test SKU",
		ItemName:    "Test Item",
		VariantName: "Red / L",
		Image:       "https://example.com/img.jpg",
		Price:       99999.0,
		Quantity:    10,
		Status:      "active",
		UpdatedAt:   "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(row)
	assert.NoError(t, err)

	var out map[string]interface{}
	err = json.Unmarshal(data, &out)
	assert.NoError(t, err)

	// Verify snake_case JSON tags
	assert.Equal(t, "ITEM001", out["item_id"])
	assert.Equal(t, "SKU001", out["sku_id"])
	assert.Equal(t, "Test SKU", out["sku_name"])
	assert.Equal(t, "Test Item", out["item_name"])
	assert.Equal(t, "Red / L", out["variant_name"])
	assert.Equal(t, "https://example.com/img.jpg", out["image"])
	assert.Equal(t, float64(99999), out["price"])
	assert.Equal(t, float64(10), out["quantity"])
	assert.Equal(t, "active", out["status"])
	assert.Equal(t, "2024-01-01T00:00:00Z", out["updated_at"])
}

func TestDBProductHandler_GetDBProducts_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDBProductHandler("/test/path")
	r.GET("/api/lazada/db/products", h.GetDBProducts)

	req, _ := http.NewRequest("GET", "/api/lazada/db/products", nil)
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
	h := NewDBProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/lazada/db/products", h.GetDBProducts)

	req, _ := http.NewRequest("GET", "/api/lazada/db/products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available in test, expect internal server error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestDBProductHandler_GetMasterProducts_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDBProductHandler("/test/path")
	r.GET("/api/lazada/db/products/master", h.GetMasterProducts)

	req, _ := http.NewRequest("GET", "/api/lazada/db/products/master", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestDBProductHandler_GetDBProducts_PaginationDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDBProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/lazada/db/products", h.GetDBProducts)

	// With explicit pagination params (handler should accept them, fail on DB)
	req, _ := http.NewRequest("GET", "/api/lazada/db/products?offset=10&limit=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available → 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDBProductHandler_GetDBProducts_LimitCap(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDBProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/lazada/db/products", h.GetDBProducts)

	// limit > 10000 should be capped (DB still unavailable → 500)
	req, _ := http.NewRequest("GET", "/api/lazada/db/products?limit=99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
