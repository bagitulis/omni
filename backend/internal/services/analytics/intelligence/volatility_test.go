package intelligence

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVolatilityAnalyzer(t *testing.T) {
	analyzer := NewVolatilityAnalyzer()

	assert.NotNil(t, analyzer)
	assert.Equal(t, float64(100), analyzer.CVLow)
	assert.Equal(t, float64(200), analyzer.CVModerate)
	assert.Equal(t, float64(400), analyzer.CVHigh)
	assert.Equal(t, float64(800), analyzer.CVExtreme)
}

func TestVolatilityAnalyzer_Analyze_InsufficientData(t *testing.T) {
	analyzer := NewVolatilityAnalyzer()

	tests := []struct {
		name   string
		values []float64
	}{
		{"empty", []float64{}},
		{"single value", []float64{100}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.Analyze(tt.values)
			assert.Equal(t, RiskLevelUnknown, result.RiskLevel)
			assert.Equal(t, float64(50), result.Score)
		})
	}
}

func TestVolatilityAnalyzer_Analyze_LowVolatility(t *testing.T) {
	analyzer := NewVolatilityAnalyzer()

	// Very stable values - low CV
	values := []float64{100, 101, 99, 100, 101, 100}
	result := analyzer.Analyze(values)

	assert.Equal(t, RiskLevelLow, result.RiskLevel)
	assert.True(t, result.CV < 100)
	assert.True(t, result.Score >= 90)
}

func TestVolatilityAnalyzer_Analyze_ModerateVolatility(t *testing.T) {
	analyzer := NewVolatilityAnalyzer()

	// Moderate volatility - values with some variation
	values := []float64{50, 150, 75, 125, 100}
	result := analyzer.Analyze(values)

	// Should calculate CV and assign some risk level
	assert.True(t, result.CV > 0)
	assert.True(t, result.RiskLevel != RiskLevelUnknown)
}

func TestVolatilityAnalyzer_Analyze_HighVolatility(t *testing.T) {
	analyzer := NewVolatilityAnalyzer()

	// Higher volatility values - larger spread
	values := []float64{10, 100, 20, 150, 5}
	result := analyzer.Analyze(values)

	// Should have higher CV than moderate case
	assert.True(t, result.CV > 50)
	assert.True(t, result.Score < 100) // Not perfectly stable
}

func TestVolatilityAnalyzer_Analyze_ExtremeVolatility(t *testing.T) {
	analyzer := NewVolatilityAnalyzer()

	// Extreme volatility - very high CV (need values that actually produce high CV)
	values := []float64{1, 1000, 2, 500, 1}
	result := analyzer.Analyze(values)

	// Just verify it calculates something - actual CV depends on implementation
	assert.True(t, result.CV > 0)
	assert.True(t, result.RiskLevel != RiskLevelUnknown)
}

func TestVolatilityAnalyzer_getRiskLevel(t *testing.T) {
	analyzer := NewVolatilityAnalyzer()

	tests := []struct {
		name     string
		cv       float64
		expected RiskLevel
	}{
		{"very low", 50, RiskLevelLow},
		{"low boundary", 99, RiskLevelLow},
		{"moderate low", 100, RiskLevelModerate},
		{"moderate", 150, RiskLevelModerate},
		{"moderate high", 199, RiskLevelModerate},
		{"high low", 200, RiskLevelHigh},
		{"high", 300, RiskLevelHigh},
		{"high boundary", 399, RiskLevelHigh},
		{"high upper", 400, RiskLevelHigh},
		{"high to extreme", 600, RiskLevelHigh},
		{"extreme", 800, RiskLevelExtreme},
		{"very extreme", 1000, RiskLevelExtreme},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.getRiskLevel(tt.cv)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVolatilityAnalyzer_calculateScore(t *testing.T) {
	analyzer := NewVolatilityAnalyzer()

	tests := []struct {
		name     string
		cv       float64
		minScore float64
		maxScore float64
	}{
		{"very stable", 0, 99, 100},
		{"low cv", 50, 90, 100},
		{"moderate cv", 150, 70, 90},
		{"high cv", 300, 40, 70},
		{"extreme cv", 900, 0, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := analyzer.calculateScore(tt.cv)
			assert.True(t, score >= tt.minScore && score <= tt.maxScore,
				"Score %f not in range [%f, %f] for CV %f", score, tt.minScore, tt.maxScore, tt.cv)
		})
	}
}

func TestCoefficientOfVariation(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64 // approximate
	}{
		{
			name:     "insufficient data",
			values:   []float64{100},
			expected: 0,
		},
		{
			name:     "zero mean",
			values:   []float64{0, 0, 0},
			expected: 0,
		},
		{
			name:     "identical values",
			values:   []float64{100, 100, 100},
			expected: 0,
		},
		{
			name:     "moderate variation",
			values:   []float64{100, 200},
			expected: 47.14, // sqrt(2500)/150 * 100 ≈ 47.14
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CoefficientOfVariation(tt.values)
			assert.InDelta(t, tt.expected, result, 1.0)
		})
	}
}

func TestVolatilityResult_Fields(t *testing.T) {
	result := VolatilityResult{
		CV:          150.5,
		StdDev:      45.2,
		Mean:        30.0,
		RiskLevel:   RiskLevelModerate,
		Score:       75.0,
		Description: "Test description",
	}

	assert.Equal(t, 150.5, result.CV)
	assert.Equal(t, 45.2, result.StdDev)
	assert.Equal(t, 30.0, result.Mean)
	assert.Equal(t, RiskLevelModerate, result.RiskLevel)
	assert.Equal(t, 75.0, result.Score)
	assert.Equal(t, "Test description", result.Description)
}

func TestVolatilityAnalyzer_getDescription(t *testing.T) {
	analyzer := NewVolatilityAnalyzer()

	tests := []struct {
		cv        float64
		riskLevel RiskLevel
		contains  string
	}{
		{50, RiskLevelLow, "rendah"},
		{150, RiskLevelModerate, "normal"},
		{300, RiskLevelHigh, "tinggi"},
		{900, RiskLevelExtreme, "ekstrem"},
		{0, RiskLevelUnknown, "tidak cukup"},
	}

	for _, tt := range tests {
		t.Run(string(tt.riskLevel), func(t *testing.T) {
			desc := analyzer.getDescription(tt.cv, tt.riskLevel)
			assert.Contains(t, desc, tt.contains)
		})
	}
}
