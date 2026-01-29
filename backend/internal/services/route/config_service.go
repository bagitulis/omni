package route

import (
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ConfigService manages route configurations
type ConfigService struct {
	db       *gorm.DB
	tenantID string
}

// NewConfigService creates a new route config service
func NewConfigService(db *gorm.DB, tenantID string) *ConfigService {
	return &ConfigService{db: db, tenantID: tenantID}
}

// GetAll retrieves all route configs
func (s *ConfigService) GetAll() ([]models.RouteExecutionConfig, error) {
	var configs []models.RouteExecutionConfig
	err := s.db.Where("tenant_id = ?", s.tenantID).Find(&configs).Error
	return configs, err
}

// Get retrieves a route config by path
func (s *ConfigService) Get(routePath string) (*models.RouteExecutionConfig, error) {
	var config models.RouteExecutionConfig
	err := s.db.Where("tenant_id = ? AND route_path = ?", s.tenantID, routePath).First(&config).Error
	if err == gorm.ErrRecordNotFound {
		// Return default config
		return s.getDefaultConfig(routePath), nil
	}
	return &config, err
}

// Update updates a route config
func (s *ConfigService) Update(routePath string, req RouteConfigRequest) (*models.RouteExecutionConfig, error) {
	var config models.RouteExecutionConfig
	err := s.db.Where("tenant_id = ? AND route_path = ?", s.tenantID, routePath).First(&config).Error

	if err == gorm.ErrRecordNotFound {
		// Create new
		config = models.RouteExecutionConfig{
			TenantID:  s.tenantID,
			RoutePath: routePath,
		}
	}

	config.IsEnabled = req.IsEnabled
	config.RateLimit = req.RateLimit
	config.Timeout = req.Timeout
	config.Category = req.Category

	if err := s.db.Save(&config).Error; err != nil {
		return nil, err
	}

	return &config, nil
}

// Enable enables a route
func (s *ConfigService) Enable(routePath string) error {
	return s.db.Model(&models.RouteExecutionConfig{}).
		Where("tenant_id = ? AND route_path = ?", s.tenantID, routePath).
		Update("is_enabled", true).Error
}

// Disable disables a route
func (s *ConfigService) Disable(routePath string) error {
	return s.db.Model(&models.RouteExecutionConfig{}).
		Where("tenant_id = ? AND route_path = ?", s.tenantID, routePath).
		Update("is_enabled", false).Error
}

// ResetToDefault resets a route config to defaults
func (s *ConfigService) ResetToDefault(routePath string) error {
	return s.db.Where("tenant_id = ? AND route_path = ?", s.tenantID, routePath).
		Delete(&models.RouteExecutionConfig{}).Error
}

// ResetAll resets all route configs to defaults
func (s *ConfigService) ResetAll() error {
	return s.db.Where("tenant_id = ?", s.tenantID).
		Delete(&models.RouteExecutionConfig{}).Error
}

// getDefaultConfig returns default config for a route
func (s *ConfigService) getDefaultConfig(routePath string) *models.RouteExecutionConfig {
	return &models.RouteExecutionConfig{
		TenantID:  s.tenantID,
		RoutePath: routePath,
		IsEnabled: true,
		RateLimit: 60, // 60 requests per minute
		Timeout:   30, // 30 seconds
	}
}

// RouteConfigRequest represents route config update request
type RouteConfigRequest struct {
	RoutePath string `json:"route_path"`
	IsEnabled bool   `json:"is_enabled"`
	RateLimit int    `json:"rate_limit"`
	Timeout   int    `json:"timeout"`
	Category  string `json:"category"`
}

// GetCategories returns distinct categories
func (s *ConfigService) GetCategories() ([]string, error) {
	var categories []string
	err := s.db.Model(&models.RouteExecutionConfig{}).
		Where("tenant_id = ?", s.tenantID).
		Distinct("category").
		Where("category IS NOT NULL AND category != ''").
		Pluck("category", &categories).Error
	return categories, err
}

// GetByCategory returns routes by category
func (s *ConfigService) GetByCategory(category string) ([]models.RouteExecutionConfig, error) {
	var configs []models.RouteExecutionConfig
	err := s.db.Where("tenant_id = ? AND category = ?", s.tenantID, category).Find(&configs).Error
	return configs, err
}

// Create creates a new route config
func (s *ConfigService) Create(req RouteConfigRequest) (*models.RouteExecutionConfig, error) {
	// Check if exists
	var existing models.RouteExecutionConfig
	err := s.db.Where("tenant_id = ? AND route_path = ?", s.tenantID, req.RoutePath).First(&existing).Error
	if err == nil {
		return nil, gorm.ErrDuplicatedKey
	}

	config := models.RouteExecutionConfig{
		TenantID:  s.tenantID,
		RoutePath: req.RoutePath,
		IsEnabled: req.IsEnabled,
		RateLimit: req.RateLimit,
		Timeout:   req.Timeout,
		Category:  req.Category,
	}

	if err := s.db.Create(&config).Error; err != nil {
		return nil, err
	}

	return &config, nil
}

// Delete deletes a route config
func (s *ConfigService) Delete(id string) (*models.RouteExecutionConfig, error) {
	var config models.RouteExecutionConfig
	err := s.db.Where("tenant_id = ? AND id = ?", s.tenantID, id).First(&config).Error
	if err != nil {
		return nil, err
	}

	if err := s.db.Delete(&config).Error; err != nil {
		return nil, err
	}

	return &config, nil
}

// BulkUpdate updates multiple route configs
func (s *ConfigService) BulkUpdate(routeIDs []string, updates map[string]interface{}) ([]models.RouteExecutionConfig, error) {
	if err := s.db.Model(&models.RouteExecutionConfig{}).
		Where("tenant_id = ? AND id IN ?", s.tenantID, routeIDs).
		Updates(updates).Error; err != nil {
		return nil, err
	}

	var configs []models.RouteExecutionConfig
	s.db.Where("tenant_id = ? AND id IN ?", s.tenantID, routeIDs).Find(&configs)
	return configs, nil
}

// ApplyPreset applies a preset configuration to routes
func (s *ConfigService) ApplyPreset(routeIDs []string, preset string) ([]models.RouteExecutionConfig, error) {
	var updates map[string]interface{}

	switch preset {
	case "high-performance":
		updates = map[string]interface{}{
			"is_enabled": true,
			"rate_limit": 120,
			"cache_ttl":  300,
			"timeout":    15,
		}
	case "standard":
		updates = map[string]interface{}{
			"is_enabled": true,
			"rate_limit": 60,
			"cache_ttl":  60,
			"timeout":    30,
		}
	case "low-traffic":
		updates = map[string]interface{}{
			"is_enabled": true,
			"rate_limit": 30,
			"cache_ttl":  600,
			"timeout":    60,
		}
	case "disabled":
		updates = map[string]interface{}{
			"is_enabled": false,
		}
	default:
		return nil, gorm.ErrInvalidData
	}

	return s.BulkUpdate(routeIDs, updates)
}
