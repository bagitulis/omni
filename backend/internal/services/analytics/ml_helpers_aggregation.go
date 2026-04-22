package analytics

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/ads"
)

// unifiedProductAgg holds aggregated data from both platforms
type unifiedProductAgg struct {
	ProductID    string
	ProductName  string
	SKU          string
	Platform     string // "tiktok", "shopee", or "combined"
	TotalCost    float64
	TotalRevenue float64
	TotalOrders  int
	Impressions  int
	Clicks       int
	PeriodCount  int
}

// aggregateTiktokProducts aggregates TikTok ads data into the product map
func (s *MLAnalyticsService) aggregateTiktokProducts(ctx context.Context, productMap map[string]*unifiedProductAgg) {
	type AggResult struct {
		ProductID    string  `gorm:"column:product_id"`
		ProductName  string  `gorm:"column:product_name"`
		TotalCost    float64 `gorm:"column:total_cost"`
		TotalRevenue float64 `gorm:"column:total_revenue"`
		TotalOrders  int     `gorm:"column:total_orders"`
		Impressions  int     `gorm:"column:impressions"`
		Clicks       int     `gorm:"column:clicks"`
		PeriodCount  int     `gorm:"column:period_count"`
	}

	var results []AggResult
	s.db.WithContext(ctx).
		Model(&models.TiktokAdsCreativeData{}).
		Select(`
			product_id,
			MAX(COALESCE(NULLIF(product_name, ''), product_id)) as product_name,
			SUM(cost) as total_cost,
			SUM(gross_revenue) as total_revenue,
			SUM(orders_sku) as total_orders,
			SUM(impressions) as impressions,
			SUM(clicks) as clicks,
			COUNT(DISTINCT period_label) as period_count
		`).
		Where("tenant_id = ? AND product_id != '' AND product_id != '-1'", s.tenantID).
		Group("product_id").
		Scan(&results)

	for _, r := range results {
		key := fmt.Sprintf("tiktok_%s", r.ProductID)
		productMap[key] = &unifiedProductAgg{
			ProductID: key, ProductName: r.ProductName, SKU: r.ProductID,
			Platform: "tiktok", TotalCost: r.TotalCost, TotalRevenue: r.TotalRevenue,
			TotalOrders: r.TotalOrders, Impressions: r.Impressions,
			Clicks: r.Clicks, PeriodCount: r.PeriodCount,
		}
	}
}

// aggregateShopeeProducts aggregates Shopee ads data into the product map
func (s *MLAnalyticsService) aggregateShopeeProducts(ctx context.Context, productMap map[string]*unifiedProductAgg) {
	type AggResult struct {
		ProductID    string  `gorm:"column:product_id"`
		ProductName  string  `gorm:"column:product_name"`
		TotalCost    float64 `gorm:"column:total_cost"`
		TotalRevenue float64 `gorm:"column:total_revenue"`
		TotalOrders  int     `gorm:"column:total_orders"`
		Impressions  int     `gorm:"column:impressions"`
		Clicks       int     `gorm:"column:clicks"`
		PeriodCount  int     `gorm:"column:period_count"`
	}

	var results []AggResult
	s.db.WithContext(ctx).
		Model(&models.ShopeeAdsProductData{}).
		Select(`
			product_id,
			MAX(COALESCE(NULLIF(product_name, ''), product_id)) as product_name,
			SUM(cost) as total_cost,
			SUM(revenue) as total_revenue,
			SUM(units_sold) as total_orders,
			SUM(impressions) as impressions,
			SUM(clicks) as clicks,
			COUNT(DISTINCT period_label) as period_count
		`).
		Where("tenant_id = ? AND product_id != ''", s.tenantID).
		Group("product_id").
		Scan(&results)

	for _, r := range results {
		key := fmt.Sprintf("shopee_%s", r.ProductID)
		productMap[key] = &unifiedProductAgg{
			ProductID: key, ProductName: r.ProductName, SKU: r.ProductID,
			Platform: "shopee", TotalCost: r.TotalCost, TotalRevenue: r.TotalRevenue,
			TotalOrders: r.TotalOrders, Impressions: r.Impressions,
			Clicks: r.Clicks, PeriodCount: r.PeriodCount,
		}
	}
}

// getHistoricalByProduct retrieves historical ROI and profit per TikTok product (last 90 days)
func (s *MLAnalyticsService) getHistoricalByProduct(ctx context.Context) map[string]ads.HistoricalValues {
	type PeriodData struct {
		ProductID   string  `gorm:"column:product_id"`
		PeriodStart string  `gorm:"column:period_start"`
		Cost        float64 `gorm:"column:cost"`
		Revenue     float64 `gorm:"column:revenue"`
	}

	var periodData []PeriodData
	s.db.WithContext(ctx).
		Model(&models.TiktokAdsCreativeData{}).
		Select("product_id, period_start, SUM(cost) as cost, SUM(gross_revenue) as revenue").
		Where("tenant_id = ?", s.tenantID).
		Where("period_start >= NOW() - INTERVAL '90 days'").
		Group("product_id, period_start").
		Order("period_start ASC").
		Scan(&periodData)

	result := make(map[string]ads.HistoricalValues)
	for _, pd := range periodData {
		roi := 0.0
		if pd.Cost > 0 {
			roi = pd.Revenue / pd.Cost
		}
		key := fmt.Sprintf("tiktok_%s", pd.ProductID)
		hist := result[key]
		hist.ROIValues = append(hist.ROIValues, roi)
		hist.ProfitValues = append(hist.ProfitValues, pd.Revenue-pd.Cost)
		result[key] = hist
	}
	return result
}

// getShopeeHistoricalByProduct retrieves historical ROI and profit per Shopee product (last 90 days)
func (s *MLAnalyticsService) getShopeeHistoricalByProduct(ctx context.Context) map[string]ads.HistoricalValues {
	type PeriodData struct {
		ProductID   string  `gorm:"column:product_id"`
		PeriodStart string  `gorm:"column:period_start"`
		Cost        float64 `gorm:"column:cost"`
		Revenue     float64 `gorm:"column:revenue"`
	}

	var periodData []PeriodData
	s.db.WithContext(ctx).
		Model(&models.ShopeeAdsProductData{}).
		Select("product_id, period_start, SUM(cost) as cost, SUM(revenue) as revenue").
		Where("tenant_id = ?", s.tenantID).
		Where("period_start >= NOW() - INTERVAL '90 days'").
		Group("product_id, period_start").
		Order("period_start ASC").
		Scan(&periodData)

	result := make(map[string]ads.HistoricalValues)
	for _, pd := range periodData {
		roi := 0.0
		if pd.Cost > 0 {
			roi = pd.Revenue / pd.Cost
		}
		key := fmt.Sprintf("shopee_%s", pd.ProductID)
		hist := result[key]
		hist.ROIValues = append(hist.ROIValues, roi)
		hist.ProfitValues = append(hist.ProfitValues, pd.Revenue-pd.Cost)
		result[key] = hist
	}
	return result
}


