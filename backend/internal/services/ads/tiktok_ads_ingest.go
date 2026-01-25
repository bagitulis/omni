package ads

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// ParseExcel parses TikTok ads Excel file
func (s *TiktokAdsService) ParseExcel(ctx context.Context, data []byte, filename string) (*TiktokParseResult, error) {
	period := ExtractTiktokPeriod(filename)
	if period == nil {
		return nil, &ParseError{Message: "Cannot extract period from filename: " + filename}
	}

	f, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	if len(rows) < 2 {
		return &TiktokParseResult{Data: []models.TiktokAdsCreativeData{}, Period: *period}, nil
	}

	// Parse header
	headers := rows[0]
	colIdx := buildColumnIndex(headers, tiktokColumnMap)

	// Parse data rows
	var creatives []models.TiktokAdsCreativeData
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		creative := parseTiktokRow(row, colIdx, period, s.tenantID)
		if creative.CampaignID != "" || creative.ProductID != "" {
			creatives = append(creatives, creative)
		}
	}

	return &TiktokParseResult{Data: creatives, Period: *period}, nil
}

// SaveBatch saves a batch of TikTok ads data
func (s *TiktokAdsService) SaveBatch(ctx context.Context, filename string, result *TiktokParseResult) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		batch := models.TiktokAdsUploadBatch{
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
			return fmt.Errorf("create batch: %w", err)
		}

		for i := range result.Data {
			result.Data[i].UploadBatchID = batch.ID
		}

		if len(result.Data) > 0 {
			if err := tx.CreateInBatches(result.Data, 100).Error; err != nil {
				return fmt.Errorf("create data (batch size %d): %w", len(result.Data), err)
			}
		}

		// Skip product summary for now - can be generated separately
		return nil
	})
}

// generateProductSummary generates product-level summary from creative data
func (s *TiktokAdsService) generateProductSummary(ctx context.Context, tx *gorm.DB, periodLabel string) error {
	var creatives []models.TiktokAdsCreativeData
	if err := tx.Where("tenant_id = ? AND period_label = ?", s.tenantID, periodLabel).Find(&creatives).Error; err != nil {
		return err
	}

	// Aggregate by product
	productMap := make(map[string]*models.TiktokAdsProductSummary)
	for _, c := range creatives {
		if c.ProductID == "" {
			continue
		}
		if _, ok := productMap[c.ProductID]; !ok {
			productMap[c.ProductID] = &models.TiktokAdsProductSummary{
				TenantID:    s.tenantID,
				ProductID:   c.ProductID,
				ProductName: c.VideoTitle, // Use VideoTitle as product name fallback
				PeriodLabel: periodLabel,
			}
		}
		p := productMap[c.ProductID]
		p.TotalCost += c.Cost
		p.TotalRevenue += c.GrossRevenue
		p.TotalConv += c.OrdersSKU
		p.TotalCreatives++
	}

	// Calculate averages and save
	for _, p := range productMap {
		if p.TotalCost > 0 {
			p.AvgROI = (p.TotalRevenue - p.TotalCost) / p.TotalCost * 100
		}
		if p.TotalConv > 0 {
			p.AvgCPA = p.TotalCost / float64(p.TotalConv)
		}
		if err := tx.Create(p).Error; err != nil {
			return err
		}
	}

	return nil
}
