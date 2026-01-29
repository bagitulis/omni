package intelligence

import (
	"math"
)

// ProductStage represents product lifecycle stage
type ProductStage string

const (
	ProductStageLaunch  ProductStage = "LAUNCH"
	ProductStageGrowth  ProductStage = "GROWTH"
	ProductStageMature  ProductStage = "MATURE"
	ProductStageDecline ProductStage = "DECLINE"
)

// LifecycleResult contains product lifecycle analysis results
type LifecycleResult struct {
	Stage          ProductStage `json:"stage"`
	DaysActive     int          `json:"days_active"`
	TrendDirection string       `json:"trend_direction"` // UP, DOWN, STABLE
	AvgRoas        float64      `json:"avg_roas"`
	RecentRoas     float64      `json:"recent_roas"`
	IsEvergreen    bool         `json:"is_evergreen"`
	Score          float64      `json:"score"`
	Description    string       `json:"description"`
}

// LifecycleAnalyzer determines product lifecycle stage
// Stages: LAUNCH -> GROWTH -> MATURE -> DECLINE
type LifecycleAnalyzer struct {
	LaunchPeriodDays int
	GrowthThreshold  float64 // ROC > threshold = growth
	DeclineThreshold float64 // ROC < threshold = decline
}

// NewLifecycleAnalyzer creates a new lifecycle analyzer
func NewLifecycleAnalyzer() *LifecycleAnalyzer {
	return &LifecycleAnalyzer{
		LaunchPeriodDays: 14,
		GrowthThreshold:  0.1,  // 10% growth
		DeclineThreshold: -0.1, // -10% decline
	}
}

// Analyze performs lifecycle analysis on ROAS history
func (l *LifecycleAnalyzer) Analyze(roasHistory []float64, daysActive int) LifecycleResult {
	if len(roasHistory) < 2 {
		return l.insufficientData()
	}

	// Calculate averages
	avgRoas := mean(roasHistory)
	recentRoas := avgRoas
	if len(roasHistory) >= 3 {
		recentRoas = mean(roasHistory[len(roasHistory)-3:])
	}

	// Determine trend direction
	trendDirection := l.determineTrend(roasHistory)

	// Determine lifecycle stage
	stage := l.determineStage(roasHistory, daysActive, trendDirection)

	// Check if evergreen (stable, long-running performer)
	isEvergreen := l.checkEvergreen(roasHistory, stage, daysActive)

	// Calculate score
	score := l.calculateScore(stage, avgRoas, trendDirection, isEvergreen)

	return LifecycleResult{
		Stage:          stage,
		DaysActive:     daysActive,
		TrendDirection: trendDirection,
		AvgRoas:        math.Round(avgRoas*100) / 100,
		RecentRoas:     math.Round(recentRoas*100) / 100,
		IsEvergreen:    isEvergreen,
		Score:          math.Round(score*10) / 10,
		Description:    l.getDescription(stage, daysActive, isEvergreen),
	}
}

// determineTrend calculates trend direction
func (l *LifecycleAnalyzer) determineTrend(values []float64) string {
	if len(values) < 3 {
		return "STABLE"
	}

	n := len(values)
	firstHalf := mean(values[:n/2])
	secondHalf := mean(values[n/2:])

	if firstHalf == 0 {
		return "STABLE"
	}

	change := (secondHalf - firstHalf) / firstHalf

	if change > l.GrowthThreshold {
		return "UP"
	}
	if change < l.DeclineThreshold {
		return "DOWN"
	}
	return "STABLE"
}

// determineStage determines lifecycle stage
func (l *LifecycleAnalyzer) determineStage(values []float64, daysActive int, trend string) ProductStage {
	// Launch phase: first N days
	if daysActive < l.LaunchPeriodDays {
		return ProductStageLaunch
	}

	// Calculate growth rate
	if len(values) >= 5 {
		recentGrowth := l.calculateRecentGrowth(values)

		if recentGrowth > l.GrowthThreshold {
			return ProductStageGrowth
		}
		if recentGrowth < l.DeclineThreshold {
			return ProductStageDecline
		}
	}

	// Default to mature if stable
	if trend == "DOWN" {
		return ProductStageDecline
	}
	if trend == "UP" {
		return ProductStageGrowth
	}

	return ProductStageMature
}

// calculateRecentGrowth calculates growth rate from recent data
func (l *LifecycleAnalyzer) calculateRecentGrowth(values []float64) float64 {
	if len(values) < 4 {
		return 0
	}

	n := len(values)
	recentAvg := mean(values[n-3:])
	olderAvg := mean(values[n-6 : n-3])

	if olderAvg == 0 {
		return 0
	}

	return (recentAvg - olderAvg) / olderAvg
}

// checkEvergreen checks if product is evergreen (stable long-term performer)
func (l *LifecycleAnalyzer) checkEvergreen(values []float64, stage ProductStage, daysActive int) bool {
	if daysActive < 30 {
		return false
	}

	// Must be in growth or mature stage
	if stage != ProductStageGrowth && stage != ProductStageMature {
		return false
	}

	// Check stability (low CV)
	cv := CoefficientOfVariation(values)
	return cv < 30
}

// calculateScore calculates lifecycle score
func (l *LifecycleAnalyzer) calculateScore(
	stage ProductStage,
	avgRoas float64,
	trend string,
	isEvergreen bool,
) float64 {
	baseScores := map[ProductStage]float64{
		ProductStageLaunch:  60,
		ProductStageGrowth:  85,
		ProductStageMature:  70,
		ProductStageDecline: 40,
	}

	score := baseScores[stage]

	// Adjust for trend
	if trend == "UP" {
		score += 10
	} else if trend == "DOWN" {
		score -= 10
	}

	// Bonus for evergreen
	if isEvergreen {
		score += 5
	}

	// Bonus for high ROAS
	if avgRoas >= 3 {
		score += 5
	}

	return math.Min(100, math.Max(0, score))
}

// getDescription returns Indonesian description
func (l *LifecycleAnalyzer) getDescription(stage ProductStage, days int, isEvergreen bool) string {
	evergreenSuffix := ""
	if isEvergreen {
		evergreenSuffix = " (Evergreen)"
	}

	daysStr := intToString(days)

	switch stage {
	case ProductStageLaunch:
		return "Launch phase (" + daysStr + " hari)" + evergreenSuffix
	case ProductStageGrowth:
		return "Growth phase (" + daysStr + " hari)" + evergreenSuffix
	case ProductStageMature:
		return "Mature phase (" + daysStr + " hari)" + evergreenSuffix
	case ProductStageDecline:
		return "Decline phase (" + daysStr + " hari)" + evergreenSuffix
	default:
		return "Unknown (" + daysStr + " hari)"
	}
}

func (l *LifecycleAnalyzer) insufficientData() LifecycleResult {
	return LifecycleResult{
		Stage:          ProductStageLaunch,
		TrendDirection: "STABLE",
		Score:          50,
		Description:    "Data tidak cukup untuk analisis lifecycle",
	}
}
