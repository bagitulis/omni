package inventory

import "context"

// SheetsClientInterface defines all Google Sheets operations needed by inventory services
// This interface is implemented by google.SheetsService
type SheetsClientInterface interface {
	// ReadRange reads data from a spreadsheet range
	ReadRange(ctx context.Context, spreadsheetID, sheetRange string) ([][]interface{}, error)
	// WriteRange writes data to a spreadsheet range
	WriteRange(ctx context.Context, spreadsheetID, sheetRange string, values [][]interface{}) error
	// ClearRange clears data from a spreadsheet range
	ClearRange(ctx context.Context, spreadsheetID, sheetRange string) error
	// GetSheetNames returns list of sheet names in a spreadsheet
	GetSheetNames(ctx context.Context, spreadsheetID string) ([]string, error)
}

// Aliases for backward compatibility
type SheetsClient = SheetsClientInterface
type SheetWriterClient = SheetsClientInterface
