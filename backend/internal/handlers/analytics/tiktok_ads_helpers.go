package analytics

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// getDashboardSummaryDirect queries dashboard summary directly (fallback)
func (h *TiktokAdsHandler) getDashboardSummaryDirect(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
) struct {
	TotalCost        float64
	TotalRevenue     float64
	TotalOrders      int64
	TotalImpressions int64
	TotalClicks      int64
} {
	var result struct {
		TotalCost        float64 `gorm:"column:total_cost"`
		TotalRevenue     float64 `gorm:"column:total_revenue"`
		TotalOrders      int64   `gorm:"column:total_orders"`
		TotalImpressions int64   `gorm:"column:total_impressions"`
		TotalClicks      int64   `gorm:"column:total_clicks"`
	}

	db.WithContext(ctx).Model(&models.TiktokAdsCreativeData{}).
		Where("tenant_id = ?", tenantID).
		Select(`
			COALESCE(SUM(cost), 0) as total_cost,
			COALESCE(SUM(gross_revenue), 0) as total_revenue,
			COALESCE(SUM(orders_sku), 0) as total_orders,
			COALESCE(SUM(impressions), 0) as total_impressions,
			COALESCE(SUM(clicks), 0) as total_clicks
		`).Scan(&result)

	return struct {
		TotalCost        float64
		TotalRevenue     float64
		TotalOrders      int64
		TotalImpressions int64
		TotalClicks      int64
	}{
		TotalCost:        result.TotalCost,
		TotalRevenue:     result.TotalRevenue,
		TotalOrders:      result.TotalOrders,
		TotalImpressions: result.TotalImpressions,
		TotalClicks:      result.TotalClicks,
	}
}

// getTopProductsFromMV retrieves top products from Materialized View
func (h *TiktokAdsHandler) getTopProductsFromMV(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
) []gin.H {
	var products []struct {
		ProductID    string  `gorm:"column:product_id"`
		ProductName  string  `gorm:"column:product_name"`
		TotalCost    float64 `gorm:"column:total_cost"`
		TotalRevenue float64 `gorm:"column:total_revenue"`
		TotalOrders  int64   `gorm:"column:total_orders"`
		Roas         float64 `gorm:"column:roas"`
	}

	// Try MV first
	err := db.WithContext(ctx).Table("mv_ml_product_analysis").
		Where("tenant_id = ?", tenantID).
		Order("total_revenue DESC").
		Limit(10).
		Find(&products).Error

	if err != nil {
		// Fallback to direct query
		return h.getTopProductsDirect(ctx, db, tenantID)
	}

	result := h.enrichProductNames(ctx, db, tenantID, products)
	return result
}

// enrichProductNames looks up real product names from the tiktok_ads_product_names
// mapping table, falling back to the MV/raw product_name if no mapping exists.
func (h *TiktokAdsHandler) enrichProductNames(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
	products []struct {
		ProductID    string  `gorm:"column:product_id"`
		ProductName  string  `gorm:"column:product_name"`
		TotalCost    float64 `gorm:"column:total_cost"`
		TotalRevenue float64 `gorm:"column:total_revenue"`
		TotalOrders  int64   `gorm:"column:total_orders"`
		Roas         float64 `gorm:"column:roas"`
	},
) []gin.H {
	// Collect product IDs for batch lookup
	productIDs := make([]string, 0, len(products))
	for _, p := range products {
		if p.ProductID != "" {
			productIDs = append(productIDs, p.ProductID)
		}
	}

	// Batch lookup from product name mapping table
	nameMap := make(map[string]string)
	if len(productIDs) > 0 {
		var mappings []models.TiktokAdsProductName
		db.WithContext(ctx).
			Where("tenant_id = ? AND product_id IN ?", tenantID, productIDs).
			Find(&mappings)
		for _, m := range mappings {
			if m.Name != "" {
				nameMap[m.ProductID] = m.Name
			}
		}
	}

	var result []gin.H
	for _, p := range products {
		displayName := p.ProductName
		if mapped, ok := nameMap[p.ProductID]; ok {
			displayName = mapped
		}
		result = append(result, gin.H{
			"product_id":   p.ProductID,
			"product_name": displayName,
			"cost":         p.TotalCost,
			"revenue":      p.TotalRevenue,
			"orders":       p.TotalOrders,
			"roi":          p.Roas,
		})
	}
	return result
}

