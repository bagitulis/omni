package shopee

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestOrderHandler_ShipOrder_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewOrderHandler("/test/path")

	r.POST("/api/shopee/orders/ship", handler.ShipOrder)

	body := `{"order_sn": "ORDER123"}`
	req, _ := http.NewRequest("POST", "/api/shopee/orders/ship", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestOrderHandler_ShipOrder_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewOrderHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/orders/ship", handler.ShipOrder)

	body := `{invalid json}`
	req, _ := http.NewRequest("POST", "/api/shopee/orders/ship", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestOrderHandler_ShipOrder_ValidRequestNoDB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewOrderHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/orders/ship", handler.ShipOrder)

	body := `{"order_sn": "ORDER123"}`
	req, _ := http.NewRequest("POST", "/api/shopee/orders/ship", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Credentials not available in test environment
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestOrderHandler_CancelOrder_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewOrderHandler("/test/path")

	r.POST("/api/shopee/orders/cancel", handler.CancelOrder)

	body := `{"order_sn": "ORDER123", "cancel_reason": "OUT_OF_STOCK"}`
	req, _ := http.NewRequest("POST", "/api/shopee/orders/cancel", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestOrderHandler_CancelOrder_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewOrderHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/orders/cancel", handler.CancelOrder)

	body := `{invalid json}`
	req, _ := http.NewRequest("POST", "/api/shopee/orders/cancel", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestOrderHandler_CancelOrder_ValidRequestNoDB(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewOrderHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/orders/cancel", handler.CancelOrder)

	body := `{"order_sn": "ORDER123", "cancel_reason": "OUT_OF_STOCK"}`
	req, _ := http.NewRequest("POST", "/api/shopee/orders/cancel", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Credentials not available in test environment
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}
