package ads

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoefficientOfVariation(t *testing.T) {
	tests := []struct {
		name       string
		values     []float64
		minSamples int
		expected   float64
	}{
		{
			name:       "insufficient samples",
			values:     []float64{100},
			minSamples: 2,
			expected:   0,
		},
		{
			name:       "identical values",
			values:     []float64{100, 100, 100},
			minSamples: 2,
			expected:   0,
		},
		{
			name:       "some variation",
			values:     []float64{90, 100, 110},
			minSamples: 2,
			expected:   8.16, // StdDev/Mean * 100
		},
		{
			name:       "high variation",
			values:     []float64{50, 100, 150},
			minSamples: 2,
			expected:   40.82, // Higher variation
		},
		{
			name:       "zero mean positive stddev",
			values:     []float64{-10, 0, 10},
			minSamples: 2,
			expected:   100, // Special case
		},
		{
			name:       "zero mean zero stddev",
			values:     []float64{0, 0, 0},
			minSamples: 2,
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CoefficientOfVariation(tt.values, tt.minSamples)
			assert.InDelta(t, tt.expected, result, 1.0)
		})
	}
}

func TestCoefficientOfVariation_MinSamplesDefault(t *testing.T) {
	values := []float64{100, 110}

	// minSamples < 2 should default to 2
	result := CoefficientOfVariation(values, 1)
	assert.True(t, result > 0) // Should calculate since we have 2 values
}

func TestGetCVLabel(t *testing.T) {
	tests := []struct {
		cv            float64
		expectedLabel CVLabel
		containsDesc  string
	}{
		{10, CVVeryConsistent, "Konsisten"},
		{19.9, CVVeryConsistent, "Konsisten"},
		{20, CVConsistent, "Konsisten"},
		{34.9, CVConsistent, "Konsisten"},
		{35, CVModerate, "Variabel"},
		{49.9, CVModerate, "Variabel"},
		{50, CVVolatile, "Volatil"},
		{100, CVVolatile, "Volatil"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			label, desc := GetCVLabel(tt.cv)
			assert.Equal(t, tt.expectedLabel, label)
			assert.Contains(t, desc, tt.containsDesc)
		})
	}
}

func TestAnalyzeVariation(t *testing.T) {
	t.Run("insufficient data", func(t *testing.T) {
		result := AnalyzeVariation([]float64{100})
		assert.False(t, result.IsValid)
		assert.Equal(t, CVModerate, result.Label)
		assert.Equal(t, float64(0), result.CV)
	})

	t.Run("valid consistent data", func(t *testing.T) {
		result := AnalyzeVariation([]float64{100, 101, 99, 100})
		assert.True(t, result.IsValid)
		assert.True(t, result.CV < 20) // Should be very consistent
		assert.Equal(t, CVVeryConsistent, result.Label)
	})

	t.Run("valid volatile data", func(t *testing.T) {
		result := AnalyzeVariation([]float64{50, 150, 25, 200})
		assert.True(t, result.IsValid)
		assert.True(t, result.CV > 50)
		assert.Equal(t, CVVolatile, result.Label)
	})
}

func TestGetConsistencyScore(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		minScore float64
		maxScore float64
	}{
		{
			name:     "insufficient data",
			values:   []float64{100},
			minScore: 50,
			maxScore: 50,
		},
		{
			name:     "identical values",
			values:   []float64{100, 100, 100},
			minScore: 99,
			maxScore: 100,
		},
		{
			name:     "consistent values",
			values:   []float64{99, 100, 101},
			minScore: 90,
			maxScore: 100,
		},
		{
			name:     "volatile values",
			values:   []float64{50, 150, 50, 150},
			minScore: 0,
			maxScore: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := GetConsistencyScore(tt.values)
			assert.True(t, score >= tt.minScore && score <= tt.maxScore,
				"Score %f not in range [%f, %f]", score, tt.minScore, tt.maxScore)
		})
	}
}

func TestCVResult_Fields(t *testing.T) {
	result := CVResult{
		CV:           25.5,
		Label:        CVConsistent,
		LabelDisplay: "Konsisten",
		IsValid:      true,
	}

	assert.Equal(t, 25.5, result.CV)
	assert.Equal(t, CVConsistent, result.Label)
	assert.Equal(t, "Konsisten", result.LabelDisplay)
	assert.True(t, result.IsValid)
}
