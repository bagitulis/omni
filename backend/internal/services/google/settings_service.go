package google

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

// SettingsService handles Google Sheets settings persistence
// NOTE: Uses schema-level multi-tenancy - tenant isolation is at DB schema level
// Each tenant has their own schema, so no tenant_id column needed in queries
type SettingsService struct {
	db       *gorm.DB
	tenantID string // Used for logging/context only
}

// NewSettingsService creates a new settings service
func NewSettingsService(db *gorm.DB, tenantID string) *SettingsService {
	return &SettingsService{db: db, tenantID: tenantID}
}

// DetailedSettingsInput represents input for updating detailed settings
type DetailedSettingsInput struct {
	WalletSpreadsheetID      string
	ShippingSpreadsheetID    string
	InventorySpreadsheetID   string
	OrderSpreadsheetID       string
	InventorySheetName       string
	WalletSheetName          string
	ShippingSheetName        string
	OrderSheetName           string
	InventorySelectedColumns []string
}

// DetailedSettingsOutput represents output for detailed settings
type DetailedSettingsOutput struct {
	WalletSpreadsheetID      string   `json:"wallet_spreadsheet_id"`
	ShippingSpreadsheetID    string   `json:"shipping_spreadsheet_id"`
	InventorySpreadsheetID   string   `json:"inventory_spreadsheet_id"`
	OrderSpreadsheetID       string   `json:"order_spreadsheet_id"`
	InventorySheetName       string   `json:"inventory_sheet_name"`
	WalletSheetName          string   `json:"wallet_sheet_name"`
	ShippingSheetName        string   `json:"shipping_sheet_name"`
	OrderSheetName           string   `json:"order_sheet_name"`
	InventorySelectedColumns []string `json:"inventory_selected_columns"`
}

// SpreadsheetLinkInput represents a spreadsheet link
type SpreadsheetLinkInput struct {
	Type          string `json:"type"`
	SpreadsheetID string `json:"spreadsheet_id"`
	URL           string `json:"url"`
	Title         string `json:"title"`
}

// getOrCreateSettings gets existing settings or creates default one
func (s *SettingsService) getOrCreateSettings(ctx context.Context) (*models.GoogleSheetsSettings, error) {
	var settings models.GoogleSheetsSettings

	// Try to find first record (schema-level isolation means only one record per tenant)
	result := s.db.WithContext(ctx).First(&settings)

	if result.Error == gorm.ErrRecordNotFound {
		// Create new record with default ID
		settings = models.GoogleSheetsSettings{
			ID:        xid.New().String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.db.WithContext(ctx).Create(&settings).Error; err != nil {
			return nil, fmt.Errorf("create settings: %w", err)
		}
		return &settings, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &settings, nil
}

// GetDetailedSettings retrieves detailed settings
func (s *SettingsService) GetDetailedSettings(ctx context.Context) (*DetailedSettingsOutput, error) {
	settings, err := s.getOrCreateSettings(ctx)
	if err != nil {
		return &DetailedSettingsOutput{}, nil
	}

	// Parse inventory selected columns from JSON
	var selectedColumns []string
	if settings.InventorySelectedColumns != "" {
		json.Unmarshal([]byte(settings.InventorySelectedColumns), &selectedColumns)
	}

	return &DetailedSettingsOutput{
		WalletSpreadsheetID:      settings.WalletSpreadsheetID,
		ShippingSpreadsheetID:    settings.ShippingSpreadsheetID,
		InventorySpreadsheetID:   settings.InventorySpreadsheetID,
		OrderSpreadsheetID:       settings.OrderSpreadsheetID,
		InventorySheetName:       settings.InventorySheetName,
		WalletSheetName:          settings.WalletSheetName,
		ShippingSheetName:        settings.ShippingSheetName,
		OrderSheetName:           settings.OrderSheetName,
		InventorySelectedColumns: selectedColumns,
	}, nil
}

// UpdateDetailedSettings updates detailed settings
func (s *SettingsService) UpdateDetailedSettings(ctx context.Context, input *DetailedSettingsInput) error {
	selectedColumnsJSON, _ := json.Marshal(input.InventorySelectedColumns)

	settings, err := s.getOrCreateSettings(ctx)
	if err != nil {
		return err
	}

	// Update fields
	settings.WalletSpreadsheetID = input.WalletSpreadsheetID
	settings.ShippingSpreadsheetID = input.ShippingSpreadsheetID
	settings.InventorySpreadsheetID = input.InventorySpreadsheetID
	settings.OrderSpreadsheetID = input.OrderSpreadsheetID
	settings.InventorySheetName = input.InventorySheetName
	settings.WalletSheetName = input.WalletSheetName
	settings.ShippingSheetName = input.ShippingSheetName
	settings.OrderSheetName = input.OrderSheetName
	settings.InventorySelectedColumns = string(selectedColumnsJSON)
	settings.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Save(settings).Error
}

// LinksByType represents spreadsheet links mapped by type
type LinksByType struct {
	Inventory string
	Wallet    string
	Shipping  string
	Order     string
}

// SaveSpreadsheetLinks saves spreadsheet links by updating individual columns
func (s *SettingsService) SaveSpreadsheetLinks(ctx context.Context, links interface{}) error {
	settings, err := s.getOrCreateSettings(ctx)
	if err != nil {
		return err
	}

	// Handle links as LinksByType struct
	if linksByType, ok := links.(*LinksByType); ok {
		settings.InventorySpreadsheetID = linksByType.Inventory
		settings.WalletSpreadsheetID = linksByType.Wallet
		settings.ShippingSpreadsheetID = linksByType.Shipping
		settings.OrderSpreadsheetID = linksByType.Order
	}

	// Also store as JSON for backward compatibility
	linksJSON, err := json.Marshal(links)
	if err != nil {
		return fmt.Errorf("marshal links: %w", err)
	}
	settings.AvailableSpreadsheets = string(linksJSON)
	settings.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Save(settings).Error
}

// SpreadsheetLinksResponse is the response format expected by frontend
type SpreadsheetLinksResponse struct {
	Inventory string `json:"inventory"`
	Wallet    string `json:"wallet"`
	Shipping  string `json:"shipping"`
	Order     string `json:"order"`
}

// GetSpreadsheetLinks retrieves saved spreadsheet links in frontend-expected format
func (s *SettingsService) GetSpreadsheetLinks(ctx context.Context) (*SpreadsheetLinksResponse, error) {
	settings, err := s.getOrCreateSettings(ctx)
	if err != nil {
		return &SpreadsheetLinksResponse{}, nil
	}

	// Return links from individual columns (this is what frontend expects)
	return &SpreadsheetLinksResponse{
		Inventory: settings.InventorySpreadsheetID,
		Wallet:    settings.WalletSpreadsheetID,
		Shipping:  settings.ShippingSpreadsheetID,
		Order:     settings.OrderSpreadsheetID,
	}, nil
}
