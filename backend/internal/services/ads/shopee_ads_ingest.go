package ads

import (
	"context"
	"encoding/csv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ParseCSV parses Shopee ads CSV file
func (s *ShopeeAdsService) ParseCSV(ctx context.Context, data []byte, filename string) (*ParseResult, error) {
	period := ExtractPeriodFromFilename(filename)
	if period == nil {
		return nil, &ParseError{Message: "Cannot extract period from filename: " + filename}
	}

	// Skip BOM if present
	if len(data) >= 3 && data[0] == 0xef && data[1] == 0xbb && data[2] == 0xbf {
		data = data[3:]
	}

	lines := splitLines(data)
	if len(lines) < 10 {
		return nil, &ParseError{Message: "CSV file too short"}
	}

	// Find header row (starts with "Urutan" or contains "Nama Iklan")
	headerIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "Urutan,") || strings.Contains(line, "Nama Iklan,") {
			headerIdx = i
			break
		}
	}
	if headerIdx == -1 {
		return nil, &ParseError{Message: "Cannot find header row in CSV"}
	}

	// Get data starting from header
	dataLines := lines[headerIdx:]
	reader := csv.NewReader(strings.NewReader(strings.Join(dataLines, "\n")))
	reader.FieldsPerRecord = -1 // Allow variable number of fields
	reader.LazyQuotes = true    // Handle quotes flexibly
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return &ParseResult{Data: []models.ShopeeAdsProductData{}, Period: *period}, nil
	}

	// Parse header
	headers := records[0]
	colIdx := buildColumnIndex(headers, shopeeColumnMap)

	// Parse data rows
	var products []models.ShopeeAdsProductData
	for i := 1; i < len(records); i++ {
		row := records[i]
		product := parseShopeeRow(row, colIdx, period, s.tenantID)
		if product.ProductID != "" {
			products = append(products, product)
		}
	}

	return &ParseResult{Data: products, Period: *period}, nil
}

// SaveBatch saves a batch of ads data
func (s *ShopeeAdsService) SaveBatch(ctx context.Context, filename string, result *ParseResult) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		batch := models.ShopeeAdsUploadBatch{
			ID:          uuid.New().String(),
			TenantID:    s.tenantID,
			FileName:    filename,
			PeriodStart: result.Period.Start,
			PeriodEnd:   result.Period.End,
			PeriodLabel: result.Period.Label,
			TotalRows:   len(result.Data),
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(&batch).Error; err != nil {
			return err
		}

		for i := range result.Data {
			result.Data[i].UploadBatchID = batch.ID
		}

		if len(result.Data) > 0 {
			return tx.CreateInBatches(result.Data, 100).Error
		}
		return nil
	})
}
