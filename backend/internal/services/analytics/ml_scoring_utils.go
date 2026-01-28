package analytics

import (
	"fmt"
	"math"
)

// roundTo rounds a float to specified decimal places
func roundTo(val float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(val*pow) / pow
}

// getHealthLabel returns human-readable health label
func getHealthLabel(score float64) string {
	switch {
	case score >= 70:
		return "Excellent"
	case score >= 55:
		return "Good"
	case score >= 40:
		return "Fair"
	default:
		return "Poor"
	}
}

// formatPercent formats float as percentage string
func formatPercent(val float64) string {
	return fmt.Sprintf("%.1f%%", val)
}

// mapToCategory maps action to category
func mapToCategory(action string) string {
	switch action {
	case "SCALE_UP":
		return "STAR"
	case "MAINTAIN":
		return "STABLE"
	case "MONITOR":
		return "WATCH"
	case "STOP":
		return "PROBLEM"
	case "EVALUATE":
		return "GROWTH"
	default:
		return "WATCH"
	}
}

// getBudgetChangePct returns recommended budget change percentage
func getBudgetChangePct(action string) float64 {
	switch action {
	case "SCALE_UP":
		return 30.0
	case "MAINTAIN":
		return 0.0
	case "MONITOR":
		return -10.0
	case "STOP":
		return -100.0
	default:
		return 0.0
	}
}

// determineFatigueStatus calculates creative fatigue status
func determineFatigueStatus(clicks, impressions, periodCount int) string {
	if periodCount < 3 {
		return "FRESH"
	}
	ctr := 0.0
	if impressions > 0 {
		ctr = float64(clicks) / float64(impressions) * 100
	}
	switch {
	case ctr >= 2.0:
		return "FRESH"
	case ctr >= 1.0:
		return "AGING"
	case ctr >= 0.5:
		return "FATIGUED"
	default:
		return "DEAD"
	}
}

// calculateChurnRisk calculates churn risk score
func calculateChurnRisk(momentum, trend, volatility float64) float64 {
	momentumRisk := math.Max(0, 50-momentum)
	trendRisk := math.Max(0, 50-trend)
	volatilityRisk := math.Max(0, volatility-50)
	risk := (momentumRisk*0.4 + trendRisk*0.4 + volatilityRisk*0.2)
	return roundTo(risk, 1)
}

// getConfidenceLevel returns confidence level based on data periods
func getConfidenceLevel(periodCount int) string {
	switch {
	case periodCount >= 8:
		return "HIGH"
	case periodCount >= 4:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// estimateSuccessProbability estimates success probability from score
func estimateSuccessProbability(score, roas float64) float64 {
	baseProbability := score / 100
	roasBonus := 0.0
	if roas >= 2.0 {
		roasBonus = 0.1
	}
	return roundTo(math.Min(baseProbability+roasBonus, 0.95), 2)
}

// getTrendDirection returns trend direction from momentum
func getTrendDirection(momentum float64) string {
	switch {
	case momentum >= 60:
		return "UP"
	case momentum <= 40:
		return "DOWN"
	default:
		return "STABLE"
	}
}
