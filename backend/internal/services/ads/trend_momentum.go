package ads

import (
	"fmt"
	"math"
)

// TrendDirection represents trend classification
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

// TrendResult holds MACD-style trend analysis result
type TrendResult struct {
	Direction    TrendDirection `json:"direction"`
	Signal       float64        `json:"signal"`
	SignalChange float64        `json:"signal_change"`
	ROC          float64        `json:"roc"`
	Score        float64        `json:"score"`
	Description  string         `json:"description"`
	IsUptrend    bool           `json:"is_uptrend"`
}

// TrendMomentumAnalyzer performs MACD-style trend analysis
type TrendMomentumAnalyzer struct {
	ShortWindow int
	LongWindow  int
	MinPeriods  int
}

// NewTrendMomentumAnalyzer creates analyzer with default windows
func NewTrendMomentumAnalyzer() *TrendMomentumAnalyzer {
	return &TrendMomentumAnalyzer{
		ShortWindow: 3,
		LongWindow:  8,
		MinPeriods:  9, // LongWindow + 1
	}
}

// Analyze performs trend analysis on time series values
func (t *TrendMomentumAnalyzer) Analyze(values []float64) TrendResult {
	if len(values) < t.MinPeriods {
		return TrendResult{
			Direction:   TrendInsufficientData,
			Score:       50,
			Description: "Insufficient data for trend analysis",
			IsUptrend:   false,
		}
	}

	// Calculate EMAs
	emaShort := t.calculateEMA(values, t.ShortWindow)
	emaLong := t.calculateEMA(values, t.LongWindow)

	// MACD Signal
	signal := emaShort[len(emaShort)-1] - emaLong[len(emaLong)-1]
	prevSignal := emaShort[len(emaShort)-2] - emaLong[len(emaLong)-2]
	signalChange := signal - prevSignal

	// Rate of Change
	roc := t.calculateROC(values, t.ShortWindow)

	// Determine direction
	direction := t.getDirection(signal, signalChange)

	// Calculate score
	score := t.calculateScore(signal, signalChange, roc, emaLong[len(emaLong)-1])

	// Generate description
	description := t.getDescription(direction, roc)

	// Determine if uptrend
	isUptrend := direction == TrendStrongUptrend || direction == TrendUptrend || direction == TrendRecovering

	return TrendResult{
		Direction:    direction,
		Signal:       roundTo(signal, 4),
		SignalChange: roundTo(signalChange, 4),
		ROC:          roundTo(roc, 2),
		Score:        roundTo(score, 1),
		Description:  description,
		IsUptrend:    isUptrend,
	}
}

// calculateEMA computes Exponential Moving Average
func (t *TrendMomentumAnalyzer) calculateEMA(values []float64, span int) []float64 {
	if len(values) == 0 {
		return []float64{}
	}

	multiplier := 2.0 / float64(span+1)
	ema := make([]float64, len(values))
	ema[0] = values[0]

	for i := 1; i < len(values); i++ {
		ema[i] = (values[i]-ema[i-1])*multiplier + ema[i-1]
	}

	return ema
}

// calculateROC computes Rate of Change percentage
func (t *TrendMomentumAnalyzer) calculateROC(values []float64, periods int) float64 {
	if len(values) < periods+1 {
		return 0
	}

	current := values[len(values)-1]
	previous := values[len(values)-(periods+1)]

	if previous == 0 {
		return 0
	}

	return ((current - previous) / math.Abs(previous)) * 100
}

// getDirection determines trend direction from signal and change
func (t *TrendMomentumAnalyzer) getDirection(signal, signalChange float64) TrendDirection {
	if signal > 0 {
		if signalChange > 0 {
			return TrendStrongUptrend
		} else if signalChange < -0.01 {
			return TrendWeakening
		}
		return TrendUptrend
	}

	// signal <= 0
	if signalChange > 0.01 {
		return TrendRecovering
	} else if signalChange < 0 {
		return TrendStrongDowntrend
	}
	return TrendDowntrend
}

// calculateScore computes trend score (0-100)
func (t *TrendMomentumAnalyzer) calculateScore(signal, signalChange, roc, emaLong float64) float64 {
	score := 50.0

	if emaLong != 0 {
		signalRatio := signal / math.Abs(emaLong)
		contribution := signalRatio * 100
		score += math.Min(25, math.Max(-25, contribution))
	}

	rocContribution := roc / 4
	score += math.Min(25, math.Max(-25, rocContribution))

	return math.Min(100, math.Max(0, score))
}

// getDescription returns human-readable description
func (t *TrendMomentumAnalyzer) getDescription(direction TrendDirection, roc float64) string {
	descriptions := map[TrendDirection]string{
		TrendStrongUptrend:    "Strong upward momentum",
		TrendUptrend:          "Upward trend",
		TrendWeakening:        "Weakening momentum",
		TrendRecovering:       "Recovery phase",
		TrendDowntrend:        "Downward trend",
		TrendStrongDowntrend:  "Strong downward momentum",
		TrendInsufficientData: "Insufficient data",
	}

	if desc, ok := descriptions[direction]; ok {
		return desc + formatROC(roc)
	}
	return formatROC(roc)
}

func formatROC(roc float64) string {
	if roc >= 0 {
		return " (ROC: +" + formatFloat(roc, 1) + "%)"
	}
	return " (ROC: " + formatFloat(roc, 1) + "%)"
}

func formatFloat(val float64, decimals int) string {
	format := "%." + string(rune('0'+decimals)) + "f"
	return fmt.Sprintf(format, val)
}

func roundTo(val float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(val*pow) / pow
}

// AnalyzeTrendMomentum is a convenience function
func AnalyzeTrendMomentum(values []float64) TrendResult {
	analyzer := NewTrendMomentumAnalyzer()
	return analyzer.Analyze(values)
}

// GetTrendScore returns just the score (0-100)
func GetTrendScore(values []float64) float64 {
	result := AnalyzeTrendMomentum(values)
	return result.Score
}

// IsUptrend checks if values show upward trend
func IsUptrend(values []float64) bool {
	result := AnalyzeTrendMomentum(values)
	return result.IsUptrend
}
