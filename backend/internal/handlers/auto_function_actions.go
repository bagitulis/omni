package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/autofunction"
	"github.com/omni/backend/internal/services/jobs"
	"gorm.io/gorm"
)

// =============================================================================
// Auto Function Action Handlers (Enable/Disable/Run/Cancel)
// =============================================================================

// Enable handles POST /api/jobs/auto-functions/:name/enable
func (h *AutoFunctionHandler) Enable(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	name := c.Param("name")
	cm := autofunction.NewConfigManager(db, tenantID)
	if err := cm.Enable(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if h.scheduler != nil {
		h.scheduler.EnableFunction(tenantID, name)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"enabled": true}})
}

// Disable handles POST /api/jobs/auto-functions/:name/disable
func (h *AutoFunctionHandler) Disable(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	name := c.Param("name")
	cm := autofunction.NewConfigManager(db, tenantID)
	if err := cm.Disable(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if h.scheduler != nil {
		h.scheduler.DisableFunction(tenantID, name)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"disabled": true}})
}

// CancelScheduledByName handles POST /api/jobs/auto-functions/:name/cancel-scheduled
func (h *AutoFunctionHandler) CancelScheduledByName(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	name := c.Param("name")
	cm := autofunction.NewConfigManager(db, tenantID)

	// Get config by name first
	cfg, err := cm.GetByName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Auto-function not found: " + name})
		return
	}

	if err := cm.CancelScheduled(cfg.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if h.scheduler != nil {
		h.scheduler.CancelScheduled(cfg.ID)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "scheduled execution cancelled"})
}

// Run handles POST /api/jobs/auto-functions/:name/run - manual execution
// Enqueues into the job queue for visibility in Current Job / Queue tabs,
// then executes the auto function in the background with full lifecycle tracking.
func (h *AutoFunctionHandler) Run(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	if h.executor == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "executor not configured"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to get tenant DB: " + err.Error()})
		return
	}

	name := c.Param("name")

	// Verify config exists
	cm := autofunction.NewConfigManager(db, tenantID)
	if _, err := cm.GetByName(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "auto-function not found: " + name})
		return
	}

	// Enqueue as a job so it's visible in Current Job / Queue tabs
	qm := jobs.NewQueueManager(db, tenantID)
	jobID, err := qm.EnqueueJob("auto_function", map[string]interface{}{
		"function_name": name,
		"trigger":       "manual",
	}, "high")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to enqueue job: " + err.Error()})
		return
	}

	// Execute in background with lifecycle tracking
	go h.executeWithTracking(db, tenantID, name, jobID)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "execution started", "job_id": jobID})
}

// GetHistory handles GET /api/jobs/auto-functions/history
func (h *AutoFunctionHandler) GetHistory(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	if h.executor == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "executor not configured"})
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	history, err := h.executor.GetAllHistory(tenantID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"history": history}})
}

// executeWithTracking runs an auto function with full job queue lifecycle tracking.
// Updates the job status through pending → running → completed/failed and records history.
func (h *AutoFunctionHandler) executeWithTracking(db *gorm.DB, tenantID, name, jobID string) {
	startTime := time.Now()
	qm := jobs.NewQueueManager(db, tenantID)

	// Mark job as running
	if err := qm.UpdateStatus(jobID, models.JobStatusRunning, ""); err != nil {
		log.Printf("[AutoFunction] Failed to update job status to running: %v", err)
	}

	// Get handler
	handler := h.executor.GetHandler(name)
	if handler == nil {
		errMsg := "no handler registered: " + name
		log.Printf("[AutoFunction] %s", errMsg)
		_ = qm.FailJob(jobID, errMsg)
		return
	}

	// Get config
	var cfg models.AutoFunctionConfig
	if err := db.Where("name = ?", name).First(&cfg).Error; err != nil {
		errMsg := "config not found: " + err.Error()
		log.Printf("[AutoFunction] %s", errMsg)
		_ = qm.FailJob(jobID, errMsg)
		return
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	log.Printf("[AutoFunction] Running '%s' for tenant: %s (job: %s)", name, tenantID, jobID)
	result, err := handler(ctx, tenantID, &cfg)

	// Update config timing (next scheduled execution)
	now := time.Now()
	nextExec := now.Add(time.Duration(cfg.IntervalMinutes) * time.Minute)
	db.Model(&cfg).Updates(map[string]interface{}{
		"last_executed":            now,
		"next_scheduled_execution": nextExec,
	})

	// Record auto_functions_history
	duration := int(time.Since(startTime).Milliseconds())
	status := "success"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
		log.Printf("[AutoFunction] %s failed: %v", name, err)
	} else {
		log.Printf("[AutoFunction] %s completed: %s", name, result)
	}

	history := &models.AutoFunctionHistory{
		FunctionName: name,
		Status:       status,
		ErrorMessage: errMsg,
		DurationMs:   duration,
		ExecutedAt:   startTime,
	}
	if histErr := db.Create(history).Error; histErr != nil {
		log.Printf("[AutoFunction] Failed to record history: %v", histErr)
	}

	// Complete or fail the job queue entry
	if err != nil {
		_ = qm.FailJob(jobID, errMsg)
	} else {
		_ = qm.CompleteJobWithResult(jobID, map[string]string{"result": result})
	}
}
