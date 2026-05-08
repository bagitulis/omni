package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/autofunction"
	"gorm.io/gorm"
)

// AutoFunctionHandler handles auto function endpoints
// Matches Node.js API format for frontend compatibility
type AutoFunctionHandler struct {
	fallbackDB *gorm.DB
	scheduler  *autofunction.Scheduler
	executor   *autofunction.Executor
}

// NewAutoFunctionHandler creates a new auto function handler
func NewAutoFunctionHandler(db *gorm.DB, scheduler *autofunction.Scheduler, executor *autofunction.Executor) *AutoFunctionHandler {
	return &AutoFunctionHandler{fallbackDB: db, scheduler: scheduler, executor: executor}
}

// getDB returns the appropriate database for the current request
func (h *AutoFunctionHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// List handles GET /api/jobs/auto-functions
// Returns all auto-function configs for the tenant
func (h *AutoFunctionHandler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	cm := autofunction.NewConfigManager(db, tenantID)
	configs, err := cm.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"configs": configs,
			"total":   len(configs),
		},
	})
}

// GetByName handles GET /api/jobs/auto-functions/:name
func (h *AutoFunctionHandler) GetByName(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	name := c.Param("name")
	cm := autofunction.NewConfigManager(db, tenantID)
	cfg, err := cm.GetByName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Auto-function not found: " + name})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": cfg})
}

// Create handles POST /api/jobs/auto-functions
// Uses CreateOrUpdate to handle both creation and updates (upsert behavior)
func (h *AutoFunctionHandler) Create(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var req models.AutoFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "name is required"})
		return
	}

	if req.IntervalMinutes < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "interval_minutes must be at least 1"})
		return
	}

	// Validate that the function name has a registered handler
	if h.executor != nil && h.executor.GetHandler(req.Name) == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "unknown function name: " + req.Name + ". Use GET /api/jobs/auto-functions/available for valid names"})
		return
	}

	// Validate time window: start_time must be before end_time (no overnight windows)
	if req.StartTime != nil && req.EndTime != nil && *req.StartTime != "" && *req.EndTime != "" {
		if *req.StartTime > *req.EndTime {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "start_time must be before end_time (overnight windows not supported)"})
			return
		}
	}

	cm := autofunction.NewConfigManager(db, tenantID)
	// Use CreateOrUpdate to handle duplicates gracefully (upsert)
	cfg, err := cm.CreateOrUpdate(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Add to scheduler if enabled
	if cfg.Enabled && h.scheduler != nil {
		h.scheduler.AddOrUpdate(cfg)
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": cfg})
}

// Update handles PUT /api/jobs/auto-functions/:name
func (h *AutoFunctionHandler) Update(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	name := c.Param("name")
	var req models.AutoFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.IntervalMinutes < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "interval_minutes must be at least 1"})
		return
	}

	// Validate time window if provided
	if (req.StartTime != nil && req.EndTime == nil) || (req.StartTime == nil && req.EndTime != nil) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Both start_time and end_time must be provided together, or neither"})
		return
	}
	// Validate time window: start_time must be before end_time
	if req.StartTime != nil && req.EndTime != nil && *req.StartTime != "" && *req.EndTime != "" {
		if *req.StartTime > *req.EndTime {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "start_time must be before end_time (overnight windows not supported)"})
			return
		}
	}

	cm := autofunction.NewConfigManager(db, tenantID)
	req.Name = name // Use path param name
	cfg, err := cm.CreateOrUpdate(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Update scheduler
	if h.scheduler != nil {
		h.scheduler.AddOrUpdate(cfg)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": cfg})
}

// Delete handles DELETE /api/jobs/auto-functions/:name
func (h *AutoFunctionHandler) Delete(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	name := c.Param("name")
	cm := autofunction.NewConfigManager(db, tenantID)

	// Get config first to get ID for scheduler removal
	cfg, err := cm.GetByName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Auto-function not found: " + name})
		return
	}

	if err := cm.DeleteByName(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Remove from scheduler
	if h.scheduler != nil {
		h.scheduler.Remove(cfg.ID)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"deleted": true}})
}

// AvailableAutoFunction describes a registered auto-function for the dropdown.
type AvailableAutoFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// availableAutoFunctions is the registry of all known auto-functions.
var availableAutoFunctions = []AvailableAutoFunction{
	{Name: "locked_today", Description: "Lock orders at end of day"},
	{Name: "auto_update_token", Description: "Refresh platform OAuth tokens"},
	{Name: "sync_from_sheets", Description: "Sync inventory from Google Sheets"},
	{Name: "sync_products", Description: "Full product sync from all platforms (Shopee, TikTok, Lazada)"},
	{Name: "sync_products_inventory", Description: "Sync products matching inventory SKUs only"},
	{Name: "retry_failed_syncs", Description: "Retry failed marketplace sync operations from last 24h"},
	{Name: "price_drift_detection", Description: "Detect SKUs with price drift between inventory and platforms"},
}

// ListAvailable handles GET /api/jobs/auto-functions/available
// Returns all registered auto-function names with descriptions for the dropdown.
func (h *AutoFunctionHandler) ListAvailable(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    availableAutoFunctions,
	})
}

// NOTE: Enable, Disable, CancelScheduledByName, Run, GetHistory handlers
// are defined in auto_function_actions.go
