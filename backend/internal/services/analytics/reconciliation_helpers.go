package analytics

import (
	"encoding/json"
	"strconv"

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

// LookupInventory finds inventory price and name for a SKU using schema prefix.
// This queries the tenant-specific inventory_records table for the given SKU
// and returns the price from the specified column.
func LookupInventory(db *gorm.DB, tenantID, sku, priceColumn string) (float64, string) {
	var record models.InventoryRecord
	tableName := models.GetTableName("InventoryRecord")
	err := db.Table(tableName).
		Where("tenant_id = ? AND key_value = ? AND key_column_name = ?", tenantID, sku, "SKU").
		First(&record).Error
	if err != nil {
		return 0, ""
	}

	var data map[string]any
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

// ComputeExpectedIncome computes expected income as:
// (inventoryPrice - formulaDeduction) * formulaMultiplier
// Returns nil if inventoryPrice is zero or negative.
func ComputeExpectedIncome(inventoryPrice, formulaDeduction, formulaMultiplier float64) *float64 {
	if inventoryPrice <= 0 {
		return nil
	}
	exp := (inventoryPrice - formulaDeduction) * formulaMultiplier
	if exp < 0 {
		exp = 0
	}
	return &exp
}

// DetermineReconciliationStatus determines the reconciliation status for a SKU group.
func DetermineReconciliationStatus(inventoryPrice *float64, hasMultiplePrices, hasPriceDifference bool) string {
	if inventoryPrice == nil {
		return "NO_INVENTORY"
	}
	if hasPriceDifference || hasMultiplePrices {
		return "PRICE_DIFF"
	}
	return "OK"
}


