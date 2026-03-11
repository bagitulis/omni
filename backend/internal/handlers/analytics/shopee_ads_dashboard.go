package analytics

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// GetDashboard handles GET /api/analytics/shopee-ads/dashboard
// Uses Materialized View for fast response (mv_shopee_ads_summary)
func (h *AdsHandler) GetDashboard(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()

	// Try to use Materialized View first (fast path)
	var mvSummary struct {
		TotalCost        float64 `gorm:"column:total_cost"`
		TotalRevenue     float64 `gorm:"column:total_revenue"`
		TotalOrders      int64   `gorm:"column:total_conversions"`
		TotalImpressions int64   `gorm:"column:total_impressions"`
		TotalClicks      int64   `gorm:"column:total_clicks"`
		OverallRoas      float64 `gorm:"column:overall_roas"`
		OverallCtr       float64 `gorm:"column:overall_ctr"`
	}

	// Query from MV (should be < 5ms)
	mvErr := db.WithContext(ctx).Table("mv_shopee_ads_summary").
		Where("tenant_id = ?", tenantID).
		First(&mvSummary).Error

	var summary ShopeeDashboardSummary
	var avgRoas, avgCtr, avgConversionRate float64

	if mvErr == nil {
		// Use MV data
		summary.TotalCost = mvSummary.TotalCost
		summary.TotalRevenue = mvSummary.TotalRevenue
		summary.TotalOrders = mvSummary.TotalOrders
		summary.TotalImpressions = mvSummary.TotalImpressions
		summary.TotalClicks = mvSummary.TotalClicks
		avgRoas = mvSummary.OverallRoas
		avgCtr = mvSummary.OverallCtr
		if summary.TotalClicks > 0 {
			avgConversionRate = float64(summary.TotalOrders) / float64(summary.TotalClicks) * 100
		}
	} else {
		// Fallback to direct query if MV not available
		summary = getDashboardSummaryDirect(ctx, db, tenantID)
		if summary.TotalCost > 0 {
			avgRoas = summary.TotalRevenue / summary.TotalCost
		}
		if summary.TotalImpressions > 0 {
			avgCtr = float64(summary.TotalClicks) / float64(summary.TotalImpressions) * 100
		}
		if summary.TotalClicks > 0 {
			avgConversionRate = float64(summary.TotalOrders) / float64(summary.TotalClicks) * 100
		}
	}

	// Get top products from MV
	topProducts := getTopShopeeProducts(ctx, db, tenantID)

	// Get bidding mode stats
	biddingModeStats := getBiddingModeStats(ctx, db, tenantID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_cost":          summary.TotalCost,
			"total_revenue":       summary.TotalRevenue,
			"total_orders":        summary.TotalOrders,
			"avg_roas":            avgRoas,
			"total_impressions":   summary.TotalImpressions,
			"total_clicks":        summary.TotalClicks,
			"avg_ctr":             avgCtr,
			"avg_conversion_rate": avgConversionRate,
			"top_products":        topProducts,
			"bidding_mode_stats":  biddingModeStats,
		},
	})
}

// ShopeeDashboardSummary holds summary data for Shopee ads dashboard
type ShopeeDashboardSummary struct {
	TotalCost        float64
	TotalRevenue     float64
	TotalOrders      int64
	TotalImpressions int64
	TotalClicks      int64
}

// getDashboardSummaryDirect queries dashboard summary directly (fallback)
func getDashboardSummaryDirect(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
) ShopeeDashboardSummary {
	var result struct {
		TotalCost        float64 `gorm:"column:total_cost"`
		TotalRevenue     float64 `gorm:"column:total_revenue"`
		TotalOrders      int64   `gorm:"column:total_orders"`
		TotalImpressions int64   `gorm:"column:total_impressions"`
		TotalClicks      int64   `gorm:"column:total_clicks"`
	}

	db.WithContext(ctx).Model(&models.ShopeeAdsProductData{}).
		Where("tenant_id = ?", tenantID).
		Select(`
			COALESCE(SUM(cost), 0) as total_cost,
			COALESCE(SUM(revenue), 0) as total_revenue,
			COALESCE(SUM(conversions), 0) as total_orders,
			COALESCE(SUM(impressions), 0) as total_impressions,
			COALESCE(SUM(clicks), 0) as total_clicks
		`).Scan(&result)

	return ShopeeDashboardSummary{
		TotalCost:        result.TotalCost,
		TotalRevenue:     result.TotalRevenue,
		TotalOrders:      result.TotalOrders,
		TotalImpressions: result.TotalImpressions,
		TotalClicks:      result.TotalClicks,
	}
}

