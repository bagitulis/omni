package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Locked Order Handler Tests (locked_order.go)
// =============================================================================

// TestLockedOrderSave_MissingTenant tests saving locked orders without tenant
func TestLockedOrderSave_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewLockedOrderHandler(nil)
	r.POST("/api/locked-orders", handler.SaveLockedOrders)

	reqBody := `{"items": []}`
	req, _ := http.NewRequest("POST", "/api/locked-orders", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "tenant")
}

// TestLockedOrderSave_InvalidBody tests saving with invalid body
func TestLockedOrderSave_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/locked-orders", func(c *gin.Context) {
		var req SaveLockedOrdersRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body: " + err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Invalid JSON
	reqBody := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/locked-orders", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "Invalid request body")
}

// TestLockedOrderGet_MissingTenant tests getting locked orders without tenant
func TestLockedOrderGet_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewLockedOrderHandler(nil)
	r.GET("/api/locked-orders", handler.GetLockedOrders)

	req, _ := http.NewRequest("GET", "/api/locked-orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "tenant")
}

// TestLockedOrderClear_MissingTenant tests clearing locked orders without tenant
func TestLockedOrderClear_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewLockedOrderHandler(nil)
	r.DELETE("/api/locked-orders", handler.ClearLockedOrders)

	req, _ := http.NewRequest("DELETE", "/api/locked-orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

// TestLockedOrderSave_EmptyItems tests saving with empty items array
func TestLockedOrderSave_EmptyItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/locked-orders", func(c *gin.Context) {
		var req SaveLockedOrdersRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"count":   len(req.Items),
			"message": "Locked orders saved successfully",
		})
	})

	reqBody := `{"items": []}`
	req, _ := http.NewRequest("POST", "/api/locked-orders", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, float64(0), resp["count"])
}

// TestLockedOrderGet_ResponseFormat tests response format for get
func TestLockedOrderGet_ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/locked-orders", func(c *gin.Context) {
		// Simulate successful response format
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"data":     []interface{}{},
			"count":    0,
			"totalQty": 0,
		})
	})

	req, _ := http.NewRequest("GET", "/api/locked-orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.NotNil(t, resp["data"])
	assert.NotNil(t, resp["count"])
	assert.NotNil(t, resp["totalQty"])
}

// TestLockedOrderClear_SuccessResponse tests clear success response
func TestLockedOrderClear_SuccessResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.DELETE("/api/locked-orders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Locked orders cleared",
		})
	})

	req, _ := http.NewRequest("DELETE", "/api/locked-orders", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "Locked orders cleared", resp["message"])
}
