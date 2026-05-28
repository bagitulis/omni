package inventory

import (
	"context"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// MaxExportRows caps the number of inventory records that can be exported
// in a single operation to prevent unbounded memory usage.
const MaxExportRows = 10000
// SheetExportService handles Google Sheets export/import operations
type SheetExportService struct {
	db           *gorm.DB
	tenantID     string
	sheetsClient SheetWriterClient
}

// NewSheetExportService creates a new SheetExportService
func NewSheetExportService(db *gorm.DB, tenantID string, sheetsClient SheetWriterClient) *SheetExportService {
	return &SheetExportService{
		db:           db,
		tenantID:     tenantID,
		sheetsClient: sheetsClient,
	}
}

// ExportToSheet exports inventory data to Google Sheets
func (s *SheetExportService) ExportToSheet(ctx context.Context, opts ExportOptions) (*ExportResult, error) {
	if s.sheetsClient == nil {
		return nil, fmt.Errorf("sheets client not configured")
	}

	var records []models.InventoryRecord
	query := s.db.WithContext(ctx).Where("tenant_id = ?", s.tenantID).Limit(MaxExportRows)
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch inventory: %w", err)
	}

	data := buildSheetData(records, opts.Fields)

	sheetName := opts.SheetName
	if sheetName == "" {
		sheetName = "Inventory"
	}

	if err := s.sheetsClient.WriteRange(ctx, opts.SpreadsheetID, sheetName, data); err != nil {
		return nil, fmt.Errorf("failed to write to sheet: %w", err)
	}

	return &ExportResult{
		SpreadsheetID: opts.SpreadsheetID,
		SheetName:     sheetName,
		RowsExported:  len(records),
		ExportedAt:    time.Now(),
	}, nil
}

// ImportFromSheet imports inventory data from Google Sheets
func (s *SheetExportService) ImportFromSheet(ctx context.Context, opts ImportOptions) (*ImportResult, error) {
	if s.sheetsClient == nil {
		return nil, fmt.Errorf("sheets client not configured")
	}

	sheetName := opts.SheetName
	if sheetName == "" {
		sheetName = "Inventory"
	}

	values, err := s.sheetsClient.ReadRange(ctx, opts.SpreadsheetID, sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read from sheet: %w", err)
	}

	result := &ImportResult{}
	startIdx := opts.StartRow - 1
	if startIdx < 0 {
		startIdx = 0
	}

	if len(values) <= startIdx {
		return result, nil
	}

	headers := parseHeaders(values[startIdx])
	if opts.HasHeader {
		startIdx++
	}

	keyColumnName := headers[0]
	if opts.KeyColumn != "" {
		keyColumnName = opts.KeyColumn
	}

	for i := startIdx; i < len(values); i++ {
		row := values[i]
		if len(row) == 0 {
			result.RowsSkipped++
			continue
		}

		record, keyValue := parseRowToJSONBRecord(headers, row, keyColumnName)
		if keyValue == "" {
			result.RowsSkipped++
			continue
		}

		if err := upsertJSONBRecord(ctx, s.db, s.tenantID, record, keyValue, keyColumnName, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %s", i+1, err.Error()))
		}
	}

	return result, nil
}

// SyncFromSheet syncs inventory data from Google Sheets
func (s *SheetExportService) SyncFromSheet(ctx context.Context, opts SyncOptions) (*SyncResult, error) {
	result := &SyncResult{SyncedAt: time.Now()}

	importOpts := ImportOptions{
		SpreadsheetID: opts.SpreadsheetID,
		SheetName:     opts.SheetName,
		KeyColumn:     opts.KeyColumn,
		StartRow:      1,
		HasHeader:     true,
	}

	importResult, err := s.ImportFromSheet(ctx, importOpts)
	if err != nil {
		return nil, err
	}

	result.Added = importResult.RowsImported
	result.Updated = importResult.RowsUpdated
	result.TotalRows = result.Added + result.Updated + importResult.RowsSkipped

	return result, nil
}

// Export exports inventory to CSV format
func (s *SheetExportService) Export(ctx context.Context, format string) ([]byte, string, string, error) {
	var records []models.InventoryRecord
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", s.tenantID).Limit(MaxExportRows).Find(&records).Error; err != nil {
		return nil, "", "", err
	}

	return exportToCSV(records)
}

// PartialSync syncs specific SKUs from sheet
func (s *SheetExportService) PartialSync(ctx context.Context, skus []string, spreadsheetID, sheetName, keyColumn string) (*SyncResult, error) {
	result := &SyncResult{SyncedAt: time.Now()}

	if s.sheetsClient == nil {
		return nil, fmt.Errorf("sheets client not configured")
	}

	if sheetName == "" {
		sheetName = "Inventory"
	}

	values, err := s.sheetsClient.ReadRange(ctx, spreadsheetID, sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read from sheet: %w", err)
	}

	if len(values) < 2 {
		return result, nil
	}

	skuSet := buildSKUSet(skus)
	headers := parseHeaders(values[0])

	keyColumnName := headers[0]
	if keyColumn != "" {
		keyColumnName = keyColumn
	}

	for i := 1; i < len(values); i++ {
		row := values[i]
		if len(row) == 0 {
			continue
		}

		data, keyValue := parseRowToJSONBRecord(headers, row, keyColumnName)
		if keyValue == "" || !skuSet[keyValue] {
			continue
		}

		syncSingleJSONBRecord(ctx, s.db, s.tenantID, data, keyValue, keyColumnName, result)
	}

	result.TotalRows = result.Added + result.Updated
	return result, nil
}
