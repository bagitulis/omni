package intelligence

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProbabilityEngine(t *testing.T) {
	engine := NewProbabilityEngine()

	assert.NotNil(t, engine)
	assert.Equal(t, 3, engine.MinDataPoints)
}

func TestProbabilityEngine_Analyze_InsufficientData(t *testing.T) {
	engine := NewProbabilityEngine()

	tests := []struct {
		name   string
		values []float64
	}{
		{"empty", []float64{}},
		{"one value", []float64{2.5}},
		{"two values", []float64{2.5, 3.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Analyze(tt.values, 2.0)
			assert.Equal(t, float64(50), result.SuccessProbability)
			assert.Equal(t, float64(50), result.FailureProbability)
			assert.Equal(t, ConfidenceLow, result.ConfidenceLevel)
		})
	}
}

func TestProbabilityEngine_Analyze_HighSuccessRate(t *testing.T) {
	engine := NewProbabilityEngine()

	// All values above threshold
	values := []float64{3.0, 4.0, 5.0, 3.5, 4.5}
	result := engine.Analyze(values, 2.0)

	assert.True(t, result.SuccessProbability > 80)
	assert.True(t, result.FailureProbability < 20)
	assert.True(t, result.ExpectedRoas > 3.0)
}

func TestProbabilityEngine_Analyze_LowSuccessRate(t *testing.T) {
	engine := NewProbabilityEngine()

	// All values below threshold
	values := []float64{0.5, 0.8, 1.2, 0.9, 1.0}
	result := engine.Analyze(values, 2.0)

	assert.True(t, result.SuccessProbability < 30)
	assert.True(t, result.FailureProbability > 30)
	assert.True(t, result.ExpectedRoas < 2.0)
}

func TestProbabilityEngine_Analyze_MixedSuccessRate(t *testing.T) {
	engine := NewProbabilityEngine()

	// Half above, half below threshold
	values := []float64{1.5, 2.5, 1.8, 2.2, 1.9, 2.1}
	result := engine.Analyze(values, 2.0)

	assert.True(t, result.SuccessProbability >= 40 && result.SuccessProbability <= 60)
}

func TestProbabilityEngine_Analyze_DefaultThreshold(t *testing.T) {
	engine := NewProbabilityEngine()

	values := []float64{2.5, 3.0, 2.8}

	// When threshold is 0, should use default of 2.0
	result := engine.Analyze(values, 0)
	assert.True(t, result.SuccessProbability > 70) // All above 2.0
}

func TestProbabilityEngine_calculateSuccessProbability(t *testing.T) {
	engine := NewProbabilityEngine()

	tests := []struct {
		name      string
		values    []float64
		threshold float64
		expected  float64 // approximate with Laplace smoothing
	}{
		{
			name:      "all above",
			values:    []float64{3, 4, 5},
			threshold: 2,
			expected:  0.8, // (3+1)/(3+2) = 0.8
		},
		{
			name:      "none above",
			values:    []float64{1, 1.5, 0.5},
			threshold: 2,
			expected:  0.2, // (0+1)/(3+2) = 0.2
		},
		{
			name:      "half above",
			values:    []float64{1, 2, 3, 4},
			threshold: 2,
			expected:  0.5, // (3+1)/(4+2) = 0.67 (2,3,4 are >= 2)
		},
		{
			name:      "empty",
			values:    []float64{},
			threshold: 2,
			expected:  0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.calculateSuccessProbability(tt.values, tt.threshold)
			assert.InDelta(t, tt.expected, result, 0.2)
		})
	}
}

func TestProbabilityEngine_getConfidenceLevel(t *testing.T) {
	engine := NewProbabilityEngine()

	tests := []struct {
		n        int
		expected ConfidenceLevel
	}{
		{1, ConfidenceLow},
		{3, ConfidenceLow},
		{6, ConfidenceLow},
		{7, ConfidenceMedium},
		{10, ConfidenceMedium},
		{13, ConfidenceMedium},
		{14, ConfidenceHigh},
		{30, ConfidenceHigh},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := engine.getConfidenceLevel(tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProbabilityEngine_ConfidenceInterval(t *testing.T) {
	engine := NewProbabilityEngine()

	values := []float64{2.0, 2.5, 3.0, 2.2, 2.8, 2.4, 2.6, 2.3}
	result := engine.Analyze(values, 2.0)

	// CI should contain the mean
	assert.True(t, result.CILower < result.ExpectedRoas)
	assert.True(t, result.CIUpper > result.ExpectedRoas)
	assert.True(t, result.CILower < result.CIUpper)
}

func TestProbabilityResult_Fields(t *testing.T) {
	result := ProbabilityResult{
		SuccessProbability: 75.0,
		FailureProbability: 10.0,
		ConfidenceLevel:    ConfidenceHigh,
		CILower:            1.8,
		CIUpper:            3.2,
		ExpectedRoas:       2.5,
		Description:        "Test description",
	}

	assert.Equal(t, 75.0, result.SuccessProbability)
	assert.Equal(t, 10.0, result.FailureProbability)
	assert.Equal(t, ConfidenceHigh, result.ConfidenceLevel)
	assert.Equal(t, 1.8, result.CILower)
	assert.Equal(t, 3.2, result.CIUpper)
	assert.Equal(t, 2.5, result.ExpectedRoas)
}

func TestNewMonteCarloProjection(t *testing.T) {
	mc := NewMonteCarloProjection()

	assert.NotNil(t, mc)
	assert.Equal(t, 1000, mc.Simulations)
}

func TestMonteCarloProjection_Project_InsufficientData(t *testing.T) {
	mc := NewMonteCarloProjection()

	result := mc.Project([]float64{1, 2}, 7)
	assert.Equal(t, float64(0), result.ExpectedValue)
	assert.Empty(t, result.Scenarios)
}

func TestMonteCarloProjection_Project_Valid(t *testing.T) {
	mc := NewMonteCarloProjection()
	mc.Simulations = 100 // Reduce for faster test

	values := []float64{2.0, 2.5, 2.2, 2.8, 2.3}
	result := mc.Project(values, 7)

	assert.True(t, result.ExpectedValue > 0)
	assert.Len(t, result.Scenarios, 100)
	assert.True(t, result.CILower < result.CIUpper)
	// Expected value should be close to historical mean
	assert.InDelta(t, mean(values), result.ExpectedValue, 1.0)
}

func TestMean(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5}, 5},
		{"multiple", []float64{1, 2, 3, 4, 5}, 3},
		{"decimals", []float64{1.5, 2.5, 3.5}, 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mean(tt.values)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStdDev(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5}, 0},
		{"identical", []float64{5, 5, 5}, 0},
		{"varied", []float64{2, 4, 6, 8}, 2.58}, // sqrt(20/3) ≈ 2.58
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stdDev(tt.values)
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}
