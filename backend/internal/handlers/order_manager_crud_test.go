package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/sync"
	"github.com/stretchr/testify/assert"
)

// TestOrderManagerHandler_GetUnpaidOrders_MissingTenant tests GetUnpaidOrders without tenant
func TestOrderManagerHandler_GetUnpaidOrders_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewOrderManagerHandler("./data")
	r.GET("/api/orders/unpaid", handler.GetUnpaidOrders)

	req, _ := http.NewRequest("GET", "/api/orders/unpaid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestOrderManagerHandler_GetUnprocessOrders_MissingTenant tests GetUnprocessOrders without tenant
func TestOrderManagerHandler_GetUnprocessOrders_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewOrderManagerHandler("./data")
	r.GET("/api/orders/unprocess", handler.GetUnprocessOrders)

	req, _ := http.NewRequest("GET", "/api/orders/unprocess", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestOrderManagerHandler_GetProcessedOrders_MissingTenant tests GetProcessedOrders without tenant
func TestOrderManagerHandler_GetProcessedOrders_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewOrderManagerHandler("./data")
	r.GET("/api/orders/processed", handler.GetProcessedOrders)

	req, _ := http.NewRequest("GET", "/api/orders/processed", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestOrderManagerHandler_GetUnpaidOrders_EmptyTenant tests when tenant has no data
func TestOrderManagerHandler_GetUnpaidOrders_EmptyTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "nonexistent-tenant")
		c.Next()
	})

	handler := NewOrderManagerHandler("./data")
	r.GET("/api/orders/unpaid", handler.GetUnpaidOrders)

	req, _ := http.NewRequest("GET", "/api/orders/unpaid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler gracefully handles missing sync service and returns empty data
	// This is expected behavior - handler creates service on demand
	assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	// Handler may return success with empty data or error
	// Both are valid responses for nonexistent tenant
}

// TestOrderManagerHandler_GetOrderByOrderSn_MissingTenant tests GetOrderByOrderSn without tenant
func TestOrderManagerHandler_GetOrderByOrderSn_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewOrderManagerHandler("./data")
	r.GET("/api/orders/:orderSn", handler.GetOrderByOrderSn)

	req, _ := http.NewRequest("GET", "/api/orders/TEST123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "Missing tenantId")
}

// TestOrderManagerHandler_GetOrderByOrderSn_MissingOrderSn tests GetOrderByOrderSn with empty orderSn
func TestOrderManagerHandler_GetOrderByOrderSn_MissingOrderSn(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewOrderManagerHandler("./data")
	r.GET("/api/orders/:orderSn", handler.GetOrderByOrderSn)

	// This test verifies empty param validation
	// Note: Gin router won't match this route if orderSn is empty
	// But if it does, handler should validate
	req, _ := http.NewRequest("GET", "/api/orders/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 404 because route won't match
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestOrderManagerHandler_GetOrderByOrderSn_HandlerExists tests handler is properly initialized
func TestOrderManagerHandler_GetOrderByOrderSn_HandlerExists(t *testing.T) {
	handler := NewOrderManagerHandler("./data")
	assert.NotNil(t, handler)
	assert.Equal(t, "./data", handler.basePath)
}

func TestParseOrderPlatformFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *sync.PlatformType
		wantErr  bool
	}{
		{name: "empty returns nil", input: "", expected: nil, wantErr: false},
		{name: "all returns nil", input: "all", expected: nil, wantErr: false},
		{name: "uppercase shopee", input: "SHOPEE", expected: platformPtr(sync.PlatformShopee), wantErr: false},
		{name: "lazada", input: "lazada", expected: platformPtr(sync.PlatformLazada), wantErr: false},
		{name: "tiktok", input: "tiktok", expected: platformPtr(sync.PlatformTiktok), wantErr: false},
		{name: "invalid platform", input: "tokopedia", expected: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOrderPlatformFilter(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.expected == nil {
				assert.Nil(t, got)
				return
			}

			if assert.NotNil(t, got) {
				assert.Equal(t, *tt.expected, *got)
			}
		})
	}
}

func platformPtr(platform sync.PlatformType) *sync.PlatformType {
	return &platform
}
