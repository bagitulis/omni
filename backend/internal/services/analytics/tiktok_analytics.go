package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/jobs"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// TiktokAnalyticsService implements handlers.TiktokAnalyticsService.
// It handles settings management, sync orchestration, reconciliation computation,
// shipping-fee analysis, item repopulation, and escrow sync with progress tracking.
type TiktokAnalyticsService struct {
	systemDB     *gorm.DB
	tenantDB     *gorm.DB
	tenantID     string
	queueManager *jobs.QueueManager
}

// NewTiktokAnalyticsService creates a new TiktokAnalyticsService.
func NewTiktokAnalyticsService(systemDB *gorm.DB, tenantDB *gorm.DB, tenantID string) *TiktokAnalyticsService {
	return &TiktokAnalyticsService{
		systemDB:     systemDB,
		tenantDB:     tenantDB,
		tenantID:     tenantID,
		queueManager: jobs.NewQueueManager(tenantDB, tenantID),
	}
}

// GetSettings retrieves analytics settings for the tenant, platform 'tiktok'.
// Returns default values when no settings are found.
func (s *TiktokAnalyticsService) GetSettings(ctx context.Context, tenantID string) (*dto.AnalyticsSettingsDTO, error) {
	var settings models.AnalyticsSettings
	err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, "tiktok").
		First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &dto.AnalyticsSettingsDTO{
				PriceColumn:       "HARGA",
				FormulaDeduction:  1500,
				FormulaMultiplier: 0.84,
			}, nil
		}
		return nil, fmt.Errorf("failed to get analytics settings: %w", err)
	}

	return &dto.AnalyticsSettingsDTO{
		PriceColumn:       settings.PriceColumn,
		FormulaDeduction:  settings.FormulaDeduction,
		FormulaMultiplier: settings.FormulaMultiplier,
	}, nil
}

// SaveSettings upserts analytics settings for the tenant, platform 'tiktok'.
func (s *TiktokAnalyticsService) SaveSettings(ctx context.Context, tenantID string, input *dto.AnalyticsSettingsDTO) (*dto.AnalyticsSettingsDTO, error) {
	var settings models.AnalyticsSettings
	err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, "tiktok").
		First(&settings).Error

	if err == gorm.ErrRecordNotFound {
		settings = models.AnalyticsSettings{
			ID:       uuid.New().String(),
			TenantID: tenantID,
			Platform: "tiktok",
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to find analytics settings: %w", err)
	}

	settings.PriceColumn = input.PriceColumn
	settings.FormulaDeduction = input.FormulaDeduction
	settings.FormulaMultiplier = input.FormulaMultiplier
	settings.UpdatedAt = time.Now()

	if err := s.tenantDB.WithContext(ctx).Save(&settings).Error; err != nil {
		return nil, fmt.Errorf("failed to save analytics settings: %w", err)
	}

	return &dto.AnalyticsSettingsDTO{
		PriceColumn:       settings.PriceColumn,
		FormulaDeduction:  settings.FormulaDeduction,
		FormulaMultiplier: settings.FormulaMultiplier,
	}, nil
}

// GetSyncStatus returns the escrow sync status for the given month/year.
func (s *TiktokAnalyticsService) GetSyncStatus(ctx context.Context, tenantID string, month, year int) (*dto.SyncStatusDTO, error) {
	var sync models.TiktokEscrowSync
	err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		First(&sync).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &dto.SyncStatusDTO{
				Synced: false,
			}, nil
		}
		return nil, fmt.Errorf("failed to get sync status: %w", err)
	}

	syncedAt := sync.SyncedAt
	return &dto.SyncStatusDTO{
		Synced:       true,
		TotalOrders:  sync.TotalOrders,
		FailedOrders: sync.FailedOrders,
		SyncedAt:     &syncedAt,
	}, nil
}

// SyncEscrow creates an escrow sync job for the given month/year and returns the job ID.
func (s *TiktokAnalyticsService) SyncEscrow(ctx context.Context, tenantID string, month, year int, forceResync bool) (string, error) {
	data := models.EscrowSyncJobData{
		TenantID:    tenantID,
		Platform:    "tiktok",
		Month:       month,
		Year:        year,
		ForceResync: forceResync,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job data: %w", err)
	}

	req := models.CreateJobRequest{
		ID:       uuid.New().String(),
		Type:     models.JobTypeTiktokEscrowSync,
		Data:     string(dataJSON),
		Priority: "normal",
	}

	job, err := s.queueManager.AddJob(req)
	if err != nil {
		return "", fmt.Errorf("failed to create sync job: %w", err)
	}

	return job.ID, nil
}

// DeleteSyncData deletes all escrow sync data for the given month/year.
// Deletes items first (via order ID subquery since TiktokEscrowItem has no Month/Year),
// then orders, then the sync record.
func (s *TiktokAnalyticsService) DeleteSyncData(ctx context.Context, tenantID string, month, year int) error {
	// Delete items via order IDs (TiktokEscrowItem has no Month/Year fields)
	if err := s.tenantDB.WithContext(ctx).
		Where("escrow_order_id IN (?)",
			s.tenantDB.WithContext(ctx).Model(&models.TiktokEscrowOrder{}).
				Select("id").
				Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year),
		).
		Delete(&models.TiktokEscrowItem{}).Error; err != nil {
		return fmt.Errorf("failed to delete escrow items: %w", err)
	}

	// Delete orders
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Delete(&models.TiktokEscrowOrder{}).Error; err != nil {
		return fmt.Errorf("failed to delete escrow orders: %w", err)
	}

	// Delete sync record
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Delete(&models.TiktokEscrowSync{}).Error; err != nil {
		return fmt.Errorf("failed to delete escrow sync record: %w", err)
	}

	return nil
}

