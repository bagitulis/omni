package analytics

import (
	"context"
	"fmt"
	"math"

	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
)

// GetReconciliation computes reconciliation summary and SKU-group details
// for the given month/year by aggregating escrow item data.
// Uses order ID subquery since TiktokEscrowItem lacks Month/Year fields.
func (s *TiktokAnalyticsService) GetReconciliation(ctx context.Context, tenantID string, month, year int) (*dto.TiktokReconciliationResultDTO, error) {
	settings, err := s.GetSettings(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	// Get order IDs for the period
	var orderIDs []string
	if err := s.tenantDB.WithContext(ctx).
		Model(&models.TiktokEscrowOrder{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Pluck("id", &orderIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to get order IDs: %w", err)
	}

	if len(orderIDs) == 0 {
		return &dto.TiktokReconciliationResultDTO{
			Summary:   dto.ReconciliationSummaryDTO{},
			SkuGroups: []dto.TiktokSkuGroupDTO{},
		}, nil
	}

	// Fetch items and orders
	var items []models.TiktokEscrowItem
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND escrow_order_id IN ?", tenantID, orderIDs).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to query escrow items: %w", err)
	}

	var orders []models.TiktokEscrowOrder
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to query escrow orders: %w", err)
	}

	// Build order map and item count per order
	orderMap := make(map[string]*models.TiktokEscrowOrder)
	orderItemCount := make(map[string]int)
	for i := range orders {
		orderMap[orders[i].ID] = &orders[i]
		orderItemCount[orders[i].ID] = 0
	}
	for _, item := range items {
		orderItemCount[item.EscrowOrderID]++
	}

	// Group items by SKU
	type skuGroup struct {
		data *SkuData
		info struct {
			productName string
			variantName string
		}
	}
	skuGroups := make(map[string]*skuGroup)

	for _, item := range items {
		// Resolve SKU: use seller_sku if available, fallback to sku_id, then UNKNOWN
		sku := GetStringValue(item.SellerSku)
		if sku == "" {
			sku = GetStringValue(item.SkuID)
		}
		if sku == "" {
			sku = "UNKNOWN"
		}

		qty := item.Quantity
		if qty <= 0 {
			qty = 1
		}

		// Per-unit price: use SalePrice if available, fallback to OriginalPrice
		price := item.SalePrice
		if price == 0 {
			price = item.OriginalPrice
		}
		unitPrice := math.Round(price / float64(qty))

		if _, exists := skuGroups[sku]; !exists {
			skuGroups[sku] = &skuGroup{
				data: NewSkuData(),
			}
			skuGroups[sku].info.productName = GetStringValue(item.ProductName)
		}

		g := skuGroups[sku]
		g.data.UnitPrices[unitPrice] = struct{}{}
		g.data.Count++

		// Actual income only for single-item orders: round(TotalSettlementAmount / quantity)
		itemCount := orderItemCount[item.EscrowOrderID]
		if order, ok := orderMap[item.EscrowOrderID]; ok {
			if itemCount == 1 && order.TotalSettlementAmount > 0 {
				unitActualIncome := math.Round(order.TotalSettlementAmount / float64(qty))
				g.data.ActualIncomes[unitActualIncome] = struct{}{}
			}
		}
	}

	// Build SKU group DTOs with inventory lookup
	result := make([]dto.TiktokSkuGroupDTO, 0, len(skuGroups))
	for sku, g := range skuGroups {
		unitPrices := MapKeysToSlice(g.data.UnitPrices)
		actualIncomes := MapKeysToSlice(g.data.ActualIncomes)

		invPrice, invName := LookupInventory(s.tenantDB, tenantID, sku, settings.PriceColumn)
		var inventoryPrice, expectedIncome *float64
		if invPrice > 0 {
			inventoryPrice = &invPrice
			expectedIncome = ComputeExpectedIncome(invPrice, settings.FormulaDeduction, settings.FormulaMultiplier)
		}

		productName := g.info.productName
		if invName != "" && productName == "" {
			productName = invName
		}

		hasMultiplePrices := len(unitPrices) > 1
		hasPriceDifference := inventoryPrice != nil && HasDifferentPrice(unitPrices, *inventoryPrice)

		status := DetermineReconciliationStatus(inventoryPrice, hasMultiplePrices, hasPriceDifference)

		result = append(result, dto.TiktokSkuGroupDTO{
			Sku:                 sku,
			SellerSku:           sku,
			ProductName:         productName,
			VariantName:         g.info.variantName,
			InventoryPrice:      inventoryPrice,
			ExpectedIncome:      expectedIncome,
			TotalTransactions:   g.data.Count,
			UniqueUnitPrices:    unitPrices,
			UniqueActualIncomes: actualIncomes,
			HasMultiplePrices:   hasMultiplePrices,
			HasPriceDifference:  hasPriceDifference,
			Status:              status,
		})
	}

	// Sort: non-OK first, then OK, then by transaction count descending
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			swap := false
			if result[i].Status == "OK" && result[j].Status != "OK" {
				swap = true
			} else if result[i].Status == result[j].Status && result[j].TotalTransactions > result[i].TotalTransactions {
				swap = true
			}
			if swap {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	// Build summary
	skuOk := 0
	skuWithPriceDiff := 0
	skuNoInventory := 0
	totalTransactions := 0
	for _, g := range result {
		totalTransactions += g.TotalTransactions
		switch g.Status {
		case "OK":
			skuOk++
		case "PRICE_DIFF":
			skuWithPriceDiff++
		case "NO_INVENTORY":
			skuNoInventory++
		}
	}

	return &dto.TiktokReconciliationResultDTO{
		Summary: dto.ReconciliationSummaryDTO{
			TotalSku:          len(result),
			TotalTransactions: totalTransactions,
			SkuOk:             skuOk,
			SkuWithPriceDiff:  skuWithPriceDiff,
			SkuNoInventory:    skuNoInventory,
		},
		SkuGroups: result,
	}, nil
}
