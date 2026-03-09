package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RouteExecutionConfigHandler handles route execution config endpoints
type RouteExecutionConfigHandler struct {
	fallbackDB *gorm.DB
}

// NewRouteExecutionConfigHandler creates a new handler
func NewRouteExecutionConfigHandler(db *gorm.DB) *RouteExecutionConfigHandler {
	return &RouteExecutionConfigHandler{fallbackDB: db}
}

// getDB returns the appropriate database for the current request
func (h *RouteExecutionConfigHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// List handles GET /api/route-execution-config
func (h *RouteExecutionConfigHandler) List(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		respondUnauthorized(c, "Missing tenantId")
		return
	}
	db, err := h.getDB(c)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	ctx := c.Request.Context()
	var configs []RouteExecutionConfig
	if err := db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("category, route_name").Find(&configs).Error; err != nil {
		respondWithConfigs(c, []RouteExecutionConfig{})
		return
	}
	respondWithConfigs(c, configs)
}

// Get handles GET /api/route-execution-config/:routeKey
func (h *RouteExecutionConfigHandler) Get(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		respondUnauthorized(c, "Missing tenantId")
		return
	}
	db, err := h.getDB(c)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	routeKey := c.Param("routeKey")
	if routeKey == "" {
		respondBadRequest(c, "routeKey required")
		return
	}
	ctx := c.Request.Context()
	var config RouteExecutionConfig
	if err := db.WithContext(ctx).Where("tenant_id = ? AND route_key = ?", tenantID, routeKey).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondNotFound(c, "config not found")
			return
		}
		respondInternalError(c, err)
		return
	}
	respondWithConfig(c, config)
}

// GetMode handles GET /api/route-execution-config/:routeKey/mode
func (h *RouteExecutionConfigHandler) GetMode(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		respondUnauthorized(c, "Missing tenantId")
		return
	}
	db, err := h.getDB(c)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	routeKey := c.Param("routeKey")
	if routeKey == "" {
		respondBadRequest(c, "routeKey required")
		return
	}

	ctx := c.Request.Context()
	var config RouteExecutionConfig
	if err := db.WithContext(ctx).Where("tenant_id = ? AND route_key = ?", tenantID, routeKey).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Default to direct mode if not found (intentional)
			c.JSON(http.StatusOK, gin.H{
				"success":        true,
				"execution_mode": "direct",
				"enabled":        true,
			})
			return
		}
		// Propagate actual DB errors instead of swallowing them (#17)
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"execution_mode": config.ExecutionMode,
		"enabled":        config.Enabled,
	})
}

// Create handles POST /api/route-execution-config
func (h *RouteExecutionConfigHandler) Create(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		respondUnauthorized(c, "Missing tenantId")
		return
	}
	db, err := h.getDB(c)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	var req RouteExecutionConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	if req.ExecutionMode != "queue" && req.ExecutionMode != "direct" {
		respondBadRequest(c, "executionMode must be 'queue' or 'direct'")
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	priority := "normal"
	if req.Priority != "" {
		priority = req.Priority
	}
	ctx := c.Request.Context()
	config := RouteExecutionConfig{
		TenantID: tenantID, RouteKey: req.RouteKey, RouteName: req.RouteName, Description: req.Description,
		ExecutionMode: req.ExecutionMode, Priority: priority, Enabled: enabled,
		Icon: req.Icon, Category: req.Category,
	}
	if err := db.WithContext(ctx).Create(&config).Error; err != nil {
		respondInternalError(c, err)
		return
	}
	respondCreated(c, config)
}

// Update handles PUT /api/route-execution-config/:routeKey
func (h *RouteExecutionConfigHandler) Update(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		respondUnauthorized(c, "Missing tenantId")
		return
	}
	db, err := h.getDB(c)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	routeKey := c.Param("routeKey")
	if routeKey == "" {
		respondBadRequest(c, "routeKey required")
		return
	}
	var req RouteExecutionConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	ctx := c.Request.Context()
	var config RouteExecutionConfig
	if err := db.WithContext(ctx).Where("tenant_id = ? AND route_key = ?", tenantID, routeKey).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondNotFound(c, "config not found")
			return
		}
		respondInternalError(c, err)
		return
	}
	// Update fields — only non-zero values
	if req.RouteName != "" {
		config.RouteName = req.RouteName
	}
	if req.Description != "" {
		config.Description = req.Description
	}
	if req.ExecutionMode != "" {
		if req.ExecutionMode != "queue" && req.ExecutionMode != "direct" {
			respondBadRequest(c, "executionMode must be 'queue' or 'direct'")
			return
		}
		config.ExecutionMode = req.ExecutionMode
	}
	if req.Priority != "" {
		config.Priority = req.Priority
	}
	if req.Enabled != nil {
		config.Enabled = *req.Enabled
	}
	if req.Icon != "" {
		config.Icon = req.Icon
	}
	if req.Category != "" {
		config.Category = req.Category
	}
	if err := db.WithContext(ctx).Save(&config).Error; err != nil {
		respondInternalError(c, err)
		return
	}
	respondWithConfig(c, config)
}

// Toggle handles POST /api/route-execution-config/:routeKey/toggle
func (h *RouteExecutionConfigHandler) Toggle(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		respondUnauthorized(c, "Missing tenantId")
		return
	}
	db, err := h.getDB(c)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	routeKey := c.Param("routeKey")
	if routeKey == "" {
		respondBadRequest(c, "routeKey required")
		return
	}
	ctx := c.Request.Context()
	var config RouteExecutionConfig
	if err := db.WithContext(ctx).Where("tenant_id = ? AND route_key = ?", tenantID, routeKey).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondNotFound(c, "config not found")
			return
		}
		respondInternalError(c, err)
		return
	}
	// Toggle execution mode
	if config.ExecutionMode == "queue" {
		config.ExecutionMode = "direct"
	} else {
		config.ExecutionMode = "queue"
	}
	if err := db.WithContext(ctx).Save(&config).Error; err != nil {
		respondInternalError(c, err)
		return
	}
	respondWithConfig(c, config)
}

// Delete handles DELETE /api/route-execution-config/:routeKey
func (h *RouteExecutionConfigHandler) Delete(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		respondUnauthorized(c, "Missing tenantId")
		return
	}
	db, err := h.getDB(c)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	routeKey := c.Param("routeKey")
	if routeKey == "" {
		respondBadRequest(c, "routeKey required")
		return
	}
	ctx := c.Request.Context()
	result := db.WithContext(ctx).Where("tenant_id = ? AND route_key = ?", tenantID, routeKey).Delete(&RouteExecutionConfig{})
	if result.Error != nil {
		respondInternalError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		respondNotFound(c, "config not found")
		return
	}
	respondDeleted(c, "config deleted")
}