// getTopProductsDirect queries top products directly (fallback)
func (h *TiktokAdsHandler) getTopProductsDirect(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
) []gin.H {
	var products []struct {
		ProductID   string  `gorm:"column:product_id"`
		ProductName string  `gorm:"column:product_name"`
		Cost        float64 `gorm:"column:cost"`
		Revenue     float64 `gorm:"column:revenue"`
		Orders      int     `gorm:"column:orders"`
	}

	db.WithContext(ctx).Model(&models.TiktokAdsCreativeData{}).
		Where("tenant_id = ? AND product_id != '' AND product_id != '-1'", tenantID).
		Select(`product_id,
			MAX(COALESCE(NULLIF(product_name, ''), product_id)) as product_name,
			COALESCE(SUM(cost), 0) as cost, 
			COALESCE(SUM(gross_revenue), 0) as revenue, 
			COALESCE(SUM(orders_sku), 0) as orders`).
		Group("product_id").
		Order("revenue DESC").
		Limit(10).
		Find(&products)

	// Batch lookup product names from mapping table
	productIDs := make([]string, 0, len(products))
	for _, p := range products {
		productIDs = append(productIDs, p.ProductID)
	}
	nameMap := make(map[string]string)
	if len(productIDs) > 0 {
		var mappings []models.TiktokAdsProductName
		db.WithContext(ctx).
			Where("tenant_id = ? AND product_id IN ?", tenantID, productIDs).
			Find(&mappings)
		for _, m := range mappings {
			if m.Name != "" {
				nameMap[m.ProductID] = m.Name
			}
		}
	}

	var result []gin.H
	for _, p := range products {
		roi := float64(0)
		if p.Cost > 0 {
			roi = p.Revenue / p.Cost
		}
		displayName := p.ProductName
		if mapped, ok := nameMap[p.ProductID]; ok {
			displayName = mapped
		}
		result = append(result, gin.H{
			"product_id":   p.ProductID,
			"product_name": displayName,
			"cost":         p.Cost,
			"revenue":      p.Revenue,
			"orders":       p.Orders,
			"roi":          roi,
		})
	}
	return result
}

// getCreativeTypeStats retrieves creative type comparison stats
func (h *TiktokAdsHandler) getCreativeTypeStats(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
) []gin.H {
	var creativeTypes []struct {
		CreativeType string  `gorm:"column:creative_type"`
		Cost         float64 `gorm:"column:cost"`
		Revenue      float64 `gorm:"column:revenue"`
		Orders       int     `gorm:"column:orders"`
	}

	db.WithContext(ctx).Model(&models.TiktokAdsCreativeData{}).
		Where("tenant_id = ?", tenantID).
		Select(`creative_type, 
			COALESCE(SUM(cost), 0) as cost, 
			COALESCE(SUM(gross_revenue), 0) as revenue, 
			COALESCE(SUM(orders_sku), 0) as orders`).
		Group("creative_type").
		Find(&creativeTypes)

	var result []gin.H
	for _, ct := range creativeTypes {
		roi := float64(0)
		costPerOrder := float64(0)
		if ct.Cost > 0 {
			roi = ct.Revenue / ct.Cost
		}
		if ct.Orders > 0 {
			costPerOrder = ct.Cost / float64(ct.Orders)
		}
		result = append(result, gin.H{
			"creative_type":  ct.CreativeType,
			"cost":           ct.Cost,
			"revenue":        ct.Revenue,
			"orders":         ct.Orders,
			"roi":            roi,
			"cost_per_order": costPerOrder,
		})
	}
	return result
}
