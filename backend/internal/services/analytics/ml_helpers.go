package analytics

import (
	"context"
	"math"
	"sort"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/ads"
)

// getProductAnalyses aggregates and analyzes all products
func (s *MLAnalyticsService) getProductAnalyses(
	ctx context.Context,
	platform string,
	limit int,
	cursor string,
) ([]models.MLProductAnalysis, error) {

	// Aggregate data by product from TikTok ads
	type AggResult struct {
		ProductID    string  `gorm:"column:product_id"`
		ProductName  string  `gorm:"column:product_name"`
		CreativeType string  `gorm:"column:creative_type"`
		TotalCost    float64 `gorm:"column:total_cost"`
		TotalRevenue float64 `gorm:"column:total_revenue"`
		TotalOrders  int     `gorm:"column:total_orders"`
		Impressions  int     `gorm:"column:impressions"`
		Clicks       int     `gorm:"column:clicks"`
		PeriodCount  int     `gorm:"column:period_count"`
	}

	var results []AggResult
	query := s.db.WithContext(ctx).
		Model(&models.TiktokAdsCreativeData{}).
		Select(`
			product_id,
			MAX(COALESCE(video_title, product_id)) as product_name,
			MAX(creative_type) as creative_type,
			SUM(cost) as total_cost,
			SUM(gross_revenue) as total_revenue,
			SUM(orders_sku) as total_orders,
			SUM(impressions) as impressions,
			SUM(clicks) as clicks,
			COUNT(DISTINCT period_label) as period_count
		`).
		Where("tenant_id = ?", s.tenantID).
		Group("product_id").
		Order("total_revenue DESC")

	if cursor != "" {
		query = query.Where("product_id > ?", cursor)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	// Get historical data for trend analysis
	historical := s.getHistoricalByProduct(ctx)

	// Build product analyses
	analyses := make([]models.MLProductAnalysis, 0, len(results))
	cfg := ads.DefaultScoringConfig()

	for _, r := range results {
		if r.ProductID == "" {
			continue
		}

		analysis := models.MLProductAnalysis{
			TenantID:     s.tenantID,
			ProductID:    r.ProductID,
			ProductName:  r.ProductName,
			CreativeType: r.CreativeType,
			TotalCost:    r.TotalCost,
			TotalRevenue: r.TotalRevenue,
			TotalProfit:  r.TotalRevenue - r.TotalCost,
			TotalOrders:  r.TotalOrders,
			PeriodCount:  r.PeriodCount,
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

// getHistoricalByProduct retrieves historical ROI and profit per product (last 90 days)
func (s *MLAnalyticsService) getHistoricalByProduct(ctx context.Context) map[string]ads.HistoricalValues {
	type PeriodData struct {
		ProductID   string  `gorm:"column:product_id"`
		PeriodStart string  `gorm:"column:period_start"`
		Cost        float64 `gorm:"column:cost"`
		Revenue     float64 `gorm:"column:revenue"`
	}

	var periodData []PeriodData
	// Optimize: Only get last 90 days of data for trend analysis
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

		hist := result[pd.ProductID]
		hist.ROIValues = append(hist.ROIValues, roi)
		hist.ProfitValues = append(hist.ProfitValues, profit)
		result[pd.ProductID] = hist
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
