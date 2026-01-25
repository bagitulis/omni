package ads

import (
	"math"
)

// ScoringConfig holds constants for scoring calculation
type ScoringConfig struct {
	WeightROI         float64
	WeightProfit      float64
	WeightMomentum    float64
	WeightConsistency float64
	WeightTrend       float64

	ScoreThresholdGood float64
	ScoreThresholdBad  float64
	ROIThresholdGood   float64
	ROIThresholdBad    float64
	LossThreshold      float64
}

// DefaultScoringConfig returns default configuration matching Python config
func DefaultScoringConfig() ScoringConfig {
	return ScoringConfig{
		WeightROI:         0.30,
		WeightProfit:      0.25,
		WeightMomentum:    0.20,
		WeightConsistency: 0.15,
		WeightTrend:       0.10,

		ScoreThresholdGood: 60.0,
		ScoreThresholdBad:  40.0,
		ROIThresholdGood:   2.0,
		ROIThresholdBad:    1.0,
		LossThreshold:      -100000.0,
	}
}

// ScoreResult holds the calculation results
type ScoreResult struct {
	ROIScore         float64 `json:"roi_score"`
	ProfitScore      float64 `json:"profit_score"`
	MomentumScore    float64 `json:"momentum_score"`
	ConsistencyScore float64 `json:"consistency_score"`
	TrendScore       float64 `json:"trend_score"`
	CompositeScore   float64 `json:"composite_score"`
	Category         string  `json:"category"`
	Action           string  `json:"action"`
}

// CalculateScores computes scores based on metrics
// metrics input should be: roi, profit, momentumPct (optional), cv (optional), trend (optional)
func CalculateScores(config ScoringConfig, roi, profit float64, momentumPct *float64, cv *float64, isUptrend *bool) ScoreResult {
	// 1. ROI Score (30%)
	// Cap at 100 (roi 10x = 100)
	roiScore := math.Min(roi*10.0, 100.0)

	// 2. Profit Score (25%)
	var profitScore float64
	if profit > 0 {
		profitScore = 100.0
	} else {
		// Example: profit -1,000,000 -> 50 + (-1000000/2000000)*50 = 25
		// Formula from python: max(0, 50 + (profit / 2000000) * 50)
		profitScore = math.Max(0, 50+(profit/2000000.0)*50.0)
	}

	// 3. Momentum Score (20%)
	momentumScore := 50.0
	if momentumPct != nil {
		// Formula: 50 + (momentum_pct / 2)
		momentumScore = 50.0 + (*momentumPct / 2.0)
		momentumScore = math.Min(math.Max(momentumScore, 0), 100)
	}

	// 4. Consistency Score (15%)
	consistencyScore := 50.0
	if cv != nil {
		// Formula: max(0, 100 - cv)
		// CV is typically > 0. CV of 0.2 (20%) -> score 80
		cvValue := *cv
		if cvValue < 0 {
			cvValue = 0 // Should not happen for Coeff of Variation
		}
		// If CV is percentage (e.g. 20), score is 80. If CV is ratio (0.2), we might need to adjust logic
		// Python logic: `cv_value = cv_result if cv_result > 0 else 50` then `100 - cv_value`
		// Assuming CV passed here is roughly 0-100 scale.
		consistencyScore = math.Max(0, 100.0-cvValue)
	}

	// 5. Trend Score (10%)
	trendScore := 50.0
	if isUptrend != nil {
		if *isUptrend {
			trendScore = 90.0
		} else {
			trendScore = 20.0
		}
	}

	// Composite
	composite := roiScore*config.WeightROI +
		profitScore*config.WeightProfit +
		momentumScore*config.WeightMomentum +
		consistencyScore*config.WeightConsistency +
		trendScore*config.WeightTrend

	// Determine Category
	category, action := determineCategory(config, composite, roi, profit)

	return ScoreResult{
		ROIScore:         round(roiScore),
		ProfitScore:      round(profitScore),
		MomentumScore:    round(momentumScore),
		ConsistencyScore: round(consistencyScore),
		TrendScore:       round(trendScore),
		CompositeScore:   round(composite),
		Category:         category,
		Action:           action,
	}
}

func determineCategory(cfg ScoringConfig, score, roi, profit float64) (string, string) {
	// Simplified reliability check (assume reliable if we are calculating score)
	// In real implementation, we might check sample size before calling this.

	if score >= cfg.ScoreThresholdGood && roi >= cfg.ROIThresholdGood {
		return "SCALE_UP", "Scale up budget 20-50%"
	} else if score >= cfg.ScoreThresholdGood {
		return "MAINTAIN", "Maintain budget"
	} else if score >= cfg.ScoreThresholdBad {
		return "MONITOR", "Monitor & optimize"
	} else if profit < cfg.LossThreshold || roi < cfg.ROIThresholdBad {
		return "STOP", "Stop immediately"
	} else {
		return "EVALUATE", "Review & decide"
	}
}

func round(val float64) float64 {
	return math.Round(val*10) / 10
}

// CalculateFullScore computes scores with automatic trend & CV analysis
// roiValues: historical ROI values for trend analysis
// profitValues: historical profit values for CV analysis
func CalculateFullScore(config ScoringConfig, roi, profit float64, roiValues, profitValues []float64) ScoreResult {
	var momentumPct *float64
	var cv *float64
	var isUptrend *bool

	// Calculate momentum from ROI trend
	if len(roiValues) >= 9 {
		trendResult := AnalyzeTrendMomentum(roiValues)
		momentumPct = &trendResult.ROC
		isUptrend = &trendResult.IsUptrend
	}

	// Calculate CV from profit consistency
	if len(profitValues) >= 2 {
		cvResult := AnalyzeVariation(profitValues)
		if cvResult.IsValid {
			cv = &cvResult.CV
		}
	}

	return CalculateScores(config, roi, profit, momentumPct, cv, isUptrend)
}

// SimpleScore is a convenience function for basic scoring without historical data
func SimpleScore(roi, profit float64) ScoreResult {
	config := DefaultScoringConfig()
	return CalculateScores(config, roi, profit, nil, nil, nil)
}
