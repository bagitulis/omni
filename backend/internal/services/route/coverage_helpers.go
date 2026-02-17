package route

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func routeKey(method, endpoint string) string {
	return strings.ToUpper(method) + " " + endpoint
}

func sourceToComponentName(source string) string {
	base := filepath.Base(source)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func toCoverageSummary(route RouteMapping) CoverageRouteSummary {
	category := "uncategorized"
	if len(route.Tags) > 0 && strings.TrimSpace(route.Tags[0]) != "" {
		category = route.Tags[0]
	}

	endpoint := NormalizeRoutePath(route.Path, false)
	return CoverageRouteSummary{
		Endpoint:  endpoint,
		Method:    strings.ToUpper(route.Method),
		IsDynamic: strings.Contains(endpoint, ":"),
		Category:  category,
	}
}

func deriveCategory(endpoint string) string {
	trimmed := strings.TrimPrefix(endpoint, "/")
	if strings.HasPrefix(trimmed, "api/") {
		trimmed = strings.TrimPrefix(trimmed, "api/")
	}

	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "general"
	}

	return parts[0]
}

func isSystemOrExternalRoute(endpoint string) bool {
	systemPrefixes := []string{
		"/api/health",
		"/api/status",
		"/api/docs",
		"/api/csrf-token",
		"/api/webhooks/",
	}

	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(endpoint, prefix) {
			return true
		}
	}

	return false
}

func sortedMapKeys[T any](input map[string]T) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortCoverageSummary(list []CoverageRouteSummary) {
	sort.Slice(list, func(i, j int) bool {
		if list[i].Endpoint == list[j].Endpoint {
			return list[i].Method < list[j].Method
		}
		return list[i].Endpoint < list[j].Endpoint
	})
}

func sortFrontendOnly(list []CoverageFrontendOnlySummary) {
	sort.Slice(list, func(i, j int) bool {
		if list[i].Endpoint == list[j].Endpoint {
			return list[i].Method < list[j].Method
		}
		return list[i].Endpoint < list[j].Endpoint
	})
}

func frontendOnlyToSummary(list []CoverageFrontendOnlySummary) []CoverageRouteSummary {
	result := make([]CoverageRouteSummary, 0, len(list))
	for _, route := range list {
		result = append(result, CoverageRouteSummary{
			Endpoint:  route.Endpoint,
			Method:    route.Method,
			IsDynamic: strings.Contains(route.Endpoint, ":"),
			Category:  route.Category,
		})
	}
	return result
}

func formatPercentage(numerator, denominator int) string {
	if denominator == 0 {
		return "0%"
	}

	value := float64(numerator) * 100 / float64(denominator)
	return fmt.Sprintf("%.1f%%", value)
}
