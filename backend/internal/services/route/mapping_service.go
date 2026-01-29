package route

import (
	"regexp"
	"sort"
	"strings"
	"sync"
)

// RouteMapping represents a single route mapping
type RouteMapping struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Handler     string   `json:"handler"`
	Middlewares []string `json:"middlewares,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// RouteAnalysis represents analysis result of routes
type RouteAnalysis struct {
	TotalRoutes       int             `json:"total_routes"`
	ByMethod          map[string]int  `json:"by_method"`
	ByTag             map[string]int  `json:"by_tag"`
	Conflicts         []RouteConflict `json:"conflicts,omitempty"`
	UnprotectedRoutes []string        `json:"unprotected_routes,omitempty"`
}

// RouteConflict represents a potential route conflict
type RouteConflict struct {
	Route1 string `json:"route1"`
	Route2 string `json:"route2"`
	Reason string `json:"reason"`
}

// MappingService provides route mapping functionality
type MappingService struct {
	routes []RouteMapping
	mu     sync.RWMutex
}

// NewMappingService creates a new mapping service
func NewMappingService() *MappingService {
	return &MappingService{
		routes: make([]RouteMapping, 0),
	}
}

// RegisterRoute registers a new route mapping
func (s *MappingService) RegisterRoute(mapping RouteMapping) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routes = append(s.routes, mapping)
}

// GetAllMappings returns all registered route mappings
func (s *MappingService) GetAllMappings() []RouteMapping {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]RouteMapping, len(s.routes))
	copy(result, s.routes)
	return result
}

// GetMappingsByTag returns routes filtered by tag
func (s *MappingService) GetMappingsByTag(tag string) []RouteMapping {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []RouteMapping
	for _, r := range s.routes {
		for _, t := range r.Tags {
			if strings.EqualFold(t, tag) {
				result = append(result, r)
				break
			}
		}
	}
	return result
}

// Analyze performs analysis on registered routes
func (s *MappingService) Analyze() RouteAnalysis {
	s.mu.RLock()
	defer s.mu.RUnlock()

	analysis := RouteAnalysis{
		TotalRoutes: len(s.routes),
		ByMethod:    make(map[string]int),
		ByTag:       make(map[string]int),
	}

	// Count by method and tag
	for _, r := range s.routes {
		analysis.ByMethod[r.Method]++
		for _, t := range r.Tags {
			analysis.ByTag[t]++
		}

		// Check for unprotected routes (no auth middleware)
		if !contains(r.Middlewares, "Auth") && !isPublicRoute(r.Path) {
			analysis.UnprotectedRoutes = append(analysis.UnprotectedRoutes, r.Method+" "+r.Path)
		}
	}

	// Detect conflicts
	analysis.Conflicts = s.detectConflicts()

	return analysis
}

// detectConflicts finds potential route conflicts
func (s *MappingService) detectConflicts() []RouteConflict {
	var conflicts []RouteConflict

	for i := 0; i < len(s.routes); i++ {
		for j := i + 1; j < len(s.routes); j++ {
			r1, r2 := s.routes[i], s.routes[j]

			if r1.Method != r2.Method {
				continue
			}

			if couldConflict(r1.Path, r2.Path) {
				conflicts = append(conflicts, RouteConflict{
					Route1: r1.Method + " " + r1.Path,
					Route2: r2.Method + " " + r2.Path,
					Reason: "Potential path parameter conflict",
				})
			}
		}
	}

	return conflicts
}

// FormatAsTable formats routes as a table string
func (s *MappingService) FormatAsTable() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Sort routes by path
	sorted := make([]RouteMapping, len(s.routes))
	copy(sorted, s.routes)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})

	var sb strings.Builder
	sb.WriteString("| Method | Path | Handler | Tags |\n")
	sb.WriteString("|--------|------|---------|------|\n")

	for _, r := range sorted {
		tags := strings.Join(r.Tags, ", ")
		sb.WriteString("| " + r.Method + " | " + r.Path + " | " + r.Handler + " | " + tags + " |\n")
	}

	return sb.String()
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func isPublicRoute(path string) bool {
	publicPaths := []string{"/health", "/csrf-token", "/auth/login", "/auth/register", "/webhooks/", "/docs"}
	for _, p := range publicPaths {
		if strings.Contains(path, p) {
			return true
		}
	}
	return false
}

func couldConflict(path1, path2 string) bool {
	// Check if paths could match same request
	// e.g., /users/:id and /users/me
	paramPattern := regexp.MustCompile(`:[^/]+`)

	normalized1 := paramPattern.ReplaceAllString(path1, ":param")
	normalized2 := paramPattern.ReplaceAllString(path2, ":param")

	if normalized1 == normalized2 && path1 != path2 {
		return true
	}
	return false
}
