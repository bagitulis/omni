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

// TestSkuBatchCheckHandler_BatchCheckSku_MissingTenant tests BatchCheckSku without tenant
func TestSkuBatchCheckHandler_BatchCheckSku_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSkuBatchCheckHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{"skus":["SKU001","SKU002"]}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/batch-check-sku", body)
	c.Request.Header.Set("Content-Type", "application/json")
	// No tenantID set

	handler.BatchCheckSku(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "Missing tenantId")
}

// TestSkuBatchCheckHandler_BatchCheckSku_InvalidBody tests BatchCheckSku with invalid body
func TestSkuBatchCheckHandler_BatchCheckSku_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSkuBatchCheckHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{}`) // Missing required skus
	c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/batch-check-sku", body)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("tenantID", "test-tenant")

	handler.BatchCheckSku(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "skus array is required")
}

// TestSkuBatchCheckHandler_BatchCheckSku_EmptySkus tests BatchCheckSku with empty array
func TestSkuBatchCheckHandler_BatchCheckSku_EmptySkus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSkuBatchCheckHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{"skus":[]}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/batch-check-sku", body)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("tenantID", "test-tenant")

	handler.BatchCheckSku(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "cannot be empty")
}

// TestSkuBatchCheckHandler_BatchSavePlatformStatus_MissingTenant tests BatchSavePlatformStatus without tenant
func TestSkuBatchCheckHandler_BatchSavePlatformStatus_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSkuBatchCheckHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{"results":[{"sku":"SKU001","lazada":true,"shopee":false,"tiktok":true}]}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/batch-save-platform-status", body)
	c.Request.Header.Set("Content-Type", "application/json")
	// No tenantID set

	handler.BatchSavePlatformStatus(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "Missing tenantId")
}

// TestSkuBatchCheckHandler_BatchSavePlatformStatus_InvalidBody tests BatchSavePlatformStatus with invalid body
func TestSkuBatchCheckHandler_BatchSavePlatformStatus_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSkuBatchCheckHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{}`) // Missing required results
	c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/batch-save-platform-status", body)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("tenantID", "test-tenant")

	handler.BatchSavePlatformStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "results array is required")
}

// TestSkuBatchCheckHandler_BatchSavePlatformStatus_EmptyResults tests BatchSavePlatformStatus with empty array
func TestSkuBatchCheckHandler_BatchSavePlatformStatus_EmptyResults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSkuBatchCheckHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{"results":[]}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/batch-save-platform-status", body)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("tenantID", "test-tenant")

	handler.BatchSavePlatformStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "cannot be empty")
}

// TestSkuBatchCheckHandler_GetPlatformStatus_MissingTenant tests GetPlatformStatus without tenant
func TestSkuBatchCheckHandler_GetPlatformStatus_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSkuBatchCheckHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/platform-status", nil)
	// No tenantID set

	handler.GetPlatformStatus(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "Missing tenantId")
}

// TestSkuCheckResult_Structure tests SkuCheckResult structure
func TestSkuCheckResult_Structure(t *testing.T) {
	result := SkuCheckResult{
		Sku:    "TEST-SKU-001",
		Lazada: true,
		Shopee: false,
		Tiktok: true,
	}

	assert.Equal(t, "TEST-SKU-001", result.Sku)
	assert.True(t, result.Lazada)
	assert.False(t, result.Shopee)
	assert.True(t, result.Tiktok)
}

// TestSkuPlatformCheckResult_TableName tests table name
func TestSkuPlatformCheckResult_TableName(t *testing.T) {
	result := SkuPlatformCheckResult{}
	assert.Equal(t, "sku_platform_check_results", result.TableName())
}

// TestNewSkuBatchCheckHandler tests handler creation
func TestNewSkuBatchCheckHandler(t *testing.T) {
	handler := NewSkuBatchCheckHandler(nil)
	assert.NotNil(t, handler)
}
