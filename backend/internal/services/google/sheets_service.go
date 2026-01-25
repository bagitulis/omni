package google

import (
	"context"
	"fmt"

	"google.golang.org/api/sheets/v4"
)

// SheetsService handles Google Sheets operations
type SheetsService struct {
	authService *AuthService
	tenantID    string
}

// NewSheetsService creates a new sheets service
func NewSheetsService(authService *AuthService, tenantID string) *SheetsService {
	return &SheetsService{
		authService: authService,
		tenantID:    tenantID,
	}
}

// SpreadsheetInfo contains spreadsheet metadata
type SpreadsheetInfo struct {
	ID         string       `json:"spreadsheet_id"`
	Title      string       `json:"title"`
	URL        string       `json:"url"`
	Sheets     []SheetInfo  `json:"sheets"`
}

// SheetInfo contains worksheet metadata
type SheetInfo struct {
	ID    int64  `json:"sheet_id"`
	Title string `json:"title"`
	Index int    `json:"index"`
	Rows  int    `json:"row_count"`
	Cols  int    `json:"column_count"`
}

// GetSpreadsheetInfo retrieves spreadsheet metadata
func (s *SheetsService) GetSpreadsheetInfo(ctx context.Context, spreadsheetID string) (*SpreadsheetInfo, error) {
	client, err := s.authService.GetClient(ctx, s.tenantID)
	if err != nil {
		return nil, err
	}

	resp, err := client.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get spreadsheet: %w", err)
	}

	info := &SpreadsheetInfo{
		ID:     resp.SpreadsheetId,
		Title:  resp.Properties.Title,
		URL:    resp.SpreadsheetUrl,
		Sheets: make([]SheetInfo, len(resp.Sheets)),
	}

	for i, sheet := range resp.Sheets {
		info.Sheets[i] = SheetInfo{
			ID:    sheet.Properties.SheetId,
			Title: sheet.Properties.Title,
			Index: int(sheet.Properties.Index),
			Rows:  int(sheet.Properties.GridProperties.RowCount),
			Cols:  int(sheet.Properties.GridProperties.ColumnCount),
		}
	}

	return info, nil
}

// ReadRange reads data from a range
func (s *SheetsService) ReadRange(ctx context.Context, spreadsheetID, sheetRange string) ([][]interface{}, error) {
	client, err := s.authService.GetClient(ctx, s.tenantID)
	if err != nil {
		return nil, err
	}

	resp, err := client.Spreadsheets.Values.Get(spreadsheetID, sheetRange).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("read range: %w", err)
	}

	return resp.Values, nil
}

// WriteRange writes data to a range
func (s *SheetsService) WriteRange(ctx context.Context, spreadsheetID, sheetRange string, values [][]interface{}) error {
	client, err := s.authService.GetClient(ctx, s.tenantID)
	if err != nil {
		return err
	}

	vr := &sheets.ValueRange{Values: values}
	_, err = client.Spreadsheets.Values.Update(spreadsheetID, sheetRange, vr).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("write range: %w", err)
	}

	return nil
}

// AppendRows appends rows to a sheet
func (s *SheetsService) AppendRows(ctx context.Context, spreadsheetID, sheetRange string, values [][]interface{}) error {
	client, err := s.authService.GetClient(ctx, s.tenantID)
	if err != nil {
		return err
	}

	vr := &sheets.ValueRange{Values: values}
	_, err = client.Spreadsheets.Values.Append(spreadsheetID, sheetRange, vr).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("append rows: %w", err)
	}

	return nil
}

// ClearRange clears data from a range
func (s *SheetsService) ClearRange(ctx context.Context, spreadsheetID, sheetRange string) error {
	client, err := s.authService.GetClient(ctx, s.tenantID)
	if err != nil {
		return err
	}

	_, err = client.Spreadsheets.Values.Clear(spreadsheetID, sheetRange, &sheets.ClearValuesRequest{}).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("clear range: %w", err)
	}

	return nil
}

// CreateSheet creates a new worksheet
func (s *SheetsService) CreateSheet(ctx context.Context, spreadsheetID, title string) (*SheetInfo, error) {
	client, err := s.authService.GetClient(ctx, s.tenantID)
	if err != nil {
		return nil, err
	}

	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{Title: title},
			},
		}},
	}

	resp, err := client.Spreadsheets.BatchUpdate(spreadsheetID, req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create sheet: %w", err)
	}

	if len(resp.Replies) > 0 && resp.Replies[0].AddSheet != nil {
		props := resp.Replies[0].AddSheet.Properties
		return &SheetInfo{
			ID:    props.SheetId,
			Title: props.Title,
			Index: int(props.Index),
		}, nil
	}

	return nil, fmt.Errorf("unexpected response")
}

// ListSpreadsheets lists accessible spreadsheets via Drive API
func (s *SheetsService) ListSpreadsheets(ctx context.Context) ([]*SpreadsheetInfo, error) {
	driveClient, err := s.authService.GetDriveClient(ctx, s.tenantID)
	if err != nil {
		return nil, fmt.Errorf("get drive client: %w", err)
	}

	// Query for spreadsheets
	files, err := driveClient.Files.List().
		Q("mimeType='application/vnd.google-apps.spreadsheet'").
		Fields("files(id,name,webViewLink)").
		PageSize(100).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}

	spreadsheets := make([]*SpreadsheetInfo, 0, len(files.Files))
	for _, f := range files.Files {
		spreadsheets = append(spreadsheets, &SpreadsheetInfo{
			ID:    f.Id,
			Title: f.Name,
			URL:   f.WebViewLink,
		})
	}

	return spreadsheets, nil
}

// CreateSpreadsheet creates a new spreadsheet and returns its info
func (s *SheetsService) CreateSpreadsheet(ctx context.Context, title string) (*SpreadsheetInfo, error) {
	client, err := s.authService.GetClient(ctx, s.tenantID)
	if err != nil {
		return nil, err
	}

	spreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{Title: title},
	}

	created, err := client.Spreadsheets.Create(spreadsheet).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create spreadsheet: %w", err)
	}

	return &SpreadsheetInfo{
		ID:    created.SpreadsheetId,
		Title: created.Properties.Title,
		URL:   created.SpreadsheetUrl,
	}, nil
}

// TestConnection tests if the service can access a spreadsheet
func (s *SheetsService) TestConnection(ctx context.Context, spreadsheetID string) error {
	client, err := s.authService.GetClient(ctx, s.tenantID)
	if err != nil {
		return err
	}

	if spreadsheetID == "" {
		// Just test if we can create a client
		return nil
	}

	// Try to get spreadsheet info
	_, err = client.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	return err
}

// GetSheetNames returns list of sheet names in a spreadsheet
func (s *SheetsService) GetSheetNames(ctx context.Context, spreadsheetID string) ([]string, error) {
	info, err := s.GetSpreadsheetInfo(ctx, spreadsheetID)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(info.Sheets))
	for i, sheet := range info.Sheets {
		names[i] = sheet.Title
	}
	return names, nil
}
