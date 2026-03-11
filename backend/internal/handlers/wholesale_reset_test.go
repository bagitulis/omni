package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestWholesaleExtendedHandler_BatchWholesaleReset_MissingTenant tests BatchWholesaleReset without tenant
func TestWholesaleExtendedHandler_BatchWholesaleReset_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-reset", handler.BatchWholesaleReset)

	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-reset", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleExtendedHandler_BatchWholesaleReset_InvalidJSON tests BatchWholesaleReset with invalid JSON
func TestWholesaleExtendedHandler_BatchWholesaleReset_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-reset", handler.BatchWholesaleReset)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-reset", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleExtendedHandler_BatchWholesaleReset_EmptyItems tests BatchWholesaleReset with empty items
func TestWholesaleExtendedHandler_BatchWholesaleReset_EmptyItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-reset", handler.BatchWholesaleReset)

	body := `{"items": []}`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-reset", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleExtendedHandler_BatchWholesaleReset_DBError tests BatchWholesaleReset with DB error
func TestWholesaleExtendedHandler_BatchWholesaleReset_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-reset", handler.BatchWholesaleReset)

	body := `{"items": [{"sku": "SKU001", "price": 100000}]}`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-reset", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Without DB, should return 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
