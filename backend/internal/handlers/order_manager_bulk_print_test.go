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

// TestBulkPrintLabels_MissingTenant verifies 401 when tenantID is absent.
func TestBulkPrintLabels_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewOrderManagerHandler("./data")
	r.POST("/api/orders/bulk-print-labels", handler.BulkPrintLabels)

	body := `{"order_sns":["ORDER-001"],"platform":"shopee"}`
	req, _ := http.NewRequest("POST", "/api/orders/bulk-print-labels", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestBulkPrintLabels_InvalidJSON verifies 400 on malformed JSON body.
func TestBulkPrintLabels_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewOrderManagerHandler("./data")
	r.POST("/api/orders/bulk-print-labels", handler.BulkPrintLabels)

	req, _ := http.NewRequest("POST", "/api/orders/bulk-print-labels", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestBulkPrintLabels_EmptyOrderSNs verifies 400 when order_sns is empty.
func TestBulkPrintLabels_EmptyOrderSNs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewOrderManagerHandler("./data")
	r.POST("/api/orders/bulk-print-labels", handler.BulkPrintLabels)

	body := `{"order_sns":[],"platform":"shopee"}`
	req, _ := http.NewRequest("POST", "/api/orders/bulk-print-labels", bytes.NewBufferString(body))
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

// TestBulkPrintLabels_ValidRequest verifies 200 with labels/failed arrays in response.
// Label retrieval will fail gracefully (no real platform client) but handler returns 200.
func TestBulkPrintLabels_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewOrderManagerHandler("./data")
	r.POST("/api/orders/bulk-print-labels", handler.BulkPrintLabels)

	body := `{"order_sns":["ORDER-001","ORDER-002"],"platform":"shopee"}`
	req, _ := http.NewRequest("POST", "/api/orders/bulk-print-labels", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok, "response data should be a map")
	assert.Contains(t, data, "labels")
	assert.Contains(t, data, "failed")
	assert.Contains(t, data, "count")
}

// TestBulkPrintLabels_WithTikTokDocumentType verifies tiktok-specific options are accepted.
func TestBulkPrintLabels_WithTikTokDocumentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewOrderManagerHandler("./data")
	r.POST("/api/orders/bulk-print-labels", handler.BulkPrintLabels)

	body := `{"order_sns":["ORDER-001"],"platform":"tiktok","tiktok_document_type":"SHIPPING_LABEL","include_products":true}`
	req, _ := http.NewRequest("POST", "/api/orders/bulk-print-labels", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))
}
