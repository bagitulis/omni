package google

import (
	"context"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSettingsService_SaveAndGetSpreadsheetLinks(t *testing.T) {
	db := setupSettingsTestDB(t)

	service := NewSettingsService(db, "test-tenant")

	input := &LinksByType{
		Inventory: "https://docs.google.com/spreadsheets/d/inventory",
		Wallet:    "https://docs.google.com/spreadsheets/d/wallet",
		Shipping:  "https://docs.google.com/spreadsheets/d/shipping",
		Order:     "https://docs.google.com/spreadsheets/d/order",
	}

	err := service.SaveSpreadsheetLinks(context.Background(), input)
	assert.NoError(t, err)

	links, err := service.GetSpreadsheetLinks(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "https://docs.google.com/spreadsheets/d/inventory/edit", links.Inventory)
	assert.Equal(t, "https://docs.google.com/spreadsheets/d/wallet/edit", links.Wallet)
	assert.Equal(t, "https://docs.google.com/spreadsheets/d/shipping/edit", links.Shipping)
	assert.Equal(t, "https://docs.google.com/spreadsheets/d/order/edit", links.Order)

	var stored models.GoogleSheetsSettings
	err = db.First(&stored).Error
	assert.NoError(t, err)
	assert.Equal(t, "inventory", stored.InventorySpreadsheetID)
	assert.Equal(t, "wallet", stored.WalletSpreadsheetID)
	assert.Equal(t, "shipping", stored.ShippingSpreadsheetID)
	assert.Equal(t, "order", stored.OrderSpreadsheetID)
}

func TestSettingsService_UpdateDetailedSettings_PersistsSheetNames(t *testing.T) {
	db := setupSettingsTestDB(t)

	service := NewSettingsService(db, "test-tenant")

	input := &DetailedSettingsInput{
		InventorySheetName: "InventorySheet",
		WalletSheetName:    "WalletSheet",
		ShippingSheetName:  "ShippingSheet",
		OrderSheetName:     "OrderSheet",
	}

	err := service.UpdateDetailedSettings(context.Background(), input)
	assert.NoError(t, err)

	settings, err := service.GetDetailedSettings(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, input.InventorySheetName, settings.InventorySheetName)
	assert.Equal(t, input.WalletSheetName, settings.WalletSheetName)
	assert.Equal(t, input.ShippingSheetName, settings.ShippingSheetName)
	assert.Equal(t, input.OrderSheetName, settings.OrderSheetName)
}

func TestSettingsService_UpdateDetailedSettings_IgnoresMissingColumns(t *testing.T) {
	db := setupLegacySettingsTestDB(t)

	service := NewSettingsService(db, "test-tenant")

	input := &DetailedSettingsInput{
		InventorySheetName: "InventorySheet",
		WalletSheetName:    "WalletSheet",
		ShippingSheetName:  "ShippingSheet",
		OrderSheetName:     "OrderSheet",
	}

	err := service.UpdateDetailedSettings(context.Background(), input)
	assert.NoError(t, err)

	settings, err := service.GetDetailedSettings(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, input.InventorySheetName, settings.InventorySheetName)
	assert.Empty(t, settings.WalletSheetName)
	assert.Empty(t, settings.ShippingSheetName)
	assert.Empty(t, settings.OrderSheetName)
}

func setupSettingsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	if err := db.AutoMigrate(&models.GoogleSheetsSettings{}); err != nil {
		t.Fatalf("failed to migrate settings model: %v", err)
	}

	return db
}

func setupLegacySettingsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	createTableSQL := `
		CREATE TABLE google_sheets_settings (
			id TEXT PRIMARY KEY,
			spreadsheet_id TEXT,
			selected_sheet TEXT,
			manual_mode BOOLEAN DEFAULT false,
			wallet_spreadsheet_id TEXT,
			shipping_spreadsheet_id TEXT,
			inventory_spreadsheet_id TEXT,
			inventory_sheet_name TEXT,
			inventory_selected_columns TEXT,
			inventory_available_worksheets TEXT,
			wallet_available_worksheets TEXT,
			shipping_available_worksheets TEXT,
			order_available_worksheets TEXT,
			order_spreadsheet_id TEXT,
			available_spreadsheets TEXT,
			available_worksheets TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`

	if err := db.Exec(createTableSQL).Error; err != nil {
		t.Fatalf("failed to create legacy google_sheets_settings table: %v", err)
	}

	if err := db.Exec(fmt.Sprintf("INSERT INTO %s (id, created_at, updated_at) VALUES (?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", models.GoogleSheetsSettings{}.TableName()), "legacy-settings").Error; err != nil {
		t.Fatalf("failed to seed legacy google_sheets_settings row: %v", err)
	}

	return db
}
