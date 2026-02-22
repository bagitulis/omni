package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestOrderSyncHandler_SyncByCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		category       string
		setupContext   func(c *gin.Context)
		expectedStatus int
		checkSuccess   *bool // nil = don't check, true/false = check specific value
	}{
		{
			name:           "missing_tenant_id_returns_401",
			category:       "unpaid",
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusUnauthorized,
			checkSuccess:   ptrBool(false),
		},
		{
			name:     "valid_category_without_platforms_returns_partial_failure",
			category: "unpaid",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			// Handler now marks success=false when any platform sync fails
			expectedStatus: http.StatusOK,
			checkSuccess:   ptrBool(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewOrderSyncHandler()

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.POST("/api/orders/sync/:category", handler.SyncByCategory)

			path := "/api/orders/sync/" + tt.category
			req, _ := http.NewRequest("POST", path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.checkSuccess != nil {
				assert.Equal(t, *tt.checkSuccess, resp["success"])
			}
		})
	}
}

func TestOrderSyncHandler_SyncPlatformOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		platform       string
		category       string
		setupContext   func(c *gin.Context)
		expectedStatus int
		checkSuccess   *bool
	}{
		{
			name:           "missing_tenant_id_returns_401",
			platform:       "shopee",
			category:       "unpaid",
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusUnauthorized,
			checkSuccess:   ptrBool(false),
		},
		{
			name:     "valid_request_without_service_returns_500",
			platform: "shopee",
			category: "unpaid",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			// Handler returns error when Shopee client is not initialized
			expectedStatus: http.StatusInternalServerError,
			checkSuccess:   ptrBool(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewOrderSyncHandler()

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/orders/sync/platform/:platform", handler.SyncPlatformOrders)

			path := "/api/orders/sync/platform/" + tt.platform
			if tt.category != "" {
				path += "?category=" + tt.category
			}
			req, _ := http.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.checkSuccess != nil {
				assert.Equal(t, *tt.checkSuccess, resp["success"])
			}
		})
	}
}

func TestOrderSyncHandler_GetOrdersByCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		category       string
		setupContext   func(c *gin.Context)
		expectedStatus int
		checkSuccess   *bool
	}{
		{
			name:           "missing_tenant_id_returns_401",
			category:       "unpaid",
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusUnauthorized,
			checkSuccess:   ptrBool(false),
		},
		{
			name:     "valid_tenant_id_returns_success",
			category: "unpaid",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			// Handler returns success with empty data when no service
			expectedStatus: http.StatusOK,
			checkSuccess:   ptrBool(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewOrderSyncHandler()

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/orders/:category", handler.GetOrdersByCategory)

			path := "/api/orders/" + tt.category
			req, _ := http.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.checkSuccess != nil {
				assert.Equal(t, *tt.checkSuccess, resp["success"])
			}
		})
	}
}

func TestOrderSyncHandler_GetOrderDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		platform       string
		orderIDs       string
		setupContext   func(c *gin.Context)
		expectedStatus int
		checkSuccess   *bool
	}{
		{
			name:           "missing_tenant_id_returns_401",
			platform:       "shopee",
			orderIDs:       "order123",
			setupContext:   func(c *gin.Context) {},
			expectedStatus: http.StatusUnauthorized,
			checkSuccess:   ptrBool(false),
		},
		{
			name:     "missing_order_ids_returns_400",
			platform: "shopee",
			orderIDs: "",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusBadRequest,
			checkSuccess:   ptrBool(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewOrderSyncHandler()

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/orders/details/:platform", handler.GetOrderDetails)

			path := "/api/orders/details/" + tt.platform
			if tt.orderIDs != "" {
				path += "?orderIds=" + tt.orderIDs
			}
			req, _ := http.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.checkSuccess != nil {
				assert.Equal(t, *tt.checkSuccess, resp["success"])
			}
		})
	}
}

// ptrBool is a helper to create a pointer to a bool
func ptrBool(b bool) *bool {
	return &b
}
