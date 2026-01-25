// Package sheets handles Google Sheets API communication
package sheets

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Service handles Google Sheets operations with credential rolling
type Service struct {
	spreadsheetID  string
	credentials    []credentialFile
	currentIndex   int
	sheetsAPI      *sheets.Service
	lastError      string
}

type credentialFile struct {
	path string
	name string
}

// NewService creates a new Google Sheets service
func NewService(spreadsheetID string) (*Service, error) {
	s := &Service{spreadsheetID: spreadsheetID}
	if err := s.loadCredentials(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Service) getConfigPath() (string, error) {
	paths := []string{
		filepath.Join("..", "..", "backend", "config", "static", "google"),
		filepath.Join("..", "..", "..", "backend", "config", "static", "google"),
		filepath.Join("backend", "config", "static", "google"),
		filepath.Join("..", "..", "backend-node", "config", "static", "google"),
		filepath.Join("..", "..", "..", "backend-node", "config", "static", "google"),
		filepath.Join("backend-node", "config", "static", "google"),
	}

	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)
	paths = append(paths, filepath.Join(exeDir, "..", "..", "backend", "config", "static", "google"))
	paths = append(paths, filepath.Join(exeDir, "..", "..", "backend-node", "config", "static", "google"))

	for _, p := range paths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p, nil
		}
	}
	return "", fmt.Errorf("google config directory not found")
}

func (s *Service) loadCredentials() error {
	configDir, err := s.getConfigPath()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(configDir)
	if err != nil {
		return fmt.Errorf("failed to read config dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") && !strings.Contains(entry.Name(), "registry") {
			s.credentials = append(s.credentials, credentialFile{
				path: filepath.Join(configDir, entry.Name()),
				name: entry.Name(),
			})
		}
	}

	if len(s.credentials) == 0 {
		return fmt.Errorf("no credential files found in config directory")
	}

	fmt.Fprintf(os.Stderr, "Loaded %d credential file(s)\n", len(s.credentials))
	return nil
}

func (s *Service) initWithCredential(cred credentialFile) error {
	ctx := context.Background()

	data, err := os.ReadFile(cred.path)
	if err != nil {
		return fmt.Errorf("failed to read credential file: %w", err)
	}

	conf, err := google.JWTConfigFromJSON(data, sheets.SpreadsheetsScope)
	if err != nil {
		return fmt.Errorf("failed to parse credentials: %w", err)
	}

	client := conf.Client(ctx)
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create sheets service: %w", err)
	}

	s.sheetsAPI = srv
	return nil
}

func (s *Service) rollToNextCredential() bool {
	startIndex := s.currentIndex
	for {
		s.currentIndex = (s.currentIndex + 1) % len(s.credentials)
		if s.currentIndex == startIndex {
			return false
		}

		if err := s.initWithCredential(s.credentials[s.currentIndex]); err == nil {
			fmt.Fprintf(os.Stderr, "Rolled to: %s\n", s.credentials[s.currentIndex].name)
			return true
		}
	}
}

func (s *Service) ensureAPI() error {
	if s.sheetsAPI == nil {
		return s.initWithCredential(s.credentials[s.currentIndex])
	}
	return nil
}

// GetValues reads values from a range
func (s *Service) GetValues(rangeStr string) ([][]string, error) {
	maxRetries := len(s.credentials)
	attempts := 0

	for attempts < maxRetries {
		if err := s.ensureAPI(); err != nil {
			s.lastError = err.Error()
			if !s.rollToNextCredential() {
				return nil, fmt.Errorf("all credentials exhausted: %s", s.lastError)
			}
			attempts++
			continue
		}

		resp, err := s.sheetsAPI.Spreadsheets.Values.Get(s.spreadsheetID, rangeStr).Do()
		if err != nil {
			s.lastError = err.Error()
			if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "429") {
				fmt.Fprintf(os.Stderr, "Error, rolling credential...\n")
				if !s.rollToNextCredential() {
					return nil, fmt.Errorf("all credentials exhausted: %s", s.lastError)
				}
				attempts++
				continue
			}
			return nil, err
		}

		result := make([][]string, len(resp.Values))
		for i, row := range resp.Values {
			strRow := make([]string, len(row))
			for j, cell := range row {
				if str, ok := cell.(string); ok {
					strRow[j] = str
				} else {
					strRow[j] = fmt.Sprintf("%v", cell)
				}
			}
			result[i] = strRow
		}
		return result, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %s", s.lastError)
}

