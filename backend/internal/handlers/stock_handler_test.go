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

func init() {
	gin.SetMode(gin.TestMode)
}

// TestStockHandler_List_MissingTenant tests List without tenant
func TestStockHandler_List_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStockHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/stock", nil)
	// No tenantID set

	handler.List(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "tenant_id")
}

// TestStockHandler_GetBySKU_MissingTenant tests GetBySKU without tenant
func TestStockHandler_GetBySKU_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStockHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/stock/SKU001", nil)
	c.Params = gin.Params{{Key: "sku", Value: "SKU001"}}
	// No tenantID set

	handler.GetBySKU(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestStockHandler_GetBySKU_MissingSKU tests GetBySKU without SKU
func TestStockHandler_GetBySKU_MissingSKU(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStockHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/stock/", nil)
	c.Params = gin.Params{{Key: "sku", Value: ""}}
	c.Set("tenant_id", "test-tenant")

	handler.GetBySKU(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "SKU required")
}

// TestStockHandler_Update_MissingTenant tests Update without tenant
func TestStockHandler_Update_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStockHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{"quantity":100}`)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/stock/SKU001", body)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "sku", Value: "SKU001"}}
	// No tenantID set

	handler.Update(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestStockHandler_Update_MissingSKU tests Update without SKU
func TestStockHandler_Update_MissingSKU(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStockHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{"quantity":100}`)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/stock/", body)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "sku", Value: ""}}
	c.Set("tenant_id", "test-tenant")

	handler.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestStockHandler_Update_InvalidBody tests Update with invalid body
func TestStockHandler_Update_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStockHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{}`) // Missing required quantity
	c.Request = httptest.NewRequest(http.MethodPut, "/api/stock/SKU001", body)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "sku", Value: "SKU001"}}
	c.Set("tenant_id", "test-tenant")

	handler.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestStockHandler_BulkUpdate_MissingTenant tests BulkUpdate without tenant
func TestStockHandler_BulkUpdate_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStockHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{"updates":[{"sku":"SKU001","quantity":100}]}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/stock/bulk", body)
	c.Request.Header.Set("Content-Type", "application/json")
	// No tenantID set

	handler.BulkUpdate(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestStockHandler_BulkUpdate_InvalidJSON tests BulkUpdate with invalid JSON
func TestStockHandler_BulkUpdate_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStockHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{invalid json}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/stock/bulk", body)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("tenant_id", "test-tenant")

	handler.BulkUpdate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestStockHandler_GetAlerts_MissingTenant tests GetAlerts without tenant
func TestStockHandler_GetAlerts_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStockHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/stock/alerts", nil)
	// No tenantID set

	handler.GetAlerts(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestNewStockHandler tests handler creation
func TestNewStockHandler(t *testing.T) {
	handler := NewStockHandler(nil)
	assert.NotNil(t, handler)
}
