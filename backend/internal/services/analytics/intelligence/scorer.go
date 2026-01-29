package intelligence

import (
	"math"
)

// ScoreCategory represents score categories
type ScoreCategory string

const (
	ScoreCategoryStar           ScoreCategory = "STAR"
	ScoreCategoryPerformer      ScoreCategory = "PERFORMER"
	ScoreCategoryAverage        ScoreCategory = "AVERAGE"
	ScoreCategoryUnderperformer ScoreCategory = "UNDERPERFORMER"
	ScoreCategoryProblem        ScoreCategory = "PROBLEM"
	ScoreCategoryStop           ScoreCategory = "STOP"
)

// RecommendationAction represents recommended actions
type RecommendationAction string

const (
	ActionScaleUpAggressive RecommendationAction = "SCALE_UP_AGGRESSIVE"
	ActionScaleUpModerate   RecommendationAction = "SCALE_UP_MODERATE"
	ActionMaintain          RecommendationAction = "MAINTAIN"
	ActionReduceBudget      RecommendationAction = "REDUCE_BUDGET"
	ActionStopImmediately   RecommendationAction = "STOP_IMMEDIATELY"
)

// ComponentScore represents a single scoring component
type ComponentScore struct {
	Name          string  `json:"name"`
	Score         float64 `json:"score"`
	Weight        float64 `json:"weight"`
	WeightedScore float64 `json:"weighted_score"`
	Description   string  `json:"description"`
}

// CompositeResult contains composite scoring results
type CompositeResult struct {
	FinalScore  float64              `json:"final_score"`
	Category    ScoreCategory        `json:"category"`
	Action      RecommendationAction `json:"action"`
	Components  []ComponentScore     `json:"components"`
	Breakdown   map[string]float64   `json:"breakdown"`
	Description string               `json:"description"`
}

// ScoringWeights defines weights for each component
type ScoringWeights struct {
	Roas       float64
	Trend      float64
	Volatility float64
	Fatigue    float64
	Elasticity float64
	Event      float64
}

// DefaultScoringWeights returns default weights
func DefaultScoringWeights() ScoringWeights {
	return ScoringWeights{
		Roas:       0.30,
		Trend:      0.20,
		Volatility: 0.15,
		Fatigue:    0.15,
		Elasticity: 0.10,
		Event:      0.10,
	}
}

// CompositeScorer calculates composite scores
type CompositeScorer struct {
	Weights ScoringWeights
}

// NewCompositeScorer creates a new scorer
func NewCompositeScorer() *CompositeScorer {
	return &CompositeScorer{
		Weights: DefaultScoringWeights(),
	}
}

// ScoreInput contains all component scores for calculation
type ScoreInput struct {
	RoasScore       float64
	TrendScore      float64
	VolatilityScore float64
	FatigueScore    float64
	ElasticityScore float64
	EventScore      float64
}

// Score calculates the composite score from components
func (s *CompositeScorer) Score(input ScoreInput) CompositeResult {
	// Default event score if not provided
	if input.EventScore == 0 {
		input.EventScore = 50
	}

	scores := map[string]float64{
		"roas":       input.RoasScore,
		"trend":      input.TrendScore,
		"volatility": input.VolatilityScore,
		"fatigue":    input.FatigueScore,
		"elasticity": input.ElasticityScore,
		"event":      input.EventScore,
	}

	weights := map[string]float64{
		"roas":       s.Weights.Roas,
		"trend":      s.Weights.Trend,
		"volatility": s.Weights.Volatility,
		"fatigue":    s.Weights.Fatigue,
		"elasticity": s.Weights.Elasticity,
		"event":      s.Weights.Event,
	}

	// Calculate weighted score
	var finalScore float64
	for k, w := range weights {
		finalScore += scores[k] * w
	}

	// Build components
	components := make([]ComponentScore, 0, len(scores))
	for k, score := range scores {
		components = append(components, ComponentScore{
			Name:          k,
			Score:         score,
			Weight:        weights[k],
			WeightedScore: score * weights[k],
			Description:   getComponentDescription(score),
		})
	}

	// Determine category and action
	category := getCategory(finalScore)
	action := getAction(finalScore, input.RoasScore)
	description := getCategoryDescription(category, finalScore)

	return CompositeResult{
		FinalScore:  math.Round(finalScore*10) / 10,
		Category:    category,
		Action:      action,
		Components:  components,
		Breakdown:   scores,
		Description: description,
	}
}

// getCategory determines score category
func getCategory(score float64) ScoreCategory {
	switch {
	case score >= 85:
		return ScoreCategoryStar
	case score >= 70:
		return ScoreCategoryPerformer
	case score >= 55:
		return ScoreCategoryAverage
	case score >= 40:
		return ScoreCategoryUnderperformer
	case score >= 25:
		return ScoreCategoryProblem
	default:
		return ScoreCategoryStop
	}
}

// getAction determines recommended action
func getAction(score, roasScore float64) RecommendationAction {
	// ROAS override
	if roasScore < 20 {
		return ActionStopImmediately
	}

	switch {
	case score >= 80:
		return ActionScaleUpAggressive
	case score >= 65:
		return ActionScaleUpModerate
	case score >= 50:
		return ActionMaintain
	case score >= 35:
		return ActionReduceBudget
	default:
		return ActionStopImmediately
	}
}

// getComponentDescription returns Indonesian description for component score
func getComponentDescription(score float64) string {
	switch {
	case score >= 80:
		return "Sangat baik"
	case score >= 60:
		return "Baik"
	case score >= 40:
		return "Cukup"
	case score >= 20:
		return "Kurang"
	default:
		return "Buruk"
	}
}

// getCategoryDescription returns Indonesian description for category
func getCategoryDescription(category ScoreCategory, score float64) string {
	scoreStr := formatFloat(score)
	switch category {
	case ScoreCategoryStar:
		return "STAR (" + scoreStr + ") - Produk unggulan, scale up agresif!"
	case ScoreCategoryPerformer:
		return "PERFORMER (" + scoreStr + ") - Performa bagus, scale up moderat."
	case ScoreCategoryAverage:
		return "AVERAGE (" + scoreStr + ") - Rata-rata, maintain dan optimasi."
	case ScoreCategoryUnderperformer:
		return "UNDERPERFORMER (" + scoreStr + ") - Di bawah rata-rata, monitor."
	case ScoreCategoryProblem:
		return "PROBLEM (" + scoreStr + ") - Perlu perbaikan segera."
	default:
		return "STOP (" + scoreStr + ") - Stop iklan, evaluasi total."
	}
}

// CalculateRoasScore calculates ROAS score (0-100)
func CalculateRoasScore(roas float64) float64 {
	// ROAS scoring tiers
	switch {
	case roas >= 5:
		return 100
	case roas >= 4:
		return 90
	case roas >= 3:
		return 80
	case roas >= 2:
		return 70
	case roas >= 1.5:
		return 60
	case roas >= 1:
		return 40
	case roas >= 0.5:
		return 20
	default:
		return 0
	}
}

// NormalizeMetric normalizes a metric to 0-100 range
func NormalizeMetric(value, minVal, maxVal float64) float64 {
	if maxVal == minVal {
		return 50
	}
	normalized := ((value - minVal) / (maxVal - minVal)) * 100
	return math.Max(0, math.Min(100, normalized))
}
