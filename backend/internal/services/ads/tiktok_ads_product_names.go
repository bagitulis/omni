package ads

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// TiktokProductNameEntry represents a product name mapping record
type TiktokProductNameEntry struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
}

// ImportProductNames imports product names from JSON map into tiktok_ads_product_names table.
// Uses upsert logic: existing product_id records get updated name, new ones get inserted.
func (s *TiktokAdsService) ImportProductNames(ctx context.Context, data []byte) (inserted int, updated int, err error) {
	var raw map[string]struct {
		Name     string `json:"name"`
		Category string `json:"category"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return 0, 0, fmt.Errorf("invalid JSON: %w", err)
	}

	if len(raw) == 0 {
		return 0, 0, fmt.Errorf("JSON contains no product entries")
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for productID, entry := range raw {
			if productID == "" || entry.Name == "" {
				continue
			}

			var existing models.TiktokAdsProductName
			found := tx.Where("tenant_id = ? AND product_id = ?", s.tenantID, productID).
				First(&existing).Error

			if found == nil {
				// Update existing
				existing.Name = entry.Name
				if entry.Category != "" {
					existing.Category = entry.Category
				}
				existing.UpdatedAt = time.Now()
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
				updated++
			} else {
				// Insert new
				newEntry := models.TiktokAdsProductName{
					TenantID:  s.tenantID,
					ProductID: productID,
					Name:      entry.Name,
					Category:  entry.Category,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				if err := tx.Create(&newEntry).Error; err != nil {
					return err
				}
				inserted++
			}
		}
		return nil
	})

	return inserted, updated, err
}

// GetProductNames returns all product name mappings for the tenant
func (s *TiktokAdsService) GetProductNames(ctx context.Context) ([]TiktokProductNameEntry, error) {
	var products []models.TiktokAdsProductName
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND name != '' AND name IS NOT NULL", s.tenantID).
		Order("name ASC").
		Find(&products).Error

	if err != nil {
		return nil, err
	}

	entries := make([]TiktokProductNameEntry, 0, len(products))
	for _, p := range products {
		entries = append(entries, TiktokProductNameEntry{
			ProductID: p.ProductID,
			Name:      p.Name,
			Category:  p.Category,
		})
	}
	return entries, nil
}

// BackfillProductNames updates product_name for existing creative data rows
func (s *TiktokAdsService) BackfillProductNames(ctx context.Context) (int64, error) {
	// Get the correct table name for the model
	stmt := &gorm.Statement{DB: s.db}
	stmt.Parse(&models.TiktokAdsProductName{})
	nameTable := stmt.Schema.Table

	stmt2 := &gorm.Statement{DB: s.db}
	stmt2.Parse(&models.TiktokAdsCreativeData{})
	dataTable := stmt2.Schema.Table

	result := s.db.WithContext(ctx).Exec(fmt.Sprintf(`
		UPDATE %s cd
		SET product_name = COALESCE(
			(SELECT pn.name FROM %s pn 
			 WHERE pn.tenant_id = cd.tenant_id AND pn.product_id = cd.product_id
			 LIMIT 1),
			CASE WHEN cd.product_id = '-1' OR cd.product_id = '' THEN 'Non-Product Creative'
			     ELSE 'Discontinued Product'
			END
		)
		WHERE cd.tenant_id = ? AND (cd.product_name IS NULL OR cd.product_name = '')
	`, dataTable, nameTable), s.tenantID)

	return result.RowsAffected, result.Error
}
