package handlers

import (
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

// TestAutoFunctionHandler_Enable_MissingTenant tests Enable without tenant
func TestAutoFunctionHandler_Enable_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.POST("/api/jobs/auto-functions/:name/enable", handler.Enable)

	req, _ := http.NewRequest("POST", "/api/jobs/auto-functions/test-func/enable", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "tenant")
}

// TestAutoFunctionHandler_Disable_MissingTenant tests Disable without tenant
func TestAutoFunctionHandler_Disable_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.POST("/api/jobs/auto-functions/:name/disable", handler.Disable)

	req, _ := http.NewRequest("POST", "/api/jobs/auto-functions/test-func/disable", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestAutoFunctionHandler_CancelScheduledByName_MissingTenant tests CancelScheduledByName without tenant
func TestAutoFunctionHandler_CancelScheduledByName_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.POST("/api/jobs/auto-functions/:name/cancel-scheduled", handler.CancelScheduledByName)

	req, _ := http.NewRequest("POST", "/api/jobs/auto-functions/test-func/cancel-scheduled", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestAutoFunctionHandler_Run_MissingTenant tests Run without tenant
func TestAutoFunctionHandler_Run_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.POST("/api/jobs/auto-functions/:name/run", handler.Run)

	req, _ := http.NewRequest("POST", "/api/jobs/auto-functions/test-func/run", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestAutoFunctionHandler_Run_NoExecutor tests Run with tenant but no executor configured
func TestAutoFunctionHandler_Run_NoExecutor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.POST("/api/jobs/auto-functions/:name/run", handler.Run)

	req, _ := http.NewRequest("POST", "/api/jobs/auto-functions/test-func/run", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "executor not configured")
}

// TestAutoFunctionHandler_GetHistory_MissingTenant tests GetHistory without tenant
func TestAutoFunctionHandler_GetHistory_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.GET("/api/jobs/auto-functions/history", handler.GetHistory)

	req, _ := http.NewRequest("GET", "/api/jobs/auto-functions/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// TestAutoFunctionHandler_GetHistory_NoExecutor tests GetHistory with tenant but no executor
func TestAutoFunctionHandler_GetHistory_NoExecutor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.GET("/api/jobs/auto-functions/history", handler.GetHistory)

	req, _ := http.NewRequest("GET", "/api/jobs/auto-functions/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"], "executor not configured")
}
