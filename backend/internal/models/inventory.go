package models

import "time"

// InventorySettings stores inventory configuration per tenant
type InventorySettings struct {
	ID                string     `gorm:"primaryKey;column:id;size:255" json:"id"`
	TenantID          string     `gorm:"column:tenant_id;uniqueIndex;not null" json:"tenant_id"`
	SpreadsheetID     string     `gorm:"column:spreadsheet_id;size:500" json:"spreadsheet_id"`
	SheetName         string     `gorm:"column:sheet_name;size:500" json:"sheet_name"`
	SelectedColumns   string     `gorm:"column:selected_columns;type:text" json:"selected_columns"` // JSON array
	AllColumns        string     `gorm:"column:all_columns;type:text" json:"all_columns"`           // JSON array
	HeaderRow         int        `gorm:"column:header_row;default:1" json:"header_row"`
	DataStartRow      int        `gorm:"column:data_start_row;default:2" json:"data_start_row"`
	KeyColumn         string     `gorm:"column:key_column;size:255" json:"key_column"`
	AutoSync          bool       `gorm:"column:auto_sync;default:false" json:"auto_sync"`
	SyncIntervalSec   int        `gorm:"column:sync_interval_seconds;default:300" json:"sync_interval_seconds"`
	LastSyncTimestamp *time.Time `gorm:"column:last_sync_timestamp" json:"last_sync_timestamp"`
	LastHeadersHash   string     `gorm:"column:last_headers_hash;size:100" json:"last_headers_hash"`
	LastSyncStatus    string     `gorm:"column:last_sync_status;size:50" json:"last_sync_status"`
	CreatedAt         time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (InventorySettings) TableName() string { return GetTableName("InventorySettings") }

// InventoryRecord represents an inventory item with dynamic columns from Google Sheets
// Data is stored as JSONB matching the spreadsheet columns
type InventoryRecord struct {
	ID            string    `gorm:"primaryKey;column:id;size:255" json:"id"`
	TenantID      string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	Data          string    `gorm:"column:data;type:jsonb;not null" json:"data"`                     // JSONB: dynamic columns from sheet
	KeyValue      string    `gorm:"column:key_value;size:500;not null" json:"key_value"`             // Value of key column (e.g., SKU value)
	KeyColumnName string    `gorm:"column:key_column_name;size:255;not null" json:"key_column_name"` // Name of key column (e.g., "SKU")
	SyncStatus    string    `gorm:"column:sync_status;size:50" json:"sync_status"`                   // Sync status: "synced", "not_synced", "error", or empty
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (InventoryRecord) TableName() string { return GetTableName("InventoryRecord") }

// InventorySyncHistory tracks sync operations
// Note: Legacy Prisma used cuid() strings; keep varchar IDs for compatibility
type InventorySyncHistory struct {
	ID               string    `gorm:"primaryKey;column:id;size:255" json:"id"`
	TenantID         string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	Status           string    `gorm:"column:status" json:"status"` // SUCCESS, ERROR, PARTIAL
	TotalRecords     int       `gorm:"column:total_records" json:"total_records"`
	NewRecords       int       `gorm:"column:new_records" json:"new_records"`
	UpdatedRecords   int       `gorm:"column:updated_records" json:"updated_records"`
	UnchangedRecords int       `gorm:"column:unchanged_records" json:"unchanged_records"`
	FailedRecords    int       `gorm:"column:failed_records" json:"failed_records"`
	Duration         int       `gorm:"column:duration" json:"duration_ms"`
	ErrorMessage     string    `gorm:"column:error_message" json:"error_message,omitempty"`
	HeadersChanged   bool      `gorm:"column:headers_changed" json:"headers_changed"`
	SyncedAt         time.Time `gorm:"column:synced_at" json:"synced_at"`
	CreatedAt        time.Time `gorm:"column:created_at" json:"created_at"`
}

func (InventorySyncHistory) TableName() string { return GetTableName("InventorySyncHistory") }

// SheetSnapshot stores a snapshot of sheet data
// AGENTS.MD: JSON tags MUST be snake_case
type SheetSnapshot struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      string    `gorm:"index;not null" json:"tenant_id"`
	SpreadsheetID string    `json:"spreadsheet_id"`
	SheetName     string    `json:"sheet_name"`
	Headers       string    `json:"headers"` // JSON array
	RowCount      int       `json:"row_count"`
	DataHash      string    `json:"data_hash"`
	CapturedAt    time.Time `json:"captured_at"`
}

func (SheetSnapshot) TableName() string { return GetTableName("SheetSnapshot") }

// InventorySkuPlatformStatus tracks SKU status on each platform
// AGENTS.MD: JSON tags MUST be snake_case
type InventorySkuPlatformStatus struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	TenantID          string    `gorm:"index;not null" json:"tenant_id"`
	SKU               string    `gorm:"index;not null" json:"sku"`
	Platform          string    `gorm:"index;not null" json:"platform"` // shopee, lazada, tiktok
	PlatformProductID string    `json:"platform_product_id"`
	PlatformItemID    string    `json:"platform_item_id"`
	PlatformSKU       string    `json:"platform_sku"`
	Status            string    `json:"status"` // active, inactive, not_found
	Stock             int       `json:"stock"`
	Price             float64   `json:"price"`
	LastCheckedAt     time.Time `json:"last_checked_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (InventorySkuPlatformStatus) TableName() string {
	return GetTableName("InventorySkuPlatformStatus")
}
