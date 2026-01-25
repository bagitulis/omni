package sheets

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// InventoryRecord represents an inventory record from sheets
type InventoryRecord struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TenantID        string    `gorm:"index;not null" json:"tenantId"`
	SKU             string    `gorm:"index;not null" json:"sku"`
	ProductName     string    `json:"productName"`
	VariationName   string    `json:"variationName,omitempty"`
	StockQuantity   int       `json:"stockQuantity"`
	ReservedQty     int       `json:"reservedQty"`
	AvailableQty    int       `json:"availableQty"`
	MinStock        int       `json:"minStock"`
	MaxStock        int       `json:"maxStock"`
	ReorderPoint    int       `json:"reorderPoint"`
	UnitCost        float64   `json:"unitCost"`
	TotalValue      float64   `json:"totalValue"`
	Location        string    `json:"location,omitempty"`
	LastCountDate   time.Time `json:"lastCountDate,omitempty"`
	Source          string    `json:"source"` // sheet, manual, api
	SheetConfigID   uint      `json:"sheetConfigId,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// TableName returns the table name for GORM
func (InventoryRecord) TableName() string {
	return "inventory_records"
}

// InventorySyncHistory represents sync history
type InventorySyncHistory struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      string    `gorm:"index;not null" json:"tenantId"`
	SheetConfigID uint      `json:"sheetConfigId,omitempty"`
	SyncType      string    `json:"syncType"` // full, delta
	RecordsAdded  int       `json:"recordsAdded"`
	RecordsUpdated int      `json:"recordsUpdated"`
	RecordsDeleted int      `json:"recordsDeleted"`
	Status        string    `json:"status"` // success, failed, partial
	ErrorMessage  string    `json:"errorMessage,omitempty"`
	Duration      int       `json:"duration"` // milliseconds
	CreatedAt     time.Time `json:"createdAt"`
}

// TableName returns the table name for GORM
func (InventorySyncHistory) TableName() string {
	return "inventory_sync_history"
}

// InventoryStats represents inventory statistics
type InventoryStats struct {
	TotalSKUs        int     `json:"totalSkus"`
	TotalStock       int     `json:"totalStock"`
	TotalValue       float64 `json:"totalValue"`
	LowStockCount    int     `json:"lowStockCount"`
	OutOfStockCount  int     `json:"outOfStockCount"`
	OverstockCount   int     `json:"overstockCount"`
	LastSyncAt       time.Time `json:"lastSyncAt,omitempty"`
}

// InventorySheetService handles inventory sheet operations
type InventorySheetService struct {
	db *gorm.DB
}

// NewInventorySheetService creates a new inventory sheet service
func NewInventorySheetService(db *gorm.DB) *InventorySheetService {
	return &InventorySheetService{db: db}
}

// GetInventoryRecords retrieves inventory records
func (s *InventorySheetService) GetInventoryRecords(
	ctx context.Context,
	tenantID string,
	limit int,
	offset int,
) ([]InventoryRecord, int64, error) {
	var records []InventoryRecord
	var total int64

	query := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if err := query.Model(&InventoryRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("sku").Find(&records).Error
	return records, total, err
}

// GetInventoryBySKU retrieves inventory by SKU
func (s *InventorySheetService) GetInventoryBySKU(
	ctx context.Context,
	tenantID string,
	sku string,
) (*InventoryRecord, error) {
	var record InventoryRecord

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND sku = ?", tenantID, sku).
		First(&record).Error

	if err != nil {
		return nil, err
	}

	return &record, nil
}

// SaveInventoryRecords saves inventory records (upsert by SKU)
func (s *InventorySheetService) SaveInventoryRecords(
	ctx context.Context,
	tenantID string,
	records []InventoryRecord,
) error {
	for i := range records {
		records[i].TenantID = tenantID
		records[i].AvailableQty = records[i].StockQuantity - records[i].ReservedQty
		records[i].TotalValue = float64(records[i].StockQuantity) * records[i].UnitCost

		err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND sku = ?", tenantID, records[i].SKU).
			Assign(records[i]).
			FirstOrCreate(&InventoryRecord{}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// GetInventoryStats returns inventory statistics
func (s *InventorySheetService) GetInventoryStats(
	ctx context.Context,
	tenantID string,
) (*InventoryStats, error) {
	stats := &InventoryStats{}

	// Total SKUs and stock
	var totalResult struct {
		Count int
		Stock int
		Value float64
	}
	s.db.WithContext(ctx).
		Model(&InventoryRecord{}).
		Where("tenant_id = ?", tenantID).
		Select("COUNT(*) as count, COALESCE(SUM(stock_quantity), 0) as stock, COALESCE(SUM(total_value), 0) as value").
		Scan(&totalResult)
	stats.TotalSKUs = totalResult.Count
	stats.TotalStock = totalResult.Stock
	stats.TotalValue = totalResult.Value

	// Low stock count
	var lowStockCount int64
	s.db.WithContext(ctx).
		Model(&InventoryRecord{}).
		Where("tenant_id = ? AND stock_quantity <= reorder_point AND stock_quantity > 0", tenantID).
		Count(&lowStockCount)
	stats.LowStockCount = int(lowStockCount)

	// Out of stock count
	var outOfStockCount int64
	s.db.WithContext(ctx).
		Model(&InventoryRecord{}).
		Where("tenant_id = ? AND stock_quantity = 0", tenantID).
		Count(&outOfStockCount)
	stats.OutOfStockCount = int(outOfStockCount)

	// Overstock count
	var overstockCount int64
	s.db.WithContext(ctx).
		Model(&InventoryRecord{}).
		Where("tenant_id = ? AND stock_quantity > max_stock AND max_stock > 0", tenantID).
		Count(&overstockCount)
	stats.OverstockCount = int(overstockCount)

	// Last sync
	var lastSync InventorySyncHistory
	if err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ?", tenantID, "success").
		Order("created_at DESC").
		First(&lastSync).Error; err == nil {
		stats.LastSyncAt = lastSync.CreatedAt
	}

	return stats, nil
}

// SaveSyncHistory saves sync history record
func (s *InventorySheetService) SaveSyncHistory(
	ctx context.Context,
	history *InventorySyncHistory,
) error {
	return s.db.WithContext(ctx).Create(history).Error
}

// GetLowStockItems returns items below reorder point
func (s *InventorySheetService) GetLowStockItems(
	ctx context.Context,
	tenantID string,
) ([]InventoryRecord, error) {
	var records []InventoryRecord

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND stock_quantity <= reorder_point", tenantID).
		Order("stock_quantity ASC").
		Find(&records).Error

	return records, err
}
