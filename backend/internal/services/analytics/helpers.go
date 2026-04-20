// Package analytics provides shared helper functions for analytics services
package analytics

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// SkuData holds intermediate data for SKU processing
type SkuData struct {
	UnitPrices    map[float64]struct{}
	ActualIncomes map[float64]struct{}
	Count         int
}

// SkuInfo holds SKU metadata
type SkuInfo struct {
	ItemName  string
	ModelSku  string
	ModelName string
}

// NewSkuData creates a new SkuData instance
func NewSkuData() *SkuData {
	return &SkuData{
		UnitPrices:    make(map[float64]struct{}),
		ActualIncomes: make(map[float64]struct{}),
	}
}

// GetStringValue safely extracts value from string pointer
func GetStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// MapKeysToSlice converts map keys to a slice
func MapKeysToSlice(m map[float64]struct{}) []float64 {
	result := make([]float64, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// HasDifferentPrice checks if any price in the slice differs from target
func HasDifferentPrice(prices []float64, target float64) bool {
	for _, p := range prices {
		if p != target {
			return true
		}
	}
	return false
}

// LookupInventory finds inventory price and name for a SKU using schema prefix
func LookupInventory(db *gorm.DB, schema, tenantID, sku, priceColumn string) (float64, string) {
	var record models.InventoryRecord
	tableName := fmt.Sprintf("%s.inventory_records", schema)
	err := db.Table(tableName).
		Where("tenant_id = ? AND key_value = ? AND key_column_name = ?", tenantID, sku, "SKU").
		First(&record).Error
	if err != nil {
		return 0, ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(record.Data), &data); err != nil {
		return 0, ""
	}

	// Parse price - handle both string and float64 values (Node.js uses parseFloat)
	price := 0.0
	if val, ok := data[priceColumn]; ok {
		switch v := val.(type) {
		case float64:
			price = v
		case string:
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				price = parsed
			}
		case int:
			price = float64(v)
		case int64:
			price = float64(v)
		}
	}

	name := ""
	if val, ok := data["Nama Barang"].(string); ok {
		name = val
	}

	return price, name
}

// ParseTiktokTimestamp converts a TikTok Unix timestamp to time.Time.
// TikTok API returns timestamps in either seconds (10-digit) or milliseconds
// (13-digit). This function handles both formats automatically.
func ParseTiktokTimestamp(ts int64) time.Time {
	if ts > 9_999_999_999 { // > ~year 2286 in seconds = must be milliseconds
		return time.Unix(ts/1000, 0)
	}
	return time.Unix(ts, 0)
}
