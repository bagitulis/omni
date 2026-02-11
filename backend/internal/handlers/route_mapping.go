package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/route"
)

// RouteHandler handles route mapping endpoints
type RouteHandler struct {
	mappingService *route.MappingService
	scannerService *route.ScannerService
}

// NewRouteHandler creates a new route handler
func NewRouteHandler(basePath string) *RouteHandler {
	return &RouteHandler{
		mappingService: route.NewMappingService(),
		scannerService: route.NewScannerService(basePath),
	}
}

// GetAllRoutes returns all registered routes
// @Summary Get all route mappings
// @Tags Routes
// @Success 200 {object} map[string]interface{}
// @Router /api/routes [get]
func (h *RouteHandler) GetAllRoutes(c *gin.Context) {
	routes := h.mappingService.GetAllMappings()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    routes,
		"total":   len(routes),
	})
}

// GetRoutesByTag returns routes filtered by tag
// @Summary Get routes by tag
// @Tags Routes
// @Param tag path string true "Tag name"
// @Success 200 {object} map[string]interface{}
// @Router /api/routes/tag/{tag} [get]
func (h *RouteHandler) GetRoutesByTag(c *gin.Context) {
	tag := c.Param("tag")
	routes := h.mappingService.GetMappingsByTag(tag)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    routes,
		"tag":     tag,
		"total":   len(routes),
	})
}

// AnalyzeRoutes performs route analysis
// @Summary Analyze route mappings
// @Tags Routes
// @Success 200 {object} map[string]interface{}
// @Router /api/routes/analyze [get]
func (h *RouteHandler) AnalyzeRoutes(c *gin.Context) {
	analysis := h.mappingService.Analyze()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analysis,
	})
}

// ScanRoutes scans codebase for routes
// @Summary Scan codebase for route definitions
// @Tags Routes
// @Success 200 {object} map[string]interface{}
// @Router /api/routes/scan [get]
func (h *RouteHandler) ScanRoutes(c *gin.Context) {
	result := h.scannerService.ScanRoutes()

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"data":         result.Routes,
		"filesScanned": result.Files,
		"errors":       result.Errors,
	})
}

// ScanMiddleware scans for middleware usage
// @Summary Scan codebase for middleware usage
// @Tags Routes
// @Success 200 {object} map[string]interface{}
// @Router /api/routes/middleware [get]
func (h *RouteHandler) ScanMiddleware(c *gin.Context) {
	result := h.scannerService.ScanForMiddleware()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// RegisterRoute registers a route mapping (used internally)
func (h *RouteHandler) RegisterRoute(mapping route.RouteMapping) {
	h.mappingService.RegisterRoute(mapping)
}

// FormatAsTable returns routes as markdown table
// @Summary Get routes as markdown table
// @Tags Routes
// @Success 200 {string} string
// @Router /api/routes/table [get]
func (h *RouteHandler) FormatAsTable(c *gin.Context) {
	table := h.mappingService.FormatAsTable()
	c.String(http.StatusOK, table)
}
