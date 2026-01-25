package google

import (
	"context"
	"fmt"
	"strings"
)

// DataOperations handles data import/export operations
type DataOperations struct {
	sheetsService *SheetsService
}

// NewDataOperations creates a new data operations service
func NewDataOperations(sheetsService *SheetsService) *DataOperations {
	return &DataOperations{sheetsService: sheetsService}
}

// ColumnMapping represents detected column mapping
type ColumnMapping struct {
	Index      int    `json:"index"`
	Header     string `json:"header"`
	SuggestedField string `json:"suggested_field"`
}

// DetectedColumns represents auto-detected columns
type DetectedColumns struct {
	Headers  []string        `json:"headers"`
	Mappings []ColumnMapping `json:"mappings"`
	RowCount int             `json:"row_count"`
}

// AutoDetectColumns auto-detects column mappings
func (d *DataOperations) AutoDetectColumns(ctx context.Context, spreadsheetID, sheetRange string) (*DetectedColumns, error) {
	data, err := d.sheetsService.ReadRange(ctx, spreadsheetID, sheetRange)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	// First row is headers
	headers := make([]string, len(data[0]))
	for i, h := range data[0] {
		headers[i] = fmt.Sprintf("%v", h)
	}

	// Auto-detect mappings
	mappings := d.detectMappings(headers)

	return &DetectedColumns{
		Headers:  headers,
		Mappings: mappings,
		RowCount: len(data) - 1, // Exclude header row
	}, nil
}

// detectMappings detects field mappings from headers
func (d *DataOperations) detectMappings(headers []string) []ColumnMapping {
	mappings := make([]ColumnMapping, len(headers))

	knownFields := map[string][]string{
		"sku":         {"sku", "kode", "code", "item_sku", "seller_sku"},
		"name":        {"name", "nama", "product_name", "item_name", "product"},
		"price":       {"price", "harga", "original_price", "selling_price"},
		"stock":       {"stock", "stok", "qty", "quantity", "available"},
		"category":    {"category", "kategori", "cat"},
		"description": {"description", "deskripsi", "desc"},
		"weight":      {"weight", "berat", "gram"},
		"status":      {"status", "active", "aktif"},
	}

	for i, header := range headers {
		headerLower := strings.ToLower(strings.TrimSpace(header))
		suggested := ""

		for field, keywords := range knownFields {
			for _, kw := range keywords {
				if strings.Contains(headerLower, kw) {
					suggested = field
					break
				}
			}
			if suggested != "" {
				break
			}
		}

		mappings[i] = ColumnMapping{
			Index:          i,
			Header:         header,
			SuggestedField: suggested,
		}
	}

	return mappings
}

// ImportData imports data from sheet with column mapping
func (d *DataOperations) ImportData(ctx context.Context, spreadsheetID, sheetRange string, mappings map[int]string) ([]map[string]interface{}, error) {
	data, err := d.sheetsService.ReadRange(ctx, spreadsheetID, sheetRange)
	if err != nil {
		return nil, err
	}

	if len(data) < 2 {
		return nil, fmt.Errorf("insufficient data")
	}

	// Skip header row
	rows := data[1:]
	result := make([]map[string]interface{}, 0, len(rows))

	for _, row := range rows {
		record := make(map[string]interface{})
		for colIdx, fieldName := range mappings {
			if colIdx < len(row) {
				record[fieldName] = row[colIdx]
			}
		}
		result = append(result, record)
	}

	return result, nil
}

// ExportData exports data to sheet
func (d *DataOperations) ExportData(ctx context.Context, spreadsheetID, sheetRange string, headers []string, data []map[string]interface{}) error {
	// Build values array
	values := make([][]interface{}, 0, len(data)+1)

	// Add headers
	headerRow := make([]interface{}, len(headers))
	for i, h := range headers {
		headerRow[i] = h
	}
	values = append(values, headerRow)

	// Add data rows
	for _, record := range data {
		row := make([]interface{}, len(headers))
		for i, h := range headers {
			if v, ok := record[h]; ok {
				row[i] = v
			} else {
				row[i] = ""
			}
		}
		values = append(values, row)
	}

	return d.sheetsService.WriteRange(ctx, spreadsheetID, sheetRange, values)
}

// BatchUpdateCells updates multiple cells
func (d *DataOperations) BatchUpdateCells(ctx context.Context, spreadsheetID string, updates map[string]interface{}) error {
	for cellRange, value := range updates {
		values := [][]interface{}{{value}}
		if err := d.sheetsService.WriteRange(ctx, spreadsheetID, cellRange, values); err != nil {
			return fmt.Errorf("update %s: %w", cellRange, err)
		}
	}
	return nil
}
