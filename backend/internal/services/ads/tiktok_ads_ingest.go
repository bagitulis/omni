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

// SaveBatch saves a batch of TikTok ads data with duplicate protection.
// The batch record is created OUTSIDE the data-insertion transaction so that
// failure status persists even if the data insert is rolled back.
func (s *TiktokAdsService) SaveBatch(ctx context.Context, filename string, result *TiktokParseResult) error {
	db := s.db.WithContext(ctx)

	// Check for duplicate period (outside transaction — read-only)
	var existingCount int64
	if err := db.Model(&models.TiktokAdsUploadBatch{}).
		Where("tenant_id = ? AND period_label = ?", s.tenantID, result.Period.Label).
		Count(&existingCount).Error; err != nil {
		return fmt.Errorf("check duplicate: %w", err)
	}
	if existingCount > 0 {
		return fmt.Errorf("period %s already uploaded — please delete existing data first before re-uploading", result.Period.Label)
	}

	// Enrich product names
	productIDs := make([]string, 0)
	for _, d := range result.Data {
		if d.ProductID != "" && d.ProductID != "-1" {
			productIDs = append(productIDs, d.ProductID)
		}
	}
	nameMap := lookupProductNames(db, s.tenantID, productIDs)

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

	// Create batch record OUTSIDE transaction — survives rollback
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
	if err := db.Create(&batch).Error; err != nil {
		return fmt.Errorf("create batch: %w", err)
	}

	for i := range result.Data {
		result.Data[i].UploadBatchID = batch.ID
	}

	// Insert data inside transaction — rollback only affects data rows
	insertErr := db.Transaction(func(tx *gorm.DB) error {
		if len(result.Data) == 0 {
			return nil
		}
		return tx.CreateInBatches(result.Data, 100).Error
	})

	if insertErr != nil {
		// Mark batch as failed — this persists because batch was created outside tx
		db.Model(&models.TiktokAdsUploadBatch{}).Where("id = ?", batch.ID).
			Updates(map[string]interface{}{
				"status":        "failed",
				"error_message": insertErr.Error(),
				"updated_at":    time.Now(),
			})
		return fmt.Errorf("insert data (batch %s): %w", batch.ID, insertErr)
	}

	// Mark batch as success
	return db.Model(&models.TiktokAdsUploadBatch{}).Where("id = ?", batch.ID).
		Updates(map[string]interface{}{
			"inserted_rows": len(result.Data),
			"skipped_rows":  skipped,
			"status":        "success",
			"updated_at":    time.Now(),
		}).Error
}

// DeleteBatch deletes a TikTok ads upload batch and its associated creative data
func (s *TiktokAdsService) DeleteBatch(ctx context.Context, batchID string) error {
	db := s.db.WithContext(ctx)

	// Verify batch belongs to this tenant
	var batch models.TiktokAdsUploadBatch
	if err := db.Where("id = ? AND tenant_id = ?", batchID, s.tenantID).First(&batch).Error; err != nil {
		return fmt.Errorf("batch not found: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// Delete creative data first (foreign key safety)
		if err := tx.Where("upload_batch_id = ? AND tenant_id = ?", batchID, s.tenantID).
			Delete(&models.TiktokAdsCreativeData{}).Error; err != nil {
			return fmt.Errorf("delete creative data: %w", err)
		}
		// Delete batch record
		if err := tx.Where("id = ? AND tenant_id = ?", batchID, s.tenantID).
			Delete(&models.TiktokAdsUploadBatch{}).Error; err != nil {
			return fmt.Errorf("delete batch: %w", err)
		}
		return nil
	})
}

// lookupProductNames fetches product names from tiktok_ads_product_names table
// NOTE: This is separate from tiktok_products (TikTok Shop sync) — ads product IDs are different
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

	var products []models.TiktokAdsProductName
	tx.Where("tenant_id = ? AND product_id IN ?", tenantID, deduped).
		Find(&products)

	for _, p := range products {
		if p.Name != "" {
			result[p.ProductID] = p.Name
		}
	}
	return result
}
