package sheets

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// SheetConfig represents a Google Sheet configuration
type SheetConfig struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      string    `gorm:"index;not null" json:"tenantId"`
	SpreadsheetID string    `gorm:"index;not null" json:"spreadsheetId"`
	SheetName     string    `json:"sheetName"`
	SheetType     string    `gorm:"index" json:"sheetType"` // inventory, wallet, shipping, etc.
	RangeStart    string    `json:"rangeStart,omitempty"`
	RangeEnd      string    `json:"rangeEnd,omitempty"`
	HeaderRow     int       `gorm:"default:1" json:"headerRow"`
	DataStartRow  int       `gorm:"default:2" json:"dataStartRow"`
	ColumnMapping string    `json:"columnMapping,omitempty"` // JSON mapping
	IsActive      bool      `gorm:"default:true" json:"isActive"`
	LastSyncAt    time.Time `json:"lastSyncAt,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// TableName returns the table name for GORM
func (SheetConfig) TableName() string {
	return "sheet_configs"
}

// SheetSnapshot represents a snapshot of sheet data
type SheetSnapshot struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      string    `gorm:"index;not null" json:"tenantId"`
	SheetConfigID uint      `gorm:"index;not null" json:"sheetConfigId"`
	DataHash      string    `json:"dataHash"`
	RowCount      int       `json:"rowCount"`
	Snapshot      string    `json:"snapshot"` // JSON data
	CreatedAt     time.Time `json:"createdAt"`
}

// TableName returns the table name for GORM
func (SheetSnapshot) TableName() string {
	return "sheet_snapshots"
}

// SheetConfigService handles sheet configuration operations
type SheetConfigService struct {
	db *gorm.DB
}

// NewSheetConfigService creates a new sheet config service
func NewSheetConfigService(db *gorm.DB) *SheetConfigService {
	return &SheetConfigService{db: db}
}

// GetSheetConfigs retrieves all sheet configurations for a tenant
func (s *SheetConfigService) GetSheetConfigs(
	ctx context.Context,
	tenantID string,
) ([]SheetConfig, error) {
	var configs []SheetConfig

	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("sheet_type, sheet_name").
		Find(&configs).Error

	return configs, err
}

// GetSheetConfigsByType retrieves sheet configurations by type
func (s *SheetConfigService) GetSheetConfigsByType(
	ctx context.Context,
	tenantID string,
	sheetType string,
) ([]SheetConfig, error) {
	var configs []SheetConfig

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND sheet_type = ?", tenantID, sheetType).
		Find(&configs).Error

	return configs, err
}

// GetSheetConfig retrieves a specific sheet configuration
func (s *SheetConfigService) GetSheetConfig(
	ctx context.Context,
	tenantID string,
	id uint,
) (*SheetConfig, error) {
	var config SheetConfig

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&config).Error

	if err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveSheetConfig saves or updates a sheet configuration
func (s *SheetConfigService) SaveSheetConfig(
	ctx context.Context,
	config *SheetConfig,
) error {
	if config.ID == 0 {
		return s.db.WithContext(ctx).Create(config).Error
	}
	return s.db.WithContext(ctx).Save(config).Error
}

// DeleteSheetConfig deletes a sheet configuration
func (s *SheetConfigService) DeleteSheetConfig(
	ctx context.Context,
	tenantID string,
	id uint,
) error {
	return s.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&SheetConfig{}).Error
}

// SaveSnapshot saves a sheet snapshot
func (s *SheetConfigService) SaveSnapshot(
	ctx context.Context,
	snapshot *SheetSnapshot,
) error {
	return s.db.WithContext(ctx).Create(snapshot).Error
}

// GetLatestSnapshot retrieves the latest snapshot for a sheet config
func (s *SheetConfigService) GetLatestSnapshot(
	ctx context.Context,
	tenantID string,
	sheetConfigID uint,
) (*SheetSnapshot, error) {
	var snapshot SheetSnapshot

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND sheet_config_id = ?", tenantID, sheetConfigID).
		Order("created_at DESC").
		First(&snapshot).Error

	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

// UpdateLastSync updates the last sync timestamp
func (s *SheetConfigService) UpdateLastSync(
	ctx context.Context,
	tenantID string,
	id uint,
) error {
	return s.db.WithContext(ctx).
		Model(&SheetConfig{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("last_sync_at", time.Now()).Error
}
