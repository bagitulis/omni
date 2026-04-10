// Package analytics provides analytics services for Shopee
package analytics

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ShopeeAnalyticsService handles Shopee analytics operations
type ShopeeAnalyticsService struct {
	db       *gorm.DB
	tenantID string
	schema   string
}

// NewShopeeAnalyticsService creates a new Shopee analytics service
func NewShopeeAnalyticsService(db *gorm.DB, tenantID string) *ShopeeAnalyticsService {
	return &ShopeeAnalyticsService{
		db:       db,
		tenantID: tenantID,
		schema:   fmt.Sprintf("tenant_%s", tenantID),
	}
}

// table returns table name with schema prefix
func (s *ShopeeAnalyticsService) table(name string) string {
	return fmt.Sprintf("%s.%s", s.schema, name)
}

// getDB returns a DB session with context
func (s *ShopeeAnalyticsService) getDB(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx)
}

// GetSyncStatus returns sync status for a month
func (s *ShopeeAnalyticsService) GetSyncStatus(ctx context.Context, month, year int) (*dto.SyncStatusDTO, error) {
	var sync models.ShopeeEscrowSync
	err := s.getDB(ctx).Table(s.table("shopee_escrow_sync")).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		First(&sync).Error

	if err == gorm.ErrRecordNotFound {
		return &dto.SyncStatusDTO{Synced: false, TotalOrders: 0, SyncedAt: nil}, nil
	}
	if err != nil {
		return nil, err
	}

	syncedAt := sync.SyncedAt.Format(time.RFC3339)
	return &dto.SyncStatusDTO{
		Synced:       true,
		TotalOrders:  sync.TotalOrders,
		FailedOrders: sync.FailedOrders,
		SyncedAt:     &syncedAt,
	}, nil
}

