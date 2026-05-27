package analytics

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
)

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
		diff := platformFee - actualFee + o.ShippingFeePlatformDiscount

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
			OrderSN:          o.OrderID,
			CustomerPaid:     platformFee,
			ActualFee:        actualFee,
			PlatformDiscount: o.ShippingFeePlatformDiscount,
			Difference:       diff,
			Status:           status,
			OrderDate:        orderDate,
			OrderStatus:      GetStringValue(o.OrderStatus),
			Currency:         o.Currency,
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

// RepopulateItems is implemented in tiktok_escrow_repopulate.go

// ComputeTiktokShippingDiff calculates the shipping fee difference for TikTok orders.
// Formula: ShippingFeeCustomerPaid - ShippingFeeActual + ShippingFeePlatformDiscount.
// Exported for deterministic testing.
func ComputeTiktokShippingDiff(customerPaid, actual, discount float64) float64 {
	return customerPaid - actual + discount
}
