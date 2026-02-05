package intelligence

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSaturationModel(t *testing.T) {
	model := NewSaturationModel()

	assert.NotNil(t, model)
	assert.Equal(t, 2000000, model.MaxBudget)
}

func TestSaturationModel_Analyze_InsufficientData(t *testing.T) {
	model := NewSaturationModel()

	tests := []struct {
		name           string
		spendHistory   []float64
		revenueHistory []float64
	}{
		{"empty", []float64{}, []float64{}},
		{"too few spend", []float64{100, 200, 300}, []float64{200, 400, 600, 800, 1000}},
		{"too few revenue", []float64{100, 200, 300, 400, 500}, []float64{200, 400, 600}},
		{"four each", []float64{100, 200, 300, 400}, []float64{200, 400, 600, 800}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.Analyze(tt.spendHistory, tt.revenueHistory)
			assert.Equal(t, SaturationInsufficientData, result.Status)
			assert.Equal(t, float64(50), result.Score)
		})
	}
}

func TestSaturationModel_Analyze_HighElasticity(t *testing.T) {
	model := NewSaturationModel()

	// Increasing spend with high marginal returns
	spend := []float64{100, 200, 300, 400, 500}
	revenue := []float64{300, 600, 900, 1200, 1500} // 3x ROAS throughout

	result := model.Analyze(spend, revenue)

	assert.True(t, result.MarginalRoas >= 2.0)
	assert.True(t, result.Headroom > 0)
	assert.True(t, result.Score >= 70)
}

func TestSaturationModel_Analyze_ApproachingSaturation(t *testing.T) {
	model := NewSaturationModel()

	// Spend increasing but marginal returns decreasing
	spend := []float64{100, 200, 300, 400, 500}
	revenue := []float64{300, 550, 750, 900, 1000} // Decreasing marginal returns

	result := model.Analyze(spend, revenue)

	// Should get valid analysis result (not insufficient data)
	assert.NotEqual(t, SaturationInsufficientData, result.Status)
	assert.True(t, result.MarginalRoas > 0)
}

func TestSaturationModel_Analyze_Saturated(t *testing.T) {
	model := NewSaturationModel()

	// Very low marginal returns
	spend := []float64{100, 200, 300, 400, 500}
	revenue := []float64{300, 400, 450, 475, 490} // Severely diminishing returns

	result := model.Analyze(spend, revenue)

	// Check for diminishing returns signals
	assert.True(t, result.MarginalRoas < 2.0 || result.Headroom < 50)
}

func TestSaturationModel_calculateMarginalRoas(t *testing.T) {
	model := NewSaturationModel()

	tests := []struct {
		name     string
		spend    []float64
		revenue  []float64
		expected float64
	}{
		{
			name:     "constant ROAS",
			spend:    []float64{100, 200},
			revenue:  []float64{300, 600},
			expected: 3.0, // (600-300)/(200-100) = 3
		},
		{
			name:     "decreasing ROAS",
			spend:    []float64{100, 200},
			revenue:  []float64{300, 500},
			expected: 2.0, // (500-300)/(200-100) = 2
		},
		{
			name:     "negative marginal",
			spend:    []float64{100, 200},
			revenue:  []float64{300, 250},
			expected: -0.5, // (250-300)/(200-100) = -0.5
		},
		{
			name:     "zero delta spend",
			spend:    []float64{100, 100},
			revenue:  []float64{300, 400},
			expected: 1.0, // Default
		},
		{
			name:     "single point",
			spend:    []float64{100},
			revenue:  []float64{300},
			expected: 1.0, // Default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.calculateMarginalRoas(tt.spend, tt.revenue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSaturationModel_calculateMarginalTrend(t *testing.T) {
	model := NewSaturationModel()

	spend := []float64{100, 200, 300, 400, 500}
	revenue := []float64{300, 550, 750, 900, 1000}

	marginals := model.calculateMarginalTrend(spend, revenue)

	assert.Len(t, marginals, 4) // n-1 marginal values
	// Marginals should be decreasing: 2.5, 2.0, 1.5, 1.0
	assert.True(t, marginals[0] > marginals[len(marginals)-1])
}

func TestSaturationModel_determineStatus(t *testing.T) {
	model := NewSaturationModel()

	tests := []struct {
		name         string
		marginalRoas float64
		headroom     float64
		expected     SaturationStatus
	}{
		{"over saturated", 0.5, -10, SaturationOverSaturated},
		{"saturated", 0.8, 5, SaturationSaturated},
		{"approaching", 1.5, 20, SaturationApproachingSaturation},
		{"moderate", 1.5, 50, SaturationModerate},
		{"high elasticity", 3.0, 50, SaturationHighElasticity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.determineStatus(tt.marginalRoas, tt.headroom)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSaturationModel_calculateScore(t *testing.T) {
	model := NewSaturationModel()

	tests := []struct {
		status   SaturationStatus
		expected float64
	}{
		{SaturationHighElasticity, 90},
		{SaturationModerate, 70},
		{SaturationApproachingSaturation, 50},
		{SaturationSaturated, 30},
		{SaturationOverSaturated, 10},
		{SaturationInsufficientData, 50},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := model.calculateScore(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSaturationResult_Fields(t *testing.T) {
	result := SaturationResult{
		Status:          SaturationModerate,
		MarginalRoas:    2.5,
		SaturationPoint: 1000000,
		CurrentSpend:    500000,
		Headroom:        100,
		Score:           70,
		Description:     "Test description",
	}

	assert.Equal(t, SaturationModerate, result.Status)
	assert.Equal(t, 2.5, result.MarginalRoas)
	assert.Equal(t, float64(1000000), result.SaturationPoint)
	assert.Equal(t, float64(500000), result.CurrentSpend)
	assert.Equal(t, float64(100), result.Headroom)
	assert.Equal(t, float64(70), result.Score)
}

func TestFormatPercent(t *testing.T) {
	result := formatPercent(50.5)
	assert.Contains(t, result, "%")
	assert.Contains(t, result, "50")
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		value    float64
		contains string
	}{
		{50.5, "50"},
		{0.0, "0"},
		{-10.5, "-10"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := formatFloat(tt.value)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestIntToString(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{123, "123"},
		{999, "999"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := intToString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
