package analytics

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
)

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
			OrderSN:       o.OrderSN,
			BuyerPaid:     platformFee,
			ActualFee:     actualFee,
			ShopeeRebate:  o.ShopeeShippingRebate,
			Difference:    diff,
			Status:        status,
			OrderDate:     orderDate,
			BuyerName:     GetStringValue(o.BuyerUserName),
			PaymentMethod: GetStringValue(o.BuyerPaymentMethod),
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
