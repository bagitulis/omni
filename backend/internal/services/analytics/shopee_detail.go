package analytics

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// GetSkuOrders returns orders associated with a SKU for the given month/year.
func (s *ShopeeAnalyticsService) GetSkuOrders(ctx context.Context, tenantID, sku string, month, year int) (*dto.ShopeeSkuOrdersResultDTO, error) {
	var items []models.ShopeeEscrowItem
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND (model_sku = ? OR sku = ?) AND month = ? AND year = ?", tenantID, sku, sku, month, year).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to query items for SKU orders: %w", err)
	}

	if len(items) == 0 {
		return &dto.ShopeeSkuOrdersResultDTO{Orders: []dto.ShopeeSkuOrderDTO{}}, nil
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

	var orders []models.ShopeeEscrowOrder
	if err := s.tenantDB.WithContext(ctx).
		Where("id IN ? AND tenant_id = ?", orderIDs, tenantID).
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to query orders for SKU orders: %w", err)
	}

	// Build order map for fast lookup
	orderMap := make(map[string]models.ShopeeEscrowOrder)
	for _, o := range orders {
		orderMap[o.ID] = o
	}

	// Build item map grouped by order ID
	itemMap := make(map[string][]models.ShopeeEscrowItem)
	for _, item := range items {
		itemMap[item.EscrowOrderID] = append(itemMap[item.EscrowOrderID], item)
	}

	result := make([]dto.ShopeeSkuOrderDTO, 0, len(orderIDs))
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

		result = append(result, dto.ShopeeSkuOrderDTO{
			ID:                   o.ID,
			OrderSN:              o.OrderSN,
			EscrowAmount:         o.EscrowAmount,
			CommissionFee:        o.CommissionFee,
			ServiceFee:           o.ServiceFee,
			SellerProcessingFee:  o.SellerProcessingFee,
			BuyerPaidShippingFee: o.BuyerPaidShippingFee,
			ActualShippingFee:    o.ActualShippingFee,
			ShopeeShippingRebate: o.ShopeeShippingRebate,
			EstimatedShippingFee: o.EstimatedShippingFee,
			BuyerTotalAmount:     o.BuyerTotalAmount,
			BuyerName:            GetStringValue(o.BuyerUserName),
			PaymentMethod:        GetStringValue(o.BuyerPaymentMethod),
			OrderDate:            orderDate,
			ItemName:             GetStringValue(item.ItemName),
			ModelName:            GetStringValue(item.ModelName),
			Sku:                  GetStringValue(item.Sku),
			ModelSku:             GetStringValue(item.ModelSku),
			Quantity:             item.Quantity,
			OriginalPrice:        item.OriginalPrice,
		})
	}

	return &dto.ShopeeSkuOrdersResultDTO{Orders: result}, nil
}

// GetOrderItems returns items for a specific order within the given month/year.
func (s *ShopeeAnalyticsService) GetOrderItems(ctx context.Context, tenantID, orderSN string, month, year int) (*dto.ShopeeOrderItemsResultDTO, error) {
	var order models.ShopeeEscrowOrder
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND order_sn = ? AND month = ? AND year = ?", tenantID, orderSN, month, year).
		First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &dto.ShopeeOrderItemsResultDTO{Items: []dto.ShopeeOrderItemDTO{}}, nil
		}
		return nil, fmt.Errorf("failed to query order for items: %w", err)
	}

	var items []models.ShopeeEscrowItem
	if err := s.tenantDB.WithContext(ctx).
		Where("escrow_order_id = ? AND tenant_id = ?", order.ID, tenantID).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}

	result := make([]dto.ShopeeOrderItemDTO, 0, len(items))
	for _, item := range items {
		result = append(result, dto.ShopeeOrderItemDTO{
			ID:                        item.ID,
			EscrowOrderID:             item.EscrowOrderID,
			ItemID:                    item.ItemID,
			ModelID:                   item.ModelID,
			Sku:                       GetStringValue(item.Sku),
			ModelSku:                  GetStringValue(item.ModelSku),
			ItemName:                  GetStringValue(item.ItemName),
			ModelName:                 GetStringValue(item.ModelName),
			Quantity:                  item.Quantity,
			OriginalPrice:             item.OriginalPrice,
			SellingPrice:              item.SellingPrice,
			DiscountedPrice:           item.DiscountedPrice,
			SellerDiscount:            item.SellerDiscount,
			ShopeeDiscount:            item.ShopeeDiscount,
			DiscountFromCoin:          item.DiscountFromCoin,
			DiscountFromVoucherSeller: item.DiscountFromVoucherSeller,
			DiscountFromVoucherShopee: item.DiscountFromVoucherShopee,
			AmsCommissionFee:          item.AmsCommissionFee,
			SellerOrderProcessingFee:  item.SellerOrderProcessingFee,
		})
	}

	return &dto.ShopeeOrderItemsResultDTO{Items: result}, nil
}
