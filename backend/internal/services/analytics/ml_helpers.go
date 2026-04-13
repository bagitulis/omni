package analytics

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/ads"
)

// getProductAnalyses aggregates and analyzes products from BOTH TikTok and Shopee ads
func (s *MLAnalyticsService) getProductAnalyses(
	ctx context.Context,
	platform string,
	limit int,
	cursor string,
) ([]models.MLProductAnalysis, error) {

	// Aggregate data from both platforms into unified map
	productMap := make(map[string]*unifiedProductAgg)

	// 1. Aggregate TikTok ads data
	if platform == "" || platform == "tiktok" {
		s.aggregateTiktokProducts(ctx, productMap)
	}

	// 2. Aggregate Shopee ads data
	if platform == "" || platform == "shopee" {
		s.aggregateShopeeProducts(ctx, productMap)
	}

	// Get historical data for trend analysis (both platforms)
	historical := s.getHistoricalByProduct(ctx)
	shopeeHistorical := s.getShopeeHistoricalByProduct(ctx)

	// Merge shopee historical into main historical map
	for k, v := range shopeeHistorical {
		if existing, ok := historical[k]; ok {
			existing.ROIValues = append(existing.ROIValues, v.ROIValues...)
			existing.ProfitValues = append(existing.ProfitValues, v.ProfitValues...)
			historical[k] = existing
		} else {
			historical[k] = v
		}
	}

	// Build product analyses
	analyses := make([]models.MLProductAnalysis, 0, len(productMap))
	cfg := ads.DefaultScoringConfig()

	for _, r := range productMap {
		if r.ProductID == "" {
			continue
		}

		analysis := models.MLProductAnalysis{
			TenantID:    s.tenantID,
			ProductID:   r.ProductID,
			ProductName: r.ProductName,
			SKU:         r.SKU,
			Platform:    r.Platform,
			TotalCost:   r.TotalCost,
			TotalRevenue: r.TotalRevenue,
			TotalProfit: r.TotalRevenue - r.TotalCost,
			TotalOrders: r.TotalOrders,
			Clicks:      r.Clicks,
			Impressions: r.Impressions,
			PeriodCount: r.PeriodCount,
		}

		// Calculate CTR
		if r.Impressions > 0 {
			analysis.CTR = float64(r.Clicks) / float64(r.Impressions)
		}

		// Calculate ROAS
		if r.TotalCost > 0 {
			analysis.ROAS = roundTo(r.TotalRevenue/r.TotalCost, 2)
		}

		// Calculate scores using existing scoring system
		var score ads.ScoreResult
		if hist, ok := historical[r.ProductID]; ok && len(hist.ROIValues) >= 2 {
			score = ads.CalculateFullScore(cfg, analysis.ROAS, analysis.TotalProfit,
				hist.ROIValues, hist.ProfitValues)
		} else {
			score = ads.SimpleScore(analysis.ROAS, analysis.TotalProfit)
		}

		// Map to unified score
		analysis.UnifiedScore = score.CompositeScore
		analysis.ROASScore = score.ROIScore
		analysis.TrendScore = score.TrendScore
		analysis.MomentumScore = score.MomentumScore
		analysis.VolatilityScore = score.ConsistencyScore

		// Map category
		analysis.Category = mapToCategory(score.Category)
		analysis.Action = score.Category
		analysis.ActionLabel = score.Action
		analysis.BudgetChangePct = getBudgetChangePct(score.Category)

		// Generate AI recommendation text
		analysis.Recommendation = generateRecommendation(analysis.Category, analysis.Action, analysis.ROAS, analysis.Platform)

		// Determine fatigue and churn
		analysis.FatigueStatus = determineFatigueStatus(r.Clicks, r.Impressions, r.PeriodCount)
		analysis.HasFatigueWarning = analysis.FatigueStatus == "FATIGUED" || analysis.FatigueStatus == "DEAD"

		analysis.ChurnRiskScore = calculateChurnRisk(analysis.MomentumScore, analysis.TrendScore, analysis.VolatilityScore)
		analysis.HasChurnRisk = analysis.ChurnRiskScore > 60

		// Confidence level
		analysis.ConfidenceLevel = getConfidenceLevel(r.PeriodCount)
		analysis.SuccessProbability = estimateSuccessProbability(analysis.UnifiedScore, analysis.ROAS)

		// Trend direction
		analysis.TrendDirection = getTrendDirection(analysis.MomentumScore)
		analysis.TrendStrength = math.Abs(analysis.MomentumScore-50) / 50

		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

// unifiedProductAgg holds aggregated data from both platforms
type unifiedProductAgg struct {
	ProductID   string
	ProductName string
	SKU         string
	Platform    string // "tiktok", "shopee", or "combined"
	TotalCost   float64
	TotalRevenue float64
	TotalOrders int
	Impressions int
	Clicks      int
	PeriodCount int
}

// aggregateTiktokProducts aggregates TikTok ads data into the product map
func (s *MLAnalyticsService) aggregateTiktokProducts(ctx context.Context, productMap map[string]*unifiedProductAgg) {
	type AggResult struct {
		ProductID   string  `gorm:"column:product_id"`
		ProductName string  `gorm:"column:product_name"`
		TotalCost   float64 `gorm:"column:total_cost"`
		TotalRevenue float64 `gorm:"column:total_revenue"`
		TotalOrders int     `gorm:"column:total_orders"`
		Impressions int     `gorm:"column:impressions"`
		Clicks      int     `gorm:"column:clicks"`
		PeriodCount int     `gorm:"column:period_count"`
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
			ProductID:    key,
			ProductName:  r.ProductName,
			SKU:          r.ProductID,
			Platform:     "tiktok",
			TotalCost:    r.TotalCost,
			TotalRevenue: r.TotalRevenue,
			TotalOrders:  r.TotalOrders,
			Impressions:  r.Impressions,
			Clicks:       r.Clicks,
			PeriodCount:  r.PeriodCount,
		}
	}
}

// aggregateShopeeProducts aggregates Shopee ads data into the product map
func (s *MLAnalyticsService) aggregateShopeeProducts(ctx context.Context, productMap map[string]*unifiedProductAgg) {
	type AggResult struct {
		ProductID   string  `gorm:"column:product_id"`
		ProductName string  `gorm:"column:product_name"`
		TotalCost   float64 `gorm:"column:total_cost"`
		TotalRevenue float64 `gorm:"column:total_revenue"`
		TotalOrders int     `gorm:"column:total_orders"`
		Impressions int     `gorm:"column:impressions"`
		Clicks      int     `gorm:"column:clicks"`
		PeriodCount int     `gorm:"column:period_count"`
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
			ProductID:    key,
			ProductName:  r.ProductName,
			SKU:          r.ProductID,
			Platform:     "shopee",
			TotalCost:    r.TotalCost,
			TotalRevenue: r.TotalRevenue,
			TotalOrders:  r.TotalOrders,
			Impressions:  r.Impressions,
			Clicks:       r.Clicks,
			PeriodCount:  r.PeriodCount,
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
		profit := pd.Revenue - pd.Cost

		key := fmt.Sprintf("tiktok_%s", pd.ProductID)
		hist := result[key]
		hist.ROIValues = append(hist.ROIValues, roi)
		hist.ProfitValues = append(hist.ProfitValues, profit)
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
		profit := pd.Revenue - pd.Cost

		key := fmt.Sprintf("shopee_%s", pd.ProductID)
		hist := result[key]
		hist.ROIValues = append(hist.ROIValues, roi)
		hist.ProfitValues = append(hist.ProfitValues, profit)
		result[key] = hist
	}

	return result
}

// applyFilters filters products by category and action
func (s *MLAnalyticsService) applyFilters(
	products []models.MLProductAnalysis,
	categoryFilter, actionFilter string,
) []models.MLProductAnalysis {
	if categoryFilter == "" && actionFilter == "" {
		return products
	}

	filtered := make([]models.MLProductAnalysis, 0)
	for _, p := range products {
		if categoryFilter != "" && p.Category != categoryFilter {
			continue
		}
		if actionFilter != "" && p.Action != actionFilter {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered
}

// sortProducts sorts products by specified field
func (s *MLAnalyticsService) sortProducts(products []models.MLProductAnalysis, sortBy, sortDir string) {
	sort.Slice(products, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "unified_score":
			less = products[i].UnifiedScore < products[j].UnifiedScore
		case "roas":
			less = products[i].ROAS < products[j].ROAS
		case "revenue":
			less = products[i].TotalRevenue < products[j].TotalRevenue
		case "profit":
			less = products[i].TotalProfit < products[j].TotalProfit
		case "cost":
			less = products[i].TotalCost < products[j].TotalCost
		default:
			less = products[i].UnifiedScore < products[j].UnifiedScore
		}
		if sortDir == "desc" {
			return !less
		}
		return less
	})
}

// generateRecommendation creates a human-readable recommendation text
func generateRecommendation(category, action string, roas float64, platform string) string {
	platformLabel := "ads"
	if platform == "tiktok" {
		platformLabel = "TikTok ads"
	} else if platform == "shopee" {
		platformLabel = "Shopee ads"
	}

	switch category {
	case "STAR":
		return fmt.Sprintf("Top performer. Scale up %s budget by 20-50%% for maximum growth. ROAS: %.1fx", platformLabel, roas)
	case "GROWTH":
		return fmt.Sprintf("Growing potential. Review %s performance and consider gradual budget increase.", platformLabel)
	case "STABLE":
		return fmt.Sprintf("Consistent performer. Maintain current %s budget and monitor trends.", platformLabel)
	case "WATCH":
		return fmt.Sprintf("Needs attention. Monitor %s metrics closely and optimize targeting.", platformLabel)
	case "PROBLEM":
		if action == "STOP" {
			return fmt.Sprintf("Critical underperformer. Consider pausing %s immediately to reduce losses.", platformLabel)
		}
		return fmt.Sprintf("Underperforming. Reduce %s budget and review creative strategy.", platformLabel)
	default:
		return fmt.Sprintf("Evaluate %s performance and decide on next steps.", platformLabel)
	}
}
