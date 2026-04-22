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
	productMap := make(map[string]*unifiedProductAgg)

	if platform == "" || platform == "tiktok" {
		s.aggregateTiktokProducts(ctx, productMap)
	}
	if platform == "" || platform == "shopee" {
		s.aggregateShopeeProducts(ctx, productMap)
	}

	historical := s.getHistoricalByProduct(ctx)
	shopeeHistorical := s.getShopeeHistoricalByProduct(ctx)

	for k, v := range shopeeHistorical {
		if existing, ok := historical[k]; ok {
			existing.ROIValues = append(existing.ROIValues, v.ROIValues...)
			existing.ProfitValues = append(existing.ProfitValues, v.ProfitValues...)
			historical[k] = existing
		} else {
			historical[k] = v
		}
	}

	analyses := make([]models.MLProductAnalysis, 0, len(productMap))
	cfg := ads.DefaultScoringConfig()

	for _, r := range productMap {
		if r.ProductID == "" {
			continue
		}
		analyses = append(analyses, buildProductAnalysis(s.tenantID, r, historical, cfg))
	}

	return analyses, nil
}

// buildProductAnalysis creates a single MLProductAnalysis from aggregated data
func buildProductAnalysis(
	tenantID string,
	r *unifiedProductAgg,
	historical map[string]ads.HistoricalValues,
	cfg ads.ScoringConfig,
) models.MLProductAnalysis {
	analysis := models.MLProductAnalysis{
		TenantID: tenantID, ProductID: r.ProductID, ProductName: r.ProductName,
		SKU: r.SKU, Platform: r.Platform, TotalCost: r.TotalCost,
		TotalRevenue: r.TotalRevenue, TotalProfit: r.TotalRevenue - r.TotalCost,
		TotalOrders: r.TotalOrders, Clicks: r.Clicks, Impressions: r.Impressions,
		PeriodCount: r.PeriodCount,
	}

	if r.Impressions > 0 {
		analysis.CTR = float64(r.Clicks) / float64(r.Impressions)
	}
	if r.TotalCost > 0 {
		analysis.ROAS = roundTo(r.TotalRevenue/r.TotalCost, 2)
	}

	var score ads.ScoreResult
	if hist, ok := historical[r.ProductID]; ok && len(hist.ROIValues) >= 2 {
		score = ads.CalculateFullScore(cfg, analysis.ROAS, analysis.TotalProfit, hist.ROIValues, hist.ProfitValues)
	} else {
		score = ads.SimpleScore(analysis.ROAS, analysis.TotalProfit)
	}

	analysis.UnifiedScore = score.CompositeScore
	analysis.ROASScore = score.ROIScore
	analysis.TrendScore = score.TrendScore
	analysis.MomentumScore = score.MomentumScore
	analysis.VolatilityScore = score.ConsistencyScore
	analysis.Category = mapToCategory(score.Category)
	analysis.Action = score.Category
	analysis.ActionLabel = score.Action
	analysis.BudgetChangePct = getBudgetChangePct(score.Category)
	analysis.Recommendation = generateRecommendation(analysis.Category, analysis.Action, analysis.ROAS, analysis.Platform)
	analysis.FatigueStatus = determineFatigueStatus(r.Clicks, r.Impressions, r.PeriodCount)
	analysis.HasFatigueWarning = analysis.FatigueStatus == "FATIGUED" || analysis.FatigueStatus == "DEAD"
	analysis.ChurnRiskScore = calculateChurnRisk(analysis.MomentumScore, analysis.TrendScore, analysis.VolatilityScore)
	analysis.HasChurnRisk = analysis.ChurnRiskScore > 60
	analysis.ConfidenceLevel = getConfidenceLevel(r.PeriodCount)
	analysis.SuccessProbability = estimateSuccessProbability(analysis.UnifiedScore, analysis.ROAS)
	analysis.TrendDirection = getTrendDirection(analysis.MomentumScore)
	analysis.TrendStrength = math.Abs(analysis.MomentumScore-50) / 50

	return analysis
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
