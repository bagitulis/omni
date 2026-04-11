package ads

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

// TiktokProductNameEntry represents a product name mapping record
type TiktokProductNameEntry struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
}

// ImportProductNames imports product names from JSON map into tiktok_products table.
// Uses upsert logic: existing product_id records get updated name, new ones get inserted.
func (s *TiktokAdsService) ImportProductNames(ctx context.Context, data []byte) (inserted int, updated int, err error) {
	// Parse JSON: map[productID] -> {name, category}
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

			// Check if product exists
			var count int64
			tx.Table("tiktok_products").
				Where("tenant_id = ? AND product_id = ?", s.tenantID, productID).
				Count(&count)

			if count > 0 {
				// Update existing
				res := tx.Table("tiktok_products").
					Where("tenant_id = ? AND product_id = ?", s.tenantID, productID).
					Updates(map[string]interface{}{
						"name": entry.Name,
					})
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected > 0 {
					updated++
				}
			} else {
				// Insert new
				res := tx.Exec(
					"INSERT INTO tiktok_products (tenant_id, product_id, name, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())",
					s.tenantID, productID, entry.Name,
				)
				if res.Error != nil {
					return res.Error
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
	var products []struct {
		ProductID string `gorm:"column:product_id"`
		Name      string `gorm:"column:name"`
	}
	err := s.db.WithContext(ctx).
		Table("tiktok_products").
		Select("product_id, name").
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
		})
	}
	return entries, nil
}

// BackfillProductNames updates product_name for existing creative data rows
// that have a product_id but empty product_name
func (s *TiktokAdsService) BackfillProductNames(ctx context.Context) (int64, error) {
	result := s.db.WithContext(ctx).Exec(`
		UPDATE tiktok_ads_creative_data cd
		SET product_name = COALESCE(
			(SELECT tp.name FROM tiktok_products tp 
			 WHERE tp.tenant_id = cd.tenant_id AND tp.product_id = cd.product_id
			 LIMIT 1),
			CASE WHEN cd.product_id = '-1' OR cd.product_id = '' THEN 'Non-Product Creative'
			     ELSE 'Discontinued Product'
			END
		)
		WHERE cd.tenant_id = ? AND (cd.product_name IS NULL OR cd.product_name = '')
	`, s.tenantID)

	return result.RowsAffected, result.Error
}
