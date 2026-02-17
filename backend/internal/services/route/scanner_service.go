package route

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ScannedRoute represents a route found during scanning
type ScannedRoute struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Handler string `json:"handler,omitempty"`
}

// ScanResult represents the result of scanning
type ScanResult struct {
	Routes []ScannedRoute `json:"routes"`
	Files  int            `json:"files_scanned"`
	Errors []string       `json:"errors,omitempty"`
}

// ScannerService provides route scanning functionality
type ScannerService struct {
	basePath string
}

// NewScannerService creates a new scanner service
func NewScannerService(basePath string) *ScannerService {
	return &ScannerService{basePath: basePath}
}

// ScanRoutes scans Go files for route definitions
func (s *ScannerService) ScanRoutes() ScanResult {
	result := ScanResult{
		Routes: make([]ScannedRoute, 0),
	}

	internalPath := s.resolveBackendInternalPath()
	if internalPath == "" {
		result.Errors = append(result.Errors, "could not locate backend internal directory from base path")
		return result
	}

	err := filepath.Walk(internalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, "Error accessing "+path+": "+err.Error())
			return nil
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		result.Files++
		routes := s.scanFile(path)
		result.Routes = append(result.Routes, routes...)

		return nil
	})

	if err != nil {
		result.Errors = append(result.Errors, "Walk error: "+err.Error())
	}

	return result
}

func (s *ScannerService) resolveBackendInternalPath() string {
	candidates := []string{
		filepath.Join(s.basePath, "internal"),
		filepath.Join(s.basePath, "backend", "internal"),
		filepath.Join(".", "internal"),
		filepath.Join(".", "backend", "internal"),
	}

	for _, candidate := range candidates {
		if s.isDirectory(candidate) {
			return candidate
		}
	}

	return ""
}

func (s *ScannerService) resolveFrontendSourcePath() string {
	configuredPath := strings.TrimSpace(os.Getenv("FRONTEND_SOURCE_PATH"))
	if configuredPath != "" && s.isDirectory(configuredPath) {
		return configuredPath
	}

	candidates := []string{
		filepath.Join(s.basePath, "..", "frontend", "src"),
		filepath.Join(s.basePath, "frontend", "src"),
		filepath.Join(s.basePath, "..", "..", "frontend", "src"),
		filepath.Join(".", "frontend", "src"),
		filepath.Join("..", "frontend", "src"),
	}

	for _, candidate := range candidates {
		if s.isDirectory(candidate) {
			return candidate
		}
	}

	return ""
}

func (s *ScannerService) isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// scanFile scans a single Go file for route definitions
func (s *ScannerService) scanFile(filePath string) []ScannedRoute {
	var routes []ScannedRoute

	content, err := os.ReadFile(filePath)
	if err != nil {
		return routes
	}

	lines := strings.Split(string(content), "\n")
	relPath, _ := filepath.Rel(s.basePath, filePath)

	// Patterns for Gin route definitions
	patterns := []struct {
		pattern *regexp.Regexp
		method  string
	}{
		{regexp.MustCompile(`\.GET\s*\(\s*"([^"]+)"`), "GET"},
		{regexp.MustCompile(`\.POST\s*\(\s*"([^"]+)"`), "POST"},
		{regexp.MustCompile(`\.PUT\s*\(\s*"([^"]+)"`), "PUT"},
		{regexp.MustCompile(`\.DELETE\s*\(\s*"([^"]+)"`), "DELETE"},
		{regexp.MustCompile(`\.PATCH\s*\(\s*"([^"]+)"`), "PATCH"},
	}

	for lineNum, line := range lines {
		for _, p := range patterns {
			matches := p.pattern.FindStringSubmatch(line)
			if len(matches) >= 2 {
				routes = append(routes, ScannedRoute{
					Method:  p.method,
					Path:    matches[1],
					File:    relPath,
					Line:    lineNum + 1,
					Handler: extractHandler(line),
				})
			}
		}
	}

	return routes
}

// ScanForMiddleware scans for middleware usage
func (s *ScannerService) ScanForMiddleware() map[string][]string {
	result := make(map[string][]string)

	internalPath := s.resolveBackendInternalPath()
	if internalPath == "" {
		return result
	}

	_ = filepath.Walk(internalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(s.basePath, path)
		middlewares := extractMiddlewares(string(content))

		if len(middlewares) > 0 {
			result[relPath] = middlewares
		}

		return nil
	})

	return result
}

// extractHandler extracts handler name from a route definition line
func extractHandler(line string) string {
	// Pattern: handlers.SomeHandler or h.SomeMethod
	pattern := regexp.MustCompile(`(\w+\.\w+)\s*\)`)
	matches := pattern.FindStringSubmatch(line)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// extractMiddlewares extracts middleware names from file content
func extractMiddlewares(content string) []string {
	var middlewares []string
	seen := make(map[string]bool)

	// Pattern: .Use(middleware.Something) or middleware.Something()
	pattern := regexp.MustCompile(`middleware\.(\w+)`)
	matches := pattern.FindAllStringSubmatch(content, -1)

	for _, m := range matches {
		if len(m) >= 2 && !seen[m[1]] {
			seen[m[1]] = true
			middlewares = append(middlewares, m[1])
		}
	}

	return middlewares
}
