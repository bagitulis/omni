package route

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var frontendAPICallPattern = regexp.MustCompile("(?i)\\b(?:api|apiClient)(?:\\.client)?\\.(get|post|put|patch|delete)\\s*(?:<[^>]*>)?\\s*\\(\\s*[\"'`]([^\"'`]+)[\"'`]")
var frontendFetchPattern = regexp.MustCompile("(?is)\\bfetch\\s*\\(\\s*[\"'`]([^\"'`]+)[\"'`]\\s*,\\s*\\{[^}]*\\bmethod\\s*:\\s*[\"'`](get|post|put|patch|delete)[\"'`]")
var frontendConstPattern = regexp.MustCompile("(?m)\\bconst\\s+([A-Z][A-Z0-9_]*)\\s*=\\s*(?:\"([^\"]+)\"|'([^']+)'|`([^`]+)`)")
var frontendPathLiteralPattern = regexp.MustCompile("(?:\"(/api/[^\"]+|/[^\"]*?/[^\"]+)\"|'(/api/[^']+|/[^']*?/[^']+)'|`(/api/[^`]+|/[^`]*?/[^`]+)`)")
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
	sourcePath := s.resolveFrontendSourcePath()
	if sourcePath == "" {
		result.Errors = append(result.Errors, "could not locate frontend src directory from base path")
		return result
	}

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, "Error accessing "+path+": "+err.Error())
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if !isFrontendCodeFile(path) {
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

		relPath, _ := filepath.Rel(sourcePath, path)
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
	constMap := parseFrontendStringConstants(content)

	matches := frontendAPICallPattern.FindAllStringSubmatch(content, -1)
	fetchMatches := frontendFetchPattern.FindAllStringSubmatch(content, -1)
	routes := make([]FrontendRouteCall, 0, len(matches)+len(fetchMatches))

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		method := strings.ToUpper(strings.TrimSpace(match[1]))
		endpoint := normalizeFrontendEndpoint(match[2], constMap)
		if endpoint == "" {
			continue
		}

		routes = append(routes, FrontendRouteCall{
			Method:   method,
			Endpoint: endpoint,
			Source:   source,
		})
	}

	for _, match := range fetchMatches {
		if len(match) < 3 {
			continue
		}

		resolvedURL := resolveTemplateConstants(match[1], constMap)
		method := strings.ToUpper(strings.TrimSpace(match[2]))
		endpoint := normalizeFetchURL(resolvedURL)
		if endpoint == "" || method == "" {
			continue
		}

		routes = append(routes, FrontendRouteCall{
			Method:   method,
			Endpoint: endpoint,
			Source:   source,
		})
	}

	routes = append(routes, parsePathLiteralRoutes(content, source, constMap)...)

	return routes
}

func isFrontendCodeFile(path string) bool {
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".test.ts") || strings.HasSuffix(lower, ".test.tsx") {
		return false
	}

	if strings.HasSuffix(lower, ".spec.ts") || strings.HasSuffix(lower, ".spec.tsx") {
		return false
	}

	return strings.HasSuffix(lower, ".ts") || strings.HasSuffix(lower, ".tsx")
}

func parseFrontendStringConstants(content string) map[string]string {
	constMap := make(map[string]string)
	matches := frontendConstPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) < 5 {
			continue
		}

		name := strings.TrimSpace(match[1])
		value := firstNonEmpty(match[2], match[3], match[4])
		value = strings.TrimSpace(value)
		if name == "" || value == "" {
			continue
		}

		constMap[name] = value
	}

	return constMap
}

func normalizeFrontendEndpoint(raw string, constMap map[string]string) string {
	resolved := resolveTemplateConstants(raw, constMap)
	return NormalizeRoutePath(resolved, true)
}

func resolveTemplateConstants(raw string, constMap map[string]string) string {
	resolved := strings.TrimSpace(raw)
	for name, value := range constMap {
		token := "${" + name + "}"
		if strings.Contains(resolved, token) {
			resolved = strings.ReplaceAll(resolved, token, value)
		}
	}

	return resolved
}

func normalizeFetchURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "${") {
		if closeIdx := strings.Index(trimmed, "}"); closeIdx >= 0 && closeIdx+1 < len(trimmed) {
			suffix := trimmed[closeIdx+1:]
			if strings.HasPrefix(suffix, "/") {
				return NormalizeRoutePath(suffix, true)
			}
		}
	}

	if idx := strings.Index(trimmed, "/api/"); idx >= 0 {
		return NormalizeRoutePath(trimmed[idx:], false)
	}

	if trimmed == "/api" {
		return "/api"
	}

	if strings.HasPrefix(trimmed, "/") {
		return NormalizeRoutePath(trimmed, true)
	}

	return ""
}

func parsePathLiteralRoutes(content, source string, constMap map[string]string) []FrontendRouteCall {
	if !strings.Contains(source, "operationMappers") {
		return nil
	}

	literals := frontendPathLiteralPattern.FindAllStringSubmatch(content, -1)
	routes := make([]FrontendRouteCall, 0, len(literals))
	seen := make(map[string]struct{})

	for _, match := range literals {
		if len(match) < 4 {
			continue
		}

		pathLiteral := firstNonEmpty(match[1], match[2], match[3])
		pathValue := resolveTemplateConstants(pathLiteral, constMap)
		endpoint := NormalizeRoutePath(pathValue, true)
		if endpoint == "" {
			continue
		}

		key := "POST " + endpoint
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		routes = append(routes, FrontendRouteCall{
			Method:   "POST",
			Endpoint: endpoint,
			Source:   source,
		})
	}

	return routes
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}

	return ""
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