// UpdateValues updates values in a range
func (s *Service) UpdateValues(rangeStr string, values [][]interface{}) error {
	if err := s.ensureAPI(); err != nil {
		return err
	}

	vr := &sheets.ValueRange{Values: values}
	_, err := s.sheetsAPI.Spreadsheets.Values.Update(s.spreadsheetID, rangeStr, vr).
		ValueInputOption("RAW").Do()
	return err
}

// AppendRow appends a row to a sheet
func (s *Service) AppendRow(sheetName string, values []interface{}) error {
	if err := s.ensureAPI(); err != nil {
		return err
	}

	rangeStr := fmt.Sprintf("%s!A:Z", sheetName)
	vr := &sheets.ValueRange{Values: [][]interface{}{values}}
	_, err := s.sheetsAPI.Spreadsheets.Values.Append(s.spreadsheetID, rangeStr, vr).
		ValueInputOption("RAW").InsertDataOption("INSERT_ROWS").Do()
	return err
}

// DeleteRow deletes a row from a sheet
func (s *Service) DeleteRow(sheetName string, rowIndex int64) error {
	if err := s.ensureAPI(); err != nil {
		return err
	}

	sheetID, err := s.getSheetID(sheetName)
	if err != nil {
		return err
	}

	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteDimension: &sheets.DeleteDimensionRequest{
					Range: &sheets.DimensionRange{
						SheetId:    sheetID,
						Dimension:  "ROWS",
						StartIndex: rowIndex,
						EndIndex:   rowIndex + 1,
					},
				},
			},
		},
	}

	_, err = s.sheetsAPI.Spreadsheets.BatchUpdate(s.spreadsheetID, req).Do()
	return err
}

// ClearCell clears a cell
func (s *Service) ClearCell(rangeStr string) error {
	if err := s.ensureAPI(); err != nil {
		return err
	}

	_, err := s.sheetsAPI.Spreadsheets.Values.Clear(s.spreadsheetID, rangeStr, &sheets.ClearValuesRequest{}).Do()
	return err
}

// DeleteColumn deletes a column from a sheet
func (s *Service) DeleteColumn(sheetName string, columnIndex int64) error {
	if err := s.ensureAPI(); err != nil {
		return err
	}

	sheetID, err := s.getSheetID(sheetName)
	if err != nil {
		return err
	}

	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteDimension: &sheets.DeleteDimensionRequest{
					Range: &sheets.DimensionRange{
						SheetId:    sheetID,
						Dimension:  "COLUMNS",
						StartIndex: columnIndex,
						EndIndex:   columnIndex + 1,
					},
				},
			},
		},
	}

	_, err = s.sheetsAPI.Spreadsheets.BatchUpdate(s.spreadsheetID, req).Do()
	return err
}

func (s *Service) getSheetID(sheetName string) (int64, error) {
	resp, err := s.sheetsAPI.Spreadsheets.Get(s.spreadsheetID).Do()
	if err != nil {
		return 0, err
	}

	for _, sheet := range resp.Sheets {
		if strings.TrimSpace(sheet.Properties.Title) == strings.TrimSpace(sheetName) {
			return sheet.Properties.SheetId, nil
		}
	}

	var available []string
	for _, sheet := range resp.Sheets {
		available = append(available, sheet.Properties.Title)
	}
	return 0, fmt.Errorf("sheet %q not found. Available: %s", sheetName, strings.Join(available, ", "))
}

// ToJSON converts any value to JSON string
func ToJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

// ColToLetter converts column index to letter (0=A, 1=B, etc)
func ColToLetter(col int) string {
	letter := ""
	for col >= 0 {
		letter = string(rune('A'+(col%26))) + letter
		col = col/26 - 1
	}
	return letter
}
