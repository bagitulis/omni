package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/route"
	"gorm.io/gorm"
)

// RouteConfigHandler handles route config endpoints
type RouteConfigHandler struct {
	fallbackDB *gorm.DB
}

// NewRouteConfigHandler creates a new route config handler
func NewRouteConfigHandler(db *gorm.DB) *RouteConfigHandler {
	return &RouteConfigHandler{fallbackDB: db}
}

// getDB returns the appropriate database for the current request
func (h *RouteConfigHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// List handles GET /api/route-config
func (h *RouteConfigHandler) List(c *gin.Context) {
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

	svc := route.NewConfigService(db, tenantID)
	configs, err := svc.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "configs": configs})
}

// Get handles GET /api/route-config/:path
func (h *RouteConfigHandler) Get(c *gin.Context) {
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

	routePath := c.Param("path")
	if routePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "route path required"})
		return
	}

	svc := route.NewConfigService(db, tenantID)
	config, err := svc.Get(routePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "config": config})
}

// Update handles PUT /api/route-config/:path
func (h *RouteConfigHandler) Update(c *gin.Context) {
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

	routePath := c.Param("path")
	if routePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "route path required"})
		return
	}

	var req route.RouteConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := route.NewConfigService(db, tenantID)
	config, err := svc.Update(routePath, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "config": config})
}

// Reset handles POST /api/routes-config/reset
func (h *RouteConfigHandler) Reset(c *gin.Context) {
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

	routePath := c.Query("path")
	svc := route.NewConfigService(db, tenantID)

	if routePath != "" {
		err = svc.ResetToDefault(routePath)
	} else {
		err = svc.ResetAll()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "reset to defaults"})
}

// ListAll handles GET /api/routes-config/all
func (h *RouteConfigHandler) ListAll(c *gin.Context) {
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

	svc := route.NewConfigService(db, tenantID)
	routes, err := svc.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total":  len(routes),
			"routes": routes,
		},
	})
}
