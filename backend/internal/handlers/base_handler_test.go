package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestBaseHandler_ErrorResponse tests error response formatting
func TestBaseHandler_ErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewBaseHandler(nil)

	r.GET("/test-error", func(c *gin.Context) {
		handler.ErrorResponse(c, http.StatusBadRequest, "Test error message")
	})

	req, _ := http.NewRequest("GET", "/test-error", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Test error message", resp["error"])
}

// TestBaseHandler_SuccessResponse tests success response formatting
func TestBaseHandler_SuccessResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewBaseHandler(nil)

	r.GET("/test-success", func(c *gin.Context) {
		handler.SuccessResponse(c, gin.H{"key": "value"})
	})

	req, _ := http.NewRequest("GET", "/test-success", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value", data["key"])
}

// TestBaseHandler_SuccessResponseWithMessage tests success response with message
func TestBaseHandler_SuccessResponseWithMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewBaseHandler(nil)

	r.GET("/test-message", func(c *gin.Context) {
		handler.SuccessResponseWithMessage(c, "Operation completed successfully")
	})

	req, _ := http.NewRequest("GET", "/test-message", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "Operation completed successfully", resp["message"])
}

// TestBaseHandler_PaginatedResponse tests paginated response formatting
func TestBaseHandler_PaginatedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewBaseHandler(nil)

	r.GET("/test-paginated", func(c *gin.Context) {
		items := []string{"item1", "item2", "item3"}
		handler.PaginatedResponse(c, items, 100, 1, 20)
	})

	req, _ := http.NewRequest("GET", "/test-paginated", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	// Check data
	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 3)

	// Check meta
	meta, ok := resp["meta"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(100), meta["total"])
	assert.Equal(t, float64(1), meta["page"])
	assert.Equal(t, float64(20), meta["limit"])
}

// TestBaseHandler_GetTenantID_Success tests successful tenant ID extraction
func TestBaseHandler_GetTenantID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewBaseHandler(nil)

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant-123")
		c.Next()
	})

	r.GET("/test-tenant", func(c *gin.Context) {
		tenantID, ok := handler.GetTenantID(c)
		if !ok {
			return
		}
		c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID})
	})

	req, _ := http.NewRequest("GET", "/test-tenant", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "test-tenant-123", resp["tenant_id"])
}

// TestBaseHandler_GetTenantID_Missing tests missing tenant ID handling
func TestBaseHandler_GetTenantID_Missing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewBaseHandler(nil)

	r.GET("/test-tenant-missing", func(c *gin.Context) {
		tenantID, ok := handler.GetTenantID(c)
		if !ok {
			return // ErrorResponse already sent
		}
		c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID})
	})

	req, _ := http.NewRequest("GET", "/test-tenant-missing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "Missing tenantID", resp["error"])
}

// TestBaseHandler_SuccessResponseRaw tests raw response formatting
func TestBaseHandler_SuccessResponseRaw(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewBaseHandler(nil)

	r.GET("/test-raw", func(c *gin.Context) {
		handler.SuccessResponseRaw(c, gin.H{
			"custom_field": "custom_value",
			"another":      123,
		})
	})

	req, _ := http.NewRequest("GET", "/test-raw", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "custom_value", resp["custom_field"])
	assert.Equal(t, float64(123), resp["another"])
}
