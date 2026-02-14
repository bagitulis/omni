package google

import (
	"context"
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
	assert.Equal(t, input.Inventory, links.Inventory)
	assert.Equal(t, input.Wallet, links.Wallet)
	assert.Equal(t, input.Shipping, links.Shipping)
	assert.Equal(t, input.Order, links.Order)
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
