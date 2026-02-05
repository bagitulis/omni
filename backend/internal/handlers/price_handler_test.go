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

func TestPriceHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewPriceHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/price", nil)
		// No tenantID set

		handler.List(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "tenantId")
	})

	t.Run("valid tenant ID with no DB returns error", func(t *testing.T) {
		handler := NewPriceHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/price", nil)
		c.Set("tenantID", "test-tenant")

		handler.List(c)

		// Without a valid DB, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPriceHandler_GetBySKU(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewPriceHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/price/SKU001", nil)
		c.Params = gin.Params{{Key: "sku", Value: "SKU001"}}
		// No tenantID set

		handler.GetBySKU(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("missing SKU returns 400", func(t *testing.T) {
		handler := NewPriceHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/price/", nil)
		c.Params = gin.Params{{Key: "sku", Value: ""}}
		c.Set("tenantID", "test-tenant")

		handler.GetBySKU(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("valid request with no DB returns error", func(t *testing.T) {
		handler := NewPriceHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/price/SKU001", nil)
		c.Params = gin.Params{{Key: "sku", Value: "SKU001"}}
		c.Set("tenantID", "test-tenant")

		handler.GetBySKU(c)

		// Without a valid DB, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPriceHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewPriceHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"price":100}`)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/price/SKU001", body)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "sku", Value: "SKU001"}}
		// No tenantID set

		handler.Update(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("missing SKU returns 400", func(t *testing.T) {
		handler := NewPriceHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"price":100}`)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/price/", body)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "sku", Value: ""}}
		c.Set("tenantID", "test-tenant")

		handler.Update(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing price returns 400", func(t *testing.T) {
		handler := NewPriceHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{}`)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/price/SKU001", body)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "sku", Value: "SKU001"}}
		c.Set("tenantID", "test-tenant")

		handler.Update(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("valid request with no DB returns error", func(t *testing.T) {
		handler := NewPriceHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"price":100}`)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/price/SKU001", body)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "sku", Value: "SKU001"}}
		c.Set("tenantID", "test-tenant")

		handler.Update(c)

		// Without a valid DB, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPriceHandler_BulkUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewPriceHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"updates":[{"sku":"SKU001","price":100}]}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/price/bulk", body)
		c.Request.Header.Set("Content-Type", "application/json")
		// No tenantID set

		handler.BulkUpdate(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		handler := NewPriceHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{invalid json}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/price/bulk", body)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.BulkUpdate(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("valid request with no DB returns error", func(t *testing.T) {
		handler := NewPriceHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"updates":[{"sku":"SKU001","price":100}]}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/price/bulk", body)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.BulkUpdate(c)

		// Without a valid DB, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
