package models

import "time"

// GoogleSheetsSettings stores Google Sheets configuration
// NOTE: This table uses schema-level multi-tenancy, so NO tenant_id column needed
// Each tenant has their own schema (e.g., tenant_yumna_bertigamart)
// Matches PostgreSQL structure from setup_schemas_v3.sql
type GoogleSheetsSettings struct {
	ID                           string    `gorm:"column:id;primaryKey;type:varchar(255)" json:"id"`
	SpreadsheetID                string    `gorm:"column:spreadsheet_id;type:varchar(500)" json:"spreadsheet_id"`
	SelectedSheet                string    `gorm:"column:selected_sheet;type:varchar(500)" json:"selected_sheet"`
	ManualMode                   bool      `gorm:"column:manual_mode;default:false" json:"manual_mode"`
	WalletSpreadsheetID          string    `gorm:"column:wallet_spreadsheet_id;type:varchar(500)" json:"wallet_spreadsheet_id"`
	ShippingSpreadsheetID        string    `gorm:"column:shipping_spreadsheet_id;type:varchar(500)" json:"shipping_spreadsheet_id"`
	InventorySpreadsheetID       string    `gorm:"column:inventory_spreadsheet_id;type:varchar(500)" json:"inventory_spreadsheet_id"`
	InventorySheetName           string    `gorm:"column:inventory_sheet_name;type:varchar(500)" json:"inventory_sheet_name"`
	WalletSheetName              string    `gorm:"column:wallet_sheet_name;type:varchar(500)" json:"wallet_sheet_name"`
	ShippingSheetName            string    `gorm:"column:shipping_sheet_name;type:varchar(500)" json:"shipping_sheet_name"`
	OrderSheetName               string    `gorm:"column:order_sheet_name;type:varchar(500)" json:"order_sheet_name"`
	InventorySelectedColumns     string    `gorm:"column:inventory_selected_columns;type:text" json:"inventory_selected_columns"`
	InventoryAvailableWorksheets string    `gorm:"column:inventory_available_worksheets;type:text" json:"inventory_available_worksheets"`
	WalletAvailableWorksheets    string    `gorm:"column:wallet_available_worksheets;type:text" json:"wallet_available_worksheets"`
	ShippingAvailableWorksheets  string    `gorm:"column:shipping_available_worksheets;type:text" json:"shipping_available_worksheets"`
	OrderAvailableWorksheets     string    `gorm:"column:order_available_worksheets;type:text" json:"order_available_worksheets"`
	OrderSpreadsheetID           string    `gorm:"column:order_spreadsheet_id;type:varchar(500)" json:"order_spreadsheet_id"`
	AvailableSpreadsheets        string    `gorm:"column:available_spreadsheets;type:text" json:"available_spreadsheets"`
	AvailableWorksheets          string    `gorm:"column:available_worksheets;type:text" json:"available_worksheets"`
	CreatedAt                    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (GoogleSheetsSettings) TableName() string { return GetTableName("GoogleSheetsSettings") }

// FilterPreference stores user filter preferences per platform
// Schema matches PostgreSQL/Prisma: id, tenant_id, platform, tab, visible_columns, column_filters, search_query, locked_columns
type FilterPreference struct {
	ID             string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	TenantID       string    `gorm:"index;not null" json:"tenant_id"`
	Platform       string    `gorm:"index;not null" json:"platform"`           // shopee, lazada, tiktok, inventory
	Tab            string    `gorm:"index;not null;default:master" json:"tab"` // product, order, master
	VisibleColumns string    `gorm:"type:text" json:"visible_columns"`         // JSON array of visible column IDs
	ColumnFilters  string    `gorm:"type:text" json:"column_filters"`          // JSON object of filter settings
	SearchQuery    string    `gorm:"type:text;default:''" json:"search_query"`
	LockedColumns  string    `gorm:"type:text;default:'[]'" json:"locked_columns"` // JSON array of locked column IDs
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (FilterPreference) TableName() string { return GetTableName("FilterPreference") }

// Spreadsheet stores registered spreadsheets
type Spreadsheet struct {
	ID             string     `gorm:"primaryKey" json:"id"`
	TenantID       string     `gorm:"index;not null" json:"tenant_id"`
	SpreadsheetID  string     `gorm:"column:spreadsheet_id;index;not null" json:"spreadsheet_id"`
	SpreadsheetURL string     `gorm:"column:spreadsheet_url" json:"spreadsheet_url"`
	Name           string     `gorm:"column:spreadsheet_name" json:"spreadsheet_name"`
	SheetName      string     `gorm:"column:sheet_name" json:"sheet_name"`
	Purpose        string     `gorm:"default:other" json:"purpose"` // inventory, orders, analytics
	Sheets         string     `gorm:"type:text" json:"sheets"`      // JSON encoded sheet metadata
	RegisteredBy   *string    `gorm:"column:registered_by" json:"registered_by,omitempty"`
	LockedBy       *string    `gorm:"column:locked_by" json:"locked_by,omitempty"`
	LockedAt       *time.Time `gorm:"column:locked_at" json:"locked_at,omitempty"`
	AutoSync       bool       `gorm:"column:auto_sync;default:false" json:"auto_sync"`
	SyncInterval   int        `gorm:"column:sync_interval;default:300" json:"sync_interval"` // seconds
	LastUsedAt     *time.Time `gorm:"column:last_used_at" json:"last_used_at,omitempty"`
	LastSyncedAt   *time.Time `gorm:"column:last_synced_at" json:"last_synced_at,omitempty"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (Spreadsheet) TableName() string { return GetTableName("Spreadsheet") }

// RouteConfig stores route configuration per tenant
type RouteConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TenantID  string    `gorm:"index;not null" json:"tenant_id"`
	RouteName string    `gorm:"index;not null" json:"route_name"`
	IsEnabled bool      `gorm:"default:true" json:"is_enabled"`
	Config    string    `json:"config"` // JSON encoded route config
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (RouteConfig) TableName() string { return GetTableName("RouteConfig") }

// GeneralSettings stores general application settings per tenant
type GeneralSettings struct {
	ID                   string    `gorm:"column:id;primaryKey;type:varchar(255)" json:"id"`
	TenantID             string    `gorm:"column:tenant_id;index" json:"tenant_id"`
	Language             string    `gorm:"column:language;type:varchar(10);default:en" json:"language"`
	Timezone             string    `gorm:"column:timezone;type:varchar(100);default:Asia/Jakarta" json:"timezone"`
	NotificationsEmail   bool      `gorm:"column:notifications_email;default:true" json:"notifications_email"`
	NotificationsBrowser bool      `gorm:"column:notifications_browser;default:true" json:"notifications_browser"`
	AutoSync             bool      `gorm:"column:auto_sync;default:true" json:"auto_sync"`
	SyncInterval         string    `gorm:"column:sync_interval;type:varchar(20);default:30" json:"sync_interval"`
	CreatedAt            time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (GeneralSettings) TableName() string { return GetTableName("GeneralSettings") }
