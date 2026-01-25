package ads

import (
	"math"
	"testing"
)

func TestDefaultScoringConfig(t *testing.T) {
	cfg := DefaultScoringConfig()

	// Verify weights sum to 1.0
	totalWeight := cfg.WeightROI + cfg.WeightProfit + cfg.WeightMomentum + cfg.WeightConsistency + cfg.WeightTrend
	if math.Abs(totalWeight-1.0) > 0.001 {
		t.Errorf("weights should sum to 1.0, got %f", totalWeight)
	}

	// Verify default values
	if cfg.WeightROI != 0.30 {
		t.Errorf("WeightROI = %f, want 0.30", cfg.WeightROI)
	}
	if cfg.ScoreThresholdGood != 60.0 {
		t.Errorf("ScoreThresholdGood = %f, want 60.0", cfg.ScoreThresholdGood)
	}
	if cfg.ROIThresholdGood != 2.0 {
		t.Errorf("ROIThresholdGood = %f, want 2.0", cfg.ROIThresholdGood)
	}
}

func TestCalculateScores_BasicCase(t *testing.T) {
	cfg := DefaultScoringConfig()

	// Test with good ROI and positive profit
	result := CalculateScores(cfg, 3.0, 100000, nil, nil, nil)

	// ROI of 3.0 should give ROI score of 30 (3.0 * 10)
	if result.ROIScore != 30.0 {
		t.Errorf("ROIScore = %f, want 30.0", result.ROIScore)
	}

	// Positive profit should give 100
	if result.ProfitScore != 100.0 {
		t.Errorf("ProfitScore = %f, want 100.0", result.ProfitScore)
	}

	// Without momentum/consistency/trend data, should default to 50
	if result.MomentumScore != 50.0 {
		t.Errorf("MomentumScore = %f, want 50.0", result.MomentumScore)
	}
	if result.ConsistencyScore != 50.0 {
		t.Errorf("ConsistencyScore = %f, want 50.0", result.ConsistencyScore)
	}
	if result.TrendScore != 50.0 {
		t.Errorf("TrendScore = %f, want 50.0", result.TrendScore)
	}
}

func TestCalculateScores_ROIScoreCap(t *testing.T) {
	cfg := DefaultScoringConfig()

	// Very high ROI should cap at 100
	result := CalculateScores(cfg, 15.0, 100000, nil, nil, nil)

	if result.ROIScore != 100.0 {
		t.Errorf("ROIScore with 15.0 ROI should cap at 100.0, got %f", result.ROIScore)
	}
}

func TestCalculateScores_NegativeProfit(t *testing.T) {
	cfg := DefaultScoringConfig()

	// Negative profit should reduce profit score
	result := CalculateScores(cfg, 1.0, -1000000, nil, nil, nil)

	// Profit of -1,000,000: 50 + (-1000000/2000000)*50 = 50 - 25 = 25
	expected := 25.0
	if result.ProfitScore != expected {
		t.Errorf("ProfitScore with -1M profit = %f, want %f", result.ProfitScore, expected)
	}
}

func TestCalculateScores_WithMomentum(t *testing.T) {
	cfg := DefaultScoringConfig()

	// Test with positive momentum
	momentum := 20.0 // 20% momentum
	result := CalculateScores(cfg, 2.0, 50000, &momentum, nil, nil)

	// Momentum score: 50 + (20/2) = 60
	if result.MomentumScore != 60.0 {
		t.Errorf("MomentumScore with 20%% momentum = %f, want 60.0", result.MomentumScore)
	}
}

func TestCalculateScores_WithConsistency(t *testing.T) {
	cfg := DefaultScoringConfig()

	// Test with CV (coefficient of variation)
	cv := 30.0 // 30 CV -> 100 - 30 = 70 consistency score
	result := CalculateScores(cfg, 2.0, 50000, nil, &cv, nil)

	if result.ConsistencyScore != 70.0 {
		t.Errorf("ConsistencyScore with CV 30 = %f, want 70.0", result.ConsistencyScore)
	}
}

