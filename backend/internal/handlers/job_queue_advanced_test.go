package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestJobQueueHandler_GetQueue_MissingTenant tests GetQueue without tenant
func TestJobQueueHandler_GetQueue_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/queue", handler.GetQueue)

	req, _ := http.NewRequest("GET", "/api/jobs/queue", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "tenant")
}

// TestJobQueueHandler_GetMonitor_MissingTenant tests GetMonitor without tenant
func TestJobQueueHandler_GetMonitor_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/monitor", handler.GetMonitor)

	req, _ := http.NewRequest("GET", "/api/jobs/monitor", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "tenant")
}

// TestJobQueueHandler_GetHistoryPaginated_MissingTenant tests GetHistoryPaginated without tenant
func TestJobQueueHandler_GetHistoryPaginated_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/history-paginated", handler.GetHistoryPaginated)

	req, _ := http.NewRequest("GET", "/api/jobs/history-paginated", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueHandler_GetHistoryJobTypes_MissingTenant tests GetHistoryJobTypes without tenant
func TestJobQueueHandler_GetHistoryJobTypes_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/history-job-types", handler.GetHistoryJobTypes)

	req, _ := http.NewRequest("GET", "/api/jobs/history-job-types", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueHandler_ClearHistory_MissingTenant tests ClearHistory without tenant
func TestJobQueueHandler_ClearHistory_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.DELETE("/api/jobs/history", handler.ClearHistory)

	req, _ := http.NewRequest("DELETE", "/api/jobs/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueHandler_CancelJobByJobId_MissingTenant tests CancelJobByJobId without tenant
func TestJobQueueHandler_CancelJobByJobId_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.POST("/api/jobs/cancel/:jobId", handler.CancelJobByJobId)

	req, _ := http.NewRequest("POST", "/api/jobs/cancel/job-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueHandler_ForceCancelJob_MissingTenant tests ForceCancelJob without tenant
func TestJobQueueHandler_ForceCancelJob_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.POST("/api/jobs/force-cancel/:jobId", handler.ForceCancelJob)

	req, _ := http.NewRequest("POST", "/api/jobs/force-cancel/job-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestJobQueueHandler_CheckTimeout_MissingTenant tests CheckTimeout without tenant
func TestJobQueueHandler_CheckTimeout_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewJobQueueHandler(nil)
	r.GET("/api/jobs/check-timeout", handler.CheckTimeout)

	req, _ := http.NewRequest("GET", "/api/jobs/check-timeout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
