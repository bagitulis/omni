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

func TestProductCloneHandler_Clone(t *testing.T) {
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
				"source_platform": "shopee",
				"target_platform": "lazada",
				"source_item_id":  "123",
			},
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "invalid_platform",
			body: map[string]interface{}{
				"source_platform": "invalid",
				"target_platform": "lazada",
				"source_item_id":  "123",
			},
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "same_platform",
			body: map[string]interface{}{
				"source_platform": "shopee",
				"target_platform": "shopee",
				"source_item_id":  "123",
			},
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "valid_request_db_not_available",
			body: map[string]interface{}{
				"source_platform": "shopee",
				"target_platform": "lazada",
				"source_item_id":  "123",
			},
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewProductCloneHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.POST("/api/products/clone", handler.Clone)

			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/products/clone", bytes.NewBuffer(body))
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

func TestProductCloneHandler_GetStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		cloneID        string
		setupContext   func(c *gin.Context)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "missing_tenant_id",
			cloneID: "clone-123",
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:    "empty_clone_id",
			cloneID: "empty",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:    "valid_request_db_not_available",
			cloneID: "clone-123",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewProductCloneHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/products/clone/status/:id", handler.GetStatus)

			url := "/api/products/clone/status/" + tt.cloneID

			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestProductCloneHandler_BatchClone(t *testing.T) {
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
				"source_platform": "shopee",
				"target_platform": "lazada",
				"source_item_ids": []string{"123", "456"},
			},
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name: "empty_source_items",
			body: map[string]interface{}{
				"source_platform": "shopee",
				"target_platform": "lazada",
				"source_item_ids": []string{},
			},
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "same_platform",
			body: map[string]interface{}{
				"source_platform": "tiktok",
				"target_platform": "tiktok",
				"source_item_ids": []string{"123"},
			},
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewProductCloneHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.POST("/api/products/clone/batch", handler.BatchClone)

			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/products/clone/batch", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestProductCloneHandler_GetProductData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		query          string
		setupContext   func(c *gin.Context)
		expectedStatus int
	}{
		{
			name:  "missing_tenant_id",
			query: "?platform=shopee&sku=SKU001",
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "missing_platform",
			query: "?sku=SKU001",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "missing_sku",
			query: "?platform=shopee",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "invalid_platform",
			query: "?platform=invalid&sku=SKU001",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewProductCloneHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/clone/product-data", handler.GetProductData)

			req, _ := http.NewRequest("GET", "/api/clone/product-data"+tt.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestProductCloneHandler_GetAvailableTargets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		query          string
		setupContext   func(c *gin.Context)
		expectedStatus int
	}{
		{
			name:  "missing_tenant_id",
			query: "?sku=SKU001",
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "missing_sku",
			query: "",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "valid_request_db_not_available",
			query: "?sku=SKU001",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewProductCloneHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/clone/available-targets", handler.GetAvailableTargets)

			req, _ := http.NewRequest("GET", "/api/clone/available-targets"+tt.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestProductCloneHandler_Preview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		query          string
		setupContext   func(c *gin.Context)
		expectedStatus int
	}{
		{
			name:  "missing_tenant_id",
			query: "?source_platform=shopee&target_platform=lazada&source_item_id=123",
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "missing_params",
			query: "?source_platform=shopee",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "same_platform",
			query: "?source_platform=shopee&target_platform=shopee&source_item_id=123",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "invalid_platform",
			query: "?source_platform=invalid&target_platform=lazada&source_item_id=123",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewProductCloneHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/clone/preview", handler.Preview)

			req, _ := http.NewRequest("GET", "/api/clone/preview"+tt.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
