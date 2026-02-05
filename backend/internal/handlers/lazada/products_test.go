package lazada

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLazadaProductHandler_GetProducts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(c *gin.Context)
		query          string
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "missing_tenant_id",
			setupContext: func(c *gin.Context) {
				// No tenant_id set
			},
			query:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "valid_request_with_defaults",
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			query:          "",
			expectedStatus: http.StatusInternalServerError, // DB not available
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewProductHandler("/test/path")

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/lazada/products", handler.GetProducts)

			req, _ := http.NewRequest("GET", "/api/lazada/products"+tt.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.expectedError {
				assert.Equal(t, false, resp["success"])
			}
		})
	}
}

func TestLazadaProductHandler_GetProductByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		itemID         string
		setupContext   func(c *gin.Context)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "missing_tenant_id",
			itemID: "ITEM123",
			setupContext: func(c *gin.Context) {
				// No tenant_id
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:   "valid_item_id",
			itemID: "ITEM123",
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError, // DB not available
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewProductHandler("/test/path")

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/lazada/products/:itemId", handler.GetProductByID)

			url := "/api/lazada/products/" + tt.itemID
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.expectedError {
				assert.Equal(t, false, resp["success"])
			}
		})
	}
}

func TestLazadaProductHandler_CreateProduct_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.POST("/api/lazada/products", handler.CreateProduct)

	body := `{"name": "Test Product", "description": "Test", "brand": "Brand", "primary_category": 123, "seller_sku": "SKU1", "price": 10000, "quantity": 10}`
	req, _ := http.NewRequest("POST", "/api/lazada/products", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestLazadaProductHandler_CreateProduct_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/lazada/products", handler.CreateProduct)

	body := `{invalid json}`
	req, _ := http.NewRequest("POST", "/api/lazada/products", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestLazadaProductHandler_UpdateProduct_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.PUT("/api/lazada/products/:itemId", handler.UpdateProduct)

	body := `{"name": "Updated Product"}`
	req, _ := http.NewRequest("PUT", "/api/lazada/products/ITEM123", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestLazadaProductHandler_UpdateProduct_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.PUT("/api/lazada/products/:itemId", handler.UpdateProduct)

	body := `{invalid json}`
	req, _ := http.NewRequest("PUT", "/api/lazada/products/ITEM123", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestLazadaProductHandler_DeleteProduct_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductHandler("/test/path")

	r.DELETE("/api/lazada/products/:itemId", handler.DeleteProduct)

	req, _ := http.NewRequest("DELETE", "/api/lazada/products/ITEM123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}
