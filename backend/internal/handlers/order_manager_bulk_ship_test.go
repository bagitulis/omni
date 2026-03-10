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

// TestBulkShipOrders_MissingTenant verifies 401 when tenantID is absent.
func TestBulkShipOrders_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewOrderManagerHandler("./data")
	r.POST("/api/orders/bulk-ship", handler.BulkShipOrders)

	body := `{"order_sns":["ORDER-001"],"platform":"shopee"}`
	req, _ := http.NewRequest("POST", "/api/orders/bulk-ship", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestBulkShipOrders_InvalidJSON verifies 400 on malformed JSON body.
func TestBulkShipOrders_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewOrderManagerHandler("./data")
	r.POST("/api/orders/bulk-ship", handler.BulkShipOrders)

	req, _ := http.NewRequest("POST", "/api/orders/bulk-ship", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestBulkShipOrders_EmptyOrderSNs verifies 400 when order_sns is empty.
func TestBulkShipOrders_EmptyOrderSNs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewOrderManagerHandler("./data")
	r.POST("/api/orders/bulk-ship", handler.BulkShipOrders)

	body := `{"order_sns":[],"platform":"shopee"}`
	req, _ := http.NewRequest("POST", "/api/orders/bulk-ship", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "order_sns cannot be empty")
}

// TestBulkShipOrders_UnsupportedPlatform verifies 400 for unknown platform.
func TestBulkShipOrders_UnsupportedPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewOrderManagerHandler("./data")
	r.POST("/api/orders/bulk-ship", handler.BulkShipOrders)

	body := `{"order_sns":["ORDER-001"],"platform":"tokopedia"}`
	req, _ := http.NewRequest("POST", "/api/orders/bulk-ship", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "Unsupported platform: tokopedia")
}

// TestBulkShipOrders_ShopeePlatform verifies 422 with summary/results for shopee (fails without DB).
func TestBulkShipOrders_ShopeePlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewOrderManagerHandler("./data")
	r.POST("/api/orders/bulk-ship", handler.BulkShipOrders)

	body := `{"order_sns":["ORDER-001"],"platform":"shopee"}`
	req, _ := http.NewRequest("POST", "/api/orders/bulk-ship", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// All fail (no credentials) → 422
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok, "response data should be a map")
	assert.Contains(t, data, "summary")
	assert.Contains(t, data, "results")
}

// TestBulkShipOrders_TikTokPlatform verifies 422 with summary/results for tiktok (fails without DB).
func TestBulkShipOrders_TikTokPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewOrderManagerHandler("./data")
	r.POST("/api/orders/bulk-ship", handler.BulkShipOrders)

	body := `{"order_sns":["ORDER-001"],"platform":"tiktok"}`
	req, _ := http.NewRequest("POST", "/api/orders/bulk-ship", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok, "response data should be a map")
	assert.Contains(t, data, "summary")
	assert.Contains(t, data, "results")
}

// TestBulkShipOrders_LazadaPlatform verifies 422 with summary/results for lazada (fails without DB).
func TestBulkShipOrders_LazadaPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewOrderManagerHandler("./data")
	r.POST("/api/orders/bulk-ship", handler.BulkShipOrders)

	body := `{"order_sns":["ORDER-001"],"platform":"lazada"}`
	req, _ := http.NewRequest("POST", "/api/orders/bulk-ship", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok, "response data should be a map")
	assert.Contains(t, data, "summary")
	assert.Contains(t, data, "results")
}

// TestBulkShipOrders_MissingPlatformReturns400 verifies that omitting platform returns 400.
func TestBulkShipOrders_MissingPlatformReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewOrderManagerHandler("./data")
	r.POST("/api/orders/bulk-ship", handler.BulkShipOrders)

	// No "platform" field — should return 400 (platform is now required)
	body := `{"order_sns":["ORDER-001"]}`
	req, _ := http.NewRequest("POST", "/api/orders/bulk-ship", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "platform is required")
}