// GetReconciliation computes reconciliation summary and SKU-group details
// for the given month/year by aggregating escrow item data.
// Uses order ID subquery since TiktokEscrowItem lacks Month/Year fields.
func (s *TiktokAnalyticsService) GetReconciliation(ctx context.Context, tenantID string, month, year int) (*dto.TiktokReconciliationResultDTO, error) {
	type skuAggregation struct {
		Sku        string
		ItemName   string
		TotalQty   int
		TotalAmt   float64
		SysAmt     float64
		OrderCount int
	}

	// Get order IDs for the period (items don't have month/year)
	var orderIDs []string
	if err := s.tenantDB.WithContext(ctx).
		Model(&models.TiktokEscrowOrder{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Pluck("id", &orderIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to get order IDs: %w", err)
	}

	if len(orderIDs) == 0 {
		return &dto.TiktokReconciliationResultDTO{
			Summary: dto.ReconciliationSummaryDTO{},
			Details: []dto.TiktokSkuGroupDTO{},
		}, nil
	}

	var agg []skuAggregation
	err := s.tenantDB.WithContext(ctx).
		Model(&models.TiktokEscrowItem{}).
		Select(`
			COALESCE(seller_sku, '') as sku,
			COALESCE(product_name, '') as item_name,
			SUM(quantity) as total_qty,
			SUM(sale_price * quantity) as total_amt,
			SUM(original_price * quantity) as sys_amt,
			COUNT(DISTINCT escrow_order_id) as order_count
		`).
		Where("escrow_order_id IN ?", orderIDs).
		Where("tenant_id = ?", tenantID).
		Group("sku, item_name").
		Order("total_amt DESC").
		Scan(&agg).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query reconciliation data: %w", err)
	}

	skuOk := 0
	skuWithPriceDiff := 0
	skuNoInventory := 0
	totalTransactions := 0

	details := make([]dto.TiktokSkuGroupDTO, 0, len(agg))
	for _, a := range agg {
		priceDiff := a.SysAmt - a.TotalAmt
		priceDiffPercent := 0.0
		if a.TotalAmt > 0 {
			priceDiffPercent = (priceDiff / a.TotalAmt) * 100
		}

		if priceDiff == 0 {
			skuOk++
		} else if a.SysAmt == 0 {
			skuNoInventory++
		} else {
			skuWithPriceDiff++
		}

		details = append(details, dto.TiktokSkuGroupDTO{
			SKU:              a.Sku,
			ItemName:         a.ItemName,
			TotalQuantity:    a.TotalQty,
			TotalAmount:      a.TotalAmt,
			SystemAmount:     a.SysAmt,
			PriceDiff:        priceDiff,
			PriceDiffPercent: priceDiffPercent,
			OrderCount:       a.OrderCount,
		})

		totalTransactions += a.TotalQty
	}

	return &dto.TiktokReconciliationResultDTO{
		Summary: dto.ReconciliationSummaryDTO{
			TotalSKU:          len(agg),
			TotalTransactions: totalTransactions,
			SKUOk:             skuOk,
			SKUWithPriceDiff:  skuWithPriceDiff,
			SKUNoInventory:    skuNoInventory,
		},
		Details: details,
	}, nil
}

// GetShippingFeeAnalysis computes shipping fee differences for the given month/year.
// Compares buyer-paid shipping fee (ShippingFeeCustomerPaid) against actual shipping fee charged.
func (s *TiktokAnalyticsService) GetShippingFeeAnalysis(ctx context.Context, tenantID string, month, year int) (*dto.TiktokShippingFeeResultDTO, error) {
	var orders []models.TiktokEscrowOrder
	err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Find(&orders).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query escrow orders: %w", err)
	}

	totalOrders := len(orders)
	ordersWithDiff := 0
	totalProfit := 0.0
	totalLoss := 0.0

	details := make([]dto.TiktokShippingOrderDTO, 0, len(orders))
	for _, o := range orders {
		platformFee := o.ShippingFeeCustomerPaid
		actualFee := o.ShippingFeeActual
		diff := platformFee - actualFee

		if diff > 0.01 || diff < -0.01 {
			ordersWithDiff++
		}
		if diff > 0 {
			totalProfit += diff
		} else {
			totalLoss += diff
		}

		orderDate := ""
		if o.OrderDate != nil {
			orderDate = o.OrderDate.Format("2006-01-02")
		}

		status := "ok"
		if diff > 0.01 {
			status = "profit"
		} else if diff < -0.01 {
			status = "loss"
		}

		details = append(details, dto.TiktokShippingOrderDTO{
			OrderSN:     o.OrderID,
			ShippingFee: platformFee,
			ActualFee:   actualFee,
			Difference:  diff,
			Status:      status,
			OrderDate:   orderDate,
		})
	}

	netImpact := totalProfit + totalLoss

	return &dto.TiktokShippingFeeResultDTO{
		Summary: dto.ShippingFeeSummaryDTO{
			TotalOrders:          totalOrders,
			OrdersWithDifference: ordersWithDiff,
			TotalProfit:          totalProfit,
			TotalLoss:            totalLoss,
			NetImpact:            netImpact,
		},
		Details: details,
	}, nil
}

// RepopulateItems triggers item repopulation for the given period.
// Placeholder: actual SDK calls will be added in a future task.
func (s *TiktokAnalyticsService) RepopulateItems(ctx context.Context, tenantID string, period string) error {
	log.Info().
		Str("tenant_id", tenantID).
		Str("period", period).
		Msg("TikTok item repopulation triggered (placeholder)")
	return nil
}
