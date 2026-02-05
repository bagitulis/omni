package analytics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSkuData(t *testing.T) {
	data := NewSkuData()

	assert.NotNil(t, data)
	assert.NotNil(t, data.UnitPrices)
	assert.NotNil(t, data.ActualIncomes)
	assert.Equal(t, 0, data.Count)
	assert.Empty(t, data.UnitPrices)
	assert.Empty(t, data.ActualIncomes)
}

func TestGetStringValue(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "nil pointer",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty string",
			input:    ptr(""),
			expected: "",
		},
		{
			name:     "non-empty string",
			input:    ptr("hello"),
			expected: "hello",
		},
		{
			name:     "string with spaces",
			input:    ptr("  spaced  "),
			expected: "  spaced  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStringValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapKeysToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    map[float64]struct{}
		expected int // length check since order is not guaranteed
	}{
		{
			name:     "empty map",
			input:    map[float64]struct{}{},
			expected: 0,
		},
		{
			name: "single element",
			input: map[float64]struct{}{
				10.5: {},
			},
			expected: 1,
		},
		{
			name: "multiple elements",
			input: map[float64]struct{}{
				10.5: {},
				20.0: {},
				30.7: {},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapKeysToSlice(tt.input)
			assert.Len(t, result, tt.expected)

			// Verify all keys are in result
			for k := range tt.input {
				assert.Contains(t, result, k)
			}
		})
	}
}

func TestHasDifferentPrice(t *testing.T) {
	tests := []struct {
		name     string
		prices   []float64
		target   float64
		expected bool
	}{
		{
			name:     "empty slice",
			prices:   []float64{},
			target:   100.0,
			expected: false,
		},
		{
			name:     "all same as target",
			prices:   []float64{100.0, 100.0, 100.0},
			target:   100.0,
			expected: false,
		},
		{
			name:     "one different",
			prices:   []float64{100.0, 99.0, 100.0},
			target:   100.0,
			expected: true,
		},
		{
			name:     "first different",
			prices:   []float64{50.0, 100.0, 100.0},
			target:   100.0,
			expected: true,
		},
		{
			name:     "last different",
			prices:   []float64{100.0, 100.0, 150.0},
			target:   100.0,
			expected: true,
		},
		{
			name:     "all different",
			prices:   []float64{50.0, 75.0, 125.0},
			target:   100.0,
			expected: true,
		},
		{
			name:     "single element same",
			prices:   []float64{100.0},
			target:   100.0,
			expected: false,
		},
		{
			name:     "single element different",
			prices:   []float64{99.0},
			target:   100.0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasDifferentPrice(tt.prices, tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSkuData_Usage(t *testing.T) {
	// Test typical usage pattern
	data := NewSkuData()

	// Add unit prices
	data.UnitPrices[100.0] = struct{}{}
	data.UnitPrices[100.0] = struct{}{} // duplicate should not increase count
	data.UnitPrices[150.0] = struct{}{}

	// Add actual incomes
	data.ActualIncomes[90.0] = struct{}{}
	data.ActualIncomes[140.0] = struct{}{}

	// Increment count
	data.Count = 5

	assert.Len(t, data.UnitPrices, 2)
	assert.Len(t, data.ActualIncomes, 2)
	assert.Equal(t, 5, data.Count)

	// Convert to slices
	prices := MapKeysToSlice(data.UnitPrices)
	assert.Len(t, prices, 2)
	assert.Contains(t, prices, 100.0)
	assert.Contains(t, prices, 150.0)
}

// ptr is a helper to create string pointers for testing
func ptr(s string) *string {
	return &s
}
