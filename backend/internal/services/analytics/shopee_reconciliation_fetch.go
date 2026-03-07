package analytics

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog/log"
)

// fetchEscrowItems retrieves escrow items for a month with order data.
// If shopee_escrow_items is empty (Shopee API sometimes omits item details for older orders),
// it falls back to shopee_order_items data joined via order_sn.
func (s *ShopeeReconciliationService) fetchEscrowItems(ctx context.Context, month, year int) ([]models.ShopeeEscrowItem, map[string]*models.ShopeeEscrowOrder, map[string]int, error) {
	itemsTable := s.table("shopee_escrow_items")
	ordersTable := s.table("shopee_escrow_orders")

	// Fetch all escrow orders for this month
	var orders []models.ShopeeEscrowOrder
	err := s.getDB(ctx).Table(ordersTable).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		Find(&orders).Error
	if err != nil {
		return nil, nil, nil, err
	}

	// Fetch escrow items joined with orders
	var items []models.ShopeeEscrowItem
	err = s.getDB(ctx).Table(itemsTable).
		Joins(fmt.Sprintf("JOIN %s ON %s.id = %s.escrow_order_id", ordersTable, ordersTable, itemsTable)).
		Where(fmt.Sprintf("%s.tenant_id = ?", itemsTable), s.tenantID).
		Where(fmt.Sprintf("%s.month = ? AND %s.year = ?", ordersTable, ordersTable), month, year).
		Find(&items).Error
	if err != nil {
		return nil, nil, nil, err
	}

	// Build order map (keyed by ID and by OrderSN for fallback)
	orderMap := make(map[string]*models.ShopeeEscrowOrder)
	orderByOrderSN := make(map[string]*models.ShopeeEscrowOrder)
	for i := range orders {
		orderMap[orders[i].ID] = &orders[i]
		orderByOrderSN[orders[i].OrderSN] = &orders[i]
	}

	// Fallback: if escrow items are empty but we have escrow orders,
	// synthesize items from shopee_order_items (populated by webhook/order sync)
	if len(items) == 0 && len(orders) > 0 {
		items, err = s.fetchFallbackItemsFromOrderItems(ctx, orders, orderByOrderSN)
		if err != nil {
			// Log but don't fail - just return empty
			log.Warn().Err(err).Str("service", "shopee_recon").Msg("Fallback fetch from order_items failed")
			items = []models.ShopeeEscrowItem{}
		}
	}

	// Count items per order
	orderItemCount := make(map[string]int)
	for _, item := range items {
		orderItemCount[item.EscrowOrderID]++
	}

	return items, orderMap, orderItemCount, nil
}

// fetchFallbackItemsFromOrderItems synthesizes ShopeeEscrowItem records from shopee_order_items.
// This is used when the Shopee escrow API did not return item-level data during sync.
func (s *ShopeeReconciliationService) fetchFallbackItemsFromOrderItems(
	ctx context.Context,
	orders []models.ShopeeEscrowOrder,
	orderByOrderSN map[string]*models.ShopeeEscrowOrder,
) ([]models.ShopeeEscrowItem, error) {
	orderItemsTable := s.table("shopee_order_items")

	// Collect all order SNs
	orderSNs := make([]string, 0, len(orders))
	for _, o := range orders {
		orderSNs = append(orderSNs, o.OrderSN)
	}

	if len(orderSNs) == 0 {
		return nil, nil
	}

	var orderItems []models.ShopeeOrderItem
	err := s.getDB(ctx).Table(orderItemsTable).
		Where("tenant_id = ? AND order_sn IN ?", s.tenantID, orderSNs).
		Find(&orderItems).Error
	if err != nil {
		return nil, err
	}

	if len(orderItems) == 0 {
		return nil, nil
	}

	// Synthesize ShopeeEscrowItem from ShopeeOrderItem
	synthesized := make([]models.ShopeeEscrowItem, 0, len(orderItems))
	for _, oi := range orderItems {
		escrowOrder, ok := orderByOrderSN[oi.OrderSN]
		if !ok {
			continue
		}

		// Calculate original price from order item price (quantity-adjusted)
		var origPrice float64
		var qty int
		if oi.Price != nil {
			origPrice = *oi.Price
		}
		if oi.Quantity != nil {
			qty = *oi.Quantity
		}
		if qty == 0 {
			qty = 1
		}

		sku := StringPtr(oi.ItemSku)
		modelSku := StringPtr(oi.ModelSku)
		itemName := StringPtr(oi.ItemName)
		modelName := StringPtr(oi.ModelName)
		monthVal := escrowOrder.Month
		yearVal := escrowOrder.Year

		synthesized = append(synthesized, models.ShopeeEscrowItem{
			ID:            fmt.Sprintf("fallback_%d", oi.ID),
			TenantID:      s.tenantID,
			EscrowOrderID: escrowOrder.ID,
			OrderSN:       StringPtr(oi.OrderSN),
			Month:         &monthVal,
			Year:          &yearVal,
			ItemID:        Int64Ptr(oi.ItemID),
			ModelID:       oi.ModelID,
			Sku:           sku,
			ModelSku:      modelSku,
			ItemName:      itemName,
			ModelName:     modelName,
			Quantity:      qty,
			OriginalPrice: origPrice * float64(qty), // total price for the item
		})
	}

	log.Info().
		Str("service", "shopee_recon").
		Int("synthesized", len(synthesized)).
		Int("order_items", len(orderItems)).
		Msg("Fallback: synthesized items from order_items")
	return synthesized, nil
}
