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

func TestInventoryHandler_UpdateStock(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"sku":"SKU001"}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/update-stock", body)
		c.Request.Header.Set("Content-Type", "application/json")
		// No tenantID set

		handler.UpdateStock(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("missing SKU returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/update-stock", body)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.UpdateStock(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("valid request with no DB returns error", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"sku":"SKU001"}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/update-stock", body)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.UpdateStock(c)

		// Without a valid DB, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestInventoryHandler_UpdateStockBatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"skus":["SKU001","SKU002"]}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/update-stock-batch", body)
		c.Request.Header.Set("Content-Type", "application/json")
		// No tenantID set

		handler.UpdateStockBatch(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("missing SKUs returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/update-stock-batch", body)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.UpdateStockBatch(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("items payload is accepted by contract", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"items":[{"sku":"SKU001","stock":9,"platforms":["shopee"]}]}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/update-stock-batch", body)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.UpdateStockBatch(c)

		// Contract binding succeeds; nil DB causes internal error in test setup
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestInventoryHandler_UpdatePrice(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"sku":"SKU001"}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/update-price", body)
		c.Request.Header.Set("Content-Type", "application/json")
		// No tenantID set

		handler.UpdatePrice(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("missing SKU returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/update-price", body)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.UpdatePrice(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestInventoryHandler_UpdatePriceBatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"items":[{"sku":"SKU001","price":100}]}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/update-price-batch", body)
		c.Request.Header.Set("Content-Type", "application/json")
		// No tenantID set

		handler.UpdatePriceBatch(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("missing items returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/update-price-batch", body)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.UpdatePriceBatch(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
