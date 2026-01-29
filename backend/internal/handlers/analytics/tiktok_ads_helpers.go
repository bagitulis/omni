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

	var result []gin.H
	for _, p := range products {
		result = append(result, gin.H{
			"product_id":   p.ProductID,
			"product_name": p.ProductName,
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
		ProductID string  `gorm:"column:product_id"`
		Cost      float64 `gorm:"column:cost"`
		Revenue   float64 `gorm:"column:revenue"`
		Orders    int     `gorm:"column:orders"`
	}

	db.WithContext(ctx).Model(&models.TiktokAdsCreativeData{}).
		Where("tenant_id = ?", tenantID).
		Select(`product_id, 
			COALESCE(SUM(cost), 0) as cost, 
			COALESCE(SUM(gross_revenue), 0) as revenue, 
			COALESCE(SUM(orders_sku), 0) as orders`).
		Group("product_id").
		Order("revenue DESC").
		Limit(10).
		Find(&products)

	var result []gin.H
	for _, p := range products {
		roi := float64(0)
		if p.Cost > 0 {
			roi = p.Revenue / p.Cost
		}
		result = append(result, gin.H{
			"product_id": p.ProductID,
			"cost":       p.Cost,
			"revenue":    p.Revenue,
			"orders":     p.Orders,
			"roi":        roi,
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
