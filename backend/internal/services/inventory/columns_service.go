package inventory

import (
	"context"
	"encoding/json"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ColumnsService handles column configuration operations
type ColumnsService struct {
	db       *gorm.DB
	tenantID string
}

// NewColumnsService creates a new columns service
func NewColumnsService(db *gorm.DB, tenantID string) *ColumnsService {
	return &ColumnsService{db: db, tenantID: tenantID}
}

// ColumnInfo represents information about an available column
type ColumnInfo struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Sortable    bool   `json:"sortable"`
	Filterable  bool   `json:"filterable"`
}

// GetAvailableColumns returns all available columns for inventory
func (s *ColumnsService) GetAvailableColumns(ctx context.Context) ([]ColumnInfo, error) {
	// Define all available columns
	columns := []ColumnInfo{
		{Key: "sku", Label: "SKU", Type: "string", Description: "Stock Keeping Unit", Sortable: true, Filterable: true},
		{Key: "product_name", Label: "Product Name", Type: "string", Description: "Product display name", Sortable: true, Filterable: true},
		{Key: "category", Label: "Category", Type: "string", Description: "Product category", Sortable: true, Filterable: true},
		{Key: "quantity", Label: "Quantity", Type: "number", Description: "Current stock quantity", Sortable: true, Filterable: true},
		{Key: "min_stock", Label: "Min Stock", Type: "number", Description: "Minimum stock threshold", Sortable: true, Filterable: false},
		{Key: "price", Label: "Price", Type: "currency", Description: "Base price", Sortable: true, Filterable: true},
		{Key: "cost", Label: "Cost", Type: "currency", Description: "Cost price", Sortable: true, Filterable: false},
		{Key: "shopee_status", Label: "Shopee Status", Type: "status", Description: "Listing status on Shopee", Sortable: false, Filterable: true},
		{Key: "lazada_status", Label: "Lazada Status", Type: "status", Description: "Listing status on Lazada", Sortable: false, Filterable: true},
		{Key: "tiktok_status", Label: "TikTok Status", Type: "status", Description: "Listing status on TikTok", Sortable: false, Filterable: true},
		{Key: "shopee_product_id", Label: "Shopee Product ID", Type: "string", Description: "Product ID on Shopee", Sortable: false, Filterable: false},
		{Key: "lazada_product_id", Label: "Lazada Product ID", Type: "string", Description: "Product ID on Lazada", Sortable: false, Filterable: false},
		{Key: "tiktok_product_id", Label: "TikTok Product ID", Type: "string", Description: "Product ID on TikTok", Sortable: false, Filterable: false},
		{Key: "last_synced", Label: "Last Synced", Type: "datetime", Description: "Last sync timestamp", Sortable: true, Filterable: false},
		{Key: "created_at", Label: "Created At", Type: "datetime", Description: "Record creation time", Sortable: true, Filterable: false},
		{Key: "updated_at", Label: "Updated At", Type: "datetime", Description: "Last update time", Sortable: true, Filterable: false},
	}
	return columns, nil
}

// GetSelectedColumns returns user's selected columns
func (s *ColumnsService) GetSelectedColumns(ctx context.Context) ([]string, error) {
	var settings models.InventorySettings
	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", s.tenantID).
		First(&settings).Error

	if err == gorm.ErrRecordNotFound {
		// Return default columns
		return []string{"sku", "product_name", "quantity", "price", "shopee_status", "lazada_status"}, nil
	}
	if err != nil {
		return nil, err
	}

	var columns []string
	if settings.SelectedColumns != "" {
		if err := json.Unmarshal([]byte(settings.SelectedColumns), &columns); err != nil {
			return []string{"sku", "product_name", "quantity", "price"}, nil
		}
	}

	if len(columns) == 0 {
		return []string{"sku", "product_name", "quantity", "price"}, nil
	}

	return columns, nil
}

// SaveSelectedColumns saves user's column selection
func (s *ColumnsService) SaveSelectedColumns(ctx context.Context, columns []string) error {
	columnsJSON, err := json.Marshal(columns)
	if err != nil {
		return err
	}

	return s.db.WithContext(ctx).
		Where(models.InventorySettings{TenantID: s.tenantID}).
		Assign(models.InventorySettings{SelectedColumns: string(columnsJSON)}).
		FirstOrCreate(&models.InventorySettings{}).Error
}
