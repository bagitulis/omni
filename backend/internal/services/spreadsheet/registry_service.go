package spreadsheet

import (
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// RegistryService manages spreadsheet registry
type RegistryService struct {
	db       *gorm.DB
	tenantID string
}

// NewRegistryService creates a new registry service
func NewRegistryService(db *gorm.DB, tenantID string) *RegistryService {
	return &RegistryService{db: db, tenantID: tenantID}
}

// Register registers a new spreadsheet
func (s *RegistryService) Register(req RegisterRequest) (*models.Spreadsheet, error) {
	spreadsheet := models.Spreadsheet{
		TenantID:      s.tenantID,
		SpreadsheetID: req.SpreadsheetID,
		Name:          req.Name,
		Purpose:       req.Purpose,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(&spreadsheet).Error; err != nil {
		return nil, err
	}

	return &spreadsheet, nil
}

// List lists all registered spreadsheets
func (s *RegistryService) List(filter ListFilter) ([]models.Spreadsheet, error) {
	var spreadsheets []models.Spreadsheet
	query := s.db.Where("tenant_id = ?", s.tenantID)

	if filter.Purpose != "" {
		query = query.Where("purpose = ?", filter.Purpose)
	}
	if filter.ActiveOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("created_at DESC").Find(&spreadsheets).Error
	return spreadsheets, err
}

// Get retrieves a spreadsheet by ID
func (s *RegistryService) Get(id uint) (*models.Spreadsheet, error) {
	var spreadsheet models.Spreadsheet
	err := s.db.Where("id = ? AND tenant_id = ?", id, s.tenantID).First(&spreadsheet).Error
	return &spreadsheet, err
}

// GetBySpreadsheetID retrieves by Google spreadsheet ID
func (s *RegistryService) GetBySpreadsheetID(spreadsheetID string) (*models.Spreadsheet, error) {
	var spreadsheet models.Spreadsheet
	err := s.db.Where("spreadsheet_id = ? AND tenant_id = ?", spreadsheetID, s.tenantID).First(&spreadsheet).Error
	return &spreadsheet, err
}

// Update updates a spreadsheet
func (s *RegistryService) Update(id uint, req UpdateRequest) (*models.Spreadsheet, error) {
	spreadsheet, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		spreadsheet.Name = req.Name
	}
	if req.Purpose != "" {
		spreadsheet.Purpose = req.Purpose
	}
	spreadsheet.IsActive = req.IsActive
	spreadsheet.UpdatedAt = time.Now()

	if err := s.db.Save(spreadsheet).Error; err != nil {
		return nil, err
	}

	return spreadsheet, nil
}

// Delete deletes a spreadsheet
func (s *RegistryService) Delete(id uint) error {
	return s.db.Where("id = ? AND tenant_id = ?", id, s.tenantID).Delete(&models.Spreadsheet{}).Error
}

// UpdateLastSync updates the last sync time
func (s *RegistryService) UpdateLastSync(id uint) error {
	now := time.Now()
	return s.db.Model(&models.Spreadsheet{}).
		Where("id = ? AND tenant_id = ?", id, s.tenantID).
		Updates(map[string]interface{}{
			"last_synced_at": now,
			"updated_at":     now,
		}).Error
}

// RegisterRequest represents spreadsheet registration request
type RegisterRequest struct {
	SpreadsheetID string `json:"spreadsheet_id" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Purpose       string `json:"purpose"` // inventory, orders, analytics
}

// UpdateRequest represents spreadsheet update request
type UpdateRequest struct {
	Name     string `json:"name"`
	Purpose  string `json:"purpose"`
	IsActive bool   `json:"is_active"`
}

// ListFilter represents list filter options
type ListFilter struct {
	Purpose    string
	ActiveOnly bool
}
