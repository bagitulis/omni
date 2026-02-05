package ads

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTrendMomentumAnalyzer(t *testing.T) {
	analyzer := NewTrendMomentumAnalyzer()

	assert.NotNil(t, analyzer)
	assert.Equal(t, 3, analyzer.ShortWindow)
	assert.Equal(t, 8, analyzer.LongWindow)
	assert.Equal(t, 9, analyzer.MinPeriods)
}

func TestTrendMomentumAnalyzer_Analyze_InsufficientData(t *testing.T) {
	analyzer := NewTrendMomentumAnalyzer()

	tests := []struct {
		name   string
		values []float64
	}{
		{"empty", []float64{}},
		{"one value", []float64{100}},
		{"eight values", []float64{1, 2, 3, 4, 5, 6, 7, 8}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.Analyze(tt.values)
			assert.Equal(t, TrendInsufficientData, result.Direction)
			assert.Equal(t, float64(50), result.Score)
			assert.False(t, result.IsUptrend)
		})
	}
}

func TestTrendMomentumAnalyzer_Analyze_Uptrend(t *testing.T) {
	analyzer := NewTrendMomentumAnalyzer()

	// Clear upward trend
	values := []float64{100, 110, 120, 130, 140, 150, 160, 170, 180}
	result := analyzer.Analyze(values)

	assert.True(t, result.Direction == TrendUptrend || result.Direction == TrendStrongUptrend)
	assert.True(t, result.IsUptrend)
	assert.True(t, result.ROC > 0)
	assert.True(t, result.Score >= 50)
}

func TestTrendMomentumAnalyzer_Analyze_Downtrend(t *testing.T) {
	analyzer := NewTrendMomentumAnalyzer()

	// Clear downward trend
	values := []float64{180, 170, 160, 150, 140, 130, 120, 110, 100}
	result := analyzer.Analyze(values)

	assert.True(t, result.Direction == TrendDowntrend || result.Direction == TrendStrongDowntrend)
	assert.False(t, result.IsUptrend)
	assert.True(t, result.ROC < 0)
	assert.True(t, result.Score <= 50)
}

func TestTrendMomentumAnalyzer_calculateEMA(t *testing.T) {
	analyzer := NewTrendMomentumAnalyzer()

	tests := []struct {
		name   string
		values []float64
		span   int
	}{
		{"empty", []float64{}, 3},
		{"single", []float64{100}, 3},
		{"ascending", []float64{100, 110, 120, 130}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.calculateEMA(tt.values, tt.span)
			assert.Len(t, result, len(tt.values))
		})
	}
}

func TestTrendMomentumAnalyzer_calculateEMA_Values(t *testing.T) {
	analyzer := NewTrendMomentumAnalyzer()

	values := []float64{100, 110, 120}
	ema := analyzer.calculateEMA(values, 3)

	// First EMA equals first value
	assert.Equal(t, 100.0, ema[0])
	// Subsequent EMAs should be weighted averages
	assert.True(t, ema[1] > 100 && ema[1] < 110)
	assert.True(t, ema[2] > ema[1])
}

func TestTrendMomentumAnalyzer_calculateROC(t *testing.T) {
	analyzer := NewTrendMomentumAnalyzer()

	tests := []struct {
		name     string
		values   []float64
		periods  int
		expected float64
	}{
		{
			name:     "positive ROC",
			values:   []float64{100, 110, 120, 150},
			periods:  3,
			expected: 50, // (150-100)/100 * 100
		},
		{
			name:     "negative ROC",
			values:   []float64{200, 180, 160, 100},
			periods:  3,
			expected: -50, // (100-200)/200 * 100
		},
		{
			name:     "insufficient data",
			values:   []float64{100, 110},
			periods:  3,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.calculateROC(tt.values, tt.periods)
			assert.InDelta(t, tt.expected, result, 1.0)
		})
	}
}

func TestTrendMomentumAnalyzer_getDirection(t *testing.T) {
	analyzer := NewTrendMomentumAnalyzer()

	tests := []struct {
		signal       float64
		signalChange float64
		expected     TrendDirection
	}{
		{1.0, 0.5, TrendStrongUptrend},     // positive signal, positive change
		{1.0, 0.0, TrendUptrend},           // positive signal, no change
		{1.0, -0.02, TrendWeakening},       // positive signal, negative change
		{-1.0, 0.02, TrendRecovering},      // negative signal, positive change
		{-1.0, -0.5, TrendStrongDowntrend}, // negative signal, negative change
		{-1.0, 0.0, TrendDowntrend},        // negative signal, no change
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			result := analyzer.getDirection(tt.signal, tt.signalChange)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrendMomentumAnalyzer_calculateScore(t *testing.T) {
	analyzer := NewTrendMomentumAnalyzer()

	// Score should be between 0-100
	testCases := []struct {
		signal       float64
		signalChange float64
		roc          float64
		emaLong      float64
	}{
		{0.5, 0.1, 10, 10},
		{-0.5, -0.1, -10, 10},
		{0, 0, 0, 10},
	}

	for _, tc := range testCases {
		score := analyzer.calculateScore(tc.signal, tc.signalChange, tc.roc, tc.emaLong)
		assert.True(t, score >= 0 && score <= 100)
	}
}

func TestAnalyzeTrendMomentum(t *testing.T) {
	values := []float64{100, 110, 120, 130, 140, 150, 160, 170, 180}
	result := AnalyzeTrendMomentum(values)

	assert.NotEqual(t, TrendInsufficientData, result.Direction)
	assert.True(t, result.Score >= 0 && result.Score <= 100)
}

func TestGetTrendScore(t *testing.T) {
	values := []float64{100, 110, 120, 130, 140, 150, 160, 170, 180}
	score := GetTrendScore(values)

	assert.True(t, score >= 0 && score <= 100)
}

func TestIsUptrend(t *testing.T) {
	uptrendValues := []float64{100, 110, 120, 130, 140, 150, 160, 170, 180}
	downtrendValues := []float64{180, 170, 160, 150, 140, 130, 120, 110, 100}

	assert.True(t, IsUptrend(uptrendValues))
	assert.False(t, IsUptrend(downtrendValues))
}

func TestTrendResult_Fields(t *testing.T) {
	result := TrendResult{
		Direction:    TrendUptrend,
		Signal:       0.5,
		SignalChange: 0.1,
		ROC:          10.5,
		Score:        75.0,
		Description:  "Test description",
		IsUptrend:    true,
	}

	assert.Equal(t, TrendUptrend, result.Direction)
	assert.Equal(t, 0.5, result.Signal)
	assert.Equal(t, 0.1, result.SignalChange)
	assert.Equal(t, 10.5, result.ROC)
	assert.Equal(t, 75.0, result.Score)
	assert.True(t, result.IsUptrend)
}

func TestRoundTo(t *testing.T) {
	tests := []struct {
		val      float64
		decimals int
		expected float64
	}{
		{1.2345, 2, 1.23},
		{1.2355, 2, 1.24},
		{10.5, 0, 11},
		{10.4, 0, 10},
	}

	for _, tt := range tests {
		result := roundTo(tt.val, tt.decimals)
		assert.Equal(t, tt.expected, result)
	}
}

func TestFormatROC(t *testing.T) {
	assert.Contains(t, formatROC(10.5), "+10.5%")
	assert.Contains(t, formatROC(-5.3), "-5.3%")
}
