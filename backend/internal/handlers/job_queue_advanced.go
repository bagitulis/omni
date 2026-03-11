package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/jobs"
)

// GetQueue handles GET /api/jobs/queue
func (h *JobQueueHandler) GetQueue(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	qm := jobs.NewQueueManager(db, tenantID)
	queue := qm.GetPendingQueue(limit)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"queue":   queue,
		"total":   len(queue),
	})
}

// GetMonitor handles GET /api/jobs/monitor
func (h *JobQueueHandler) GetMonitor(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	qm := jobs.NewQueueManager(db, tenantID)
	status := qm.GetQueueStatus()
	recentHistory := qm.GetRecentHistory(10)

	// Also get auto_functions_history and merge
	var autoFuncHistory []models.AutoFunctionHistory
	db.Order("executed_at DESC").Limit(10).Find(&autoFuncHistory)

	// Convert auto_functions_history to unified format and merge
	combinedHistory := mergeHistories(recentHistory, autoFuncHistory)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"current_job":     status.CurrentJob,
			"pending_queue":   status.PendingQueue,
			"total_pending":   status.TotalPending,
			"total_completed": status.TotalCompleted,
			"recent_history":  combinedHistory,
		},
	})
}

// GetHistoryPaginated handles GET /api/jobs/history-paginated
func (h *JobQueueHandler) GetHistoryPaginated(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	filter := jobs.PaginatedFilter{
		Page:      page,
		PageSize:  pageSize,
		Status:    c.Query("status"),
		JobType:   c.Query("jobType"),
		Search:    c.Query("search"),
		SortBy:    c.Query("sortBy"),
		SortOrder: c.DefaultQuery("sortOrder", "desc"),
	}

	qm := jobs.NewQueueManager(db, tenantID)
	result, err := qm.GetHistoryPaginated(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetHistoryJobTypes handles GET /api/jobs/history-job-types
func (h *JobQueueHandler) GetHistoryJobTypes(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	qm := jobs.NewQueueManager(db, tenantID)
	jobTypes := qm.GetDistinctJobTypes()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"job_types": jobTypes,
		},
	})
}

// ClearHistory handles DELETE /api/jobs/history
func (h *JobQueueHandler) ClearHistory(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	qm := jobs.NewQueueManager(db, tenantID)
	deleted := qm.ClearJobHistory()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"deleted": deleted,
		"message": "history records deleted",
	})
}

// CancelJobByJobId handles POST /api/jobs/cancel/:jobId
func (h *JobQueueHandler) CancelJobByJobId(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	jobID := c.Param("jobId")
	qm := jobs.NewQueueManager(db, tenantID)
	success := qm.CancelJobByStringID(jobID, false)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"cancelled": success,
	})
}

// ForceCancelJob handles POST /api/jobs/force-cancel/:jobId
func (h *JobQueueHandler) ForceCancelJob(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	jobID := c.Param("jobId")
	qm := jobs.NewQueueManager(db, tenantID)
	success := qm.CancelJobByStringID(jobID, true)

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"force_cancelled": success,
	})
}

// CheckTimeout handles GET /api/jobs/check-timeout
func (h *JobQueueHandler) CheckTimeout(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	timeoutMinutes := 5
	if t := c.Query("timeout"); t != "" {
		if parsed, err := strconv.Atoi(t); err == nil && parsed > 0 {
			timeoutMinutes = parsed
		}
	}

	qm := jobs.NewQueueManager(db, tenantID)
	timedOutJobs := qm.CheckAndTimeoutStuckJobs(timeoutMinutes)

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"timed_out_jobs":  timedOutJobs,
		"count":           len(timedOutJobs),
		"timeout_minutes": timeoutMinutes,
	})
}
