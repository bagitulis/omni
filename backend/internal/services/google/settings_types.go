package google

import "time"

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

// SheetMetaEntry represents a single sheet's metadata
type SheetMetaEntry struct {
	Name        string `json:"name"`
	SheetID     int    `json:"sheet_id"`
	Index       int    `json:"index"`
	ColumnCount int    `json:"column_count"`
	RowCount    int    `json:"row_count"`
}

// DetailedSettingsOutput represents output for detailed settings
type DetailedSettingsOutput struct {
	WalletSpreadsheetID      string                      `json:"wallet_spreadsheet_id"`
	ShippingSpreadsheetID    string                      `json:"shipping_spreadsheet_id"`
	InventorySpreadsheetID   string                      `json:"inventory_spreadsheet_id"`
	OrderSpreadsheetID       string                      `json:"order_spreadsheet_id"`
	InventorySheetName       string                      `json:"inventory_sheet_name"`
	WalletSheetName          string                      `json:"wallet_sheet_name"`
	ShippingSheetName        string                      `json:"shipping_sheet_name"`
	OrderSheetName           string                      `json:"order_sheet_name"`
	InventorySelectedColumns []string                    `json:"inventory_selected_columns"`
	SheetsMetadata           map[string][]SheetMetaEntry `json:"sheets_metadata,omitempty"`
	LastUpdated              *time.Time                  `json:"last_updated,omitempty"`
}

// SpreadsheetLinkInput represents a spreadsheet link
type SpreadsheetLinkInput struct {
	Type          string `json:"type"`
	SpreadsheetID string `json:"spreadsheet_id"`
	URL           string `json:"url"`
	Title         string `json:"title"`
}

// LinksByType represents spreadsheet links mapped by type
type LinksByType struct {
	Inventory string
	Wallet    string
	Shipping  string
	Order     string
}

// SpreadsheetLinksResponse is the response format expected by frontend
type SpreadsheetLinksResponse struct {
	Inventory string `json:"inventory"`
	Wallet    string `json:"wallet"`
	Shipping  string `json:"shipping"`
	Order     string `json:"order"`
}
