package inventory

import "time"

// ExportOptions contains options for exporting inventory to Google Sheets
type ExportOptions struct {
	SpreadsheetID string   `json:"spreadsheet_id"`
	SheetName     string   `json:"sheet_name"`
	Fields        []string `json:"fields"` // Which fields to export
}

// ImportOptions contains options for importing from Google Sheets
type ImportOptions struct {
	SpreadsheetID string `json:"spreadsheet_id"`
	SheetName     string `json:"sheet_name"`
	StartRow      int    `json:"start_row"`  // 1-based row number
	HasHeader     bool   `json:"has_header"` // Whether first row is header
	KeyColumn     string `json:"key_column"` // Column to use as unique key
}

// SyncOptions contains options for syncing with Google Sheets
type SyncOptions struct {
	SpreadsheetID string `json:"spreadsheet_id"`
	SheetName     string `json:"sheet_name"`
	SyncMode      string `json:"sync_mode"`  // "full", "incremental"
	KeyColumn     string `json:"key_column"` // Column to use as unique key
}

// ExportResult contains the result of an export operation
type ExportResult struct {
	SpreadsheetID string    `json:"spreadsheet_id"`
	SheetName     string    `json:"sheet_name"`
	RowsExported  int       `json:"rows_exported"`
	ExportedAt    time.Time `json:"exported_at"`
}

// ImportResult contains the result of an import operation
type ImportResult struct {
	RowsImported int      `json:"rows_imported"`
	RowsUpdated  int      `json:"rows_updated"`
	RowsSkipped  int      `json:"rows_skipped"`
	Errors       []string `json:"errors,omitempty"`
}

// SyncResult contains the result of a sync operation
type SyncResult struct {
	TotalRows int       `json:"total_rows"`
	Added     int       `json:"added"`
	Updated   int       `json:"updated"`
	Deleted   int       `json:"deleted"`
	Errors    []string  `json:"errors,omitempty"`
	SyncedAt  time.Time `json:"synced_at"`
}
