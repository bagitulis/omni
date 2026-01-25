package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/autofunction"
)

// =============================================================================
// Auto Function Action Handlers (Enable/Disable/Run/Cancel)
// =============================================================================

// Enable handles POST /api/jobs/auto-functions/:name/enable
func (h *AutoFunctionHandler) Enable(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant ID required"})
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
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant ID required"})
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
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant ID required"})
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
func (h *AutoFunctionHandler) Run(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant ID required"})
		return
	}

	if h.executor == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "executor not configured"})
		return
	}

	name := c.Param("name")
	if err := h.executor.ExecuteByName(tenantID, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "execution started"})
}

// GetHistory handles GET /api/jobs/auto-functions/history
func (h *AutoFunctionHandler) GetHistory(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant ID required"})
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
