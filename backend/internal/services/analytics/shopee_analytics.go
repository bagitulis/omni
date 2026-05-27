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
	"gorm.io/gorm"
)

// ShopeeAnalyticsService implements handlers.ShopeeAnalyticsService.
// It handles settings management, sync orchestration, reconciliation computation,
// shipping-fee analysis, item repopulation, and escrow sync with progress tracking.
type ShopeeAnalyticsService struct {
	systemDB     *gorm.DB
	tenantDB     *gorm.DB
	tenantID     string
	queueManager *jobs.QueueManager
}

// NewShopeeAnalyticsService creates a new ShopeeAnalyticsService.
func NewShopeeAnalyticsService(systemDB *gorm.DB, tenantDB *gorm.DB, tenantID string) *ShopeeAnalyticsService {
	return &ShopeeAnalyticsService{
		systemDB:     systemDB,
		tenantDB:     tenantDB,
		tenantID:     tenantID,
		queueManager: jobs.NewQueueManager(tenantDB, tenantID),
	}
}

// GetSettings retrieves analytics settings for the tenant, platform 'shopee'.
// Returns default values when no settings are found.
func (s *ShopeeAnalyticsService) GetSettings(ctx context.Context, tenantID string) (*dto.AnalyticsSettingsDTO, error) {
	var settings models.AnalyticsSettings
	err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, "shopee").
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

// SaveSettings upserts analytics settings for the tenant, platform 'shopee'.
func (s *ShopeeAnalyticsService) SaveSettings(ctx context.Context, tenantID string, input *dto.AnalyticsSettingsDTO) (*dto.AnalyticsSettingsDTO, error) {
	var settings models.AnalyticsSettings
	err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, "shopee").
		First(&settings).Error

	if err == gorm.ErrRecordNotFound {
		settings = models.AnalyticsSettings{
			ID:       uuid.New().String(),
			TenantID: tenantID,
			Platform: "shopee",
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
func (s *ShopeeAnalyticsService) GetSyncStatus(ctx context.Context, tenantID string, month, year int) (*dto.SyncStatusDTO, error) {
	var sync models.ShopeeEscrowSync
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
func (s *ShopeeAnalyticsService) SyncEscrow(ctx context.Context, tenantID string, month, year int, forceResync bool) (string, error) {
	data := models.EscrowSyncJobData{
		TenantID:    tenantID,
		Platform:    "shopee",
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
		Type:     models.JobTypeShopeeEscrowSync,
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
// Deletes items first (foreign key constraint), then orders, then the sync record.
func (s *ShopeeAnalyticsService) DeleteSyncData(ctx context.Context, tenantID string, month, year int) error {
	// Delete items first (foreign key constraint from ShopeeEscrowItem -> ShopeeEscrowOrder)
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Delete(&models.ShopeeEscrowItem{}).Error; err != nil {
		return fmt.Errorf("failed to delete escrow items: %w", err)
	}

	// Delete orders
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Delete(&models.ShopeeEscrowOrder{}).Error; err != nil {
		return fmt.Errorf("failed to delete escrow orders: %w", err)
	}

	// Delete sync record
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Delete(&models.ShopeeEscrowSync{}).Error; err != nil {
		return fmt.Errorf("failed to delete escrow sync record: %w", err)
	}

	return nil
}

// GetReconciliation computes reconciliation summary and SKU-group details
// for the given month/year by aggregating escrow item data.
func (s *ShopeeAnalyticsService) GetReconciliation(ctx context.Context, tenantID string, month, year int) (*dto.ReconciliationResultDTO, error) {
	type skuAggregation struct {
		Sku        string
		ItemName   string
		TotalQty   int
		TotalAmt   float64
		SysAmt     float64
		OrderCount int
	}

	var agg []skuAggregation
	err := s.tenantDB.WithContext(ctx).
		Model(&models.ShopeeEscrowItem{}).
		Select(`
			COALESCE(model_sku, sku, '') as sku,
			COALESCE(item_name, '') as item_name,
			SUM(quantity) as total_qty,
			SUM(selling_price * quantity) as total_amt,
			SUM(original_price * quantity) as sys_amt,
			COUNT(DISTINCT escrow_order_id) as order_count
		`).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Group("COALESCE(model_sku, sku, ''), COALESCE(item_name, '')").
		Order("total_amt DESC").
		Scan(&agg).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query reconciliation data: %w", err)
	}

	skuOk := 0
	skuWithPriceDiff := 0
	skuNoInventory := 0
	totalTransactions := 0

	details := make([]dto.SkuGroupDTO, 0, len(agg))
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

		details = append(details, dto.SkuGroupDTO{
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

	return &dto.ReconciliationResultDTO{
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
// Compares buyer-paid shipping fee (platform fee) against actual shipping fee charged.
func (s *ShopeeAnalyticsService) GetShippingFeeAnalysis(ctx context.Context, tenantID string, month, year int) (*dto.ShopeeShippingFeeResultDTO, error) {
	var orders []models.ShopeeEscrowOrder
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

	details := make([]dto.ShopeeShippingOrderDTO, 0, len(orders))
	for _, o := range orders {
		platformFee := o.BuyerPaidShippingFee
		actualFee := o.ActualShippingFee
		diff := platformFee - actualFee + o.ShopeeShippingRebate

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

		details = append(details, dto.ShopeeShippingOrderDTO{
			OrderSN:     o.OrderSN,
			PlatformFee: platformFee,
			ActualFee:   actualFee,
			Difference:  diff,
			Status:      status,
			OrderDate:   orderDate,
		})
	}

	netImpact := totalProfit + totalLoss

	return &dto.ShopeeShippingFeeResultDTO{
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

// ComputeShopeeShippingDiff calculates the shipping fee difference for Shopee orders.
// Formula: BuyerPaidShippingFee - ActualShippingFee + ShopeeShippingRebate.
// Exported for deterministic testing.
func ComputeShopeeShippingDiff(buyerPaid, actual, rebate float64) float64 {
	return buyerPaid - actual + rebate
}
