package master_product

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

func parseImportRows(filename string, file io.Reader) ([]FileImportRow, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".csv":
		return parseCSVRows(file)
	case ".xlsx":
		return parseXLSXRows(file)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s (supported: .xlsx, .csv)", ext)
	}
}

func parseCSVRows(file io.Reader) ([]FileImportRow, error) {
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv file: %w", err)
	}
	if len(records) == 0 {
		return nil, errors.New("import file is empty")
	}
	if len(records)-1 > maxImportTemplateRows {
		return nil, fmt.Errorf("import file exceeds maximum rows (%d)", maxImportTemplateRows)
	}

	headers := normalizeHeaders(records[0])
	rows := make([]FileImportRow, 0, len(records)-1)
	for idx, record := range records[1:] {
		mapped := mapRow(headers, record)
		rows = append(rows, rowFromMappedValues(idx+2, mapped))
	}
	return rows, nil
}

func parseXLSXRows(file io.Reader) ([]FileImportRow, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read xlsx file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, errors.New("xlsx file has no sheets")
	}

	records, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read xlsx rows: %w", err)
	}
	if len(records) == 0 {
		return nil, errors.New("import file is empty")
	}
	if len(records)-1 > maxImportTemplateRows {
		return nil, fmt.Errorf("import file exceeds maximum rows (%d)", maxImportTemplateRows)
	}

	headers := normalizeHeaders(records[0])
	rows := make([]FileImportRow, 0, len(records)-1)
	for idx, record := range records[1:] {
		mapped := mapRow(headers, record)
		rows = append(rows, rowFromMappedValues(idx+2, mapped))
	}
	return rows, nil
}

func mapRow(headers []string, record []string) map[string]string {
	mapped := make(map[string]string, len(headers))
	for idx, header := range headers {
		if idx >= len(record) {
			mapped[header] = ""
			continue
		}
		mapped[header] = strings.TrimSpace(record[idx])
	}
	return mapped
}

func rowFromMappedValues(rowNumber int, mapped map[string]string) FileImportRow {
	priceValue, _ := parsePrice(mappedValue(mapped, "price"))
	stockValue, _ := parseStock(mappedValue(mapped, "stock"))

	return FileImportRow{
		RowNumber:   rowNumber,
		ItemName:    mappedValue(mapped, "item_name", "title", "name"),
		ItemSku:     mappedValue(mapped, "item_sku", "seller_sku", "sku"),
		VariantName: mappedValue(mapped, "variant_name", "variant"),
		Price:       priceValue,
		Stock:       stockValue,
		BatchKey:    mappedValue(mapped, "batch_key", "batch"),
		Description: mappedValue(mapped, "description", "desc"),
		ImageUrls:   parseImageURLs(mappedValue(mapped, "image_urls", "images", "image_url")),
	}
}

func parsePrice(raw string) (float64, error) {
	value := strings.TrimSpace(strings.ReplaceAll(raw, ",", ""))
	if value == "" {
		return 0, errors.New("empty price")
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func parseStock(raw string) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, errors.New("empty stock")
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func parseImageURLs(raw string) []string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return []string{}
	}

	separator := "|"
	if !strings.Contains(value, "|") && strings.Contains(value, ",") {
		separator = ","
	}

	parts := strings.Split(value, separator)
	urls := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		urls = append(urls, trimmed)
	}
	return urls
}

func isHTTPURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func normalizeHeaders(rawHeaders []string) []string {
	headers := make([]string, 0, len(rawHeaders))
	for _, header := range rawHeaders {
		normalized := strings.TrimSpace(strings.ToLower(header))
		normalized = strings.ReplaceAll(normalized, " ", "_")
		headers = append(headers, normalized)
	}
	return headers
}

func mappedValue(mapped map[string]string, keys ...string) string {
	for _, key := range keys {
		if value, exists := mapped[key]; exists {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
