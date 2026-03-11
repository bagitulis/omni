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
		return nil, fmt.Errorf("get settings: %w", err)
	}

	// Parse inventory selected columns from JSON
	var selectedColumns []string
	if settings.InventorySelectedColumns != "" {
		if err := json.Unmarshal([]byte(settings.InventorySelectedColumns), &selectedColumns); err != nil {
			return nil, fmt.Errorf("parse inventory_selected_columns: %w", err)
		}
	}

	// Parse available worksheets from stored JSON columns
	sheetsMetadata := make(map[string][]SheetMetaEntry)
	parseWorksheets := func(key, raw string) {
		if raw == "" {
			return
		}
		var entries []SheetMetaEntry
		if err := json.Unmarshal([]byte(raw), &entries); err == nil && len(entries) > 0 {
			sheetsMetadata[key] = entries
		}
	}
	parseWorksheets("inventory", settings.InventoryAvailableWorksheets)
	parseWorksheets("wallet", settings.WalletAvailableWorksheets)
	parseWorksheets("shipping", settings.ShippingAvailableWorksheets)
	parseWorksheets("order", settings.OrderAvailableWorksheets)

	var lastUpdated *time.Time
	if !settings.UpdatedAt.IsZero() {
		lastUpdated = &settings.UpdatedAt
	}

	output := &DetailedSettingsOutput{
		WalletSpreadsheetID:      settings.WalletSpreadsheetID,
		ShippingSpreadsheetID:    settings.ShippingSpreadsheetID,
		InventorySpreadsheetID:   settings.InventorySpreadsheetID,
		OrderSpreadsheetID:       settings.OrderSpreadsheetID,
		InventorySheetName:       settings.InventorySheetName,
		WalletSheetName:          settings.WalletSheetName,
		ShippingSheetName:        settings.ShippingSheetName,
		OrderSheetName:           settings.OrderSheetName,
		InventorySelectedColumns: selectedColumns,
		LastUpdated:              lastUpdated,
	}

	if len(sheetsMetadata) > 0 {
		output.SheetsMetadata = sheetsMetadata
	}

	return output, nil
}

// UpdateDetailedSettings updates detailed settings
func (s *SettingsService) UpdateDetailedSettings(ctx context.Context, input *DetailedSettingsInput) error {
	settings, err := s.getOrCreateSettings(ctx)
	if err != nil {
		return err
	}

	// Build map of fields to update (only non-empty values)
	updates := make(map[string]interface{})

	if input.WalletSpreadsheetID != "" {
		updates["wallet_spreadsheet_id"] = input.WalletSpreadsheetID
	}
	if input.ShippingSpreadsheetID != "" {
		updates["shipping_spreadsheet_id"] = input.ShippingSpreadsheetID
	}
	if input.InventorySpreadsheetID != "" {
		updates["inventory_spreadsheet_id"] = input.InventorySpreadsheetID
	}
	if input.OrderSpreadsheetID != "" {
		updates["order_spreadsheet_id"] = input.OrderSpreadsheetID
	}
	if input.InventorySheetName != "" {
		updates["inventory_sheet_name"] = input.InventorySheetName
	}
	if input.WalletSheetName != "" {
		updates["wallet_sheet_name"] = input.WalletSheetName
	}
	if input.ShippingSheetName != "" {
		updates["shipping_sheet_name"] = input.ShippingSheetName
	}
	if input.OrderSheetName != "" {
		updates["order_sheet_name"] = input.OrderSheetName
	}
	if input.InventorySelectedColumns != nil {
		selectedColumnsJSON, err := json.Marshal(input.InventorySelectedColumns)
		if err != nil {
			return fmt.Errorf("marshal inventory_selected_columns: %w", err)
		}
		updates["inventory_selected_columns"] = string(selectedColumnsJSON)
	}

	updates["updated_at"] = time.Now()

	filteredUpdates, err := s.filterUpdatesByExistingColumns(ctx, settings, updates)
	if err != nil {
		return fmt.Errorf("inspect google_sheets_settings columns: %w", err)
	}

	if len(filteredUpdates) == 0 {
		return nil
	}

	return s.db.WithContext(ctx).Model(settings).Updates(filteredUpdates).Error
}

