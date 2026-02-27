package orders

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// RecalculateResult holds stats from a Locked/Sellable recalculation.
type RecalculateResult struct {
	UpdatedRecords int
	TotalRecords   int
	LockedSKUs     int
	TotalColumn    string
}

// RecalculateLockedSellable reads locked_orders from DB and updates
// Locked/Sellable columns in all inventory_records for the given tenant.
//
// Sellable = rawTotal - Locked (clamped to 0).
// If there are no locked orders, every record gets Locked=0 and Sellable=rawTotal.
//
// This should be called:
//  1. After saving locked orders  (POST /api/orders/locked-today)
//  2. After SyncFromSheets         (POST /api/inventory/sync/from-sheets)
//  3. On inventory page load       (triggered by frontend)
func RecalculateLockedSellable(ctx context.Context, db *gorm.DB, tenantID string) (*RecalculateResult, error) {
	result := &RecalculateResult{}

	// 1. Read locked orders from DB and build SKU→qty map
	lockedService := NewLockedOrderService(db)
	lockedOrders, err := lockedService.GetLockedOrders(ctx, tenantID)
	if err != nil {
		// If table doesn't exist or other DB error, treat as no locked orders
		lockedOrders = nil
	}

	lockedBySku := make(map[string]int, len(lockedOrders))
	for _, item := range lockedOrders {
		if item.SKU != "" {
			lockedBySku[item.SKU] += item.Qty
		}
	}
	result.LockedSKUs = len(lockedBySku)

	// 2. Fetch all inventory records for this tenant
	var inventoryRecords []models.InventoryRecord
	if err := db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&inventoryRecords).Error; err != nil {
		return result, err
	}

	result.TotalRecords = len(inventoryRecords)
	if len(inventoryRecords) == 0 {
		return result, nil
	}

	// 3. Read RawTotalColumn from settings
	totalColumnName := ""
	var invSettings models.InventorySettings
	if err := db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		First(&invSettings).Error; err == nil && invSettings.RawTotalColumn != "" {
		totalColumnName = invSettings.RawTotalColumn
	}

	// Fallback: auto-detect from JSONB keys if user hasn't configured via UI
	if totalColumnName == "" {
		totalColumnName = "Total" // ultimate fallback
		candidates := []string{"TOTAL", "Total", "total", "Stock", "stock", "STOCK", "Quantity", "quantity", "QTY", "qty", "Stok", "stok"}
		var sampleData map[string]interface{}
		if err := json.Unmarshal([]byte(inventoryRecords[0].Data), &sampleData); err == nil {
			for _, c := range candidates {
				if _, ok := sampleData[c]; ok {
					totalColumnName = c
					break
				}
			}
		}
	}
	result.TotalColumn = totalColumnName

	// 4. Update each record with Locked/Sellable
	for _, record := range inventoryRecords {
		var dataMap map[string]interface{}
		if err := json.Unmarshal([]byte(record.Data), &dataMap); err != nil {
			continue
		}

		// Get Total value from the detected column
		totalVal := 0.0
		if val, ok := dataMap[totalColumnName]; ok {
			switch v := val.(type) {
			case float64:
				totalVal = v
			case string:
				if parsed, e := strconv.ParseFloat(v, 64); e == nil {
					totalVal = parsed
				}
			}
		}

		// Locked = qty from locked_orders (0 if no locked orders for this SKU)
		lockedQty := lockedBySku[record.KeyValue]
		sellable := int(totalVal) - lockedQty
		if sellable < 0 {
			sellable = 0
		}

		dataMap["Locked"] = lockedQty
		dataMap["Sellable"] = sellable

		updated, err := json.Marshal(dataMap)
		if err != nil {
			continue
		}

		if err := db.WithContext(ctx).
			Model(&record).
			Update("data", string(updated)).Error; err != nil {
			continue
		}
		result.UpdatedRecords++
	}

	return result, nil
}
