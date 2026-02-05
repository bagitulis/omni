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

// TestInventoryHandler_GetRecordByKey_MissingTenant tests GetRecordByKey without tenant
func TestInventoryHandler_GetRecordByKey_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewInventoryHandler(nil)
	r.GET("/api/inventory/:keyValue", handler.GetRecordByKey)

	req, _ := http.NewRequest("GET", "/api/inventory/SKU123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestInventoryHandler_GetRecordByKey_EmptyKeyValue tests GetRecordByKey with empty key
func TestInventoryHandler_GetRecordByKey_EmptyKeyValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewInventoryHandler(nil)
	// Route with empty param would be different path, test path param behavior
	r.GET("/api/inventory/:keyValue", handler.GetRecordByKey)

	req, _ := http.NewRequest("GET", "/api/inventory/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Empty param returns 404 (route not matched)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestInventoryHandler_UpdateRecordByKey_MissingTenant tests UpdateRecordByKey without tenant
func TestInventoryHandler_UpdateRecordByKey_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewInventoryHandler(nil)
	r.PUT("/api/inventory/:keyValue", handler.UpdateRecordByKey)

	req, _ := http.NewRequest("PUT", "/api/inventory/SKU123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestInventoryHandler_UpdateRecordByKey_InvalidBody tests UpdateRecordByKey with invalid JSON
func TestInventoryHandler_UpdateRecordByKey_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewInventoryHandler(nil)
	r.PUT("/api/inventory/:keyValue", handler.UpdateRecordByKey)

	req, _ := http.NewRequest("PUT", "/api/inventory/SKU123", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestInventoryHandler_CreateRecord_MissingTenant tests CreateRecord without tenant
func TestInventoryHandler_CreateRecord_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewInventoryHandler(nil)
	r.POST("/api/inventory", handler.CreateRecord)

	req, _ := http.NewRequest("POST", "/api/inventory", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestInventoryHandler_CreateRecord_InvalidBody tests CreateRecord with invalid JSON
func TestInventoryHandler_CreateRecord_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewInventoryHandler(nil)
	r.POST("/api/inventory", handler.CreateRecord)

	req, _ := http.NewRequest("POST", "/api/inventory", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestInventoryHandler_CreateRecord_MissingKeyValue tests CreateRecord without keyValue
func TestInventoryHandler_CreateRecord_MissingKeyValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewInventoryHandler(nil)
	r.POST("/api/inventory", handler.CreateRecord)

	body := `{"name": "Test Product"}`
	req, _ := http.NewRequest("POST", "/api/inventory", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "keyValue")
}

// TestInventoryHandler_DeleteRecordByKey_MissingTenant tests DeleteRecordByKey without tenant
func TestInventoryHandler_DeleteRecordByKey_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewInventoryHandler(nil)
	r.DELETE("/api/inventory/:keyValue", handler.DeleteRecordByKey)

	req, _ := http.NewRequest("DELETE", "/api/inventory/SKU123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
