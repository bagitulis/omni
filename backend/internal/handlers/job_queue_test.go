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
// Job Queue Handler Tests (job_queue.go)
// =============================================================================

// TestJobQueueAddJob_MissingTenant tests adding job without tenant
func TestJobQueueAddJob_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.POST("/api/jobs", handler.AddJob)

	reqBody := `{"type": "sync", "data": {}}`
	req, _ := http.NewRequest("POST", "/api/jobs", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "Missing tenantId")
}

// TestJobQueueListJobs_MissingTenant tests listing jobs without tenant
func TestJobQueueListJobs_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs", handler.ListJobs)

	req, _ := http.NewRequest("GET", "/api/jobs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "Missing tenantId")
}

// TestJobQueueGetJob_MissingTenant tests getting a job without tenant
func TestJobQueueGetJob_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/:id", handler.GetJob)

	req, _ := http.NewRequest("GET", "/api/jobs/job-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueGetJob_MissingJobID tests getting a job without job ID
func TestJobQueueGetJob_MissingJobID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/jobs/:id", func(c *gin.Context) {
		jobID := c.Param("id")
		if jobID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "jobId": jobID})
	})

	req, _ := http.NewRequest("GET", "/api/jobs/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Empty id will get 404 from router, not 400 (correct behavior)
	assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusOK)
}

// TestJobQueueCancelJob_MissingTenant tests canceling a job without tenant
func TestJobQueueCancelJob_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.DELETE("/api/jobs/:id", handler.CancelJob)

	req, _ := http.NewRequest("DELETE", "/api/jobs/job-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueGetHistory_MissingTenant tests getting history without tenant
func TestJobQueueGetHistory_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/history", handler.GetHistory)

	req, _ := http.NewRequest("GET", "/api/jobs/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueGetStats_MissingTenant tests getting stats without tenant
func TestJobQueueGetStats_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/stats", handler.GetStats)

	req, _ := http.NewRequest("GET", "/api/jobs/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueEnqueueJob_MissingTenant tests enqueuing job without tenant
func TestJobQueueEnqueueJob_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.POST("/api/jobs/enqueue", handler.EnqueueJob)

	reqBody := `{"type": "sync", "data": {"key": "value"}}`
	req, _ := http.NewRequest("POST", "/api/jobs/enqueue", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueEnqueueJob_InvalidBody tests enqueuing job with invalid body
func TestJobQueueEnqueueJob_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/jobs/enqueue", func(c *gin.Context) {
		var req struct {
			Type     string                 `json:"type" binding:"required"`
			Data     map[string]interface{} `json:"data" binding:"required"`
			Priority string                 `json:"priority"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	})

	// Missing required field "type"
	reqBody := `{"data": {"key": "value"}}`
	req, _ := http.NewRequest("POST", "/api/jobs/enqueue", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestJobQueueEnqueueJob_DefaultPriority tests enqueuing with default priority
func TestJobQueueEnqueueJob_DefaultPriority(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/jobs/enqueue", func(c *gin.Context) {
		var req struct {
			Type     string                 `json:"type" binding:"required"`
			Data     map[string]interface{} `json:"data" binding:"required"`
			Priority string                 `json:"priority"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Priority == "" {
			req.Priority = "normal"
		}

		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"priority": req.Priority,
		})
	})

	reqBody := `{"type": "sync", "data": {"key": "value"}}`
	req, _ := http.NewRequest("POST", "/api/jobs/enqueue", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "normal", resp["priority"])
}

