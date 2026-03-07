package sheets

import (
	"context"
	"fmt"
)

// SheetsClient provides Google Sheets integration
type SheetsClient struct {
	credentials string
}

// NewSheetsClient creates a new Google Sheets client
func NewSheetsClient(credentials string) (*SheetsClient, error) {
	return &SheetsClient{
		credentials: credentials,
	}, nil
}

// ReadFromSheet reads data from a Google Sheet
// WARNING: This is a stub — Google Sheets API integration is not yet implemented.
func (c *SheetsClient) ReadFromSheet(ctx context.Context, spreadsheetID, sheetName string) ([][]interface{}, error) {
	return nil, fmt.Errorf("sheets ReadFromSheet is not yet implemented (stub)")
}

// WriteToSheet writes data to a Google Sheet
// WARNING: This is a stub — Google Sheets API integration is not yet implemented.
func (c *SheetsClient) WriteToSheet(ctx context.Context, spreadsheetID, sheetName string, data [][]interface{}) error {
	if spreadsheetID == "" {
		return fmt.Errorf("spreadsheet ID is required")
	}
	return fmt.Errorf("sheets WriteToSheet is not yet implemented (stub)")
}

// ClearSheet clears all data from a sheet
// WARNING: This is a stub — Google Sheets API integration is not yet implemented.
func (c *SheetsClient) ClearSheet(ctx context.Context, spreadsheetID, sheetName string) error {
	return fmt.Errorf("sheets ClearSheet is not yet implemented (stub)")
}

// AppendToSheet appends data to a sheet
// WARNING: This is a stub — Google Sheets API integration is not yet implemented.
func (c *SheetsClient) AppendToSheet(ctx context.Context, spreadsheetID, sheetName string, data [][]interface{}) error {
	return fmt.Errorf("sheets AppendToSheet is not yet implemented (stub)")
}
