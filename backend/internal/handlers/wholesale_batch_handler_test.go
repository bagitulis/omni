package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestWholesaleBatchHandler_NewHandler tests handler creation
func TestWholesaleBatchHandler_NewHandler(t *testing.T) {
	handler := NewWholesaleBatchHandler("", nil)
	assert.NotNil(t, handler)
}

// TestWholesaleBatchHandler_BatchUpdateBySkus_MissingTenant tests BatchUpdateBySkus without tenant
func TestWholesaleBatchHandler_BatchUpdateBySkus_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleBatchHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-update-skus", handler.BatchUpdateBySkus)

	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-update-skus", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleBatchHandler_BatchUpdateBySkus_InvalidJSON tests BatchUpdateBySkus with invalid JSON
func TestWholesaleBatchHandler_BatchUpdateBySkus_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleBatchHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-update-skus", handler.BatchUpdateBySkus)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-update-skus", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleBatchHandler_BatchUpdateBySkus_EmptyItems tests BatchUpdateBySkus with empty items
func TestWholesaleBatchHandler_BatchUpdateBySkus_EmptyItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleBatchHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-update-skus", handler.BatchUpdateBySkus)

	body := `{"items": []}`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-update-skus", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleBatchHandler_BatchDeleteByItemIds_MissingTenant tests BatchDeleteByItemIds without tenant
func TestWholesaleBatchHandler_BatchDeleteByItemIds_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleBatchHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-delete", handler.BatchDeleteByItemIds)

	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-delete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleBatchHandler_BatchDeleteByItemIds_InvalidJSON tests BatchDeleteByItemIds with invalid JSON
func TestWholesaleBatchHandler_BatchDeleteByItemIds_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleBatchHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-delete", handler.BatchDeleteByItemIds)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-delete", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleBatchHandler_BatchDeleteBySkus_MissingTenant tests BatchDeleteBySkus without tenant
func TestWholesaleBatchHandler_BatchDeleteBySkus_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleBatchHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-delete-skus", handler.BatchDeleteBySkus)

	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-delete-skus", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleBatchHandler_BatchDeleteBySkus_InvalidJSON tests BatchDeleteBySkus with invalid JSON
func TestWholesaleBatchHandler_BatchDeleteBySkus_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleBatchHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-delete-skus", handler.BatchDeleteBySkus)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-delete-skus", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleBatchHandler_BatchDeleteBySkus_EmptySkus tests BatchDeleteBySkus with empty SKUs
func TestWholesaleBatchHandler_BatchDeleteBySkus_EmptySkus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleBatchHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-delete-skus", handler.BatchDeleteBySkus)

	body := `{"skus": []}`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-delete-skus", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleBatchHandler_Preview_MissingTenant tests Preview without tenant
func TestWholesaleBatchHandler_Preview_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleBatchHandler("", nil)
	r.POST("/api/wholesale/preview", handler.Preview)

	req, _ := http.NewRequest("POST", "/api/wholesale/preview", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleBatchHandler_Preview_InvalidJSON tests Preview with invalid JSON
func TestWholesaleBatchHandler_Preview_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleBatchHandler("", nil)
	r.POST("/api/wholesale/preview", handler.Preview)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/wholesale/preview", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
