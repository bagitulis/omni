package ads

import (
	"math"
)

// CVLabel represents consistency classification
type CVLabel string

const (
	CVVeryConsistent CVLabel = "very_consistent"
	CVConsistent     CVLabel = "consistent"
	CVModerate       CVLabel = "moderate"
	CVVolatile       CVLabel = "volatile"
)

// CVResult holds Coefficient of Variation analysis result
type CVResult struct {
	CV           float64 `json:"cv"`
	Label        CVLabel `json:"label"`
	LabelDisplay string  `json:"label_display"`
	IsValid      bool    `json:"is_valid"`
}

// CoefficientOfVariation calculates CV = (StdDev / Mean) * 100
// Lower CV = more consistent, higher CV = more volatile
func CoefficientOfVariation(values []float64, minSamples int) float64 {
	if minSamples < 2 {
		minSamples = 2
	}

	n := len(values)
	if n < minSamples {
		return 0.0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(n)

	// Calculate standard deviation
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	stdDev := math.Sqrt(sumSquares / float64(n))

	// Avoid division by zero
	if mean == 0 {
		if stdDev > 0 {
			return 100.0
		}
		return 0.0
	}

	cv := (stdDev / math.Abs(mean)) * 100
	return math.Round(cv*100) / 100 // Round to 2 decimals
}

// GetCVLabel converts CV to human-readable label
// Thresholds:
// - < 20%: Sangat Konsisten (very consistent)
// - 20-35%: Konsisten (consistent)
// - 35-50%: Cukup Variabel (moderate)
// - > 50%: Volatil (volatile)
func GetCVLabel(cv float64) (CVLabel, string) {
	if cv < 20 {
		return CVVeryConsistent, "Sangat Konsisten"
	} else if cv < 35 {
		return CVConsistent, "Konsisten"
	} else if cv < 50 {
		return CVModerate, "Cukup Variabel"
	}
	return CVVolatile, "Volatil"
}

// AnalyzeVariation performs full CV analysis on values
func AnalyzeVariation(values []float64) CVResult {
	minSamples := 2

	if len(values) < minSamples {
		return CVResult{
			CV:           0,
			Label:        CVModerate,
			LabelDisplay: "Data tidak cukup",
			IsValid:      false,
		}
	}

	cv := CoefficientOfVariation(values, minSamples)
	label, display := GetCVLabel(cv)

	return CVResult{
		CV:           cv,
		Label:        label,
		LabelDisplay: display,
		IsValid:      true,
	}
}

// GetConsistencyScore converts CV to a 0-100 score
// CV 0% = 100 score (very consistent)
// CV 100%+ = 0 score (very volatile)
func GetConsistencyScore(values []float64) float64 {
	if len(values) < 2 {
		return 50 // Default neutral score
	}

	cv := CoefficientOfVariation(values, 2)
	score := math.Max(0, 100-cv)
	return math.Round(score*10) / 10
}
