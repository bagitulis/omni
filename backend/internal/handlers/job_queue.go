package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/jobs"
	"gorm.io/gorm"
)

// JobQueueHandler handles job queue endpoints
type JobQueueHandler struct {
	fallbackDB *gorm.DB
}

// NewJobQueueHandler creates a new job queue handler
func NewJobQueueHandler(db *gorm.DB) *JobQueueHandler {
	return &JobQueueHandler{fallbackDB: db}
}

// getDB returns the appropriate database for the current request
func (h *JobQueueHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// AddJob handles POST /api/jobs
func (h *JobQueueHandler) AddJob(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req models.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	qm := jobs.NewQueueManager(db, tenantID)
	job, err := qm.AddJob(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "job": job})
}

// ListJobs handles GET /api/jobs
func (h *JobQueueHandler) ListJobs(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var filter models.JobFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	qm := jobs.NewQueueManager(db, tenantID)
	jobsList, total, err := qm.ListJobs(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"jobs":    jobsList,
		"total":   total,
		"page":    filter.Page,
	})
}

// GetJob handles GET /api/jobs/:id
func (h *JobQueueHandler) GetJob(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}

	qm := jobs.NewQueueManager(db, tenantID)
	job, err := qm.GetJob(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "job": job})
}

// CancelJob handles DELETE /api/jobs/:id
func (h *JobQueueHandler) CancelJob(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job ID"})
		return
	}

	qm := jobs.NewQueueManager(db, tenantID)
	if err := qm.CancelJob(jobID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "job cancelled"})
}

// GetHistory handles GET /api/jobs/history
func (h *JobQueueHandler) GetHistory(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	qm := jobs.NewQueueManager(db, tenantID)
	history := qm.GetRecentHistory(100)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"history": history,
		"total":   len(history),
	})
}

// GetStats handles GET /api/jobs/stats
func (h *JobQueueHandler) GetStats(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	qm := jobs.NewQueueManager(db, tenantID)
	stats, err := qm.GetJobStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "stats": stats})
}

// EnqueueJob handles POST /api/jobs/enqueue
func (h *JobQueueHandler) EnqueueJob(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

	qm := jobs.NewQueueManager(db, tenantID)
	jobID, err := qm.EnqueueJob(req.Type, req.Data, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "jobId": jobID, "status": "pending"})
}

// GetStatus handles GET /api/jobs/status
func (h *JobQueueHandler) GetStatus(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	qm := jobs.NewQueueManager(db, tenantID)
	currentJob := qm.GetCurrentRunningJob()
	stats, _ := qm.GetJobStats()

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"currentJob": currentJob,
		"stats":      stats,
	})
}

// GetJobStatus handles GET /api/jobs/status/:jobId
func (h *JobQueueHandler) GetJobStatus(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	jobID := c.Param("jobId")
	qm := jobs.NewQueueManager(db, tenantID)

	currentJob := qm.GetCurrentRunningJob()
	history := qm.GetJobHistoryByID(jobID, 10)

	status := "unknown"
	if currentJob != nil && currentJob.ID == jobID {
		status = "running"
	} else if len(history) > 0 {
		status = string(history[0].Status)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"jobId":      jobID,
		"status":     status,
		"currentJob": currentJob,
		"history":    history,
	})
}