// GetSettings returns analytics settings for Shopee
func (s *ShopeeAnalyticsService) GetSettings(ctx context.Context) (*dto.AnalyticsSettingsDTO, error) {
	var settings models.AnalyticsSettings
	err := s.getDB(ctx).Table(s.table("analytics_settings")).
		Where("tenant_id = ? AND platform = ?", s.tenantID, "shopee").
		First(&settings).Error

	if err == gorm.ErrRecordNotFound {
		return &dto.AnalyticsSettingsDTO{
			PriceColumn:       "HARGA",
			FormulaDeduction:  1500,
			FormulaMultiplier: 0.84,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return &dto.AnalyticsSettingsDTO{
		PriceColumn:       settings.PriceColumn,
		FormulaDeduction:  settings.FormulaDeduction,
		FormulaMultiplier: settings.FormulaMultiplier,
	}, nil
}

// SaveSettings saves analytics settings for Shopee
func (s *ShopeeAnalyticsService) SaveSettings(ctx context.Context, req *dto.AnalyticsSettingsDTO) error {
	db := s.getDB(ctx)
	tbl := s.table("analytics_settings")
	var existing models.AnalyticsSettings
	err := db.Table(tbl).Where("tenant_id = ? AND platform = ?", s.tenantID, "shopee").
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		settings := models.AnalyticsSettings{
			ID:                uuid.New().String(),
			TenantID:          s.tenantID,
			Platform:          "shopee",
			PriceColumn:       req.PriceColumn,
			FormulaDeduction:  req.FormulaDeduction,
			FormulaMultiplier: req.FormulaMultiplier,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
		return db.Table(tbl).Create(&settings).Error
	}
	if err != nil {
		return err
	}

	return db.Table(tbl).Where("id = ?", existing.ID).Updates(map[string]interface{}{
		"price_column":       req.PriceColumn,
		"formula_deduction":  req.FormulaDeduction,
		"formula_multiplier": req.FormulaMultiplier,
		"updated_at":         time.Now(),
	}).Error
}

// SyncEscrow creates or updates sync record for a month
// For now, this creates a sync record. Actual API sync should be implemented separately.
func (s *ShopeeAnalyticsService) SyncEscrow(ctx context.Context, month, year int, forceResync bool) (*dto.SyncResultDTO, error) {
	db := s.getDB(ctx)
	syncTable := s.table("shopee_escrow_sync")
	ordersTable := s.table("shopee_orders")

	// Check if already synced
	var existingSync models.ShopeeEscrowSync
	err := db.Table(syncTable).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		First(&existingSync).Error

	if err == nil && !forceResync {
		return &dto.SyncResultDTO{
			TotalOrders: existingSync.TotalOrders,
			TotalItems:  0,
			Message:     "Already synced. Use forceResync to update.",
		}, nil
	}

	// Count orders for this month from shopee_orders
	var orderCount int64
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)
	db.Table(ordersTable).
		Where("tenant_id = ? AND created_at >= ? AND created_at < ?", s.tenantID, startDate, endDate).
		Count(&orderCount)

	// Upsert sync record
	now := time.Now()
	if err == gorm.ErrRecordNotFound {
		sync := models.ShopeeEscrowSync{
			ID:          uuid.New().String(),
			TenantID:    s.tenantID,
			Month:       month,
			Year:        year,
			TotalOrders: int(orderCount),
			SyncedAt:    now,
		}
		if err := db.Table(syncTable).Create(&sync).Error; err != nil {
			return nil, err
		}
	} else if err == nil {
		if err := db.Table(syncTable).Where("id = ?", existingSync.ID).Updates(map[string]interface{}{
			"total_orders": orderCount,
			"synced_at":    now,
		}).Error; err != nil {
			return nil, err
		}
	}

	return &dto.SyncResultDTO{
		TotalOrders: int(orderCount),
		TotalItems:  0,
		Message:     fmt.Sprintf("Synced %d orders for %d-%02d", orderCount, year, month),
	}, nil
}

// DeleteSyncData deletes all sync data for a month
func (s *ShopeeAnalyticsService) DeleteSyncData(ctx context.Context, month, year int) error {
	db := s.getDB(ctx)
	return db.Transaction(func(tx *gorm.DB) error {
		// Get order IDs for items deletion
		var orderIDs []string
		tx.Table(s.table("shopee_escrow_orders")).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			Pluck("id", &orderIDs)

		// Delete items
		if len(orderIDs) > 0 {
			if err := tx.Table(s.table("shopee_escrow_items")).
				Where("escrow_order_id IN ?", orderIDs).
				Delete(&models.ShopeeEscrowItem{}).Error; err != nil {
				return err
			}
		}

		// Delete orders
		if err := tx.Table(s.table("shopee_escrow_orders")).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			Delete(&models.ShopeeEscrowOrder{}).Error; err != nil {
			return err
		}

		// Delete sync record
		return tx.Table(s.table("shopee_escrow_sync")).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			Delete(&models.ShopeeEscrowSync{}).Error
	})
}

// GetReconciliation analyzes price reconciliation for a month
func (s *ShopeeAnalyticsService) GetReconciliation(ctx context.Context, month, year int) (*dto.ReconciliationResultDTO, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	reconService := NewShopeeReconciliationService(s.db, s.tenantID, s.schema)
	return reconService.GetReconciliation(ctx, month, year, settings)
}

// GetShippingFeeAnalysis analyzes shipping fee differences
func (s *ShopeeAnalyticsService) GetShippingFeeAnalysis(ctx context.Context, month, year int) (*dto.ShopeeShippingFeeResultDTO, error) {
	var orders []models.ShopeeEscrowOrder
	err := s.getDB(ctx).Table(s.table("shopee_escrow_orders")).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		Find(&orders).Error
	if err != nil {
		return nil, err
	}

	var resultOrders []dto.ShopeeShippingOrderDTO
	var totalProfit, totalLoss float64

	for _, order := range orders {
		difference := order.BuyerPaidShippingFee - order.ActualShippingFee + order.ShopeeShippingRebate
		if math.Abs(difference) < 0.01 {
			continue
		}

		var orderDate *string
		if order.OrderDate != nil {
			formatted := order.OrderDate.Format("2006-01-02")
			orderDate = &formatted
		}

		resultOrders = append(resultOrders, dto.ShopeeShippingOrderDTO{
			OrderSn:       order.OrderSN,
			OrderDate:     orderDate,
			BuyerPaid:     order.BuyerPaidShippingFee,
			ActualFee:     order.ActualShippingFee,
			ShopeeRebate:  order.ShopeeShippingRebate,
			Difference:    difference,
			BuyerName:     order.BuyerUserName,
			PaymentMethod: order.BuyerPaymentMethod,
		})

		if difference > 0 {
			totalProfit += difference
		} else {
			totalLoss += difference
		}
	}

	return &dto.ShopeeShippingFeeResultDTO{
		Summary: dto.ShippingFeeSummaryDTO{
			TotalOrders:          len(orders),
			OrdersWithDifference: len(resultOrders),
			TotalProfit:          totalProfit,
			TotalLoss:            totalLoss,
			NetImpact:            totalProfit + totalLoss,
		},
		Orders: resultOrders,
	}, nil
}

// RepopulateItems repopulates shopee_escrow_items from raw_order_income stored in shopee_escrow_orders
func (s *ShopeeAnalyticsService) RepopulateItems(ctx context.Context, month, year int) (int, int, error) {
	syncSvc := NewShopeeEscrowSyncService(s.db, s.tenantID, "")
	return syncSvc.RepopulateEscrowItems(ctx, month, year)
}

