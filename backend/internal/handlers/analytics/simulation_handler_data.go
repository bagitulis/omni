package analytics

import (
	"context"
	"math"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/analytics/intelligence"
	"gorm.io/gorm"
)

// periodAgg holds per-period aggregated data with actual date ranges
type periodAgg struct {
	PeriodLabel string    `gorm:"column:period_label"`
	PeriodStart time.Time `gorm:"column:period_start"`
	PeriodEnd   time.Time `gorm:"column:period_end"`
	Spend       float64   `gorm:"column:spend"`
	Revenue     float64   `gorm:"column:revenue"`
	Clicks      int       `gorm:"column:clicks"`
	Orders      int       `gorm:"column:orders"`
	Impressions int       `gorm:"column:impressions"`
}

// buildHistorical converts period aggregations into ProductHistoricalData
// with daily normalization based on actual date span (min start -> max end).
func buildHistorical(productID string, periods []periodAgg) intelligence.ProductHistoricalData {
	result := intelligence.ProductHistoricalData{ProductID: productID}

	if len(periods) == 0 {
		return result
	}

	var totalSpend, totalRevenue float64
	var totalClicks, totalOrders, totalImpressions int
	var minStart, maxEnd time.Time

	for _, p := range periods {
		periodDays := math.Max(1.0, p.PeriodEnd.Sub(p.PeriodStart).Hours()/24.0)
		result.SpendHistory = append(result.SpendHistory, p.Spend/periodDays)
		result.RevenueHistory = append(result.RevenueHistory, p.Revenue/periodDays)
		roas := 0.0
		if p.Spend > 0 {
			roas = p.Revenue / p.Spend
		}
		result.RoasHistory = append(result.RoasHistory, roas)

		totalSpend += p.Spend
		totalRevenue += p.Revenue
		totalClicks += p.Clicks
		totalOrders += p.Orders
		totalImpressions += p.Impressions

		if minStart.IsZero() || p.PeriodStart.Before(minStart) {
			minStart = p.PeriodStart
		}
		if p.PeriodEnd.After(maxEnd) {
			maxEnd = p.PeriodEnd
		}
	}

	totalActualDays := math.Max(1.0, maxEnd.Sub(minStart).Hours()/24.0)
	result.DaysOfData = int(math.Round(totalActualDays))

	if totalSpend > 0 {
		result.CurrentRoas = totalRevenue / totalSpend
	}
	result.CurrentSpend = totalSpend / totalActualDays
	result.TotalClicks = totalClicks
	result.TotalOrders = totalOrders
	result.TotalImpressions = totalImpressions

	if totalImpressions > 0 && totalClicks > 0 {
		result.AvgCTR = float64(totalClicks) / float64(totalImpressions) * 100
	}
	if totalClicks > 0 && totalOrders > 0 {
		result.AvgCVR = float64(totalOrders) / float64(totalClicks) * 100
	}
	if totalClicks > 0 && totalSpend > 0 {
		result.AvgCPC = totalSpend / float64(totalClicks)
	}

	return result
}

// getProductHistoricalData retrieves historical data for a product
func (h *SimulationHandler) getProductHistoricalData(
	ctx context.Context, db *gorm.DB, tenantID, productID string,
) intelligence.ProductHistoricalData {
	tiktokData := h.getTiktokProductData(ctx, db, tenantID, productID)
	if len(tiktokData.RoasHistory) >= 3 {
		return tiktokData
	}

	shopeeData := h.getShopeeProductData(ctx, db, tenantID, productID)
	if len(shopeeData.RoasHistory) >= 3 {
		return shopeeData
	}

	if len(tiktokData.RoasHistory) > len(shopeeData.RoasHistory) {
		return tiktokData
	}
	return shopeeData
}

// getTiktokProductData gets TikTok product historical data per period
func (h *SimulationHandler) getTiktokProductData(
	ctx context.Context, db *gorm.DB, tenantID, productID string,
) intelligence.ProductHistoricalData {
	var periods []periodAgg
	db.WithContext(ctx).Model(&models.TiktokAdsCreativeData{}).
		Where("tenant_id = ? AND product_id = ?", tenantID, productID).
		Select(`period_label, MIN(period_start) as period_start, MAX(period_end) as period_end,
			COALESCE(SUM(cost), 0) as spend, COALESCE(SUM(gross_revenue), 0) as revenue,
			COALESCE(SUM(clicks), 0) as clicks, COALESCE(SUM(orders_sku), 0) as orders,
			COALESCE(SUM(impressions), 0) as impressions`).
		Group("period_label").Order("period_label ASC").Scan(&periods)
	return buildHistorical(productID, periods)
}

// getShopeeProductData gets Shopee product historical data per period
func (h *SimulationHandler) getShopeeProductData(
	ctx context.Context, db *gorm.DB, tenantID, productID string,
) intelligence.ProductHistoricalData {
	var periods []periodAgg
	db.WithContext(ctx).Model(&models.ShopeeAdsProductData{}).
		Where("tenant_id = ? AND product_id = ?", tenantID, productID).
		Select(`period_label, MIN(period_start) as period_start, MAX(period_end) as period_end,
			COALESCE(SUM(cost), 0) as spend, COALESCE(SUM(revenue), 0) as revenue,
			COALESCE(SUM(clicks), 0) as clicks, COALESCE(SUM(conversions), 0) as orders,
			COALESCE(SUM(impressions), 0) as impressions`).
		Group("period_label").Order("period_label ASC").Scan(&periods)
	return buildHistorical(productID, periods)
}
