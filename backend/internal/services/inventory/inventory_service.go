// Package inventory provides inventory management services
package inventory

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// InventoryService handles inventory operations
type InventoryService struct {
	db       *gorm.DB
	tenantID string
}

// NewInventoryService creates a new inventory service
func NewInventoryService(db *gorm.DB, tenantID string) *InventoryService {
	return &InventoryService{db: db, tenantID: tenantID}
}

// ListFilter represents filter for inventory listing
type ListFilter struct {
	Search            string
	SyncStatus        []string // "synced", "not_synced", "error"
	StockStatus       string   // "in_stock", "low_stock", "out_of_stock", ""
	Platform          []string // "shopee", "lazada", "tiktok"
	LowStockThreshold int      // For stock_status filtering
	Limit             int
	Offset            int
}

// ListResult represents paginated list result
type ListResult struct {
	Total    int64                    `json:"total"`
	Returned int                      `json:"returned"`
	Offset   int                      `json:"offset"`
	Data     []models.InventoryRecord `json:"data"`
}

// GetRecords retrieves inventory records with filtering
func (s *InventoryService) GetRecords(ctx context.Context, filter ListFilter) (*ListResult, error) {
	query := s.db.WithContext(ctx).Where("tenant_id = ?", s.tenantID)

	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where("key_value ILIKE ? OR data::text ILIKE ?", search, search)
	}

	query = buildSyncStatusFilter(query, filter)
	query = buildPlatformFilter(query, filter, s.tenantID)
	query = buildStockStatusFilter(query, filter)

	var total int64
	if err := query.Model(&models.InventoryRecord{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if filter.Limit == 0 {
		filter.Limit = 50
	}
	query = query.Limit(filter.Limit).Offset(filter.Offset).Order("key_value ASC")

	var records []models.InventoryRecord
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}

	return &ListResult{
		Total:    total,
		Returned: len(records),
		Offset:   filter.Offset,
		Data:     records,
	}, nil
}

// PlatformStatusItem represents platform sync status for a SKU
type PlatformStatusItem struct {
	Platform          string  `json:"platform"`
	PlatformProductID string  `json:"platform_product_id"`
	PlatformItemID    string  `json:"platform_item_id"`
	PlatformSKU       string  `json:"platform_sku"`
	Status            string  `json:"status"`
	Stock             int     `json:"stock"`
	Price             float64 `json:"price"`
	LastSyncAt        string  `json:"last_sync_at"`
	ErrorMessage      string  `json:"error_message,omitempty"`
}

// GetRecordsWithPlatformStatus retrieves inventory records with platform status enrichment
func (s *InventoryService) GetRecordsWithPlatformStatus(ctx context.Context, filter ListFilter) (*ListResult, map[string][]PlatformStatusItem, error) {
	// Get base records
	result, err := s.GetRecords(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	// Extract SKUs
	skus := make([]string, 0, len(result.Data))
	for _, rec := range result.Data {
		skus = append(skus, rec.KeyValue)
	}

	// Fetch platform statuses
	platformMap := make(map[string][]PlatformStatusItem)
	if len(skus) > 0 {
		var platformStatuses []models.InventorySkuPlatformStatus
		err = s.db.WithContext(ctx).
			Where("tenant_id = ? AND sku IN ?", s.tenantID, skus).
			Find(&platformStatuses).Error
		if err != nil {
			return nil, nil, err
		}

		// Group by SKU
		for _, ps := range platformStatuses {
			platformMap[ps.SKU] = append(platformMap[ps.SKU], PlatformStatusItem{
				Platform:          ps.Platform,
				PlatformProductID: ps.PlatformProductID,
				PlatformItemID:    ps.PlatformItemID,
				PlatformSKU:       ps.PlatformSKU,
				Status:            ps.Status,
				Stock:             ps.Stock,
				Price:             ps.Price,
				LastSyncAt:        ps.LastCheckedAt.Format(time.RFC3339),
			})
		}
	}

	return result, platformMap, nil
}

// GetBySKU retrieves a single inventory record by SKU
func (s *InventoryService) GetBySKU(ctx context.Context, sku string) (*models.InventoryRecord, error) {
	var record models.InventoryRecord
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", s.tenantID, sku).
		First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &record, err
}

// Update updates an inventory record
func (s *InventoryService) Update(ctx context.Context, sku string, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).
		Model(&models.InventoryRecord{}).
		Where("tenant_id = ? AND key_value = ?", s.tenantID, sku).
		Updates(updates).Error
}

// Delete deletes an inventory record
func (s *InventoryService) Delete(ctx context.Context, sku string) error {
	return s.db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", s.tenantID, sku).
		Delete(&models.InventoryRecord{}).Error
}

// GetCategories retrieves distinct categories
func (s *InventoryService) GetCategories(ctx context.Context) ([]string, error) {
	// Categories are stored in JSONB data field, not as a direct column
	return []string{}, nil
}

// GetSettings retrieves inventory settings — defined in inventory_settings_service.go
// GetByKeyValue / UpdateByKeyValue / CreateRecord / DeleteByKeyValue — defined in inventory_crud_service.go
