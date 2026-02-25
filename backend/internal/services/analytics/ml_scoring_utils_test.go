package analytics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundTo(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		decimals int
		expected float64
	}{
		{name: "round to two decimals", value: 12.3456, decimals: 2, expected: 12.35},
		{name: "round to zero decimals", value: 12.5, decimals: 0, expected: 13},
		{name: "negative decimals", value: 1234.5, decimals: -2, expected: 1200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, roundTo(tt.value, tt.decimals))
		})
	}
}

func TestGetHealthLabel(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected string
	}{
		{name: "excellent threshold", score: 70, expected: "Excellent"},
		{name: "good below excellent", score: 69.9, expected: "Good"},
		{name: "good threshold", score: 55, expected: "Good"},
		{name: "fair below good", score: 54.9, expected: "Fair"},
		{name: "fair threshold", score: 40, expected: "Fair"},
		{name: "poor below fair", score: 39.9, expected: "Poor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getHealthLabel(tt.score))
		})
	}
}

func TestFormatPercent(t *testing.T) {
	assert.Equal(t, "0.0%", formatPercent(0))
	assert.Equal(t, "12.3%", formatPercent(12.34))
	assert.Equal(t, "-1.2%", formatPercent(-1.25))
}

func TestMapToCategory(t *testing.T) {
	tests := []struct {
		action   string
		expected string
	}{
		{action: "SCALE_UP", expected: "STAR"},
		{action: "MAINTAIN", expected: "STABLE"},
		{action: "MONITOR", expected: "WATCH"},
		{action: "STOP", expected: "PROBLEM"},
		{action: "EVALUATE", expected: "GROWTH"},
		{action: "UNKNOWN", expected: "WATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			assert.Equal(t, tt.expected, mapToCategory(tt.action))
		})
	}
}

func TestGetBudgetChangePct(t *testing.T) {
	tests := []struct {
		action   string
		expected float64
	}{
		{action: "SCALE_UP", expected: 30.0},
		{action: "MAINTAIN", expected: 0.0},
		{action: "MONITOR", expected: -10.0},
		{action: "STOP", expected: -100.0},
		{action: "UNKNOWN", expected: 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			assert.Equal(t, tt.expected, getBudgetChangePct(tt.action))
		})
	}
}

func TestDetermineFatigueStatus(t *testing.T) {
	tests := []struct {
		name        string
		clicks      int
		impressions int
		periodCount int
		expected    string
	}{
		{name: "fresh when insufficient periods", clicks: 1, impressions: 1000, periodCount: 2, expected: "FRESH"},
		{name: "dead when no impressions", clicks: 10, impressions: 0, periodCount: 3, expected: "DEAD"},
		{name: "fresh on high ctr", clicks: 20, impressions: 1000, periodCount: 5, expected: "FRESH"},
		{name: "aging on medium ctr", clicks: 10, impressions: 1000, periodCount: 5, expected: "AGING"},
		{name: "fatigued on low ctr", clicks: 5, impressions: 1000, periodCount: 5, expected: "FATIGUED"},
		{name: "dead on very low ctr", clicks: 4, impressions: 1000, periodCount: 5, expected: "DEAD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, determineFatigueStatus(tt.clicks, tt.impressions, tt.periodCount))
		})
	}
}

func TestCalculateChurnRisk(t *testing.T) {
	assert.Equal(t, 0.0, calculateChurnRisk(60, 55, 40))
	assert.Equal(t, 16.0, calculateChurnRisk(40, 30, 70))
	assert.Equal(t, 12.2, calculateChurnRisk(33.3, 44.4, 66.6))
}

func TestGetConfidenceLevel(t *testing.T) {
	assert.Equal(t, "HIGH", getConfidenceLevel(8))
	assert.Equal(t, "MEDIUM", getConfidenceLevel(4))
	assert.Equal(t, "LOW", getConfidenceLevel(3))
}

func TestEstimateSuccessProbability(t *testing.T) {
	assert.Equal(t, 0.6, estimateSuccessProbability(60, 1.5))
	assert.Equal(t, 0.7, estimateSuccessProbability(60, 2.0))
	assert.Equal(t, 0.95, estimateSuccessProbability(99, 3.0))
	assert.Equal(t, 0.33, estimateSuccessProbability(33.3, 1.0))
}

func TestGetTrendDirection(t *testing.T) {
	assert.Equal(t, "UP", getTrendDirection(60))
	assert.Equal(t, "STABLE", getTrendDirection(59.9))
	assert.Equal(t, "DOWN", getTrendDirection(40))
	assert.Equal(t, "STABLE", getTrendDirection(40.1))
}
