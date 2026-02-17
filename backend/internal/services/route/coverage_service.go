package route

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// BuildCoverageReport builds FE↔BE route coverage from backend routes and scanned frontend API calls.
func (s *ScannerService) BuildCoverageReport(backendRoutes []RouteMapping) CoverageReport {
	frontendScan := s.ScanFrontendRoutes()

	backendIndex := make(map[string]RouteMapping)
	backendByEndpoint := make(map[string]map[string]struct{})
	matchedBackend := make(map[string]struct{})
	backendMethodCount := make(map[string]int)
	dynamicCount := 0

	for _, backendRoute := range backendRoutes {
		endpoint := NormalizeRoutePath(backendRoute.Path, false)
		if endpoint == "" {
			continue
		}

		method := strings.ToUpper(strings.TrimSpace(backendRoute.Method))
		if method == "" {
			continue
		}

		key := routeKey(method, endpoint)
		if _, exists := backendIndex[key]; exists {
			continue
		}

		normalizedRoute := backendRoute
		normalizedRoute.Method = method
		normalizedRoute.Path = endpoint
		if len(normalizedRoute.Tags) == 0 {
			normalizedRoute.Tags = []string{deriveCategory(endpoint)}
		}

		backendIndex[key] = normalizedRoute
		backendMethodCount[method]++

		if strings.Contains(endpoint, ":") {
			dynamicCount++
		}

		if _, exists := backendByEndpoint[endpoint]; !exists {
			backendByEndpoint[endpoint] = make(map[string]struct{})
		}
		backendByEndpoint[endpoint][method] = struct{}{}
	}

	frontendIndex := make(map[string]*frontendAggregate)
	componentsAgg := make(map[string]*componentAggregate)
	methodMismatchCount := 0

	for _, call := range frontendScan.Routes {
		endpoint := NormalizeRoutePath(call.Endpoint, true)
		if endpoint == "" {
			continue
		}

		method := strings.ToUpper(strings.TrimSpace(call.Method))
		if method == "" {
			continue
		}

		componentName := sourceToComponentName(call.Source)
		componentPath := filepath.ToSlash(filepath.Join("src", "api", call.Source))

		if _, exists := componentsAgg[componentName]; !exists {
			componentsAgg[componentName] = &componentAggregate{
				Path:         componentPath,
				RoutesCalled: make(map[string]struct{}),
			}
		}

		componentsAgg[componentName].RoutesCalled[fmt.Sprintf("%s %s", method, endpoint)] = struct{}{}

		key := routeKey(method, endpoint)
		if _, exists := frontendIndex[key]; !exists {
			frontendIndex[key] = &frontendAggregate{
				Method:     method,
				Endpoint:   endpoint,
				Category:   deriveCategory(endpoint),
				Components: make(map[string]struct{}),
			}
		}

		frontendIndex[key].Components[componentName] = struct{}{}
	}

	connected := make([]CoverageRouteSummary, 0)
	frontendOnly := make([]CoverageFrontendOnlySummary, 0)

	for key, frontRoute := range frontendIndex {
		if backendRoute, exists := backendIndex[key]; exists {
			matchedBackend[key] = struct{}{}
			connected = append(connected, toCoverageSummary(backendRoute))
			continue
		}

		status := "missing_backend_route"
		if methods, exists := backendByEndpoint[frontRoute.Endpoint]; exists && len(methods) > 0 {
			status = "method_mismatch"
			methodMismatchCount++
		}

		frontendOnly = append(frontendOnly, CoverageFrontendOnlySummary{
			Endpoint:   frontRoute.Endpoint,
			Method:     frontRoute.Method,
			Components: sortedMapKeys(frontRoute.Components),
			Status:     status,
			Category:   frontRoute.Category,
		})
	}

	backendOnly := make([]CoverageRouteSummary, 0)
	unused := make([]CoverageRouteSummary, 0)

	for key, backendRoute := range backendIndex {
		if _, matched := matchedBackend[key]; matched {
			continue
		}

		summary := toCoverageSummary(backendRoute)
		backendOnly = append(backendOnly, summary)
		if !isSystemOrExternalRoute(summary.Endpoint) {
			unused = append(unused, summary)
		}
	}

	sortCoverageSummary(connected)
	sortFrontendOnly(frontendOnly)
	sortCoverageSummary(backendOnly)
	sortCoverageSummary(unused)

	components := make(map[string]CoverageComponentDetail, len(componentsAgg))
	for name, detail := range componentsAgg {
		components[name] = CoverageComponentDetail{
			Path:         detail.Path,
			RoutesCalled: sortedMapKeys(detail.RoutesCalled),
			Buttons:      []string{},
		}
	}

	disconnectedMap := make(map[string]CoverageFrontendOnlySummary, len(frontendOnly))
	for _, route := range frontendOnly {
		disconnectedMap[route.Method+" "+route.Endpoint] = route
	}

	backendOnlyMap := make(map[string]CoverageRouteSummary, len(backendOnly))
	for _, route := range backendOnly {
		backendOnlyMap[route.Method+" "+route.Endpoint] = route
	}

	unusedMap := make(map[string]CoverageRouteSummary, len(unused))
	for _, route := range unused {
		unusedMap[route.Method+" "+route.Endpoint] = route
	}

	calledRoutes := len(frontendIndex)
	connectionRate := formatPercentage(len(connected), calledRoutes)

	categories := CoverageCategories{
		Connected:    connected,
		FrontendOnly: frontendOnly,
		BackendOnly:  backendOnly,
		Unused:       unused,
	}

	byCategory := map[string][]CoverageRouteSummary{
		"connected":     connected,
		"frontend_only": frontendOnlyToSummary(frontendOnly),
		"backend_only":  backendOnly,
		"unused":        unused,
	}

	categoryStats := map[string]int{
		"connected":     len(connected),
		"frontend_only": len(frontendOnly),
		"backend_only":  len(backendOnly),
		"unused":        len(unused),
	}

	categoryLabels := map[string]CoverageCategoryLabel{
		"connected": {
			Title:  "Connected",
			Detail: "Routes called by frontend and implemented by backend",
		},
		"frontend_only": {
			Title:  "Frontend Only",
			Detail: "Routes called by frontend but not implemented in backend",
		},
		"backend_only": {
			Title:  "Backend Only",
			Detail: "Routes available in backend but not called by frontend",
		},
		"unused": {
			Title:  "Unused",
			Detail: "Backend-only routes excluding known system or external endpoints",
		},
	}

	statistics := map[string]interface{}{
		"backend_route_total":       len(backendIndex),
		"frontend_route_total":      calledRoutes,
		"connected_route_total":     len(connected),
		"frontend_scan_files":       frontendScan.Files,
		"frontend_scan_error_count": len(frontendScan.Errors),
		"frontend_scan_errors":      frontendScan.Errors,
		"backend_methods":           backendMethodCount,
		"method_mismatch_total":     methodMismatchCount,
	}

	return CoverageReport{
		TotalRoutes:             len(backendIndex),
		TotalComponents:         len(components),
		TotalCategories:         len(categoryLabels),
		TotalDynamicRoutes:      dynamicCount,
		TotalCalledRoutes:       calledRoutes,
		TotalDisconnectedRoutes: len(frontendOnly),
		TotalUnusedRoutes:       len(unused),
		ConnectionRate:          connectionRate,
		ByCategory:              byCategory,
		Categories:              categories,
		CategoryLabels:          categoryLabels,
		CategoryStats:           categoryStats,
		Components:              components,
		DisconnectedRoutes:      disconnectedMap,
		BackendOnlyRoutes:       backendOnlyMap,
		UnusedRoutes:            unusedMap,
		ButtonToEndpoints:       map[string]map[string][]string{},
		Statistics:              statistics,
		Timestamp:               time.Now().UTC().Format(time.RFC3339),
	}
}
