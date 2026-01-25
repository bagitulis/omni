package handlers

// =============================================================================
// Route Mapping DTOs (Data Transfer Objects)
// =============================================================================

// RouteInfo represents route information
type RouteInfo struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Handler     string   `json:"handler"`
	Middleware  []string `json:"middleware,omitempty"`
	Platform    string   `json:"platform,omitempty"`
	Category    string   `json:"category,omitempty"`
	Description string   `json:"description,omitempty"`
}

// HandlerInfo represents handler information
type HandlerInfo struct {
	Name    string `json:"name"`
	Package string `json:"package"`
	File    string `json:"file,omitempty"`
	Routes  int    `json:"routes"`
}

// ServiceInfo represents service information
type ServiceInfo struct {
	Name     string   `json:"name"`
	Package  string   `json:"package"`
	Methods  []string `json:"methods,omitempty"`
	Handlers []string `json:"handlers,omitempty"`
}

// MiddlewareInfo represents middleware information
type MiddlewareInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	AppliesTo   []string `json:"applies_to,omitempty"`
	Order       int      `json:"order"`
}

// UnusedRouteInfo represents unused route info
type UnusedRouteInfo struct {
	Path    string `json:"path"`
	Method  string `json:"method"`
	Handler string `json:"handler"`
	Reason  string `json:"reason"`
}

// DuplicateRouteInfo represents duplicate route info
type DuplicateRouteInfo struct {
	Path     string   `json:"path"`
	Methods  []string `json:"methods"`
	Handlers []string `json:"handlers"`
}

// AnalyzeRequest represents analyze request
type AnalyzeRequest struct {
	IncludeStats    bool `json:"include_stats"`
	IncludeMetrics  bool `json:"include_metrics"`
	IncludeCoverage bool `json:"include_coverage"`
}

// RouteAnalysis represents route analysis result
type RouteAnalysis struct {
	TotalRoutes         int                    `json:"total_routes"`
	ByMethod            map[string]int         `json:"by_method"`
	ByPlatform          map[string]int         `json:"by_platform"`
	ByCategory          map[string]int         `json:"by_category"`
	AuthenticatedRoutes int                    `json:"authenticated_routes"`
	PublicRoutes        int                    `json:"public_routes"`
	DeprecatedRoutes    int                    `json:"deprecated_routes"`
	Stats               map[string]interface{} `json:"stats,omitempty"`
}
