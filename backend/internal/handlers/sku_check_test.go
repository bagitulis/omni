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

// TestSKUCheckHandler_CheckSingle_MissingTenant tests CheckSingle without tenant
func TestSKUCheckHandler_CheckSingle_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSKUCheckHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/sku/check/SKU001", nil)
	c.Params = gin.Params{{Key: "sku", Value: "SKU001"}}
	// No tenantID set

	handler.CheckSingle(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "tenant ID required")
}

// TestSKUCheckHandler_CheckSingle_MissingSKU tests CheckSingle without SKU
func TestSKUCheckHandler_CheckSingle_MissingSKU(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSKUCheckHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/sku/check/", nil)
	c.Params = gin.Params{{Key: "sku", Value: ""}}
	c.Set("tenantID", "test-tenant")

	handler.CheckSingle(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "SKU required")
}

// TestSKUCheckHandler_CheckSingle_NoAPIs tests CheckSingle without platform APIs
func TestSKUCheckHandler_CheckSingle_NoAPIs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSKUCheckHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/sku/check/SKU001", nil)
	c.Params = gin.Params{{Key: "sku", Value: "SKU001"}}
	c.Set("tenantID", "test-tenant")
	// No platform APIs set

	handler.CheckSingle(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "no platform APIs")
}

// TestSKUCheckHandler_CheckBatch_MissingTenant tests CheckBatch without tenant
func TestSKUCheckHandler_CheckBatch_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSKUCheckHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{"skus":["SKU001","SKU002"]}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/sku/batch-check", body)
	c.Request.Header.Set("Content-Type", "application/json")
	// No tenantID set

	handler.CheckBatch(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestSKUCheckHandler_CheckBatch_InvalidBody tests CheckBatch with invalid body
func TestSKUCheckHandler_CheckBatch_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSKUCheckHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := bytes.NewBufferString(`{}`) // Missing required skus
	c.Request = httptest.NewRequest(http.MethodPost, "/api/sku/batch-check", body)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("tenantID", "test-tenant")

	handler.CheckBatch(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSKUCheckHandler_CheckBatch_TooManySKUs tests CheckBatch with too many SKUs
func TestSKUCheckHandler_CheckBatch_TooManySKUs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSKUCheckHandler(nil)

	// Create array of 101 SKUs
	skus := make([]string, 101)
	for i := 0; i < 101; i++ {
		skus[i] = "SKU" + string(rune('0'+i%10))
	}
	reqBody, _ := json.Marshal(map[string]interface{}{"skus": skus})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/sku/batch-check", bytes.NewBuffer(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("tenantID", "test-tenant")

	handler.CheckBatch(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["error"], "max 100")
}

// TestSKUCheckHandler_GetCachedStatus_MissingTenant tests GetCachedStatus without tenant
func TestSKUCheckHandler_GetCachedStatus_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSKUCheckHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/sku/status/SKU001", nil)
	c.Params = gin.Params{{Key: "sku", Value: "SKU001"}}
	// No tenantID set

	handler.GetCachedStatus(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestSKUCheckHandler_GetCachedStatus_MissingSKU tests GetCachedStatus without SKU
func TestSKUCheckHandler_GetCachedStatus_MissingSKU(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewSKUCheckHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/sku/status/", nil)
	c.Params = gin.Params{{Key: "sku", Value: ""}}
	c.Set("tenantID", "test-tenant")

	handler.GetCachedStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestNewSKUCheckHandler tests handler creation
func TestNewSKUCheckHandler(t *testing.T) {
	handler := NewSKUCheckHandler(nil)
	assert.NotNil(t, handler)
}
