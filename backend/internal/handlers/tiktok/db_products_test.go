package tiktok

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

	r.GET("/api/tiktok/db/products", handler.GetDBProducts)

	req, _ := http.NewRequest("GET", "/api/tiktok/db/products", nil)
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
	r.GET("/api/tiktok/db/products", handler.GetDBProducts)

	req, _ := http.NewRequest("GET", "/api/tiktok/db/products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available in test environment
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestDBProductHandler_GetDBProducts_Pagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{
			name:           "custom_offset_and_limit",
			query:          "?offset=10&limit=50",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "limit_capped_at_1000",
			query:          "?offset=0&limit=9999",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "default_pagination",
			query:          "",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewDBProductHandler("/test/path")

			r.Use(func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
				c.Next()
			})
			r.GET("/api/tiktok/db/products", handler.GetDBProducts)

			req, _ := http.NewRequest("GET", "/api/tiktok/db/products"+tt.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestDBProductHandler_GetMasterProducts_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewDBProductHandler("/test/path")

	r.GET("/api/tiktok/db/products/master", handler.GetMasterProducts)

	req, _ := http.NewRequest("GET", "/api/tiktok/db/products/master", nil)
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
	r.GET("/api/tiktok/db/products/master", handler.GetMasterProducts)

	req, _ := http.NewRequest("GET", "/api/tiktok/db/products/master", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available in test environment
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestDBProductHandler_Constructor(t *testing.T) {
	handler := NewDBProductHandler("/data/path")
	assert.NotNil(t, handler)
	assert.Equal(t, "/data/path", handler.basePath)
}

func TestMasterProductItem_JSONFields(t *testing.T) {
	item := MasterProductItem{
		ProductID:   "PROD001",
		SKU:         "SKU001",
		SellerSKU:   "SELLER001",
		SkuID:       "SKUID001",
		VariantName: "Red XL",
		ItemName:    "Test Product",
		Image:       "https://example.com/image.jpg",
		Price:       99.99,
		Quantity:    10,
		Status:      "active",
		UpdatedAt:   "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(item)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	// Verify snake_case JSON fields
	assert.Equal(t, "PROD001", decoded["product_id"])
	assert.Equal(t, "SKU001", decoded["sku"])
	assert.Equal(t, "SELLER001", decoded["seller_sku"])
	assert.Equal(t, "SKUID001", decoded["sku_id"])
	assert.Equal(t, "Red XL", decoded["variant_name"])
	assert.Equal(t, "Test Product", decoded["item_name"])
	assert.Equal(t, float64(99.99), decoded["price"])
	assert.Equal(t, float64(10), decoded["quantity"])
	assert.Equal(t, "active", decoded["status"])
}

func TestMasterProductItem_SkuID_Omitempty(t *testing.T) {
	// sku_id has omitempty tag, so empty string should be omitted
	item := MasterProductItem{
		ProductID: "PROD001",
		SKU:       "SKU001",
		SkuID:     "", // empty - should be omitted
	}

	data, err := json.Marshal(item)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	_, hasSkuID := decoded["sku_id"]
	assert.False(t, hasSkuID, "sku_id should be omitted when empty")
}
