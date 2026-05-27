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
func (s *ShopeeAnalyticsService) GetReconciliation(ctx context.Context, tenantID string, month, year int) (*dto.ReconciliationResultDTO, error) {
	settings, err := s.GetSettings(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	// Fetch items and orders for the period
	var items []models.ShopeeEscrowItem
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to query escrow items: %w", err)
	}

	var orders []models.ShopeeEscrowOrder
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to query escrow orders: %w", err)
	}

	// Build order map and item count per order
	orderMap := make(map[string]*models.ShopeeEscrowOrder)
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
		info *SkuInfo
	}
	skuGroups := make(map[string]*skuGroup)

	for _, item := range items {
		sku := GetStringValue(item.ModelSku)
		if sku == "" {
			sku = GetStringValue(item.Sku)
		}
		if sku == "" {
			sku = "UNKNOWN"
		}

		qty := item.Quantity
		if qty <= 0 {
			qty = 1
		}

		// Per-unit price calculation: round(OriginalPrice / quantity)
		unitPrice := math.Round(item.OriginalPrice / float64(qty))

		if _, exists := skuGroups[sku]; !exists {
			skuGroups[sku] = &skuGroup{
				data: NewSkuData(),
				info: &SkuInfo{
					ItemName:  GetStringValue(item.ItemName),
					ModelSku:  sku,
					ModelName: GetStringValue(item.ModelName),
				},
			}
		}

		g := skuGroups[sku]
		g.data.UnitPrices[unitPrice] = struct{}{}
		g.data.Count++

		// Actual income only for single-item orders: round(EscrowAmount / quantity)
		itemCount := orderItemCount[item.EscrowOrderID]
		if order, ok := orderMap[item.EscrowOrderID]; ok {
			if itemCount == 1 && order.EscrowAmount > 0 {
				unitActualIncome := math.Round(order.EscrowAmount / float64(qty))
				g.data.ActualIncomes[unitActualIncome] = struct{}{}
			}
		}
	}

	// Build SKU group DTOs with inventory lookup
	result := make([]dto.SkuGroupDTO, 0, len(skuGroups))
	for sku, g := range skuGroups {
		unitPrices := MapKeysToSlice(g.data.UnitPrices)
		actualIncomes := MapKeysToSlice(g.data.ActualIncomes)

		invPrice, invName := LookupInventory(s.tenantDB, tenantID, sku, settings.PriceColumn)
		var inventoryPrice, expectedIncome *float64
		if invPrice > 0 {
			inventoryPrice = &invPrice
			expectedIncome = ComputeExpectedIncome(invPrice, settings.FormulaDeduction, settings.FormulaMultiplier)
		}

		itemName := g.info.ItemName
		if invName != "" && itemName == "" {
			itemName = invName
		}

		hasMultiplePrices := len(unitPrices) > 1
		hasPriceDifference := inventoryPrice != nil && HasDifferentPrice(unitPrices, *inventoryPrice)

		status := DetermineReconciliationStatus(inventoryPrice, hasMultiplePrices, hasPriceDifference)

		result = append(result, dto.SkuGroupDTO{
			Sku:                 sku,
			ModelSku:            g.info.ModelSku,
			ItemName:            itemName,
			ModelName:           g.info.ModelName,
			VariantName:         g.info.ModelName,
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

	// Sort: non-OK first (priority NO_INVENTORY, PRICE_DIFF), then OK, then by transaction count descending
	// This matches historical behavior
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

	return &dto.ReconciliationResultDTO{
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