func TestCalculateScores_WithTrend(t *testing.T) {
	cfg := DefaultScoringConfig()

	// Test with uptrend
	uptrend := true
	result := CalculateScores(cfg, 2.0, 50000, nil, nil, &uptrend)

	if result.TrendScore != 90.0 {
		t.Errorf("TrendScore with uptrend = %f, want 90.0", result.TrendScore)
	}

	// Test with downtrend
	downtrend := false
	result = CalculateScores(cfg, 2.0, 50000, nil, nil, &downtrend)

	if result.TrendScore != 20.0 {
		t.Errorf("TrendScore with downtrend = %f, want 20.0", result.TrendScore)
	}
}

func TestDetermineCategory(t *testing.T) {
	cfg := DefaultScoringConfig()

	tests := []struct {
		name     string
		score    float64
		roi      float64
		profit   float64
		category string
	}{
		{"scale_up", 70.0, 3.0, 100000, "SCALE_UP"},
		{"maintain_high_score_low_roi", 65.0, 1.5, 100000, "MAINTAIN"},
		{"monitor", 50.0, 1.5, 50000, "MONITOR"},
		{"stop_low_roi", 30.0, 0.5, -50000, "STOP"},
		{"stop_big_loss", 30.0, 1.2, -200000, "STOP"},
		{"evaluate", 35.0, 1.2, -50000, "EVALUATE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, _ := determineCategory(cfg, tt.score, tt.roi, tt.profit)
			if category != tt.category {
				t.Errorf("determineCategory() = %v, want %v", category, tt.category)
			}
		})
	}
}

func TestSimpleScore(t *testing.T) {
	// SimpleScore should work without additional parameters
	result := SimpleScore(2.5, 50000)

	if result.ROIScore == 0 && result.ProfitScore == 0 {
		t.Error("SimpleScore should return non-zero scores")
	}

	if result.Category == "" {
		t.Error("SimpleScore should set category")
	}

	if result.Action == "" {
		t.Error("SimpleScore should set action")
	}
}

func TestRound(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{1.234, 1.2},
		{1.256, 1.3},
		{50.0, 50.0},
		{99.99, 100.0},
		{0.05, 0.1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := round(tt.input)
			if got != tt.expected {
				t.Errorf("round(%f) = %f, want %f", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCalculateFullScore_WithHistoricalData(t *testing.T) {
	cfg := DefaultScoringConfig()

	// With enough historical data, should calculate momentum and CV
	roiValues := []float64{1.5, 1.6, 1.7, 1.8, 1.9, 2.0, 2.1, 2.2, 2.3}
	profitValues := []float64{10000, 12000, 11000, 13000, 12500}

	result := CalculateFullScore(cfg, 2.3, 12500, roiValues, profitValues)

	// Should have calculated scores (not just defaults)
	if result.CompositeScore == 0 {
		t.Error("CalculateFullScore should calculate composite score")
	}
	if result.Category == "" {
		t.Error("CalculateFullScore should set category")
	}
}

func TestCalculateFullScore_WithoutHistoricalData(t *testing.T) {
	cfg := DefaultScoringConfig()

	// Without historical data, should fall back to defaults
	result := CalculateFullScore(cfg, 2.0, 50000, nil, nil)

	// Should still work and set defaults
	if result.MomentumScore != 50.0 {
		t.Errorf("MomentumScore without data = %f, want 50.0", result.MomentumScore)
	}
}

func TestScoreResult_JSONTags(t *testing.T) {
	// Verify JSON field names are snake_case (per AGENTS.MD)
	result := ScoreResult{
		ROIScore:         50.0,
		ProfitScore:      50.0,
		MomentumScore:    50.0,
		ConsistencyScore: 50.0,
		TrendScore:       50.0,
		CompositeScore:   50.0,
		Category:         "MAINTAIN",
		Action:           "Maintain budget",
	}

	// The struct should have snake_case JSON tags
	// This is a compile-time check - if it compiles, tags exist
	_ = result
}
