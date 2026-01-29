package handlers

// SyncStatusRequest represents sync status request body
type SyncStatusRequest struct {
	SheetID string `json:"sheet_id"`
}

// ExportToSheetRequest represents export to sheet request
type ExportToSheetRequest struct {
	SheetID string                   `json:"sheet_id"`
	Data    []map[string]interface{} `json:"data"`
}

// ImportFromSheetRequest represents import from sheet request
type ImportFromSheetRequest struct {
	SheetID   string `json:"sheet_id"`
	SheetName string `json:"sheet_name"`
}
