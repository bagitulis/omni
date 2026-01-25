package google

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

// SheetConfigService handles sheet configuration persistence
type SheetConfigService struct {
	db       *gorm.DB
	tenantID string
}

// NewSheetConfigService creates a new sheet config service
func NewSheetConfigService(db *gorm.DB, tenantID string) *SheetConfigService {
	return &SheetConfigService{db: db, tenantID: tenantID}
}

// SheetConfigInput represents input for sheet configuration
type SheetConfigInput struct {
	SheetID      string
	Type         string // inventory, wallet, shipping, order
	Mappings     map[string]string
	HeaderRow    int
	DataStartRow int
}

// SheetConfigOutput represents stored sheet configuration
type SheetConfigOutput struct {
	ID           int64             `json:"id"`
	SheetID      string            `json:"sheet_id"`
	Type         string            `json:"type"`
	Mappings     map[string]string `json:"mappings"`
	HeaderRow    int               `json:"header_row"`
	DataStartRow int               `json:"data_start_row"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
}

// SaveConfig saves or updates a sheet configuration
func (s *SheetConfigService) SaveConfig(ctx context.Context, input *SheetConfigInput) error {
	mappingsJSON, err := json.Marshal(input.Mappings)
	if err != nil {
		return fmt.Errorf("marshal mappings: %w", err)
	}

	result := s.db.WithContext(ctx).Exec(`
		INSERT INTO SheetConfig (tenant_id, sheet_id, type, mappings, header_row, data_start_row, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(tenant_id, sheet_id) DO UPDATE SET
			type = excluded.type,
			mappings = excluded.mappings,
			header_row = excluded.header_row,
			data_start_row = excluded.data_start_row,
			updated_at = CURRENT_TIMESTAMP
	`, s.tenantID, input.SheetID, input.Type, string(mappingsJSON), input.HeaderRow, input.DataStartRow)

	return result.Error
}

// ListConfigs retrieves all sheet configurations for tenant
func (s *SheetConfigService) ListConfigs(ctx context.Context) ([]SheetConfigOutput, error) {
	var configs []struct {
		ID           int64
		SheetID      string
		Type         string
		Mappings     string
		HeaderRow    int
		DataStartRow int
		CreatedAt    string
		UpdatedAt    string
	}

	result := s.db.WithContext(ctx).Raw(`
		SELECT id, sheet_id, type, mappings, header_row, data_start_row, created_at, updated_at
		FROM SheetConfig
		WHERE tenant_id = ?
		ORDER BY updated_at DESC
	`, s.tenantID).Scan(&configs)

	if result.Error != nil {
		return nil, result.Error
	}

	output := make([]SheetConfigOutput, len(configs))
	for i, cfg := range configs {
		var mappings map[string]string
		if cfg.Mappings != "" {
			json.Unmarshal([]byte(cfg.Mappings), &mappings)
		}

		output[i] = SheetConfigOutput{
			ID:           cfg.ID,
			SheetID:      cfg.SheetID,
			Type:         cfg.Type,
			Mappings:     mappings,
			HeaderRow:    cfg.HeaderRow,
			DataStartRow: cfg.DataStartRow,
			CreatedAt:    cfg.CreatedAt,
			UpdatedAt:    cfg.UpdatedAt,
		}
	}

	return output, nil
}

// DeleteConfig deletes a sheet configuration
func (s *SheetConfigService) DeleteConfig(ctx context.Context, sheetID string) error {
	result := s.db.WithContext(ctx).Exec(`
		DELETE FROM SheetConfig
		WHERE tenant_id = ? AND sheet_id = ?
	`, s.tenantID, sheetID)

	return result.Error
}

// GetConfig retrieves a specific sheet configuration
func (s *SheetConfigService) GetConfig(ctx context.Context, sheetID string) (*SheetConfigOutput, error) {
	var cfg struct {
		ID           int64
		SheetID      string
		Type         string
		Mappings     string
		HeaderRow    int
		DataStartRow int
		CreatedAt    string
		UpdatedAt    string
	}

	result := s.db.WithContext(ctx).Raw(`
		SELECT id, sheet_id, type, mappings, header_row, data_start_row, created_at, updated_at
		FROM SheetConfig
		WHERE tenant_id = ? AND sheet_id = ?
	`, s.tenantID, sheetID).Scan(&cfg)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	var mappings map[string]string
	if cfg.Mappings != "" {
		json.Unmarshal([]byte(cfg.Mappings), &mappings)
	}

	return &SheetConfigOutput{
		ID:           cfg.ID,
		SheetID:      cfg.SheetID,
		Type:         cfg.Type,
		Mappings:     mappings,
		HeaderRow:    cfg.HeaderRow,
		DataStartRow: cfg.DataStartRow,
		CreatedAt:    cfg.CreatedAt,
		UpdatedAt:    cfg.UpdatedAt,
	}, nil
}
