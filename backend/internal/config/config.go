package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Config holds application configuration
type Config struct {
	Environment   string
	Port          string
	DatabasePath  string
	JWTSecret     string
	EncryptionKey string

	// CORS settings
	CORSOrigins []string

	// Rate limiting
	RateLimitRate  int // requests per interval
	RateLimitBurst int // max burst

	// Server settings
	ReadTimeout     int // seconds
	WriteTimeout    int // seconds
	ShutdownTimeout int // seconds

	// Database driver selection
	DBDriver string // "sqlite" or "postgres"

	// PostgreSQL settings (used when DBDriver = "postgres")
	PGHost     string
	PGPort     int
	PGUser     string
	PGPassword string
	PGDatabase string
	PGSSLMode  string
}

// TenantConfig represents a tenant configuration
// Matches Node.js tenants.json structure
type TenantConfig struct {
	ID       string `json:"-"`       // Populated from map key
	DBPath   string `json:"db_path"` // Path to tenant database
	ShopName string `json:"shop_name"`
	IsGlobal bool   `json:"is_global,omitempty"` // True for system tenant
}

// TenantsConfig holds all tenant configurations
// Node.js format: { "tenant_id": { "dbPath": "...", "shopName": "..." }, ... }
type TenantsConfig struct {
	Tenants map[string]*TenantConfig // Map of tenant ID to config
}

var (
	tenantsConfig *TenantsConfig
	tenantsMu     sync.RWMutex
)

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		Environment:     getEnv("GO_ENV", "development"),
		Port:            getEnv("PORT", "8080"),
		DatabasePath:    getEnv("DATABASE_PATH", "../backend/data"),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		EncryptionKey:   getEnv("ENCRYPTION_KEY", ""),
		CORSOrigins:     getEnvList("CORS_ORIGINS", []string{"http://localhost:5173", "http://localhost:3000"}),
		RateLimitRate:   getEnvInt("RATE_LIMIT_RATE", 100),
		RateLimitBurst:  getEnvInt("RATE_LIMIT_BURST", 200),
		ReadTimeout:     getEnvInt("READ_TIMEOUT", 30),
		WriteTimeout:    getEnvInt("WRITE_TIMEOUT", 30),
		ShutdownTimeout: getEnvInt("SHUTDOWN_TIMEOUT", 30),

		// Database driver (default: sqlite, switch to postgres for production)
		DBDriver: getEnv("DB_DRIVER", "sqlite"),

		// PostgreSQL settings
		PGHost:     getEnv("PG_HOST", "localhost"),
		PGPort:     getEnvInt("PG_PORT", 5432),
		PGUser:     getEnv("PG_USER", "omni"),
		PGPassword: getEnv("PG_PASSWORD", ""),
		PGDatabase: getEnv("PG_DATABASE", "omni_main"),
		PGSSLMode:  getEnv("PG_SSLMODE", "disable"),
	}
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.EncryptionKey == "" {
		return fmt.Errorf("ENCRYPTION_KEY is required")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.DBDriver == "postgres" {
		if c.PGPassword == "" {
			return fmt.Errorf("PG_PASSWORD is required when using postgres driver")
		}
		if c.PGHost == "" {
			return fmt.Errorf("PG_HOST is required when using postgres driver")
		}
	}
	return nil
}

// IsProduction returns true if running in production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// LoadTenants loads tenant configuration from file
// Node.js format: { "tenant_id": { "dbPath": "...", "shopName": "..." }, ... }
func LoadTenants(basePath string) (*TenantsConfig, error) {
	tenantsMu.Lock()
	defer tenantsMu.Unlock()

	if tenantsConfig != nil {
		return tenantsConfig, nil
	}

	configPath := filepath.Join(basePath, "config", "static", "tenants.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tenants.json: %w", err)
	}

	// Parse as map (Node.js format)
	var rawConfig map[string]TenantConfig
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse tenants.json: %w", err)
	}

	// Convert to TenantsConfig with IDs populated
	cfg := &TenantsConfig{
		Tenants: make(map[string]*TenantConfig),
	}
	for id, tenant := range rawConfig {
		t := tenant // Copy to avoid pointer issues
		t.ID = id
		cfg.Tenants[id] = &t
	}

	tenantsConfig = cfg
	return tenantsConfig, nil
}

// GetTenant returns tenant config by ID
func GetTenant(tenantID string) (*TenantConfig, error) {
	tenantsMu.RLock()
	defer tenantsMu.RUnlock()

	if tenantsConfig == nil {
		return nil, fmt.Errorf("tenants not loaded")
	}

	if tenant, exists := tenantsConfig.Tenants[tenantID]; exists {
		return tenant, nil
	}
	return nil, fmt.Errorf("tenant not found: %s", tenantID)
}

// GetTenantDBPath returns the database path for a tenant
// Reads from tenants.json dbPath field
func GetTenantDBPath(tenantID, basePath string) (string, error) {
	tenant, err := GetTenant(tenantID)
	if err != nil {
		return "", err
	}
	if tenant.DBPath == "" {
		// Fallback to convention if dbPath not specified
		return filepath.Join(basePath, "config", "databases", fmt.Sprintf("%s.db", tenantID)), nil
	}
	// dbPath in tenants.json is relative to basePath
	return filepath.Join(basePath, tenant.DBPath), nil
}

// ValidateTenant checks if tenant ID is valid (not global/system)
func ValidateTenant(tenantID string) bool {
	tenant, err := GetTenant(tenantID)
	if err != nil || tenant == nil {
		return false
	}
	// Global tenants (like "system") are not valid for user operations
	return !tenant.IsGlobal
}

// GetRealTenants returns only non-global tenants (excludes "system")
func GetRealTenants() []string {
	tenantsMu.RLock()
	defer tenantsMu.RUnlock()

	if tenantsConfig == nil {
		return nil
	}

	var tenants []string
	for id, t := range tenantsConfig.Tenants {
		if !t.IsGlobal && id != "system" {
			tenants = append(tenants, id)
		}
	}
	return tenants
}

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultVal
}

func getEnvList(key string, defaultVal []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		var result []string
		for _, v := range splitString(value, ",") {
			if trimmed := trimSpace(v); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultVal
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// GetDataDir returns the database/data directory path from environment
func GetDataDir() string {
	return getEnv("DATABASE_PATH", "./data")
}
