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

// SaveBatch saves a batch of TikTok ads data with duplicate protection
func (s *TiktokAdsService) SaveBatch(ctx context.Context, filename string, result *TiktokParseResult) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check for duplicate period
		var existingCount int64
		if err := tx.Model(&models.TiktokAdsUploadBatch{}).
			Where("tenant_id = ? AND period_label = ?", s.tenantID, result.Period.Label).
			Count(&existingCount).Error; err != nil {
			return fmt.Errorf("check duplicate: %w", err)
		}
		if existingCount > 0 {
			return fmt.Errorf("period %s already uploaded — please delete existing data first before re-uploading", result.Period.Label)
		}

		// Collect product IDs for name enrichment
		productIDs := make([]string, 0)
		for _, d := range result.Data {
			if d.ProductID != "" && d.ProductID != "-1" {
				productIDs = append(productIDs, d.ProductID)
			}
		}

		// Lookup product names from tiktok_products table
		nameMap := lookupProductNames(tx, s.tenantID, productIDs)

		// Enrich product names and mark special cases
		skipped := 0
		for i := range result.Data {
			pid := result.Data[i].ProductID
			if pid == "" || pid == "-1" {
				result.Data[i].ProductName = "Non-Product Creative"
				skipped++
			} else if name, ok := nameMap[pid]; ok {
				result.Data[i].ProductName = name
			} else {
				result.Data[i].ProductName = "Discontinued Product"
			}
		}

		batch := models.TiktokAdsUploadBatch{
			ID:          uuid.New().String(),
			TenantID:    s.tenantID,
			FileName:    filename,
			PeriodStart: result.Period.Start,
			PeriodEnd:   result.Period.End,
			PeriodLabel: result.Period.Label,
			TotalRows:   len(result.Data),
			SkippedRows: skipped,
			Status:      "processing",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := tx.Create(&batch).Error; err != nil {
			return fmt.Errorf("create batch: %w", err)
		}

		for i := range result.Data {
			result.Data[i].UploadBatchID = batch.ID
		}

		inserted := 0
		if len(result.Data) > 0 {
			if err := tx.CreateInBatches(result.Data, 100).Error; err != nil {
				tx.Model(&models.TiktokAdsUploadBatch{}).Where("id = ?", batch.ID).
					Updates(map[string]interface{}{
						"status":        "failed",
						"error_message": err.Error(),
						"updated_at":    time.Now(),
					})
				return fmt.Errorf("create data (batch size %d): %w", len(result.Data), err)
			}
			inserted = len(result.Data)
		}

		// Update batch as success
		return tx.Model(&models.TiktokAdsUploadBatch{}).Where("id = ?", batch.ID).
			Updates(map[string]interface{}{
				"inserted_rows": inserted,
				"skipped_rows":  skipped,
				"status":        "success",
				"updated_at":    time.Now(),
			}).Error
	})
}

// lookupProductNames fetches product names from tiktok_products table
func lookupProductNames(tx *gorm.DB, tenantID string, productIDs []string) map[string]string {
	result := make(map[string]string)
	if len(productIDs) == 0 {
		return result
	}

	// Deduplicate
	unique := make(map[string]bool)
	deduped := make([]string, 0)
	for _, id := range productIDs {
		if !unique[id] {
			unique[id] = true
			deduped = append(deduped, id)
		}
	}

	var products []struct {
		ProductID string `gorm:"column:product_id"`
		Name      string `gorm:"column:name"`
	}
	tx.Table("tiktok_products").
		Select("product_id, name").
		Where("tenant_id = ? AND product_id IN ?", tenantID, deduped).
		Find(&products)

	for _, p := range products {
		if p.Name != "" {
			result[p.ProductID] = p.Name
		}
	}
	return result
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
				ProductName: c.ProductName, // Use enriched product name
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
