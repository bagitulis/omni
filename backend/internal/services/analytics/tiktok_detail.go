package analytics

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// GetSkuOrders returns orders associated with a SKU for the given month/year.
func (s *TiktokAnalyticsService) GetSkuOrders(ctx context.Context, tenantID, sku string, month, year int) (*dto.TiktokSkuOrdersResultDTO, error) {
	var items []models.TiktokEscrowItem
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND (seller_sku = ? OR sku_id = ?)", tenantID, sku, sku).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to query items for SKU orders: %w", err)
	}

	if len(items) == 0 {
		return &dto.TiktokSkuOrdersResultDTO{Orders: []dto.TiktokSkuOrderDTO{}}, nil
	}

	// Collect unique order IDs
	orderIDSet := make(map[string]bool)
	for _, item := range items {
		orderIDSet[item.EscrowOrderID] = true
	}
	orderIDs := make([]string, 0, len(orderIDSet))
	for oid := range orderIDSet {
		orderIDs = append(orderIDs, oid)
	}

	// Get orders matching month/year and tenant
	var orders []models.TiktokEscrowOrder
	if err := s.tenantDB.WithContext(ctx).
		Where("id IN ? AND tenant_id = ? AND month = ? AND year = ?", orderIDs, tenantID, month, year).
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to query orders for SKU orders: %w", err)
	}

	// Build order map
	orderMap := make(map[string]models.TiktokEscrowOrder)
	for _, o := range orders {
		orderMap[o.ID] = o
	}

	// Build item map grouped by order ID
	itemMap := make(map[string][]models.TiktokEscrowItem)
	for _, item := range items {
		itemMap[item.EscrowOrderID] = append(itemMap[item.EscrowOrderID], item)
	}

	result := make([]dto.TiktokSkuOrderDTO, 0, len(orderIDs))
	for _, oid := range orderIDs {
		o, ok := orderMap[oid]
		if !ok {
			continue
		}
		orderItems := itemMap[oid]
		if len(orderItems) == 0 {
			continue
		}

		// Use the first matching item for SKU-specific fields
		item := orderItems[0]
		orderDate := ""
		if o.OrderDate != nil {
			orderDate = o.OrderDate.Format("2006-01-02")
		}

		result = append(result, dto.TiktokSkuOrderDTO{
			ID:                          o.ID,
			OrderID:                     o.OrderID,
			OrderStatus:                 GetStringValue(o.OrderStatus),
			TotalSettlementAmount:       o.TotalSettlementAmount,
			ProductRevenue:              o.ProductRevenue,
			PlatformCommission:           o.PlatformCommission,
			TransactionFee:              o.TransactionFee,
			ShippingFeeCustomerPaid:     o.ShippingFeeCustomerPaid,
			ShippingFeeActual:           o.ShippingFeeActual,
			ShippingFeePlatformDiscount: o.ShippingFeePlatformDiscount,
			SellerShippingDiscount:      o.SellerShippingDiscount,
			RefundAmount:                o.RefundAmount,
			Currency:                    o.Currency,
			BuyerName:                   GetStringValue(o.BuyerName),
			OrderDate:                   orderDate,
			ProductName:                 GetStringValue(item.ProductName),
			SellerSku:                   GetStringValue(item.SellerSku),
			Quantity:                    item.Quantity,
			SalePrice:                   item.SalePrice,
			OriginalPrice:               item.OriginalPrice,
		})
	}

	return &dto.TiktokSkuOrdersResultDTO{Orders: result}, nil
}

// GetOrderItems returns items for a specific order within the given month/year.
func (s *TiktokAnalyticsService) GetOrderItems(ctx context.Context, tenantID, orderSN string, month, year int) (*dto.TiktokOrderItemsResultDTO, error) {
	var order models.TiktokEscrowOrder
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND order_id = ? AND month = ? AND year = ?", tenantID, orderSN, month, year).
		First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &dto.TiktokOrderItemsResultDTO{Items: []dto.TiktokOrderItemDTO{}}, nil
		}
		return nil, fmt.Errorf("failed to query order for items: %w", err)
	}

	var items []models.TiktokEscrowItem
	if err := s.tenantDB.WithContext(ctx).
		Where("escrow_order_id = ? AND tenant_id = ?", order.ID, tenantID).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}

	result := make([]dto.TiktokOrderItemDTO, 0, len(items))
	for _, item := range items {
		result = append(result, dto.TiktokOrderItemDTO{
			ID:                          item.ID,
			EscrowOrderID:               item.EscrowOrderID,
			ProductName:                 GetStringValue(item.ProductName),
			SkuID:                       GetStringValue(item.SkuID),
			SellerSku:                   GetStringValue(item.SellerSku),
			Quantity:                    item.Quantity,
			SalePrice:                   item.SalePrice,
			OriginalPrice:               item.OriginalPrice,
			SubtotalAfterSellerDiscount: item.SubtotalAfterSellerDiscount,
			PlatformDiscount:            item.PlatformDiscount,
			SellerDiscount:              item.SellerDiscount,
			Commission:                  item.Commission,
			TransactionFeeItem:           item.TransactionFeeItem,
			SettlementAmount:            item.SettlementAmount,
		})
	}

	return &dto.TiktokOrderItemsResultDTO{Items: result}, nil
}
