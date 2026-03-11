package handlers

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
)

// RouteMappingHandler handles route mapping analysis endpoints
type RouteMappingHandler struct {
	engine     *gin.Engine
	routesOnce sync.Once
	routeCache []RouteInfo
}

// NewRouteMappingHandler creates a new route mapping handler
func NewRouteMappingHandler(engine *gin.Engine) *RouteMappingHandler {
	return &RouteMappingHandler{engine: engine}
}

// GetRoutes handles GET /api/route-mapping/routes
func (h *RouteMappingHandler) GetRoutes(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	routes := h.extractRoutes()
	c.JSON(http.StatusOK, response.Success(gin.H{
		"total":  len(routes),
		"routes": routes,
	}))
}

// GetRoutesByPlatform handles GET /api/route-mapping/routes/:platform
func (h *RouteMappingHandler) GetRoutesByPlatform(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	platform := c.Param("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing platform"))
		return
	}

	routes := h.extractRoutes()
	filtered := make([]RouteInfo, 0)

	for _, route := range routes {
		if strings.Contains(strings.ToLower(route.Path), strings.ToLower(platform)) {
			route.Platform = platform
			filtered = append(filtered, route)
		}
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"platform": platform,
		"total":    len(filtered),
		"routes":   filtered,
	}))
}

// GetHandlers handles GET /api/route-mapping/handlers
func (h *RouteMappingHandler) GetHandlers(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	routes := h.extractRoutes()
	handlerMap := make(map[string]int)

	for _, route := range routes {
		handlerMap[route.Handler]++
	}

	handlers := make([]HandlerInfo, 0, len(handlerMap))
	for name, count := range handlerMap {
		parts := strings.Split(name, ".")
		pkg := ""
		if len(parts) > 1 {
			pkg = strings.Join(parts[:len(parts)-1], ".")
		}
		handlers = append(handlers, HandlerInfo{
			Name:    name,
			Package: pkg,
			Routes:  count,
		})
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total":    len(handlers),
		"handlers": handlers,
	}))
}

// GetServices handles GET /api/route-mapping/services
func (h *RouteMappingHandler) GetServices(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	c.JSON(http.StatusNotImplemented, response.Error("Service introspection not yet implemented"))
}

// GetMiddleware handles GET /api/route-mapping/middleware
func (h *RouteMappingHandler) GetMiddleware(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	c.JSON(http.StatusNotImplemented, response.Error("Middleware introspection not yet implemented"))
}

// GetUnused handles GET /api/route-mapping/unused
func (h *RouteMappingHandler) GetUnused(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	c.JSON(http.StatusNotImplemented, response.Error("Unused route detection requires access-log analysis and is not yet implemented"))
}

// GetDuplicates handles GET /api/route-mapping/duplicates
func (h *RouteMappingHandler) GetDuplicates(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	routes := h.extractRoutes()
	pathMap := make(map[string][]RouteInfo)

	for _, route := range routes {
		key := route.Path
		pathMap[key] = append(pathMap[key], route)
	}

	duplicates := make([]DuplicateRouteInfo, 0)
	for path, rs := range pathMap {
		if len(rs) > 1 {
			methods := make([]string, 0, len(rs))
			handlers := make([]string, 0, len(rs))
			for _, r := range rs {
				methods = append(methods, r.Method)
				handlers = append(handlers, r.Handler)
			}
			duplicates = append(duplicates, DuplicateRouteInfo{
				Path:     path,
				Methods:  methods,
				Handlers: handlers,
			})
		}
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total":      len(duplicates),
		"duplicates": duplicates,
	}))
}

// AnalyzeRoutes handles POST /api/route-mapping/analyze
func (h *RouteMappingHandler) AnalyzeRoutes(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request body"))
		return
	}

	routes := h.extractRoutes()

	analysis := RouteAnalysis{
		TotalRoutes: len(routes),
		ByMethod:    make(map[string]int),
		ByPlatform:  make(map[string]int),
		ByCategory:  make(map[string]int),
	}

	publicPaths := []string{"/api/auth/login", "/api/auth/register", "/api/health", "/api/webhooks/"}

	for _, route := range routes {
		analysis.ByMethod[route.Method]++

		// Detect platform
		if strings.Contains(route.Path, "/shopee/") {
			analysis.ByPlatform["shopee"]++
		} else if strings.Contains(route.Path, "/lazada/") {
			analysis.ByPlatform["lazada"]++
		} else if strings.Contains(route.Path, "/tiktok/") {
			analysis.ByPlatform["tiktok"]++
		} else {
			analysis.ByPlatform["general"]++
		}

		// Check if public
		isPublic := false
		for _, pp := range publicPaths {
			if strings.HasPrefix(route.Path, pp) {
				isPublic = true
				break
			}
		}
		if isPublic {
			analysis.PublicRoutes++
		} else {
			analysis.AuthenticatedRoutes++
		}
	}

	if req.IncludeStats {
		avgRoutesPerHandler := 0.0
		if len(analysis.ByMethod) > 0 {
			avgRoutesPerHandler = float64(len(routes)) / float64(len(analysis.ByMethod))
		}
		analysis.Stats = map[string]interface{}{
			"avgRoutesPerHandler": avgRoutesPerHandler,
		}
	}

	c.JSON(http.StatusOK, response.Success(analysis))
}

// extractRoutes extracts route information from gin engine.
// Uses sync.Once to cache the result — routes are registered at startup and never change.
func (h *RouteMappingHandler) extractRoutes() []RouteInfo {
	h.routesOnce.Do(func() {
		if h.engine == nil {
			h.routeCache = []RouteInfo{}
			return
		}

		ginRoutes := h.engine.Routes()
		routes := make([]RouteInfo, 0, len(ginRoutes))

		for _, r := range ginRoutes {
			handlerName := runtime.FuncForPC(reflect.ValueOf(r.Handler).Pointer()).Name()
			routes = append(routes, RouteInfo{
				Method:  r.Method,
				Path:    r.Path,
				Handler: handlerName,
			})
		}

		h.routeCache = routes
	})

	return h.routeCache
}
