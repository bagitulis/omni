package route

// CoverageRouteSummary represents a route item for connected/backend-only/unused buckets.
type CoverageRouteSummary struct {
	Endpoint  string `json:"endpoint"`
	Method    string `json:"method,omitempty"`
	IsDynamic bool   `json:"is_dynamic,omitempty"`
	Category  string `json:"category,omitempty"`
}

// CoverageFrontendOnlySummary represents a frontend route that has no backend match.
type CoverageFrontendOnlySummary struct {
	Endpoint   string   `json:"endpoint"`
	Method     string   `json:"method,omitempty"`
	Components []string `json:"components"`
	Status     string   `json:"status,omitempty"`
	Category   string   `json:"category,omitempty"`
}

// CoverageCategories groups route coverage buckets.
type CoverageCategories struct {
	Connected    []CoverageRouteSummary        `json:"connected"`
	FrontendOnly []CoverageFrontendOnlySummary `json:"frontend_only"`
	BackendOnly  []CoverageRouteSummary        `json:"backend_only"`
	Unused       []CoverageRouteSummary        `json:"unused"`
}

// CoverageComponentDetail represents frontend API adapter usage detail.
type CoverageComponentDetail struct {
	Path         string   `json:"path,omitempty"`
	RoutesCalled []string `json:"routes_called,omitempty"`
	Buttons      []string `json:"buttons,omitempty"`
}

// CoverageCategoryLabel provides UI label text per category.
type CoverageCategoryLabel struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// CoverageReport represents the full payload consumed by frontend route-mapping page.
type CoverageReport struct {
	TotalRoutes             int                                    `json:"total_routes"`
	TotalComponents         int                                    `json:"total_components"`
	TotalCategories         int                                    `json:"total_categories"`
	TotalDynamicRoutes      int                                    `json:"total_dynamic_routes"`
	TotalCalledRoutes       int                                    `json:"total_called_routes"`
	TotalDisconnectedRoutes int                                    `json:"total_disconnected_routes"`
	TotalUnusedRoutes       int                                    `json:"total_unused_routes"`
	ConnectionRate          string                                 `json:"connection_rate"`
	ByCategory              map[string][]CoverageRouteSummary      `json:"by_category"`
	Categories              CoverageCategories                     `json:"categories"`
	CategoryLabels          map[string]CoverageCategoryLabel       `json:"category_labels"`
	CategoryStats           map[string]int                         `json:"category_stats"`
	Components              map[string]CoverageComponentDetail     `json:"components"`
	DisconnectedRoutes      map[string]CoverageFrontendOnlySummary `json:"disconnected_routes"`
	BackendOnlyRoutes       map[string]CoverageRouteSummary        `json:"backend_only_routes"`
	UnusedRoutes            map[string]CoverageRouteSummary        `json:"unused_routes"`
	ButtonToEndpoints       map[string]map[string][]string         `json:"button_to_endpoints"`
	Statistics              map[string]interface{}                 `json:"statistics"`
	Timestamp               string                                 `json:"timestamp"`
}

type frontendAggregate struct {
	Method     string
	Endpoint   string
	Category   string
	Components map[string]struct{}
}

type componentAggregate struct {
	Path         string
	RoutesCalled map[string]struct{}
}
