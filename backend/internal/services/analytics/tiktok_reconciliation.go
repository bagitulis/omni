// Package analytics provides reconciliation services for TikTok
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

// TiktokReconciliationService handles TikTok reconciliation operations
type TiktokReconciliationService struct {
	db       *gorm.DB
	tenantID string
	schema   string
}

// NewTiktokReconciliationService creates a new TikTok reconciliation service
func NewTiktokReconciliationService(db *gorm.DB, tenantID, schema string) *TiktokReconciliationService {
	return &TiktokReconciliationService{db: db, tenantID: tenantID, schema: schema}
}

// table returns table name with schema prefix
func (s *TiktokReconciliationService) table(name string) string {
	return fmt.Sprintf("%s.%s", s.schema, name)
}

// getDB returns a DB session with context
func (s *TiktokReconciliationService) getDB(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx)
}

// tiktokSkuLookup holds resolved SKU data from the tiktok_skus table
type tiktokSkuLookup struct {
	SellerSku   string
	VariantName string
}

// loadSkuLookupMap pre-fetches all tiktok_skus for this tenant, keyed by sku_id.
// This is used to resolve the actual seller_sku when the escrow item has
// variant names stored in seller_sku (legacy bug from Finance API sku_name mapping).
func (s *TiktokReconciliationService) loadSkuLookupMap(ctx context.Context) map[string]tiktokSkuLookup {
	var skus []models.TiktokSku
	s.getDB(ctx).Table(s.table("tiktok_skus")).
		Where("tenant_id = ?", s.tenantID).
		Find(&skus)

	result := make(map[string]tiktokSkuLookup, len(skus))
	for _, sk := range skus {
		result[sk.SkuID] = tiktokSkuLookup{
			SellerSku:   sk.SellerSku,
			VariantName: sk.VariantName,
		}
	}
	return result
}

// GetReconciliation analyzes price reconciliation for TikTok
func (s *TiktokReconciliationService) GetReconciliation(ctx context.Context, month, year int, settings *dto.AnalyticsSettingsDTO) (*dto.ReconciliationResultDTO, error) {
	items, orderMap, orderItemCount, err := s.fetchEscrowItems(ctx, month, year)
	if err != nil {
		return nil, err
	}

	// Pre-load SKU lookup map for resolving correct seller_sku
	skuLookup := s.loadSkuLookupMap(ctx)

	skuMap, skuInfoMap := s.groupItemsBySKU(items, orderMap, orderItemCount, skuLookup)
	skuGroups := s.buildSkuGroups(ctx, skuMap, skuInfoMap, settings)
	s.sortSkuGroups(skuGroups)

	summary := dto.ReconciliationSummaryDTO{TotalSku: len(skuGroups), TotalTransactions: len(items)}
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

	return &dto.ReconciliationResultDTO{Summary: summary, SkuGroups: skuGroups}, nil
}

// fetchEscrowItems retrieves escrow items for a month with order data
func (s *TiktokReconciliationService) fetchEscrowItems(ctx context.Context, month, year int) ([]models.TiktokEscrowItem, map[string]*models.TiktokEscrowOrder, map[string]int, error) {
	var items []models.TiktokEscrowItem
	itemsTable := s.table("tiktok_escrow_items")
	ordersTable := s.table("tiktok_escrow_orders")
	err := s.getDB(ctx).Table(itemsTable).
		Select(itemsTable+".*").
		Joins(fmt.Sprintf("JOIN %s ON %s.id = %s.escrow_order_id", ordersTable, ordersTable, itemsTable)).
		Where(fmt.Sprintf("%s.tenant_id = ?", itemsTable), s.tenantID).
		Where(fmt.Sprintf("%s.month = ? AND %s.year = ?", ordersTable, ordersTable), month, year).
		Find(&items).Error
	if err != nil {
		return nil, nil, nil, err
	}

	// Fetch all orders to get TotalSettlementAmount
	var orders []models.TiktokEscrowOrder
	err = s.getDB(ctx).Table(ordersTable).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		Find(&orders).Error
	if err != nil {
		return nil, nil, nil, err
	}

	// Build order map
	orderMap := make(map[string]*models.TiktokEscrowOrder)
	for i := range orders {
		orderMap[orders[i].ID] = &orders[i]
	}

	// Count items per order
	orderItemCount := make(map[string]int)
	for _, item := range items {
		orderItemCount[item.EscrowOrderID]++
	}

	return items, orderMap, orderItemCount, nil
}

