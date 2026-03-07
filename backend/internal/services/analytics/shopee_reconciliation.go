// Package analytics provides reconciliation services for Shopee
package analytics

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ShopeeReconciliationService handles Shopee price reconciliation
type ShopeeReconciliationService struct {
	db       *gorm.DB
	tenantID string
	schema   string
}

// NewShopeeReconciliationService creates a new reconciliation service
func NewShopeeReconciliationService(db *gorm.DB, tenantID, schema string) *ShopeeReconciliationService {
	return &ShopeeReconciliationService{db: db, tenantID: tenantID, schema: schema}
}

// table returns table name with schema prefix
func (s *ShopeeReconciliationService) table(name string) string {
	return fmt.Sprintf("%s.%s", s.schema, name)
}

// getDB returns a DB session with context
func (s *ShopeeReconciliationService) getDB(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx)
}

// GetReconciliation analyzes price reconciliation for a month
func (s *ShopeeReconciliationService) GetReconciliation(ctx context.Context, month, year int, settings *dto.AnalyticsSettingsDTO) (*dto.ReconciliationResultDTO, error) {
	items, orderMap, orderItemCount, err := s.fetchEscrowItems(ctx, month, year)
	if err != nil {
		return nil, err
	}

	skuMap, skuInfoMap := s.groupItemsBySKU(items, orderMap, orderItemCount)
	skuGroups := s.buildSkuGroups(ctx, skuMap, skuInfoMap, settings)
	s.sortSkuGroups(skuGroups)

	return s.buildResult(skuGroups, len(items)), nil
}

// groupItemsBySKU groups items by SKU code
func (s *ShopeeReconciliationService) groupItemsBySKU(items []models.ShopeeEscrowItem, orderMap map[string]*models.ShopeeEscrowOrder, orderItemCount map[string]int) (map[string]*SkuData, map[string]*SkuInfo) {
	skuMap := make(map[string]*SkuData)
	skuInfoMap := make(map[string]*SkuInfo)

	for _, item := range items {
		sku := GetStringValue(item.ModelSku)
		if sku == "" {
			sku = GetStringValue(item.Sku)
		}
		if sku == "" {
			sku = "UNKNOWN"
		}

		quantity := item.Quantity
		if quantity == 0 {
			quantity = 1
		}
		unitPrice := math.Round(item.OriginalPrice / float64(quantity))

		if _, exists := skuMap[sku]; !exists {
			skuMap[sku] = NewSkuData()
		}

		data := skuMap[sku]
		data.UnitPrices[unitPrice] = struct{}{}
		data.Count++

		// Calculate actual income for SINGLE-ITEM orders only (like Node.js)
		itemCount := orderItemCount[item.EscrowOrderID]
		if order, ok := orderMap[item.EscrowOrderID]; ok {
			if itemCount == 1 && order.EscrowAmount > 0 {
				unitActualIncome := math.Round(order.EscrowAmount / float64(quantity))
				data.ActualIncomes[unitActualIncome] = struct{}{}
			}
		}

		if _, exists := skuInfoMap[sku]; !exists {
			skuInfoMap[sku] = &SkuInfo{
				ItemName:  GetStringValue(item.ItemName),
				ModelSku:  GetStringValue(item.ModelSku),
				ModelName: GetStringValue(item.ModelName),
			}
		}
	}
	return skuMap, skuInfoMap
}

// buildSkuGroups builds SKU group DTOs with inventory lookup
func (s *ShopeeReconciliationService) buildSkuGroups(ctx context.Context, skuMap map[string]*SkuData, skuInfoMap map[string]*SkuInfo, settings *dto.AnalyticsSettingsDTO) []dto.SkuGroupDTO {
	skuGroups := make([]dto.SkuGroupDTO, 0, len(skuMap))
	db := s.getDB(ctx)

	for sku, data := range skuMap {
		info := skuInfoMap[sku]
		if info == nil {
			info = &SkuInfo{ItemName: "Unknown", ModelSku: "", ModelName: ""}
		}

		unitPrices := MapKeysToSlice(data.UnitPrices)
		actualIncomes := MapKeysToSlice(data.ActualIncomes)
		sort.Float64s(unitPrices)
		sort.Float64s(actualIncomes)

		invPrice, invName := LookupInventory(db, s.schema, s.tenantID, sku, settings.PriceColumn)
		var inventoryPrice, expectedIncome *float64
		if invPrice > 0 {
			inventoryPrice = &invPrice
			exp := (invPrice - settings.FormulaDeduction) * settings.FormulaMultiplier
			expectedIncome = &exp
		}
		if invName != "" && info.ItemName == "" {
			info.ItemName = invName
		}

		hasMultiplePrices := len(unitPrices) > 1
		hasPriceDifference := inventoryPrice != nil && HasDifferentPrice(unitPrices, *inventoryPrice)

		status := s.determineStatus(inventoryPrice, hasMultiplePrices, hasPriceDifference)

		skuGroups = append(skuGroups, dto.SkuGroupDTO{
			Sku:                 sku,
			ModelSku:            info.ModelSku,
			ItemName:            info.ItemName,
			ModelName:           info.ModelName,
			InventoryPrice:      inventoryPrice,
			ExpectedIncome:      expectedIncome,
			TotalTransactions:   data.Count,
			UniqueUnitPrices:    unitPrices,
			UniqueActualIncomes: actualIncomes,
			HasMultiplePrices:   hasMultiplePrices,
			HasPriceDifference:  hasPriceDifference,
			Status:              status,
		})
	}
	return skuGroups
}

// determineStatus determines SKU reconciliation status
func (s *ShopeeReconciliationService) determineStatus(inventoryPrice *float64, hasMultiplePrices, hasPriceDifference bool) string {
	if inventoryPrice == nil {
		return "NO_INVENTORY"
	}
	if hasPriceDifference || hasMultiplePrices {
		return "PRICE_DIFF"
	}
	return "OK"
}

// sortSkuGroups sorts SKU groups by status and transaction count
func (s *ShopeeReconciliationService) sortSkuGroups(skuGroups []dto.SkuGroupDTO) {
	sort.Slice(skuGroups, func(i, j int) bool {
		if skuGroups[i].Status != "OK" && skuGroups[j].Status == "OK" {
			return true
		}
		if skuGroups[i].Status == "OK" && skuGroups[j].Status != "OK" {
			return false
		}
		return skuGroups[i].TotalTransactions > skuGroups[j].TotalTransactions
	})
}

// buildResult builds the final reconciliation result DTO
func (s *ShopeeReconciliationService) buildResult(skuGroups []dto.SkuGroupDTO, totalItems int) *dto.ReconciliationResultDTO {
	summary := dto.ReconciliationSummaryDTO{
		TotalSku:          len(skuGroups),
		TotalTransactions: totalItems,
	}
	for _, g := range skuGroups {
		switch g.Status {
		case "OK":
			summary.SkuOk++
		case "PRICE_DIFF":
			summary.SkuWithPriceDiff++
		case "NO_INVENTORY":
			summary.SkuNoInventory++
		}
	}
	return &dto.ReconciliationResultDTO{Summary: summary, SkuGroups: skuGroups}
}
