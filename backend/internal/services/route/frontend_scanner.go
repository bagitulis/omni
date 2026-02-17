package route

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var frontendAPICallPattern = regexp.MustCompile("(?i)\\b(?:api|apiClient)\\.(get|post|put|patch|delete)\\s*(?:<[^>]*>)?\\s*\\(\\s*[\"'`]([^\"'`]+)[\"'`]")
var dynamicSegmentPattern = regexp.MustCompile(`\$\{[^}]+\}`)
var backendParamPattern = regexp.MustCompile(`:[^/]+`)

// FrontendRouteCall represents a frontend endpoint invocation.
type FrontendRouteCall struct {
	Method   string `json:"method"`
	Endpoint string `json:"endpoint"`
	Source   string `json:"source"`
}

// FrontendScanResult represents frontend scan output.
type FrontendScanResult struct {
	Routes []FrontendRouteCall `json:"routes"`
	Files  int                 `json:"files_scanned"`
	Errors []string            `json:"errors,omitempty"`
}

// ScanFrontendRoutes scans frontend API adapter files for route calls.
func (s *ScannerService) ScanFrontendRoutes() FrontendScanResult {
	result := FrontendScanResult{Routes: make([]FrontendRouteCall, 0)}
	apiPath := s.resolveFrontendAPIPath()
	if apiPath == "" {
		result.Errors = append(result.Errors, "could not locate frontend src/api directory from base path")
		return result
	}

	err := filepath.Walk(apiPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, "Error accessing "+path+": "+err.Error())
			return nil
		}

		if info.IsDir() || !strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".test.ts") {
			return nil
		}

		base := strings.ToLower(filepath.Base(path))
		if base == "client.ts" || base == "queryclient.ts" {
			return nil
		}

		result.Files++
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			result.Errors = append(result.Errors, "Read error "+path+": "+readErr.Error())
			return nil
		}

		relPath, _ := filepath.Rel(apiPath, path)
		routes := parseFrontendRoutes(string(content), relPath)
		result.Routes = append(result.Routes, routes...)
		return nil
	})

	if err != nil {
		result.Errors = append(result.Errors, "Walk error: "+err.Error())
	}

	result.Routes = uniqueFrontendRoutes(result.Routes)
	return result
}

func parseFrontendRoutes(content, source string) []FrontendRouteCall {
	matches := frontendAPICallPattern.FindAllStringSubmatch(content, -1)
	routes := make([]FrontendRouteCall, 0, len(matches))

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		method := strings.ToUpper(strings.TrimSpace(match[1]))
		endpoint := NormalizeRoutePath(match[2], true)
		if endpoint == "" {
			continue
		}

		routes = append(routes, FrontendRouteCall{
			Method:   method,
			Endpoint: endpoint,
			Source:   source,
		})
	}

	return routes
}

func NormalizeRoutePath(path string, ensureAPIPrefix bool) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return ""
	}

	trimmed = strings.Split(trimmed, "?")[0]
	trimmed = dynamicSegmentPattern.ReplaceAllString(trimmed, ":param")
	trimmed = backendParamPattern.ReplaceAllString(trimmed, ":param")

	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}

	if ensureAPIPrefix && !strings.HasPrefix(trimmed, "/api/") && trimmed != "/api" {
		trimmed = "/api" + trimmed
	}

	if len(trimmed) > 1 {
		trimmed = strings.TrimSuffix(trimmed, "/")
	}

	return trimmed
}

func uniqueFrontendRoutes(routes []FrontendRouteCall) []FrontendRouteCall {
	unique := make(map[string]FrontendRouteCall, len(routes))
	for _, route := range routes {
		key := route.Method + " " + route.Endpoint + " " + route.Source
		if _, exists := unique[key]; !exists {
			unique[key] = route
		}
	}

	result := make([]FrontendRouteCall, 0, len(unique))
	for _, route := range unique {
		result = append(result, route)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Endpoint == result[j].Endpoint {
			if result[i].Method == result[j].Method {
				return result[i].Source < result[j].Source
			}
			return result[i].Method < result[j].Method
		}
		return result[i].Endpoint < result[j].Endpoint
	})

	return result
}
