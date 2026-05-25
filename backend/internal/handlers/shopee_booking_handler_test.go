package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- ShopeeBookingHandler.GetBookingOrders tests ---

func TestShopeeBookingHandler_GetBookingOrders_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/booking", nil)

	h := NewShopeeBookingHandler()
	h.GetBookingOrders(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Missing tenant_id", resp["error"])
}

func TestShopeeBookingHandler_GetBookingOrders_NonShopeePlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/booking?platform=lazada", nil)
	c.Set("tenant_id", "test-tenant")

	h := NewShopeeBookingHandler()
	h.GetBookingOrders(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "Booking orders are only available for Shopee")
}

func TestShopeeBookingHandler_GetBookingOrders_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/booking?page=1&page_size=10&search=test&booking_status=confirmed&match_status=matched&platform=shopee", nil)
	c.Set("tenant_id", "test-tenant")

	h := NewShopeeBookingHandler()
	h.GetBookingOrders(c)

	// Handler passes request parsing/validation; service returns 500 because no DB connection.
	// This proves query param parsing, tenant validation, and platform validation all work.
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShopeeBookingHandler_GetBookingOrders_EmptyPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/booking?platform=", nil)
	c.Set("tenant_id", "test-tenant")

	h := NewShopeeBookingHandler()
	h.GetBookingOrders(c)

	// Empty platform string should pass validation (only non-empty non-shopee is rejected).
	// Service will return 500 because no DB.
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

// --- ShopeeBookingHandler.GetBookingOrderDetail tests ---

func TestShopeeBookingHandler_GetBookingOrderDetail_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/booking/SN123", nil)
	c.Params = gin.Params{{Key: "booking_sn", Value: "SN123"}}

	h := NewShopeeBookingHandler()
	h.GetBookingOrderDetail(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Missing tenant_id", resp["error"])
}

func TestShopeeBookingHandler_GetBookingOrderDetail_MissingBookingSN(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/booking/", nil)
	c.Set("tenant_id", "test-tenant")

	h := NewShopeeBookingHandler()
	h.GetBookingOrderDetail(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Missing booking_sn", resp["error"])
}

func TestShopeeBookingHandler_GetBookingOrderDetail_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/booking/SN123", nil)
	c.Params = gin.Params{{Key: "booking_sn", Value: "SN123"}}
	c.Set("tenant_id", "test-tenant")

	h := NewShopeeBookingHandler()
	h.GetBookingOrderDetail(c)

	// Handler passes request parsing/validation; service returns 500 because no DB/repository configured.
	// This proves tenant validation and booking_sn param parsing work.
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

// --- OrderSyncHandler.SyncByCategory booking routing test ---

func TestOrderSyncHandler_SyncByCategory_BookingRouting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		category       string
		query          string
		hasTenant      bool
		expectedStatus int
		checkResp      func(t *testing.T, resp map[string]interface{})
	}{
		{
			name:           "missing_tenant_id_returns_401",
			category:       "booking",
			hasTenant:      false,
			expectedStatus: http.StatusUnauthorized,
			checkResp: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, false, resp["success"])
				assert.Equal(t, "Missing tenant_id", resp["error"])
			},
		},
		{
			name:           "booking_category_routes_to_booking_sync",
			category:       "booking",
			hasTenant:      true,
			expectedStatus: http.StatusOK,
			checkResp: func(t *testing.T, resp map[string]interface{}) {
				// Booking sync service without DB returns partial failure
				assert.Equal(t, false, resp["success"])
				assert.Equal(t, "PARTIAL_SYNC_FAILURE", resp["code"])
				assert.Equal(t, "booking", resp["category"])
			},
		},
		{
			name:           "booking_category_rejects_non_shopee_platform",
			category:       "booking",
			query:          "?platforms=lazada",
			hasTenant:      true,
			expectedStatus: http.StatusBadRequest,
			checkResp: func(t *testing.T, resp map[string]interface{}) {
				assert.Equal(t, false, resp["success"])
				assert.Equal(t, "Booking sync is only available for Shopee", resp["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewOrderSyncHandler()

			r.Use(func(c *gin.Context) {
				if tt.hasTenant {
					c.Set("tenant_id", "test-tenant")
				}
				c.Next()
			})
			r.POST("/api/orders/sync/:category", handler.SyncByCategory)

			path := "/api/orders/sync/" + tt.category + tt.query
			req, _ := http.NewRequest("POST", path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.checkResp != nil {
				tt.checkResp(t, resp)
			}
		})
	}
}
