package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/route"
)

// GetCategories handles GET /api/routes-config/categories
func (h *RouteConfigHandler) GetCategories(c *gin.Context) {
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
	categories, err := svc.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"categories": categories},
	})
}

// GetByCategory handles GET /api/routes-config/by-category/:category
func (h *RouteConfigHandler) GetByCategory(c *gin.Context) {
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

	category := c.Param("category")
	svc := route.NewConfigService(db, tenantID)
	routes, err := svc.GetByCategory(category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"category": category,
			"total":    len(routes),
			"routes":   routes,
		},
	})
}

// Create handles POST /api/routes-config
func (h *RouteConfigHandler) Create(c *gin.Context) {
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

	var req route.RouteConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.RoutePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "routePath is required"})
		return
	}

	svc := route.NewConfigService(db, tenantID)
	config, err := svc.Create(req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "route already exists" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": config})
}

// Delete handles DELETE /api/routes-config/:id
func (h *RouteConfigHandler) Delete(c *gin.Context) {
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

	id := c.Param("id")
	svc := route.NewConfigService(db, tenantID)
	config, err := svc.Delete(id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "route not found" {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": config})
}

// BulkUpdate handles POST /api/routes-config/bulk-update
func (h *RouteConfigHandler) BulkUpdate(c *gin.Context) {
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
		RouteIDs []string               `json:"routeIds"`
		Updates  map[string]interface{} `json:"updates"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := route.NewConfigService(db, tenantID)
	routes, err := svc.BulkUpdate(req.RouteIDs, req.Updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"updated": len(routes),
			"routes":  routes,
		},
	})
}

// ApplyPreset handles POST /api/routes-config/apply-preset/:preset
func (h *RouteConfigHandler) ApplyPreset(c *gin.Context) {
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

	preset := c.Param("preset")
	var req struct {
		RouteIDs []string `json:"routeIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := route.NewConfigService(db, tenantID)
	routes, err := svc.ApplyPreset(req.RouteIDs, preset)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "unknown preset" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"preset":  preset,
			"applied": len(routes),
			"routes":  routes,
		},
	})
}
