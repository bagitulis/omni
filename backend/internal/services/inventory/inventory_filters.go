package inventory

import "gorm.io/gorm"

func buildSyncStatusFilter(query *gorm.DB, filter ListFilter) *gorm.DB {
	if len(filter.SyncStatus) == 0 {
		return query
	}

	hasNotSynced := false
	for _, status := range filter.SyncStatus {
		if status == "not_synced" {
			hasNotSynced = true
			break
		}
	}

	if hasNotSynced {
		return query.Where("(sync_status IN ? OR sync_status IS NULL OR sync_status = '')", filter.SyncStatus)
	}

	return query.Where("sync_status IN ?", filter.SyncStatus)
}

func buildPlatformFilter(query *gorm.DB, filter ListFilter, tenantID string) *gorm.DB {
	if len(filter.Platform) == 0 {
		return query
	}

	return query.Joins("LEFT JOIN inventory_sku_platform_status ON inventory_sku_platform_status.sku = inventory_records.key_value AND inventory_sku_platform_status.tenant_id = ?", tenantID).
		Where("inventory_sku_platform_status.platform IN ?", filter.Platform).
		Distinct()
}

func buildStockStatusFilter(query *gorm.DB, filter ListFilter) *gorm.DB {
	if filter.StockStatus == "" {
		return query
	}

	threshold := filter.LowStockThreshold
	if threshold == 0 {
		threshold = 10
	}

	switch filter.StockStatus {
	case "in_stock":
		return query.Where(`
			CASE
				WHEN data->>'stock' ~ '^\d+$' THEN (data->>'stock')::INTEGER > ?
				ELSE FALSE
			END`, threshold)
	case "low_stock":
		return query.Where(`
			CASE
				WHEN data->>'stock' ~ '^\d+$' THEN (data->>'stock')::INTEGER BETWEEN 1 AND ?
				ELSE FALSE
			END`, threshold)
	case "out_of_stock":
		return query.Where(`
			CASE
				WHEN data->>'stock' ~ '^\d+$' THEN (data->>'stock')::INTEGER = 0
				ELSE FALSE
			END`)
	default:
		return query
	}
}
