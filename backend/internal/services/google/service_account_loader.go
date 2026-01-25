package google

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/omni/backend/internal/utils/logger"
)

// ServiceAccountLoader handles loading service account credentials
type ServiceAccountLoader struct {
	basePath string
}

// NewServiceAccountLoader creates a new service account loader
func NewServiceAccountLoader(basePath string) *ServiceAccountLoader {
	return &ServiceAccountLoader{basePath: basePath}
}

// LoadFirstAvailable loads the first available service account JSON file
func (l *ServiceAccountLoader) LoadFirstAvailable() ([]byte, error) {
	accounts, err := l.DiscoverAccounts()
	if err != nil {
		return nil, err
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("no service account files found in %s", l.basePath)
	}

	// Load the first account
	credentials, err := os.ReadFile(accounts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read service account file %s: %w", accounts[0], err)
	}

	logger.Infof("Loaded service account: %s", filepath.Base(accounts[0]))
	return credentials, nil
}

// DiscoverAccounts finds all service account JSON files in the base path
func (l *ServiceAccountLoader) DiscoverAccounts() ([]string, error) {
	var accounts []string

	// Check if path exists
	if _, err := os.Stat(l.basePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("service account directory does not exist: %s", l.basePath)
	}

	entries, err := os.ReadDir(l.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", l.basePath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip non-JSON files and registry.json
		if !strings.HasSuffix(name, ".json") || name == "registry.json" {
			continue
		}

		fullPath := filepath.Join(l.basePath, name)
		accounts = append(accounts, fullPath)
		logger.Debugf("Discovered service account: %s", name)
	}

	return accounts, nil
}

// LoadAllAccounts loads all available service account credentials
func (l *ServiceAccountLoader) LoadAllAccounts() ([][]byte, error) {
	accounts, err := l.DiscoverAccounts()
	if err != nil {
		return nil, err
	}

	var credentials [][]byte
	for _, accountPath := range accounts {
		data, err := os.ReadFile(accountPath)
		if err != nil {
			logger.Warnf("Failed to read service account %s: %v", accountPath, err)
			continue
		}
		credentials = append(credentials, data)
		logger.Infof("Loaded service account: %s", filepath.Base(accountPath))
	}

	if len(credentials) == 0 {
		return nil, fmt.Errorf("no valid service account files found")
	}

	return credentials, nil
}

// GetDefaultPath returns the default path for service account files
func GetDefaultServiceAccountPath() string {
	// Check environment variable first
	if path := os.Getenv("GOOGLE_SERVICE_ACCOUNT_PATH"); path != "" {
		return path
	}

	// Check relative paths from working directory
	possiblePaths := []string{
		"config/static/google",
		"../config/static/google",
		"/app/config/static/google", // Docker path
	}

	for _, p := range possiblePaths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}

	// Default fallback
	return "config/static/google"
}
