package handlers

import "github.com/omni/backend/internal/services/sync"

// AggregatedLockedItem represents an aggregated locked order item
type AggregatedLockedItem struct {
	SKU           string   `json:"sku"`
	ProductName   string   `json:"product_name"`
	VariationName string   `json:"variation_name,omitempty"`
	Qty           int      `json:"qty"`
	Platforms     []string `json:"platforms"`
}

// aggregateLockedOrders aggregates orders by SKU + ProductName + Variation.
// Groups items and sums quantities, tracking which platforms they came from.
// NOTE: Orders are flattened (each row = one item), so we read from order fields directly.
func aggregateLockedOrders(unprocessOrders, processedOrders []sync.Order) []AggregatedLockedItem {
	// Map to aggregate by key (SKU|ProductName|Variation)
	aggregated := make(map[string]*AggregatedLockedItem)
	platformSets := make(map[string]map[string]bool)

	// Process all orders (each order is a flattened item row)
	allOrders := append(unprocessOrders, processedOrders...)

	for _, order := range allOrders {
		// In flattened format, item data is directly on the order object
		sku := order.SKU
		productName := order.ProductName
		variationName := order.VariationName
		qty := order.Quantity // This is the "qty" field in flattened format
		platform := order.Platform

		if sku == "" && productName == "" {
			continue // Skip empty items
		}

		// Ensure minimum qty of 1 if item exists
		if qty <= 0 {
			qty = 1
		}

		key := sku + "|" + productName + "|" + variationName

		if _, exists := aggregated[key]; !exists {
			aggregated[key] = &AggregatedLockedItem{
				SKU:           sku,
				ProductName:   productName,
				VariationName: variationName,
				Qty:           0,
				Platforms:     []string{},
			}
			platformSets[key] = make(map[string]bool)
		}

		aggregated[key].Qty += qty
		platformSets[key][platform] = true
	}

	// Convert map to slice and add platforms
	result := make([]AggregatedLockedItem, 0, len(aggregated))
	for key, item := range aggregated {
		platforms := make([]string, 0, len(platformSets[key]))
		for p := range platformSets[key] {
			platforms = append(platforms, p)
		}
		item.Platforms = platforms
		result = append(result, *item)
	}

	// Sort by qty descending
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Qty > result[i].Qty {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}
