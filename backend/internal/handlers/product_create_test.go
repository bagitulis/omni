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

func TestProductCreateHandler_CreateOnShopee(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           map[string]interface{}
		setupContext   func(c *gin.Context)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "missing_tenant_id",
			body: map[string]interface{}{
				"name":        "Test Product",
				"price":       10000,
				"category_id": 123,
			},
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "valid_tenant_no_api_configured",
			body: map[string]interface{}{
				"name":        "Test Product",
				"price":       10000,
				"category_id": 123,
			},
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
				// No shopeeAPI set
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewProductCreateHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.POST("/api/products/create/shopee", handler.CreateOnShopee)

			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/products/create/shopee", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.expectedError {
				_, hasError := resp["error"]
				assert.True(t, hasError || resp["success"] == false)
			}
		})
	}
}

func TestProductCreateHandler_CreateOnLazada(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           map[string]interface{}
		setupContext   func(c *gin.Context)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "missing_tenant_id",
			body: map[string]interface{}{
				"name":        "Test Product",
				"price":       10000,
				"category_id": 123,
			},
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "valid_tenant_no_api_configured",
			body: map[string]interface{}{
				"name":        "Test Product",
				"price":       10000,
				"category_id": 123,
			},
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
				// No lazadaAPI set
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewProductCreateHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.POST("/api/products/create/lazada", handler.CreateOnLazada)

			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/products/create/lazada", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.expectedError {
				_, hasError := resp["error"]
				assert.True(t, hasError || resp["success"] == false)
			}
		})
	}
}

func TestProductCreateHandler_CreateOnTiktok(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           map[string]interface{}
		setupContext   func(c *gin.Context)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "missing_tenant_id",
			body: map[string]interface{}{
				"name":        "Test Product",
				"price":       10000,
				"category_id": 123,
			},
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "valid_tenant_no_api_configured",
			body: map[string]interface{}{
				"name":        "Test Product",
				"price":       10000,
				"category_id": 123,
			},
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
				// No tiktokAPI set
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewProductCreateHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.POST("/api/products/create/tiktok", handler.CreateOnTiktok)

			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/products/create/tiktok", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.expectedError {
				_, hasError := resp["error"]
				assert.True(t, hasError || resp["success"] == false)
			}
		})
	}
}

func TestProductCreateHandler_GetCategories(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		platform       string
		query          string
		setupContext   func(c *gin.Context)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:     "valid_platform_with_empty_api",
			platform: "tiktok",
			query:    "",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:     "valid_platform_no_api",
			platform: "shopee",
			query:    "",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
				// No shopeeAPI set
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:     "valid_platform_with_parent_id",
			platform: "lazada",
			query:    "?parent_id=123",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
				// No lazadaAPI set
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewProductCreateHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/products/categories/:platform", handler.GetCategories)

			url := "/api/products/categories/" + tt.platform + tt.query

			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