// resolveSkuAndVariant resolves the actual seller_sku and variant name for an escrow item.
// For old data where seller_sku contains a variant name (e.g. "Hitam"),
// it looks up the correct seller_sku from the tiktok_skus table using sku_id.
func resolveSkuAndVariant(item models.TiktokEscrowItem, skuLookup map[string]tiktokSkuLookup) (sku, variantName string) {
	rawSku := GetStringValue(item.SellerSku)
	skuID := GetStringValue(item.SkuID)
	variantName = ""

	// Try to resolve from tiktok_skus table using sku_id
	if skuID != "" {
		if lookup, ok := skuLookup[skuID]; ok && lookup.SellerSku != "" {
			// If stored seller_sku differs from the real one, it's likely a variant name
			if rawSku != "" && rawSku != lookup.SellerSku {
				variantName = rawSku // The "wrong" value is actually the variant name
			}
			if variantName == "" {
				variantName = lookup.VariantName
			}
			return lookup.SellerSku, variantName
		}
	}

	// Fallback: use whatever is stored
	sku = rawSku
	if sku == "" {
		sku = skuID
	}
	if sku == "" {
		sku = "UNKNOWN"
	}
	return sku, variantName
}

// groupItemsBySKU groups items by SKU, resolving actual seller_sku from tiktok_skus
func (s *TiktokReconciliationService) groupItemsBySKU(
	items []models.TiktokEscrowItem,
	orderMap map[string]*models.TiktokEscrowOrder,
	orderItemCount map[string]int,
	skuLookup map[string]tiktokSkuLookup,
) (map[string]*SkuData, map[string]*SkuInfo) {
	skuMap := make(map[string]*SkuData)
	skuInfoMap := make(map[string]*SkuInfo)

	for _, item := range items {
		sku, variantName := resolveSkuAndVariant(item, skuLookup)

		quantity := item.Quantity
		if quantity == 0 {
			quantity = 1
		}

		// Calculate per-unit price (use SalePrice if available, fallback to OriginalPrice)
		price := item.SalePrice
		if price == 0 {
			price = item.OriginalPrice
		}
		unitPrice := math.Round(price / float64(quantity))

		if _, exists := skuMap[sku]; !exists {
			skuMap[sku] = &SkuData{
				UnitPrices:    make(map[float64]struct{}),
				ActualIncomes: make(map[float64]struct{}),
			}
		}

		data := skuMap[sku]
		data.UnitPrices[unitPrice] = struct{}{}
		data.Count++

		// Calculate actual income for SINGLE-ITEM orders only (like Node.js)
		itemCount := orderItemCount[item.EscrowOrderID]
		if order, ok := orderMap[item.EscrowOrderID]; ok {
			if itemCount == 1 && order.TotalSettlementAmount > 0 {
				unitActualIncome := math.Round(order.TotalSettlementAmount / float64(quantity))
				data.ActualIncomes[unitActualIncome] = struct{}{}
			}
		}

		if _, exists := skuInfoMap[sku]; !exists {
			skuInfoMap[sku] = &SkuInfo{
				ItemName:  GetStringValue(item.ProductName),
				ModelSku:  sku,
				ModelName: variantName,
			}
		}
	}

	return skuMap, skuInfoMap
}

// buildSkuGroups builds SKU groups from mapped data
func (s *TiktokReconciliationService) buildSkuGroups(ctx context.Context, skuMap map[string]*SkuData, skuInfoMap map[string]*SkuInfo, settings *dto.AnalyticsSettingsDTO) []dto.SkuGroupDTO {
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

		var status string
		if inventoryPrice == nil {
			status = "NO_INVENTORY"
		} else if hasPriceDifference || hasMultiplePrices {
			status = "PRICE_DIFF"
		} else {
			status = "OK"
		}

		skuGroups = append(skuGroups, dto.SkuGroupDTO{
			Sku:                 sku,
			ModelSku:            info.ModelSku,
			ItemName:            info.ItemName,
			ModelName:           info.ModelName,
			VariantName:         info.ModelName, // Variation display name
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

// sortSkuGroups sorts groups by status and transaction count
func (s *TiktokReconciliationService) sortSkuGroups(skuGroups []dto.SkuGroupDTO) {
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

