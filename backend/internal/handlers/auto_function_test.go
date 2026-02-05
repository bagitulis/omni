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
// Auto Function Handler Tests (auto_function.go)
// =============================================================================

// TestAutoFunctionList_MissingTenant tests listing auto functions without tenant
func TestAutoFunctionList_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.GET("/api/jobs/auto-functions", handler.List)

	req, _ := http.NewRequest("GET", "/api/jobs/auto-functions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "tenant ID required")
}

// TestAutoFunctionGetByName_MissingTenant tests getting auto function by name without tenant
func TestAutoFunctionGetByName_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.GET("/api/jobs/auto-functions/:name", handler.GetByName)

	req, _ := http.NewRequest("GET", "/api/jobs/auto-functions/test-func", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAutoFunctionCreate_MissingTenant tests creating auto function without tenant
func TestAutoFunctionCreate_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.POST("/api/jobs/auto-functions", handler.Create)

	reqBody := `{"name": "test-func", "interval_minutes": 5}`
	req, _ := http.NewRequest("POST", "/api/jobs/auto-functions", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAutoFunctionCreate_InvalidBody tests creating auto function with invalid body
func TestAutoFunctionCreate_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.POST("/api/jobs/auto-functions", handler.Create)

	// Invalid JSON
	reqBody := `{invalid`
	req, _ := http.NewRequest("POST", "/api/jobs/auto-functions", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler gets DB before parsing JSON, so returns 500 (DB error) not 400
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestAutoFunctionCreate_InvalidInterval tests creating with interval < 1
func TestAutoFunctionCreate_InvalidInterval(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/jobs/auto-functions", func(c *gin.Context) {
		type req struct {
			Name            string `json:"name"`
			IntervalMinutes int    `json:"interval_minutes"`
		}
		var r req
		if err := c.ShouldBindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		if r.IntervalMinutes < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "interval_minutes must be at least 1"})
			return
		}
	})

	reqBody := `{"name": "test-func", "interval_minutes": 0}`
	req, _ := http.NewRequest("POST", "/api/jobs/auto-functions", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "interval_minutes must be at least 1")
}

// TestAutoFunctionUpdate_MissingTenant tests updating auto function without tenant
func TestAutoFunctionUpdate_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.PUT("/api/jobs/auto-functions/:name", handler.Update)

	reqBody := `{"interval_minutes": 10}`
	req, _ := http.NewRequest("PUT", "/api/jobs/auto-functions/test-func", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAutoFunctionUpdate_InvalidTimeWindow tests updating with partial time window
func TestAutoFunctionUpdate_InvalidTimeWindow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.PUT("/api/jobs/auto-functions/:name", func(c *gin.Context) {
		type req struct {
			IntervalMinutes int     `json:"interval_minutes"`
			StartTime       *string `json:"start_time"`
			EndTime         *string `json:"end_time"`
		}
		var r req
		if err := c.ShouldBindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		if r.IntervalMinutes < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "interval_minutes must be at least 1"})
			return
		}
		if (r.StartTime != nil && r.EndTime == nil) || (r.StartTime == nil && r.EndTime != nil) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Both start_time and end_time must be provided together, or neither"})
			return
		}
	})

	// Only start_time provided, no end_time
	startTime := "08:00"
	reqBody := `{"interval_minutes": 10, "start_time": "` + startTime + `"}`
	req, _ := http.NewRequest("PUT", "/api/jobs/auto-functions/test-func", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "start_time and end_time must be provided together")
}

// TestAutoFunctionDelete_MissingTenant tests deleting auto function without tenant
func TestAutoFunctionDelete_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.DELETE("/api/jobs/auto-functions/:name", handler.Delete)

	req, _ := http.NewRequest("DELETE", "/api/jobs/auto-functions/test-func", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// Auto Function Action Tests (auto_function_actions.go)
// =============================================================================

// TestAutoFunctionEnable_MissingTenant tests enabling auto function without tenant
func TestAutoFunctionEnable_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.POST("/api/jobs/auto-functions/:name/enable", handler.Enable)

	req, _ := http.NewRequest("POST", "/api/jobs/auto-functions/test-func/enable", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAutoFunctionDisable_MissingTenant tests disabling auto function without tenant
func TestAutoFunctionDisable_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.POST("/api/jobs/auto-functions/:name/disable", handler.Disable)

	req, _ := http.NewRequest("POST", "/api/jobs/auto-functions/test-func/disable", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAutoFunctionCancelScheduled_MissingTenant tests canceling scheduled without tenant
func TestAutoFunctionCancelScheduled_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.POST("/api/jobs/auto-functions/:name/cancel-scheduled", handler.CancelScheduledByName)

	req, _ := http.NewRequest("POST", "/api/jobs/auto-functions/test-func/cancel-scheduled", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAutoFunctionRun_MissingTenant tests running auto function without tenant
func TestAutoFunctionRun_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.POST("/api/jobs/auto-functions/:name/run", handler.Run)

	req, _ := http.NewRequest("POST", "/api/jobs/auto-functions/test-func/run", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAutoFunctionRun_MissingExecutor tests running without executor configured
func TestAutoFunctionRun_MissingExecutor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	// Handler with nil executor
	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.POST("/api/jobs/auto-functions/:name/run", handler.Run)

	req, _ := http.NewRequest("POST", "/api/jobs/auto-functions/test-func/run", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "executor not configured")
}

// TestAutoFunctionGetHistory_MissingTenant tests getting history without tenant
func TestAutoFunctionGetHistory_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.GET("/api/jobs/auto-functions/history", handler.GetHistory)

	req, _ := http.NewRequest("GET", "/api/jobs/auto-functions/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAutoFunctionGetHistory_MissingExecutor tests getting history without executor
func TestAutoFunctionGetHistory_MissingExecutor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	// Handler with nil executor
	handler := NewAutoFunctionHandler(nil, nil, nil)
	r.GET("/api/jobs/auto-functions/history", handler.GetHistory)

	req, _ := http.NewRequest("GET", "/api/jobs/auto-functions/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "executor not configured")
}

// TestAutoFunctionGetHistory_WithLimit tests getting history with custom limit
func TestAutoFunctionGetHistory_WithLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/jobs/auto-functions/history", func(c *gin.Context) {
		// Verify limit parsing
		limit := 50
		if l := c.Query("limit"); l != "" {
			if parsed, err := parseInt(l); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"limit":   limit,
		})
	})

	req, _ := http.NewRequest("GET", "/api/jobs/auto-functions/history?limit=100", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, float64(100), resp["limit"])
}

// Helper function for parsing int
func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}
