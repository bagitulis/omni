package shopee

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestShopeeOrderHandler_GetOrders(t *testing.T) {
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
				// No tenant_id set - simulates missing auth
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
			expectedStatus: http.StatusInternalServerError, // DB not available in test
			expectedError:  true,
		},
		{
			name: "valid_request_with_pagination",
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			query:          "?page=2&pageSize=50",
			expectedStatus: http.StatusInternalServerError, // DB not available in test
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
			r.GET("/api/shopee/orders", handler.GetOrders)

			req, _ := http.NewRequest("GET", "/api/shopee/orders"+tt.query, nil)
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

func TestShopeeOrderHandler_GetOrderByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		orderSN        string
		setupContext   func(c *gin.Context)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:    "valid_order_sn",
			orderSN: "ORDER123",
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
			r.GET("/api/shopee/orders/:orderSn", handler.GetOrderByID)

			url := "/api/shopee/orders/" + tt.orderSN

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

func TestShopeeOrderHandler_GetOrderByID_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewOrderHandler("/test/path")

	r.GET("/api/shopee/orders/:orderSn", handler.GetOrderByID)

	req, _ := http.NewRequest("GET", "/api/shopee/orders/ORDER123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}
