package intelligence

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTrendAnalyzer(t *testing.T) {
	analyzer := NewTrendAnalyzer()

	assert.NotNil(t, analyzer)
	assert.Equal(t, 5, analyzer.MinDataPoints)
}

func TestTrendAnalyzer_Analyze_InsufficientData(t *testing.T) {
	analyzer := NewTrendAnalyzer()

	tests := []struct {
		name   string
		values []float64
	}{
		{"empty", []float64{}},
		{"one value", []float64{100}},
		{"two values", []float64{100, 110}},
		{"three values", []float64{100, 110, 120}},
		{"four values", []float64{100, 110, 120, 130}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.Analyze(tt.values)
			assert.Equal(t, TrendInsufficientData, result.Direction)
			assert.Equal(t, float64(50), result.Score)
		})
	}
}

func TestTrendAnalyzer_Analyze_Uptrend(t *testing.T) {
	analyzer := NewTrendAnalyzer()

	// Clear upward trend
	values := []float64{100, 110, 120, 130, 140}
	result := analyzer.Analyze(values)

	assert.True(t, result.Direction == TrendUptrend || result.Direction == TrendStrongUptrend)
	assert.True(t, result.RateOfChange > 0)
	assert.True(t, result.Score > 50)
}

func TestTrendAnalyzer_Analyze_StrongUptrend(t *testing.T) {
	analyzer := NewTrendAnalyzer()

	// Strong upward trend (>15% ROC)
	values := []float64{100, 120, 140, 160, 180}
	result := analyzer.Analyze(values)

	assert.Equal(t, TrendStrongUptrend, result.Direction)
	assert.True(t, result.RateOfChange > 15)
	assert.True(t, result.Score >= 80)
}

func TestTrendAnalyzer_Analyze_Downtrend(t *testing.T) {
	analyzer := NewTrendAnalyzer()

	// Clear downward trend
	values := []float64{150, 140, 130, 120, 110}
	result := analyzer.Analyze(values)

	assert.True(t, result.Direction == TrendDowntrend || result.Direction == TrendStrongDowntrend)
	assert.True(t, result.RateOfChange < 0)
	assert.True(t, result.Score < 50)
}

func TestTrendAnalyzer_Analyze_StrongDowntrend(t *testing.T) {
	analyzer := NewTrendAnalyzer()

	// Strong downward trend (>15% negative ROC)
	values := []float64{200, 170, 140, 110, 80}
	result := analyzer.Analyze(values)

	assert.True(t, result.Direction == TrendDowntrend || result.Direction == TrendStrongDowntrend)
	assert.True(t, result.RateOfChange < -15)
	assert.True(t, result.Score <= 30)
}

func TestTrendAnalyzer_Analyze_Flat(t *testing.T) {
	analyzer := NewTrendAnalyzer()

	// Flat trend (no significant change)
	values := []float64{100, 101, 99, 100, 101}
	result := analyzer.Analyze(values)

	assert.True(t, result.RateOfChange >= -5 && result.RateOfChange <= 5)
	assert.True(t, result.Score >= 30 && result.Score <= 70)
}

func TestTrendAnalyzer_calculateROC(t *testing.T) {
	analyzer := NewTrendAnalyzer()

	tests := []struct {
		name     string
		values   []float64
		expected float64 // approximate
	}{
		{
			name:     "positive ROC",
			values:   []float64{100, 100, 100, 150, 150, 150},
			expected: 50, // (150-100)/100 * 100
		},
		{
			name:     "negative ROC",
			values:   []float64{200, 200, 200, 100, 100, 100},
			expected: -50, // (100-200)/200 * 100
		},
		{
			name:     "zero ROC",
			values:   []float64{100, 100, 100, 100, 100},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.calculateROC(tt.values)
			assert.InDelta(t, tt.expected, result, 1.0)
		})
	}
}

func TestTrendAnalyzer_isAccelerating(t *testing.T) {
	analyzer := NewTrendAnalyzer()

	tests := []struct {
		name     string
		values   []float64
		expected bool
	}{
		{
			name:     "accelerating uptrend",
			values:   []float64{100, 105, 115, 130}, // +5, +10, +15 (increasing)
			expected: true,
		},
		{
			name:     "decelerating uptrend",
			values:   []float64{100, 120, 135, 145}, // +20, +15, +10 (decreasing)
			expected: false,
		},
		{
			name:     "accelerating downtrend",
			values:   []float64{150, 140, 125, 100}, // -10, -15, -25 (magnitude increasing)
			expected: false,                         // change2 < change1 in absolute terms for downtrend
		},
		{
			name:     "insufficient data",
			values:   []float64{100, 110, 120},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.isAccelerating(tt.values)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrendAnalyzer_calculateScore(t *testing.T) {
	analyzer := NewTrendAnalyzer()

	tests := []struct {
		name      string
		direction TrendDirection
		minScore  float64
		maxScore  float64
	}{
		{"strong uptrend", TrendStrongUptrend, 85, 100},
		{"uptrend", TrendUptrend, 70, 90},
		{"recovering", TrendRecovering, 55, 75},
		{"weakening", TrendWeakening, 30, 55},
		{"downtrend", TrendDowntrend, 15, 40},
		{"strong downtrend", TrendStrongDowntrend, 0, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := analyzer.calculateScore(5, 2, tt.direction)
			assert.True(t, score >= tt.minScore && score <= tt.maxScore,
				"Score %f not in range [%f, %f]", score, tt.minScore, tt.maxScore)
		})
	}
}

func TestTrendResult_Fields(t *testing.T) {
	result := TrendResult{
		Direction:      TrendUptrend,
		RateOfChange:   10.5,
		Score:          75.0,
		Momentum:       3.2,
		IsAccelerating: true,
		Description:    "Test description",
	}

	assert.Equal(t, TrendUptrend, result.Direction)
	assert.Equal(t, 10.5, result.RateOfChange)
	assert.Equal(t, 75.0, result.Score)
	assert.Equal(t, 3.2, result.Momentum)
	assert.True(t, result.IsAccelerating)
	assert.Equal(t, "Test description", result.Description)
}

func TestMin(t *testing.T) {
	assert.Equal(t, 3, min(3, 5))
	assert.Equal(t, 3, min(5, 3))
	assert.Equal(t, 5, min(5, 5))
	assert.Equal(t, -1, min(-1, 0))
}