func (s *SettingsService) filterUpdatesByExistingColumns(ctx context.Context, model interface{}, updates map[string]interface{}) (map[string]interface{}, error) {
	columnTypes, err := s.db.WithContext(ctx).Migrator().ColumnTypes(model)
	if err != nil {
		return nil, err
	}

	existingColumns := make(map[string]struct{}, len(columnTypes))
	for _, columnType := range columnTypes {
		existingColumns[columnType.Name()] = struct{}{}
	}

	filteredUpdates := make(map[string]interface{}, len(updates))
	for column, value := range updates {
		if _, exists := existingColumns[column]; !exists {
			continue
		}
		filteredUpdates[column] = value
	}

	if len(filteredUpdates) == 0 {
		return filteredUpdates, nil
	}

	if _, hasUpdatedAt := filteredUpdates["updated_at"]; !hasUpdatedAt {
		if _, exists := existingColumns["updated_at"]; exists {
			filteredUpdates["updated_at"] = time.Now()
		}
	}

	return orderedMap(filteredUpdates), nil
}



// SaveWorksheetMetadata persists discovered worksheet metadata for a given type
// so the Sheet Metadata card can display it without re-validating.
func (s *SettingsService) SaveWorksheetMetadata(ctx context.Context, linkType string, sheets []SheetInfo) error {
	settings, err := s.getOrCreateSettings(ctx)
	if err != nil {
		return err
	}

	// Convert SheetInfo to SheetMetaEntry for storage
	entries := make([]SheetMetaEntry, len(sheets))
	for i, sheet := range sheets {
		entries[i] = SheetMetaEntry{
			Name:        sheet.Title,
			SheetID:     int(sheet.ID),
			Index:       sheet.Index,
			ColumnCount: sheet.Cols,
			RowCount:    sheet.Rows,
		}
	}

	metaJSON, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("marshal worksheet metadata: %w", err)
	}

	// Map type to column name
	columnMap := map[string]string{
		"inventory": "inventory_available_worksheets",
		"wallet":    "wallet_available_worksheets",
		"shipping":  "shipping_available_worksheets",
		"order":     "order_available_worksheets",
	}

	column, ok := columnMap[linkType]
	if !ok {
		return fmt.Errorf("unknown link type: %s", linkType)
	}

	updates := map[string]interface{}{
		column:       string(metaJSON),
		"updated_at": time.Now(),
	}

	filteredUpdates, err := s.filterUpdatesByExistingColumns(ctx, settings, updates)
	if err != nil {
		return err
	}

	if len(filteredUpdates) == 0 {
		return nil
	}

	return s.db.WithContext(ctx).Model(settings).Updates(filteredUpdates).Error
}



// SaveSpreadsheetLinks saves spreadsheet links by updating individual columns
// Extracts spreadsheet IDs from URLs before saving
func (s *SettingsService) SaveSpreadsheetLinks(ctx context.Context, links interface{}) error {
	settings, err := s.getOrCreateSettings(ctx)
	if err != nil {
		return err
	}

	// Build map of fields to update (only non-empty values)
	updates := make(map[string]interface{})

	// Handle links as LinksByType struct
	if linksByType, ok := links.(*LinksByType); ok {
		if linksByType.Inventory != "" {
			updates["inventory_spreadsheet_id"] = extractSpreadsheetID(linksByType.Inventory)
		}
		if linksByType.Wallet != "" {
			updates["wallet_spreadsheet_id"] = extractSpreadsheetID(linksByType.Wallet)
		}
		if linksByType.Shipping != "" {
			updates["shipping_spreadsheet_id"] = extractSpreadsheetID(linksByType.Shipping)
		}
		if linksByType.Order != "" {
			updates["order_spreadsheet_id"] = extractSpreadsheetID(linksByType.Order)
		}
	}

	// Also store as JSON for backward compatibility
	linksJSON, err := json.Marshal(links)
	if err != nil {
		return fmt.Errorf("marshal links: %w", err)
	}
	updates["available_spreadsheets"] = string(linksJSON)
	updates["updated_at"] = time.Now()

	return s.db.WithContext(ctx).Model(settings).Updates(updates).Error
}



// GetSpreadsheetLinks retrieves saved spreadsheet links in frontend-expected format
func (s *SettingsService) GetSpreadsheetLinks(ctx context.Context) (*SpreadsheetLinksResponse, error) {
	settings, err := s.getOrCreateSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}

	// Reconstruct full URLs from saved IDs (frontend expects full URLs)
	inventoryURL := reconstructURL(settings.InventorySpreadsheetID)
	walletURL := reconstructURL(settings.WalletSpreadsheetID)
	shippingURL := reconstructURL(settings.ShippingSpreadsheetID)
	orderURL := reconstructURL(settings.OrderSpreadsheetID)

	return &SpreadsheetLinksResponse{
		Inventory: inventoryURL,
		Wallet:    walletURL,
		Shipping:  shippingURL,
		Order:     orderURL,
	}, nil
}
