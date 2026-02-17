package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/route"
)

// RouteHandler handles route mapping endpoints.
type RouteHandler struct {
	mappingService *route.MappingService
	scannerService *route.ScannerService
	engine         *gin.Engine
}

// NewRouteHandler creates a new route handler.
func NewRouteHandler(basePath string) *RouteHandler {
	return &RouteHandler{
		mappingService: route.NewMappingService(),
		scannerService: route.NewScannerService(basePath),
	}
}

// SetEngine attaches gin engine for runtime route extraction.
func (h *RouteHandler) SetEngine(engine *gin.Engine) {
	h.engine = engine
}

// GetAllRoutes returns all runtime backend routes.
func (h *RouteHandler) GetAllRoutes(c *gin.Context) {
	routes := h.extractBackendRoutes()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    routes,
		"total":   len(routes),
	})
}

// GetRoutesByTag returns routes filtered by tag.
func (h *RouteHandler) GetRoutesByTag(c *gin.Context) {
	tag := strings.TrimSpace(c.Param("tag"))
	routes := h.extractBackendRoutes()
	filtered := make([]route.RouteMapping, 0)

	for _, routeItem := range routes {
		for _, routeTag := range routeItem.Tags {
			if strings.EqualFold(routeTag, tag) {
				filtered = append(filtered, routeItem)
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    filtered,
		"tag":     tag,
		"total":   len(filtered),
	})
}

// AnalyzeRoutes returns summarized route analysis.
func (h *RouteHandler) AnalyzeRoutes(c *gin.Context) {
	backendRoutes := h.extractBackendRoutes()
	legacyAnalysisService := route.NewMappingService()
	for _, routeItem := range backendRoutes {
		routeItem.Middlewares = []string{"Auth"}
		legacyAnalysisService.RegisterRoute(routeItem)
	}
	analysis := legacyAnalysisService.Analyze()
	if analysis.Conflicts == nil {
		analysis.Conflicts = []route.RouteConflict{}
	}
	if analysis.UnprotectedRoutes == nil {
		analysis.UnprotectedRoutes = []string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_routes":       analysis.TotalRoutes,
			"by_method":          analysis.ByMethod,
			"by_tag":             analysis.ByTag,
			"conflicts":          analysis.Conflicts,
			"unprotected_routes": analysis.UnprotectedRoutes,
		},
	})
}

// GetRouteCoverage returns FE↔BE mapping coverage payload.
func (h *RouteHandler) GetRouteCoverage(c *gin.Context) {
	backendRoutes := h.extractBackendRoutes()
	report := h.scannerService.BuildCoverageReport(backendRoutes)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    report,
	})
}

// GetRouteCoverageStatistics returns compact route-mapping statistics.
func (h *RouteHandler) GetRouteCoverageStatistics(c *gin.Context) {
	backendRoutes := h.extractBackendRoutes()
	report := h.scannerService.BuildCoverageReport(backendRoutes)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"statistics":          report.Statistics,
			"connection_rate":     report.ConnectionRate,
			"total_routes":        report.TotalRoutes,
			"total_called_routes": report.TotalCalledRoutes,
			"total_disconnected":  report.TotalDisconnectedRoutes,
			"total_backend_only":  len(report.Categories.BackendOnly),
			"total_unused":        report.TotalUnusedRoutes,
			"total_components":    report.TotalComponents,
			"last_calculated_at":  report.Timestamp,
		},
	})
}

// ScanRoutes scans codebase for route definitions.
func (h *RouteHandler) ScanRoutes(c *gin.Context) {
	result := h.scannerService.ScanRoutes()

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"data":          result.Routes,
		"files_scanned": result.Files,
		"errors":        result.Errors,
	})
}

// ScanMiddleware scans for middleware usage.
func (h *RouteHandler) ScanMiddleware(c *gin.Context) {
	result := h.scannerService.ScanForMiddleware()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// RegisterRoute registers a route mapping (used internally).
func (h *RouteHandler) RegisterRoute(mapping route.RouteMapping) {
	h.mappingService.RegisterRoute(mapping)
}

// FormatAsTable returns backend routes as markdown table.
func (h *RouteHandler) FormatAsTable(c *gin.Context) {
	routes := h.extractBackendRoutes()

	var builder strings.Builder
	builder.WriteString("| Method | Path | Handler | Tags |\n")
	builder.WriteString("|--------|------|---------|------|\n")

	for _, routeItem := range routes {
		tags := strings.Join(routeItem.Tags, ", ")
		builder.WriteString("| " + routeItem.Method + " | " + routeItem.Path + " | " + routeItem.Handler + " | " + tags + " |\n")
	}

	c.String(http.StatusOK, builder.String())
}

func (h *RouteHandler) extractBackendRoutes() []route.RouteMapping {
	if h.engine == nil {
		fallback := h.scannerService.ScanRoutes()
		routes := make([]route.RouteMapping, 0, len(fallback.Routes))
		for _, scanned := range fallback.Routes {
			endpoint := route.NormalizeRoutePath(scanned.Path, true)
			if endpoint == "" {
				continue
			}
			routes = append(routes, route.RouteMapping{
				Method:  strings.ToUpper(scanned.Method),
				Path:    endpoint,
				Handler: scanned.Handler,
				Tags:    []string{deriveTagFromPath(endpoint)},
			})
		}
		sortRouteMappings(routes)
		return routes
	}

	ginRoutes := h.engine.Routes()
	routes := make([]route.RouteMapping, 0, len(ginRoutes))
	seen := make(map[string]struct{}, len(ginRoutes))

	for _, ginRoute := range ginRoutes {
		endpoint := route.NormalizeRoutePath(ginRoute.Path, false)
		if endpoint == "" {
			continue
		}

		method := strings.ToUpper(ginRoute.Method)
		key := method + " " + endpoint
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		routes = append(routes, route.RouteMapping{
			Method:  method,
			Path:    endpoint,
			Handler: ginRoute.Handler,
			Tags:    []string{deriveTagFromPath(endpoint)},
		})
	}

	sortRouteMappings(routes)
	return routes
}

func sortRouteMappings(routes []route.RouteMapping) {
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})
}

func deriveTagFromPath(path string) string {
	trimmed := strings.TrimPrefix(path, "/")
	trimmed = strings.TrimPrefix(trimmed, "api/")
	if trimmed == "" {
		return "general"
	}

	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "general"
	}

	return parts[0]
}
