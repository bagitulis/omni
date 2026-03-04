package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestWholesaleExtendedHandler_NewHandler tests handler creation
func TestWholesaleExtendedHandler_NewHandler(t *testing.T) {
	handler := NewWholesaleExtendedHandler("", nil)
	assert.NotNil(t, handler)
}

// TestWholesaleExtendedHandler_DeleteWholesale_MissingTenant tests DeleteWholesale without tenant
func TestWholesaleExtendedHandler_DeleteWholesale_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleExtendedHandler("", nil)
	r.DELETE("/api/wholesale/shopee/:itemId", handler.DeleteWholesale)

	req, _ := http.NewRequest("DELETE", "/api/wholesale/shopee/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleExtendedHandler_DeleteWholesale_InvalidItemID tests DeleteWholesale with invalid item ID
func TestWholesaleExtendedHandler_DeleteWholesale_InvalidItemID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.DELETE("/api/wholesale/shopee/:itemId", handler.DeleteWholesale)

	req, _ := http.NewRequest("DELETE", "/api/wholesale/shopee/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleExtendedHandler_UpdateWholesale_MissingTenant tests UpdateWholesale without tenant
func TestWholesaleExtendedHandler_UpdateWholesale_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleExtendedHandler("", nil)
	r.PUT("/api/wholesale/shopee/:itemId", handler.UpdateWholesale)

	req, _ := http.NewRequest("PUT", "/api/wholesale/shopee/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleExtendedHandler_UpdateWholesale_InvalidItemID tests UpdateWholesale with invalid item ID
func TestWholesaleExtendedHandler_UpdateWholesale_InvalidItemID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.PUT("/api/wholesale/shopee/:itemId", handler.UpdateWholesale)

	req, _ := http.NewRequest("PUT", "/api/wholesale/shopee/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleExtendedHandler_UpdateWholesale_InvalidJSON tests UpdateWholesale with invalid JSON
func TestWholesaleExtendedHandler_UpdateWholesale_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.PUT("/api/wholesale/shopee/:itemId", handler.UpdateWholesale)

	body := `{invalid json`
	req, _ := http.NewRequest("PUT", "/api/wholesale/shopee/123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleExtendedHandler_UpdateWholesale_ValidRequest tests UpdateWholesale with valid request
// Now requires DB and Shopee API, so without them it returns 500
func TestWholesaleExtendedHandler_UpdateWholesale_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.PUT("/api/wholesale/shopee/:itemId", handler.UpdateWholesale)

	body := `{"tiers": [{"min_count": 5, "max_count": 10, "unit_price": 90000}]}`
	req, _ := http.NewRequest("PUT", "/api/wholesale/shopee/123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Real handler needs DB/API, so returns 500 without them
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestWholesaleExtendedHandler_GetWholesaleInfo_MissingTenant tests GetWholesaleInfo without tenant
func TestWholesaleExtendedHandler_GetWholesaleInfo_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleExtendedHandler("", nil)
	r.GET("/api/wholesale/shopee/:itemId/info", handler.GetWholesaleInfo)

	req, _ := http.NewRequest("GET", "/api/wholesale/shopee/123/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleExtendedHandler_GetWholesaleInfo_InvalidItemID tests GetWholesaleInfo with invalid item ID
func TestWholesaleExtendedHandler_GetWholesaleInfo_InvalidItemID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.GET("/api/wholesale/shopee/:itemId/info", handler.GetWholesaleInfo)

	req, _ := http.NewRequest("GET", "/api/wholesale/shopee/invalid/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleExtendedHandler_GetWholesaleInfo_ValidItemID tests GetWholesaleInfo with valid item ID
// Now requires DB/API so returns 500 without them
func TestWholesaleExtendedHandler_GetWholesaleInfo_ValidItemID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.GET("/api/wholesale/shopee/:itemId/info", handler.GetWholesaleInfo)

	req, _ := http.NewRequest("GET", "/api/wholesale/shopee/123/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestWholesaleExtendedHandler_LookupItemId_MissingTenant tests LookupItemId without tenant
func TestWholesaleExtendedHandler_LookupItemId_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleExtendedHandler("", nil)
	r.GET("/api/wholesale/shopee/lookup/:sku", handler.LookupItemId)

	req, _ := http.NewRequest("GET", "/api/wholesale/shopee/lookup/SKU001", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleExtendedHandler_SetTiktokWholesale_MissingTenant tests SetTiktokWholesale without tenant
func TestWholesaleExtendedHandler_SetTiktokWholesale_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/tiktok/:productId", handler.SetTiktokWholesale)

	req, _ := http.NewRequest("POST", "/api/wholesale/tiktok/product123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleExtendedHandler_SetTiktokWholesale_InvalidJSON tests SetTiktokWholesale with invalid JSON
func TestWholesaleExtendedHandler_SetTiktokWholesale_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/tiktok/:productId", handler.SetTiktokWholesale)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/wholesale/tiktok/product123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleExtendedHandler_SetTiktokWholesale_ValidRequest tests SetTiktokWholesale returns 501
func TestWholesaleExtendedHandler_SetTiktokWholesale_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/tiktok/:productId", handler.SetTiktokWholesale)

	body := `{"tiers": [{"min_count": 5, "max_count": 10, "unit_price": 90000}]}`
	req, _ := http.NewRequest("POST", "/api/wholesale/tiktok/product123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// TikTok wholesale returns 501 Not Implemented
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

// TestWholesaleExtendedHandler_BatchDeleteByItemIds_MissingTenant tests BatchDeleteByItemIds without tenant
func TestWholesaleExtendedHandler_BatchDeleteByItemIds_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-delete", handler.BatchDeleteByItemIds)

	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-delete", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleExtendedHandler_BatchAdd_MissingTenant tests BatchAdd without tenant
func TestWholesaleExtendedHandler_BatchAdd_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-add", handler.BatchAdd)

	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-add", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleExtendedHandler_BatchAdd_InvalidJSON tests BatchAdd with invalid JSON
func TestWholesaleExtendedHandler_BatchAdd_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-add", handler.BatchAdd)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-add", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleExtendedHandler_BatchAdd_ValidRequest tests BatchAdd needs DB/API
func TestWholesaleExtendedHandler_BatchAdd_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-add", handler.BatchAdd)

	body := `{"items": [{"item_id": 123, "sku": "SKU001", "tiers": [{"min_count": 5, "max_count": 10, "unit_price": 90000}]}]}`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-add", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Real handler needs DB/API
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestWholesaleExtendedHandler_Preview_MissingTenant tests Preview without tenant
func TestWholesaleExtendedHandler_Preview_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/preview", handler.Preview)

	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/preview", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleExtendedHandler_ImportWholesale_MissingTenant tests ImportWholesale without tenant
func TestWholesaleExtendedHandler_ImportWholesale_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/import", handler.ImportWholesale)

	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/import", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleExtendedHandler_ImportWholesale_InvalidJSON tests ImportWholesale with invalid JSON
func TestWholesaleExtendedHandler_ImportWholesale_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/import", handler.ImportWholesale)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/import", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleExtendedHandler_ImportWholesale_ValidRequest tests ImportWholesale needs DB/API
func TestWholesaleExtendedHandler_ImportWholesale_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/import", handler.ImportWholesale)

	body := `{"data": [{"sku": "SKU001", "tiers": [{"min_count": 5, "max_count": 10, "unit_price": 90000}]}]}`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/import", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Real handler needs DB/API
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestWholesaleExtendedHandler_BatchSetMpq_MissingTenant tests BatchSetMpq without tenant
func TestWholesaleExtendedHandler_BatchSetMpq_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-mpq", handler.BatchSetMpq)

	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-mpq", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleExtendedHandler_BatchSetMpq_InvalidJSON tests BatchSetMpq with invalid JSON
func TestWholesaleExtendedHandler_BatchSetMpq_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-mpq", handler.BatchSetMpq)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-mpq", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleExtendedHandler_BatchSetMpq_EmptyItems tests BatchSetMpq with empty items
func TestWholesaleExtendedHandler_BatchSetMpq_EmptyItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/shopee/batch-mpq", handler.BatchSetMpq)

	body := `{"items": [], "mpq": 5}`
	req, _ := http.NewRequest("POST", "/api/wholesale/shopee/batch-mpq", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleExtendedHandler_BatchSetTiktokMpq_MissingTenant tests BatchSetTiktokMpq without tenant
func TestWholesaleExtendedHandler_BatchSetTiktokMpq_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/tiktok/batch-mpq", handler.BatchSetTiktokMpq)

	req, _ := http.NewRequest("POST", "/api/wholesale/tiktok/batch-mpq", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleExtendedHandler_BatchSetTiktokMpq_InvalidJSON tests BatchSetTiktokMpq with invalid JSON
func TestWholesaleExtendedHandler_BatchSetTiktokMpq_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/tiktok/batch-mpq", handler.BatchSetTiktokMpq)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/wholesale/tiktok/batch-mpq", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWholesaleExtendedHandler_BatchSetTiktokMpq_ValidRequest tests returns 501
func TestWholesaleExtendedHandler_BatchSetTiktokMpq_ValidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleExtendedHandler("", nil)
	r.POST("/api/wholesale/tiktok/batch-mpq", handler.BatchSetTiktokMpq)

	body := `{"products": [{"product_id": "product123", "mpq": 5}]}`
	req, _ := http.NewRequest("POST", "/api/wholesale/tiktok/batch-mpq", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// TikTok MPQ returns 501 Not Implemented
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}
