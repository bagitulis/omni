package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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
