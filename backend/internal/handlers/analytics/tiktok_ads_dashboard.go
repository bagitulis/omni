package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
)

// GetDashboard handles GET /api/analytics/tiktok-ads/dashboard
// Uses Materialized View for fast response (mv_tiktok_ads_summary)
func (h *TiktokAdsHandler) GetDashboard(c *gin.Context) {
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
		TotalOrders      int64   `gorm:"column:total_orders"`
		TotalImpressions int64   `gorm:"column:total_impressions"`
		TotalClicks      int64   `gorm:"column:total_clicks"`
		OverallRoas      float64 `gorm:"column:overall_roas"`
		AvgCtr           float64 `gorm:"column:avg_ctr"`
	}

	// Query from MV (should be < 5ms)
	mvErr := db.WithContext(ctx).Table("mv_tiktok_ads_summary").
		Where("tenant_id = ?", tenantID).
		First(&mvSummary).Error

	var summary struct {
		TotalCost        float64
		TotalRevenue     float64
		TotalOrders      int64
		TotalImpressions int64
		TotalClicks      int64
	}
	var avgRoi, avgCtr, avgConversionRate float64

	if mvErr == nil {
		// Use MV data
		summary.TotalCost = mvSummary.TotalCost
		summary.TotalRevenue = mvSummary.TotalRevenue
		summary.TotalOrders = mvSummary.TotalOrders
		summary.TotalImpressions = mvSummary.TotalImpressions
		summary.TotalClicks = mvSummary.TotalClicks
		avgRoi = mvSummary.OverallRoas
		avgCtr = mvSummary.AvgCtr
		if summary.TotalClicks > 0 {
			avgConversionRate = float64(summary.TotalOrders) / float64(summary.TotalClicks) * 100
		}
	} else {
		// Fallback to direct query if MV not available
		summary = h.getDashboardSummaryDirect(ctx, db, tenantID)
		if summary.TotalCost > 0 {
			avgRoi = summary.TotalRevenue / summary.TotalCost
		}
		if summary.TotalImpressions > 0 {
			avgCtr = float64(summary.TotalClicks) / float64(summary.TotalImpressions) * 100
		}
		if summary.TotalClicks > 0 {
			avgConversionRate = float64(summary.TotalOrders) / float64(summary.TotalClicks) * 100
		}
	}

	// Get top products from MV
	topProducts := h.getTopProductsFromMV(ctx, db, tenantID)

	// Get creative type stats
	creativeTypeStats := h.getCreativeTypeStats(ctx, db, tenantID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_cost":               summary.TotalCost,
			"total_revenue":            summary.TotalRevenue,
			"total_orders":             summary.TotalOrders,
			"avg_roi":                  avgRoi,
			"total_impressions":        summary.TotalImpressions,
			"total_clicks":             summary.TotalClicks,
			"avg_ctr":                  avgCtr,
			"avg_conversion_rate":      avgConversionRate,
			"top_products":             topProducts,
			"creative_type_comparison": creativeTypeStats,
		},
	})
}