// getTopShopeeProducts retrieves top products from MV or direct query
func getTopShopeeProducts(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
) []gin.H {
	var products []struct {
		ProductID   string  `gorm:"column:product_id"`
		ProductName string  `gorm:"column:product_name"`
		TotalCost   float64 `gorm:"column:total_cost"`
		TotalRev    float64 `gorm:"column:total_revenue"`
		Conversions int64   `gorm:"column:total_conversions"`
		AvgRoas     float64 `gorm:"column:avg_roas"`
	}

	// Try MV first
	err := db.WithContext(ctx).Table("mv_shopee_ads_product_analysis").
		Where("tenant_id = ?", tenantID).
		Order("total_revenue DESC").
		Limit(10).
		Find(&products).Error

	if err != nil {
		// Fallback to direct query
		return getTopShopeeProductsDirect(ctx, db, tenantID)
	}

	var result []gin.H
	for _, p := range products {
		result = append(result, gin.H{
			"product_id":   p.ProductID,
			"product_name": p.ProductName,
			"cost":         p.TotalCost,
			"revenue":      p.TotalRev,
			"orders":       p.Conversions,
			"roas":         p.AvgRoas,
		})
	}
	return result
}

// getTopShopeeProductsDirect queries top products directly (fallback)
func getTopShopeeProductsDirect(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
) []gin.H {
	var products []struct {
		ProductID   string  `gorm:"column:product_id"`
		ProductName string  `gorm:"column:product_name"`
		Cost        float64 `gorm:"column:cost"`
		Revenue     float64 `gorm:"column:revenue"`
		Orders      int64   `gorm:"column:orders"`
	}

	db.WithContext(ctx).Model(&models.ShopeeAdsProductData{}).
		Where("tenant_id = ?", tenantID).
		Select(`product_id, product_name,
			COALESCE(SUM(cost), 0) as cost, 
			COALESCE(SUM(revenue), 0) as revenue, 
			COALESCE(SUM(conversions), 0) as orders`).
		Group("product_id, product_name").
		Order("revenue DESC").
		Limit(10).
		Find(&products)

	var result []gin.H
	for _, p := range products {
		roas := float64(0)
		if p.Cost > 0 {
			roas = p.Revenue / p.Cost
		}
		result = append(result, gin.H{
			"product_id":   p.ProductID,
			"product_name": p.ProductName,
			"cost":         p.Cost,
			"revenue":      p.Revenue,
			"orders":       p.Orders,
			"roas":         roas,
		})
	}
	return result
}

// getBiddingModeStats retrieves bidding mode comparison stats
func getBiddingModeStats(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
) []gin.H {
	var modes []struct {
		BiddingMode string  `gorm:"column:bidding_mode"`
		Cost        float64 `gorm:"column:cost"`
		Revenue     float64 `gorm:"column:revenue"`
		Orders      int64   `gorm:"column:orders"`
	}

	db.WithContext(ctx).Model(&models.ShopeeAdsProductData{}).
		Where("tenant_id = ?", tenantID).
		Select(`bidding_mode, 
			COALESCE(SUM(cost), 0) as cost, 
			COALESCE(SUM(revenue), 0) as revenue, 
			COALESCE(SUM(conversions), 0) as orders`).
		Group("bidding_mode").
		Find(&modes)

	var result []gin.H
	for _, m := range modes {
		roas := float64(0)
		costPerOrder := float64(0)
		if m.Cost > 0 {
			roas = m.Revenue / m.Cost
		}
		if m.Orders > 0 {
			costPerOrder = m.Cost / float64(m.Orders)
		}
		result = append(result, gin.H{
			"bidding_mode":   m.BiddingMode,
			"cost":           m.Cost,
			"revenue":        m.Revenue,
			"orders":         m.Orders,
			"roas":           roas,
			"cost_per_order": costPerOrder,
		})
	}
	return result
}
