package intelligence

import (
	"math"
)

// TrendDirection represents the direction of a trend
type TrendDirection string

const (
	TrendStrongUptrend    TrendDirection = "STRONG_UPTREND"
	TrendUptrend          TrendDirection = "UPTREND"
	TrendWeakening        TrendDirection = "WEAKENING"
	TrendRecovering       TrendDirection = "RECOVERING"
	TrendDowntrend        TrendDirection = "DOWNTREND"
	TrendStrongDowntrend  TrendDirection = "STRONG_DOWNTREND"
	TrendInsufficientData TrendDirection = "INSUFFICIENT_DATA"
)

// TrendResult contains trend analysis results
type TrendResult struct {
	Direction      TrendDirection `json:"direction"`
	RateOfChange   float64        `json:"rate_of_change"`
	Score          float64        `json:"score"`
	Momentum       float64        `json:"momentum"`
	IsAccelerating bool           `json:"is_accelerating"`
	Description    string         `json:"description"`
}

// TrendAnalyzer analyzes trend momentum
type TrendAnalyzer struct {
	MinDataPoints int
}

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer() *TrendAnalyzer {
	return &TrendAnalyzer{MinDataPoints: 5}
}

// Analyze performs trend analysis on a series of values
func (t *TrendAnalyzer) Analyze(values []float64) TrendResult {
	if len(values) < t.MinDataPoints {
		return TrendResult{
			Direction:   TrendInsufficientData,
			Score:       50,
			Description: "Data tidak cukup untuk analisis trend",
		}
	}

	// Calculate Rate of Change (ROC)
	roc := t.calculateROC(values)

	// Calculate Simple Moving Average Momentum
	momentum := t.calculateMomentum(values)

	// Determine if accelerating or decelerating
	isAccelerating := t.isAccelerating(values)

	// Determine direction
	direction := t.getDirection(roc, momentum, isAccelerating)

	// Calculate score (0-100)
	score := t.calculateScore(roc, momentum, direction)

	return TrendResult{
		Direction:      direction,
		RateOfChange:   math.Round(roc*100) / 100,
		Score:          math.Round(score*10) / 10,
		Momentum:       math.Round(momentum*100) / 100,
		IsAccelerating: isAccelerating,
		Description:    t.getDescription(direction, roc),
	}
}

// calculateROC calculates Rate of Change as percentage
func (t *TrendAnalyzer) calculateROC(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	// Compare recent average (last 3) to older average (first 3)
	recentCount := min(3, len(values))
	olderCount := min(3, len(values))

	var recentSum, olderSum float64
	for i := len(values) - recentCount; i < len(values); i++ {
		recentSum += values[i]
	}
	for i := 0; i < olderCount; i++ {
		olderSum += values[i]
	}

	recentAvg := recentSum / float64(recentCount)
	olderAvg := olderSum / float64(olderCount)

	if olderAvg == 0 {
		return 0
	}

	return ((recentAvg - olderAvg) / olderAvg) * 100
}

// calculateMomentum calculates momentum using simple moving average difference
func (t *TrendAnalyzer) calculateMomentum(values []float64) float64 {
	if len(values) < 5 {
		return 0
	}

	// Short-term SMA (3 periods)
	shortCount := 3
	var shortSum float64
	for i := len(values) - shortCount; i < len(values); i++ {
		shortSum += values[i]
	}
	shortSMA := shortSum / float64(shortCount)

	// Long-term SMA (5 periods or all if less)
	longCount := min(5, len(values))
	var longSum float64
	for i := len(values) - longCount; i < len(values); i++ {
		longSum += values[i]
	}
	longSMA := longSum / float64(longCount)

	if longSMA == 0 {
		return 0
	}

	return ((shortSMA - longSMA) / longSMA) * 100
}

// isAccelerating checks if trend is accelerating
func (t *TrendAnalyzer) isAccelerating(values []float64) bool {
	if len(values) < 4 {
		return false
	}

	// Compare second derivative (change in rate of change)
	n := len(values)
	change1 := values[n-2] - values[n-3]
	change2 := values[n-1] - values[n-2]

	return change2 > change1
}

// getDirection determines trend direction
func (t *TrendAnalyzer) getDirection(roc, momentum float64, isAccelerating bool) TrendDirection {
	// Strong thresholds
	if roc > 15 && momentum > 5 {
		return TrendStrongUptrend
	}
	if roc < -15 && momentum < -5 {
		return TrendStrongDowntrend
	}

	// Moderate thresholds
	if roc > 5 {
		if isAccelerating {
			return TrendStrongUptrend
		}
		return TrendUptrend
	}
	if roc < -5 {
		if !isAccelerating {
			return TrendStrongDowntrend
		}
		return TrendDowntrend
	}

	// Weak/transition states
	if momentum > 0 && roc > 0 {
		return TrendRecovering
	}
	if momentum < 0 && roc > -5 {
		return TrendWeakening
	}

	return TrendWeakening
}

// calculateScore calculates trend score (0-100)
func (t *TrendAnalyzer) calculateScore(roc, momentum float64, direction TrendDirection) float64 {
	// Base score based on direction
	baseScores := map[TrendDirection]float64{
		TrendStrongUptrend:    90,
		TrendUptrend:          75,
		TrendRecovering:       60,
		TrendWeakening:        40,
		TrendDowntrend:        25,
		TrendStrongDowntrend:  10,
		TrendInsufficientData: 50,
	}

	base := baseScores[direction]

	// Adjust by ROC magnitude (capped at +/-10)
	rocAdjustment := math.Min(10, math.Max(-10, roc/2))

	return math.Min(100, math.Max(0, base+rocAdjustment))
}

// getDescription returns Indonesian description for trend
func (t *TrendAnalyzer) getDescription(direction TrendDirection, roc float64) string {
	descriptions := map[TrendDirection]string{
		TrendStrongUptrend:    "Momentum sangat kuat (ROC: %+.1f%%)",
		TrendUptrend:          "Trend naik (ROC: %+.1f%%)",
		TrendWeakening:        "Momentum melemah (ROC: %+.1f%%)",
		TrendRecovering:       "Sedang recovery (ROC: %+.1f%%)",
		TrendDowntrend:        "Trend turun (ROC: %+.1f%%)",
		TrendStrongDowntrend:  "Momentum sangat lemah (ROC: %+.1f%%)",
		TrendInsufficientData: "Data tidak cukup",
	}

	format, ok := descriptions[direction]
	if !ok {
		return "Unknown"
	}

	if direction == TrendInsufficientData {
		return format
	}

	return formatString(format, roc)
}

// formatString is a simple string formatter for single float
func formatString(format string, value float64) string {
	// Manual formatting since we can't use fmt in simple cases
	sign := ""
	if value >= 0 {
		sign = "+"
	}
	return format[:len(format)-6] + sign + floatToString(value, 1) + "%)"
}

// floatToString converts float to string with decimal places
func floatToString(f float64, decimals int) string {
	if decimals == 1 {
		return string(rune(int(math.Abs(f))+48)) + "." + string(rune(int(math.Abs(f)*10)%10+48))
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
