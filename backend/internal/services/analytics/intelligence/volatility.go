package intelligence

import (
	"math"
)

// RiskLevel represents volatility risk level
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "LOW"
	RiskLevelModerate RiskLevel = "MODERATE"
	RiskLevelHigh     RiskLevel = "HIGH"
	RiskLevelExtreme  RiskLevel = "EXTREME"
	RiskLevelUnknown  RiskLevel = "UNKNOWN"
)

// VolatilityResult contains volatility analysis results
type VolatilityResult struct {
	CV          float64   `json:"cv"`
	StdDev      float64   `json:"std_dev"`
	Mean        float64   `json:"mean"`
	RiskLevel   RiskLevel `json:"risk_level"`
	Score       float64   `json:"score"`
	Description string    `json:"description"`
}

// VolatilityAnalyzer analyzes ROAS volatility to measure risk
// Uses Coefficient of Variation (CV = StdDev / Mean * 100)
// NOTE: TikTok Ads naturally have high CV (100-300% is normal)
type VolatilityAnalyzer struct {
	CVLow      float64 // <100% is very stable for TikTok
	CVModerate float64 // 100-200% is normal
	CVHigh     float64 // 200-400% is moderate risk
	CVExtreme  float64 // >800% is genuinely problematic
}

// NewVolatilityAnalyzer creates a new volatility analyzer
func NewVolatilityAnalyzer() *VolatilityAnalyzer {
	return &VolatilityAnalyzer{
		CVLow:      100,
		CVModerate: 200,
		CVHigh:     400,
		CVExtreme:  800,
	}
}

// Analyze performs volatility analysis on a series of values
func (v *VolatilityAnalyzer) Analyze(values []float64) VolatilityResult {
	if len(values) < 2 {
		return VolatilityResult{
			RiskLevel:   RiskLevelUnknown,
			Score:       50,
			Description: "Data tidak cukup untuk analisis volatilitas",
		}
	}

	// Calculate mean
	meanVal := mean(values)

	// Calculate standard deviation
	stdDevVal := stdDev(values)

	// Calculate Coefficient of Variation
	cv := float64(0)
	if meanVal > 0 {
		cv = (stdDevVal / meanVal) * 100
	}

	// Get risk level (TikTok-adjusted)
	riskLevel := v.getRiskLevel(cv)

	// Calculate score (higher = more stable)
	score := v.calculateScore(cv)

	return VolatilityResult{
		CV:          math.Round(cv*10) / 10,
		StdDev:      math.Round(stdDevVal*100) / 100,
		Mean:        math.Round(meanVal*100) / 100,
		RiskLevel:   riskLevel,
		Score:       math.Round(score*10) / 10,
		Description: v.getDescription(cv, riskLevel),
	}
}

// getRiskLevel determines risk level from CV (TikTok-adjusted)
func (v *VolatilityAnalyzer) getRiskLevel(cv float64) RiskLevel {
	if cv < v.CVLow {
		return RiskLevelLow
	}
	if cv < v.CVModerate {
		return RiskLevelModerate
	}
	if cv < v.CVHigh {
		return RiskLevelHigh
	}
	if cv < v.CVExtreme {
		return RiskLevelHigh
	}
	return RiskLevelExtreme
}

// calculateScore converts CV to stability score (0-100, higher = more stable)
func (v *VolatilityAnalyzer) calculateScore(cv float64) float64 {
	if cv < v.CVLow {
		return 90 + (v.CVLow-cv)/v.CVLow*10
	}
	if cv < v.CVModerate {
		return 70 + (v.CVModerate-cv)/(v.CVModerate-v.CVLow)*20
	}
	if cv < v.CVHigh {
		return 40 + (v.CVHigh-cv)/(v.CVHigh-v.CVModerate)*30
	}
	if cv < v.CVExtreme {
		return 20 + (v.CVExtreme-cv)/(v.CVExtreme-v.CVHigh)*20
	}
	return math.Max(0, 20-(cv-v.CVExtreme)/100)
}

// getDescription returns Indonesian description
func (v *VolatilityAnalyzer) getDescription(cv float64, riskLevel RiskLevel) string {
	cvStr := formatFloat(cv)
	switch riskLevel {
	case RiskLevelLow:
		return "Volatilitas rendah (CV: " + cvStr + "%) - Sangat stabil"
	case RiskLevelModerate:
		return "Volatilitas normal (CV: " + cvStr + "%) - Normal untuk TikTok"
	case RiskLevelHigh:
		return "Volatilitas tinggi (CV: " + cvStr + "%) - Monitor ketat"
	case RiskLevelExtreme:
		return "Volatilitas ekstrem (CV: " + cvStr + "%) - Risiko tinggi"
	default:
		return "Data tidak cukup"
	}
}

// CoefficientOfVariation calculates CV from values
func CoefficientOfVariation(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	m := mean(values)
	if m == 0 {
		return 0
	}
	return (stdDev(values) / m) * 100
}
