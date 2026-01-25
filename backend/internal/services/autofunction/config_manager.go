package autofunction

import (
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ConfigManager manages auto function configurations
// Uses schema-based multi-tenancy (tenant_id column not needed)
type ConfigManager struct {
	db       *gorm.DB
	tenantID string // Kept for reference but not used in queries
}

// NewConfigManager creates a new config manager
func NewConfigManager(db *gorm.DB, tenantID string) *ConfigManager {
	return &ConfigManager{db: db, tenantID: tenantID}
}

// Create creates a new auto function config
func (m *ConfigManager) Create(req models.AutoFunctionRequest) (*models.AutoFunctionConfig, error) {
	now := time.Now()
	var nextExecution *time.Time
	if req.Enabled {
		next := now.Add(time.Duration(req.IntervalMinutes) * time.Minute)
		nextExecution = &next
	}

	cfg := &models.AutoFunctionConfig{
		Name:                   req.Name,
		Enabled:                req.Enabled,
		IntervalMinutes:        req.IntervalMinutes,
		StartTime:              req.StartTime,
		EndTime:                req.EndTime,
		NextScheduledExecution: nextExecution,
	}

	if err := m.db.Create(cfg).Error; err != nil {
		return nil, err
	}

	return cfg, nil
}

// Update updates an existing config by ID
func (m *ConfigManager) Update(id uint, req models.AutoFunctionRequest) (*models.AutoFunctionConfig, error) {
	var cfg models.AutoFunctionConfig
	if err := m.db.Where("id = ?", id).First(&cfg).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	var nextExecution *time.Time
	if req.Enabled {
		next := now.Add(time.Duration(req.IntervalMinutes) * time.Minute)
		nextExecution = &next
	}

	cfg.Name = req.Name
	cfg.Enabled = req.Enabled
	cfg.IntervalMinutes = req.IntervalMinutes
	cfg.StartTime = req.StartTime
	cfg.EndTime = req.EndTime
	cfg.NextScheduledExecution = nextExecution
	cfg.UpdatedAt = now

	if err := m.db.Save(&cfg).Error; err != nil {
		return nil, err
	}

	return &cfg, nil
}

// UpdateByName updates config by name
func (m *ConfigManager) UpdateByName(name string, req models.AutoFunctionRequest) (*models.AutoFunctionConfig, error) {
	var cfg models.AutoFunctionConfig
	if err := m.db.Where("name = ?", name).First(&cfg).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	var nextExecution *time.Time
	if req.Enabled {
		next := now.Add(time.Duration(req.IntervalMinutes) * time.Minute)
		nextExecution = &next
	}

	cfg.Enabled = req.Enabled
	cfg.IntervalMinutes = req.IntervalMinutes
	cfg.StartTime = req.StartTime
	cfg.EndTime = req.EndTime
	cfg.NextScheduledExecution = nextExecution
	cfg.UpdatedAt = now

	if err := m.db.Save(&cfg).Error; err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Get retrieves a config by ID
func (m *ConfigManager) Get(id uint) (*models.AutoFunctionConfig, error) {
	var cfg models.AutoFunctionConfig
	err := m.db.Where("id = ?", id).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetByName retrieves a config by name
func (m *ConfigManager) GetByName(name string) (*models.AutoFunctionConfig, error) {
	var cfg models.AutoFunctionConfig
	err := m.db.Where("name = ?", name).First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// List retrieves all configs
func (m *ConfigManager) List() ([]models.AutoFunctionConfig, error) {
	var configs []models.AutoFunctionConfig
	err := m.db.Order("created_at DESC").Find(&configs).Error
	return configs, err
}

// Delete deletes a config by ID
func (m *ConfigManager) Delete(id uint) error {
	return m.db.Where("id = ?", id).Delete(&models.AutoFunctionConfig{}).Error
}

// DeleteByName deletes a config by name
func (m *ConfigManager) DeleteByName(name string) error {
	return m.db.Where("name = ?", name).Delete(&models.AutoFunctionConfig{}).Error
}

// Enable enables a config by name
func (m *ConfigManager) Enable(name string) error {
	now := time.Now()
	var cfg models.AutoFunctionConfig
	if err := m.db.Where("name = ?", name).First(&cfg).Error; err != nil {
		return err
	}
	next := now.Add(time.Duration(cfg.IntervalMinutes) * time.Minute)
	return m.db.Model(&models.AutoFunctionConfig{}).
		Where("name = ?", name).
		Updates(map[string]interface{}{
			"enabled":                  true,
			"next_scheduled_execution": next,
		}).Error
}

// Disable disables a config by name
func (m *ConfigManager) Disable(name string) error {
	return m.db.Model(&models.AutoFunctionConfig{}).
		Where("name = ?", name).
		Updates(map[string]interface{}{
			"enabled":                  false,
			"next_scheduled_execution": nil,
		}).Error
}

// CancelScheduled cancels a scheduled execution by ID
func (m *ConfigManager) CancelScheduled(id uint) error {
	return m.db.Model(&models.AutoFunctionConfig{}).
		Where("id = ?", id).
		Update("next_scheduled_execution", nil).Error
}

// CancelScheduledByName cancels scheduled execution by name
func (m *ConfigManager) CancelScheduledByName(name string) error {
	return m.db.Model(&models.AutoFunctionConfig{}).
		Where("name = ?", name).
		Update("next_scheduled_execution", nil).Error
}

// UpdateLastExecuted updates last execution time and schedules next
func (m *ConfigManager) UpdateLastExecuted(name string) error {
	now := time.Now()
	var cfg models.AutoFunctionConfig
	if err := m.db.Where("name = ?", name).First(&cfg).Error; err != nil {
		return err
	}

	next := now.Add(time.Duration(cfg.IntervalMinutes) * time.Minute)
	return m.db.Model(&models.AutoFunctionConfig{}).
		Where("name = ?", name).
		Updates(map[string]interface{}{
			"last_executed":            now,
			"next_scheduled_execution": next,
		}).Error
}

// CreateOrUpdate creates or updates a config by name
func (m *ConfigManager) CreateOrUpdate(req models.AutoFunctionRequest) (*models.AutoFunctionConfig, error) {
	existing, err := m.GetByName(req.Name)
	if err == gorm.ErrRecordNotFound {
		return m.Create(req)
	}
	if err != nil {
		return nil, err
	}
	return m.Update(existing.ID, req)
}
