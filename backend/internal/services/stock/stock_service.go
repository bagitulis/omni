// Package stock provides stock management services
package stock

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	inventorySvc "github.com/omni/backend/internal/services/inventory"
	"gorm.io/gorm"
)

// StockService handles stock operations
type StockService struct {
	db          *gorm.DB
	tenantID    string
	dbPath      string
	credService *services.CredentialService
}

// NewStockService creates a new stock service
func NewStockService(db *gorm.DB, tenantID string) *StockService {
	return &StockService{db: db, tenantID: tenantID}
}

// NewStockServiceWithCreds creates a stock service with credential support
func NewStockServiceWithCreds(db *gorm.DB, tenantID, dbPath string) *StockService {
	return &StockService{
		db:          db,
		tenantID:    tenantID,
		dbPath:      dbPath,
		credService: services.NewCredentialService(dbPath),
	}
}

// GetStock retrieves stock for a SKU
func (s *StockService) GetStock(ctx context.Context, sku string) (*models.StockDTO, error) {
	var record models.InventoryRecord
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", s.tenantID, sku).
		First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Get platform status
	platforms, err := s.getPlatformStock(ctx, sku)
	if err != nil {
		return nil, err
	}

	// Extract values from JSONB data
	skuVal := inventorySvc.GetSKU(record)
	productName := inventorySvc.GetProductName(record)
	quantity := inventorySvc.GetQuantity(record)
	minStock := inventorySvc.GetDataInt(record, "MinStock")

	return &models.StockDTO{
		SKU:          skuVal,
		ProductName:  productName,
		CurrentStock: quantity,
		MinStock:     minStock,
		IsLowStock:   quantity <= minStock,
		Platforms:    platforms,
	}, nil
}

// GetAllStock retrieves all stock
func (s *StockService) GetAllStock(ctx context.Context, lowStockOnly bool, limit, offset int) ([]models.StockDTO, int64, error) {
	var total int64
	query := s.db.WithContext(ctx).Model(&models.InventoryRecord{}).Where("tenant_id = ?", s.tenantID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit == 0 {
		limit = 50
	}

	var records []models.InventoryRecord
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", s.tenantID).Limit(limit).Offset(offset).Order("key_value ASC").Find(&records).Error; err != nil {
		return nil, 0, err
	}

	stocks := make([]models.StockDTO, 0, len(records))
	for _, r := range records {
		skuVal := inventorySvc.GetSKU(r)
		productName := inventorySvc.GetProductName(r)
		quantity := inventorySvc.GetQuantity(r)
		minStock := inventorySvc.GetDataInt(r, "MinStock")

		// Filter low stock if needed
		if lowStockOnly && minStock > 0 && quantity > minStock {
			continue
		}

		platforms, _ := s.getPlatformStock(ctx, skuVal)
		stocks = append(stocks, models.StockDTO{
			SKU:          skuVal,
			ProductName:  productName,
			CurrentStock: quantity,
			MinStock:     minStock,
			IsLowStock:   minStock > 0 && quantity <= minStock,
			Platforms:    platforms,
		})
	}

	return stocks, total, nil
}

// UpdateStock updates stock for a SKU
func (s *StockService) UpdateStock(ctx context.Context, req models.StockUpdateRequest) (*models.StockUpdateResult, error) {
	result := &models.StockUpdateResult{SKU: req.SKU, Success: true}

	// Find the record
	var record models.InventoryRecord
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", s.tenantID, req.SKU).
		First(&record).Error
	if err != nil {
		result.Success = false
		result.Message = "SKU not found"
		return result, nil
	}

	// Update quantity in JSONB data
	data := inventorySvc.GetDataMap(record)
	stockUpdated := false
	for _, key := range []string{"Stock", "stock", "STOCK", "Quantity", "quantity", "QTY", "qty", "Stok", "stok"} {
		if _, exists := data[key]; exists {
			data[key] = req.Quantity
			stockUpdated = true
			break
		}
	}
	if !stockUpdated {
		data["Stock"] = req.Quantity
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		result.Success = false
		result.Message = "failed to marshal stock data"
		return result, fmt.Errorf("failed to marshal stock data: %w", err)
	}
	record.Data = string(jsonData)

	if err := s.db.WithContext(ctx).Save(&record).Error; err != nil {
		result.Success = false
		result.Message = err.Error()
		return result, fmt.Errorf("failed to save stock: %w", err)
	}

	// Sync to platforms if specified
	platforms := req.Platforms
	if len(platforms) == 0 {
		platforms = []string{"shopee", "lazada", "tiktok"}
	}

	for _, platform := range platforms {
		platformResult := s.syncToPlatform(ctx, req.SKU, req.Quantity, platform)
		result.PlatformResults = append(result.PlatformResults, platformResult)
	}

	return result, nil
}

// BulkUpdateStock updates stock for multiple SKUs
func (s *StockService) BulkUpdateStock(ctx context.Context, req models.BulkStockUpdateRequest) (*models.BulkStockUpdateResult, error) {
	result := &models.BulkStockUpdateResult{
		TotalRequested: len(req.Updates),
		Results:        make([]models.StockUpdateResult, 0, len(req.Updates)),
	}

	for _, update := range req.Updates {
		res, err := s.UpdateStock(ctx, update)
		if err != nil {
			result.TotalFailed++
			result.Results = append(result.Results, models.StockUpdateResult{
				SKU:     update.SKU,
				Success: false,
				Message: err.Error(),
			})
			continue
		}
		result.Results = append(result.Results, *res)
		if res.Success {
			result.TotalSuccess++
		} else {
			result.TotalFailed++
		}
	}

	return result, nil
}

// GetLowStockAlerts returns low stock alerts
func (s *StockService) GetLowStockAlerts(ctx context.Context) ([]models.StockAlert, error) {
	var records []models.InventoryRecord
	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", s.tenantID).
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	alerts := make([]models.StockAlert, 0)
	for _, r := range records {
		skuVal := inventorySvc.GetSKU(r)
		productName := inventorySvc.GetProductName(r)
		quantity := inventorySvc.GetQuantity(r)
		minStock := inventorySvc.GetDataInt(r, "MinStock")

		if minStock > 0 && quantity <= minStock {
			alerts = append(alerts, models.StockAlert{
				SKU:          skuVal,
				ProductName:  productName,
				CurrentStock: quantity,
				MinStock:     minStock,
				Shortage:     minStock - quantity,
			})
		}
	}

	return alerts, nil
}

// Note: Platform sync methods (syncToPlatform, syncToShopee, syncToLazada, syncToTiktok)
// are defined in stock_sync.go