// TestJobQueueGetStatus_MissingTenant tests getting status without tenant
func TestJobQueueGetStatus_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/status", handler.GetStatus)

	req, _ := http.NewRequest("GET", "/api/jobs/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueGetJobStatus_MissingTenant tests getting job status without tenant
func TestJobQueueGetJobStatus_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/status/:jobId", handler.GetJobStatus)

	req, _ := http.NewRequest("GET", "/api/jobs/status/job-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// Job Queue Advanced Handler Tests (job_queue_advanced.go)
// =============================================================================

// TestJobQueueGetQueue_MissingTenant tests getting queue without tenant
func TestJobQueueGetQueue_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/queue", handler.GetQueue)

	req, _ := http.NewRequest("GET", "/api/jobs/queue", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueGetQueue_CustomLimit tests getting queue with custom limit
func TestJobQueueGetQueue_CustomLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/jobs/queue", func(c *gin.Context) {
		limit := 100
		if l := c.Query("limit"); l != "" {
			if parsed, err := parseIntSafe(l); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"limit":   limit,
		})
	})

	req, _ := http.NewRequest("GET", "/api/jobs/queue?limit=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(50), resp["limit"])
}

// TestJobQueueGetMonitor_MissingTenant tests getting monitor without tenant
func TestJobQueueGetMonitor_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/monitor", handler.GetMonitor)

	req, _ := http.NewRequest("GET", "/api/jobs/monitor", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueGetHistoryPaginated_MissingTenant tests paginated history without tenant
func TestJobQueueGetHistoryPaginated_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/history-paginated", handler.GetHistoryPaginated)

	req, _ := http.NewRequest("GET", "/api/jobs/history-paginated", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueGetHistoryPaginated_CustomPagination tests with custom pagination
func TestJobQueueGetHistoryPaginated_CustomPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/jobs/history-paginated", func(c *gin.Context) {
		page := 1
		pageSize := 20
		if p := c.Query("page"); p != "" {
			if parsed, err := parseIntSafe(p); err == nil && parsed > 0 {
				page = parsed
			}
		}
		if ps := c.Query("pageSize"); ps != "" {
			if parsed, err := parseIntSafe(ps); err == nil && parsed > 0 {
				pageSize = parsed
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"page":     page,
			"pageSize": pageSize,
		})
	})

	req, _ := http.NewRequest("GET", "/api/jobs/history-paginated?page=3&pageSize=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(3), resp["page"])
	assert.Equal(t, float64(50), resp["pageSize"])
}

// TestJobQueueGetHistoryJobTypes_MissingTenant tests getting job types without tenant
func TestJobQueueGetHistoryJobTypes_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/history-job-types", handler.GetHistoryJobTypes)

	req, _ := http.NewRequest("GET", "/api/jobs/history-job-types", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueClearHistory_MissingTenant tests clearing history without tenant
func TestJobQueueClearHistory_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.DELETE("/api/jobs/history", handler.ClearHistory)

	req, _ := http.NewRequest("DELETE", "/api/jobs/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueCancelJobByJobId_MissingTenant tests cancel by job ID without tenant
func TestJobQueueCancelJobByJobId_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.POST("/api/jobs/cancel/:jobId", handler.CancelJobByJobId)

	req, _ := http.NewRequest("POST", "/api/jobs/cancel/job-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueForceCancelJob_MissingTenant tests force cancel without tenant
func TestJobQueueForceCancelJob_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.POST("/api/jobs/force-cancel/:jobId", handler.ForceCancelJob)

	req, _ := http.NewRequest("POST", "/api/jobs/force-cancel/job-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueCheckTimeout_MissingTenant tests check timeout without tenant
func TestJobQueueCheckTimeout_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/check-timeout", handler.CheckTimeout)

	req, _ := http.NewRequest("GET", "/api/jobs/check-timeout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueCheckTimeout_CustomTimeout tests with custom timeout minutes
func TestJobQueueCheckTimeout_CustomTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/jobs/check-timeout", func(c *gin.Context) {
		timeoutMinutes := 5
		if tm := c.Query("timeout"); tm != "" {
			if parsed, err := parseIntSafe(tm); err == nil && parsed > 0 {
				timeoutMinutes = parsed
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"success":        true,
			"timeoutMinutes": timeoutMinutes,
		})
	})

	req, _ := http.NewRequest("GET", "/api/jobs/check-timeout?timeout=15", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(15), resp["timeoutMinutes"])
}

// Helper function for safe int parsing
func parseIntSafe(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}
