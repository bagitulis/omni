// Package analytics provides analytics services for TikTok
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

// TiktokAnalyticsService handles TikTok analytics operations
type TiktokAnalyticsService struct {
	db       *gorm.DB
	tenantID string
	schema   string
}

// NewTiktokAnalyticsService creates a new TikTok analytics service
func NewTiktokAnalyticsService(db *gorm.DB, tenantID string) *TiktokAnalyticsService {
	return &TiktokAnalyticsService{
		db:       db,
		tenantID: tenantID,
		schema:   fmt.Sprintf("tenant_%s", tenantID),
	}
}

// table returns table name with schema prefix
func (s *TiktokAnalyticsService) table(name string) string {
	return fmt.Sprintf("%s.%s", s.schema, name)
}

// getDB returns a DB session with context
func (s *TiktokAnalyticsService) getDB(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx)
}

// GetSyncStatus returns sync status for a month
func (s *TiktokAnalyticsService) GetSyncStatus(ctx context.Context, month, year int) (*dto.SyncStatusDTO, error) {
	var sync models.TiktokEscrowSync
	err := s.getDB(ctx).Table(s.table("tiktok_escrow_sync")).
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

// GetSettings returns analytics settings for TikTok
func (s *TiktokAnalyticsService) GetSettings(ctx context.Context) (*dto.AnalyticsSettingsDTO, error) {
	var settings models.AnalyticsSettings
	err := s.getDB(ctx).Table(s.table("analytics_settings")).
		Where("tenant_id = ? AND platform = ?", s.tenantID, "tiktok").
		First(&settings).Error

	if err == gorm.ErrRecordNotFound {
		return &dto.AnalyticsSettingsDTO{
			PriceColumn:       "HARGA",
			FormulaDeduction:  1500,
			FormulaMultiplier: 0.86,
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

// SaveSettings saves analytics settings for TikTok
func (s *TiktokAnalyticsService) SaveSettings(ctx context.Context, req *dto.AnalyticsSettingsDTO) error {
	db := s.getDB(ctx)
	tbl := s.table("analytics_settings")
	var existing models.AnalyticsSettings
	err := db.Table(tbl).Where("tenant_id = ? AND platform = ?", s.tenantID, "tiktok").
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		settings := models.AnalyticsSettings{
			ID:                uuid.New().String(),
			TenantID:          s.tenantID,
			Platform:          "tiktok",
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
func (s *TiktokAnalyticsService) SyncEscrow(ctx context.Context, month, year int, forceResync bool) (*dto.SyncResultDTO, error) {
	db := s.getDB(ctx)
	syncTable := s.table("tiktok_escrow_sync")
	ordersTable := s.table("tiktok_orders")

	// Check if already synced
	var existingSync models.TiktokEscrowSync
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

	// Count orders for this month from tiktok_orders
	var orderCount int64
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)
	db.Table(ordersTable).
		Where("tenant_id = ? AND created_at >= ? AND created_at < ?", s.tenantID, startDate, endDate).
		Count(&orderCount)

	// Upsert sync record
	now := time.Now()
	if err == gorm.ErrRecordNotFound {
		sync := models.TiktokEscrowSync{
			ID:          uuid.New().String(),
			TenantID:    s.tenantID,
			Month:       month,
			Year:        year,
			TotalOrders: int(orderCount),
			SyncedAt:    now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := db.Table(syncTable).Create(&sync).Error; err != nil {
			return nil, err
		}
	} else if err == nil {
		if err := db.Table(syncTable).Where("id = ?", existingSync.ID).Updates(map[string]interface{}{
			"total_orders": orderCount,
			"synced_at":    now,
			"updated_at":   now,
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
func (s *TiktokAnalyticsService) DeleteSyncData(ctx context.Context, month, year int) error {
	db := s.getDB(ctx)
	return db.Transaction(func(tx *gorm.DB) error {
		var orderIDs []string
		tx.Table(s.table("tiktok_escrow_orders")).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			Pluck("id", &orderIDs)

		if len(orderIDs) > 0 {
			if err := tx.Table(s.table("tiktok_escrow_items")).
				Where("escrow_order_id IN ?", orderIDs).
				Delete(&models.TiktokEscrowItem{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Table(s.table("tiktok_escrow_orders")).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			Delete(&models.TiktokEscrowOrder{}).Error; err != nil {
			return err
		}

		return tx.Table(s.table("tiktok_escrow_sync")).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			Delete(&models.TiktokEscrowSync{}).Error
	})
}

// GetReconciliation analyzes price reconciliation for TikTok
func (s *TiktokAnalyticsService) GetReconciliation(ctx context.Context, month, year int) (*dto.ReconciliationResultDTO, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	reconService := NewTiktokReconciliationService(s.db, s.tenantID, s.schema)
	return reconService.GetReconciliation(ctx, month, year, settings)
}

// RepopulateItems re-populates tiktok_escrow_items from raw_order_data already in DB.
// Use this when escrow_items is empty but orders were already synced.
func (s *TiktokAnalyticsService) RepopulateItems(ctx context.Context, month, year int) (int, int, error) {
	dbPath := fmt.Sprintf("tenant_%s", s.tenantID)
	syncSvc := NewTiktokEscrowSyncService(s.db, s.tenantID, dbPath)
	return syncSvc.RepopulateEscrowItems(ctx, month, year)
}

// GetShippingFeeAnalysis analyzes TikTok shipping fee differences
func (s *TiktokAnalyticsService) GetShippingFeeAnalysis(ctx context.Context, month, year int) (*dto.TiktokShippingFeeResultDTO, error) {
	var orders []models.TiktokEscrowOrder
	err := s.getDB(ctx).Table(s.table("tiktok_escrow_orders")).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		Find(&orders).Error
	if err != nil {
		return nil, err
	}

	var resultOrders []dto.TiktokShippingOrderDTO
	var totalProfit, totalLoss float64

	for _, order := range orders {
		difference := order.ShippingFeeCustomerPaid - order.ShippingFeeActual + order.ShippingFeePlatformDiscount
		if math.Abs(difference) < 0.01 {
			continue
		}

		var orderDate *string
		if order.OrderDate != nil {
			formatted := order.OrderDate.Format("2006-01-02")
			orderDate = &formatted
		}

		resultOrders = append(resultOrders, dto.TiktokShippingOrderDTO{
			OrderID:         order.OrderID,
			OrderDate:       orderDate,
			CustomerPaid:    order.ShippingFeeCustomerPaid,
			ActualCost:      order.ShippingFeeActual,
			PlatformSubsidy: order.ShippingFeePlatformDiscount,
			Difference:      difference,
		})

		if difference > 0 {
			totalProfit += difference
		} else {
			totalLoss += difference
		}
	}

	return &dto.TiktokShippingFeeResultDTO{
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
