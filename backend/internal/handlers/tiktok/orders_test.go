package tiktok

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTikTokOrderHandler_GetOrders(t *testing.T) {
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
		{
			name: "valid_request_with_pagination",
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			query:          "?page=2&pageSize=50",
			expectedStatus: http.StatusInternalServerError, // DB not available
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewOrderHandler("/test/path")

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/tiktok/orders", handler.GetOrders)

			req, _ := http.NewRequest("GET", "/api/tiktok/orders"+tt.query, nil)
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

func TestTikTokOrderHandler_GetOrderByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		orderID        string
		setupContext   func(c *gin.Context)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "missing_tenant_id",
			orderID: "ORDER123",
			setupContext: func(c *gin.Context) {
				// No tenant_id
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:    "valid_order_id",
			orderID: "ORDER123",
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
			handler := NewOrderHandler("/test/path")

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/tiktok/orders/:orderId", handler.GetOrderByID)

			url := "/api/tiktok/orders/" + tt.orderID
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

func TestTikTokOrderHandler_ShipOrder_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewOrderHandler("/test/path")

	r.POST("/api/tiktok/orders/ship", handler.ShipOrder)

	body := `{"order_id": "ORDER123", "package_id": "PKG123"}`
	req, _ := http.NewRequest("POST", "/api/tiktok/orders/ship", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestTikTokOrderHandler_ShipOrder_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewOrderHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/tiktok/orders/ship", handler.ShipOrder)

	body := `{invalid json}`
	req, _ := http.NewRequest("POST", "/api/tiktok/orders/ship", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestTikTokOrderHandler_ShipOrder_MissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewOrderHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/tiktok/orders/ship", handler.ShipOrder)

	// Missing package_id
	body := `{"order_id": "ORDER123"}`
	req, _ := http.NewRequest("POST", "/api/tiktok/orders/ship", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestTikTokOrderHandler_CancelOrder_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewOrderHandler("/test/path")

	r.POST("/api/tiktok/orders/cancel", handler.CancelOrder)

	body := `{"order_id": "ORDER123", "cancel_reason": "Out of stock"}`
	req, _ := http.NewRequest("POST", "/api/tiktok/orders/cancel", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestTikTokOrderHandler_CancelOrder_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewOrderHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/tiktok/orders/cancel", handler.CancelOrder)

	body := `{invalid json}`
	req, _ := http.NewRequest("POST", "/api/tiktok/orders/cancel", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestParsePagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		query            string
		expectedPage     int
		expectedPageSize int
	}{
		{
			name:             "default_values",
			query:            "",
			expectedPage:     1,
			expectedPageSize: 20,
		},
		{
			name:             "custom_values",
			query:            "?page=5&pageSize=50",
			expectedPage:     5,
			expectedPageSize: 50,
		},
		{
			name:             "invalid_page_defaults_to_1",
			query:            "?page=-1&pageSize=20",
			expectedPage:     1,
			expectedPageSize: 20,
		},
		{
			name:             "page_size_over_100_capped",
			query:            "?page=1&pageSize=150",
			expectedPage:     1,
			expectedPageSize: 20, // Capped by validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			var actualPage, actualPageSize int

			r.GET("/test", func(c *gin.Context) {
				actualPage, actualPageSize = parsePagination(c)
				c.JSON(http.StatusOK, gin.H{"page": actualPage, "pageSize": actualPageSize})
			})

			req, _ := http.NewRequest("GET", "/test"+tt.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedPage, actualPage)
			assert.Equal(t, tt.expectedPageSize, actualPageSize)
		})
	}
}
