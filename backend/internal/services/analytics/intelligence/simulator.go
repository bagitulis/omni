package intelligence

import (
	"math"
	"time"
)

// FeasibilityStatus represents budget feasibility
type FeasibilityStatus string

const (
	FeasibilityAchievable    FeasibilityStatus = "ACHIEVABLE"
	FeasibilityDifficult     FeasibilityStatus = "DIFFICULT"
	FeasibilityNotAchievable FeasibilityStatus = "NOT_ACHIEVABLE"
)

// TrendPrediction represents trend prediction
type TrendPrediction string

const (
	TrendPredictionUp       TrendPrediction = "UP"
	TrendPredictionDown     TrendPrediction = "DOWN"
	TrendPredictionStagnant TrendPrediction = "STAGNANT"
)

// BudgetAlternative represents an alternative budget scenario
type BudgetAlternative struct {
	TargetRoas     float64 `json:"target_roas"`
	RequiredBudget float64 `json:"required_budget"`
	ExpectedRoas   float64 `json:"expected_roas"`
}

// SimulationRequest contains simulation input parameters
type SimulationRequest struct {
	ProductID    string  `json:"product_id"`
	TargetRoas   float64 `json:"target_roas"`
	BudgetPerDay float64 `json:"budget_per_day"`
	PeriodDays   int     `json:"period_days"`
}

// SimulationResult contains simulation output
type SimulationResult struct {
	Feasibility       FeasibilityStatus   `json:"feasibility"`
	ConfidencePercent float64             `json:"confidence_percent"`
	CurrentRoas       float64             `json:"current_roas"`
	ProjectedRoas     float64             `json:"projected_roas"`
	TrendPrediction   TrendPrediction     `json:"trend_prediction"`
	OptimalBudget     float64             `json:"optimal_budget"`
	Recommendation    string              `json:"recommendation"`
	Alternatives      []BudgetAlternative `json:"alternatives"`
}

// ProductHistoricalData contains historical data for a product
type ProductHistoricalData struct {
	ProductID      string
	ProductName    string
	SpendHistory   []float64
	RevenueHistory []float64
	RoasHistory    []float64
	CurrentRoas    float64
	CurrentSpend   float64
	DaysOfData     int
}

// BudgetSimulator simulates budget scenarios
type BudgetSimulator struct {
	calendar    *IndonesianCalendar
	saturation  *SaturationModel
	trend       *TrendAnalyzer
	probability *ProbabilityEngine
	scorer      *CompositeScorer
}

// NewBudgetSimulator creates a new budget simulator
func NewBudgetSimulator() *BudgetSimulator {
	return &BudgetSimulator{
		calendar:    NewIndonesianCalendar(),
		saturation:  NewSaturationModel(),
		trend:       NewTrendAnalyzer(),
		probability: NewProbabilityEngine(),
		scorer:      NewCompositeScorer(),
	}
}

// Simulate runs budget simulation with target ROAS and budget
func (s *BudgetSimulator) Simulate(req SimulationRequest, data ProductHistoricalData) SimulationResult {
	if len(data.RoasHistory) < 3 {
		return s.insufficientData()
	}

	// Analyze current state
	trendResult := s.trend.Analyze(data.RoasHistory)
	satResult := s.saturation.Analyze(data.SpendHistory, data.RevenueHistory)
	probResult := s.probability.Analyze(data.RoasHistory, req.TargetRoas)

	// Calculate projected ROAS with new budget
	projectedRoas := s.projectRoasWithBudget(
		data.CurrentRoas,
		data.CurrentSpend,
		req.BudgetPerDay,
		satResult,
	)

	// Determine feasibility
	feasibility := s.determineFeasibility(req.TargetRoas, projectedRoas, probResult)

	// Calculate confidence
	confidence := s.calculateConfidence(probResult, trendResult, len(data.RoasHistory))

	// Determine trend prediction
	trend := s.determineTrendPrediction(projectedRoas, data.CurrentRoas, trendResult)

	// Calculate optimal budget for target ROAS
	optimalBudget := s.calculateOptimalBudget(req.TargetRoas, data, satResult)

	// Generate recommendation
	recommendation := s.generateRecommendation(
		req.TargetRoas,
		req.BudgetPerDay,
		projectedRoas,
		optimalBudget,
		feasibility,
	)

	// Generate alternatives
	alternatives := s.generateAlternatives(req.TargetRoas, req.BudgetPerDay, data, satResult)

	return SimulationResult{
		Feasibility:       feasibility,
		ConfidencePercent: confidence,
		CurrentRoas:       math.Round(data.CurrentRoas*100) / 100,
		ProjectedRoas:     math.Round(projectedRoas*100) / 100,
		TrendPrediction:   trend,
		OptimalBudget:     math.Round(optimalBudget),
		Recommendation:    recommendation,
		Alternatives:      alternatives,
	}
}

// projectRoasWithBudget projects ROAS based on budget change
func (s *BudgetSimulator) projectRoasWithBudget(
	currentRoas, currentSpend, newBudget float64,
	satResult SaturationResult,
) float64 {
	if currentSpend == 0 {
		return currentRoas
	}

	budgetChangeRatio := newBudget / currentSpend

	// Apply diminishing returns based on saturation
	var efficiencyFactor float64
	switch satResult.Status {
	case SaturationHighElasticity:
		efficiencyFactor = 0.95
	case SaturationModerate:
		efficiencyFactor = 0.85
	case SaturationApproachingSaturation:
		efficiencyFactor = 0.70
	case SaturationSaturated:
		efficiencyFactor = 0.50
	case SaturationOverSaturated:
		efficiencyFactor = 0.30
	default:
		efficiencyFactor = 0.80
	}

	if budgetChangeRatio > 1 {
		// Increasing budget: diminishing returns
		roasDecrease := (budgetChangeRatio - 1) * (1 - efficiencyFactor)
		return currentRoas * math.Max(0.3, 1-roasDecrease)
	} else if budgetChangeRatio < 1 {
		// Decreasing budget: may improve efficiency, but capped at 20% max improvement
		// Rationale: reducing budget doesn't linearly increase ROAS — fewer impressions
		// means worse placement and less data for optimization
		rawIncrease := (1 - budgetChangeRatio) * efficiencyFactor * 0.3
		cappedIncrease := math.Min(rawIncrease, 0.20) // Max 20% ROAS improvement
		return currentRoas * (1 + cappedIncrease)
	}

	return currentRoas
}

// determineFeasibility determines if target ROAS is achievable
func (s *BudgetSimulator) determineFeasibility(
	targetRoas, projectedRoas float64,
	probResult ProbabilityResult,
) FeasibilityStatus {
	// If projected ROAS meets or exceeds target, it's achievable or at worst difficult
	if projectedRoas >= targetRoas {
		if probResult.SuccessProbability >= 50 {
			return FeasibilityAchievable
		}
		return FeasibilityDifficult
	}

	roasGap := targetRoas - projectedRoas
	roasGapPct := (roasGap / targetRoas) * 100

	if roasGapPct <= 5 && probResult.SuccessProbability >= 50 {
		return FeasibilityAchievable
	}
	if roasGapPct <= 20 && probResult.SuccessProbability >= 30 {
		return FeasibilityDifficult
	}
	return FeasibilityNotAchievable
}

// calculateConfidence calculates confidence percentage
func (s *BudgetSimulator) calculateConfidence(
	probResult ProbabilityResult,
	trendResult TrendResult,
	dataPoints int,
) float64 {
	// Base confidence from data volume
	var dataConfidence float64
	if dataPoints >= 30 {
		dataConfidence = 90
	} else if dataPoints >= 14 {
		dataConfidence = 75
	} else if dataPoints >= 7 {
		dataConfidence = 60
	} else {
		dataConfidence = 40
	}

	// Adjust by trend stability
	trendAdjustment := (trendResult.Score - 50) / 5

	// Combine
	confidence := dataConfidence + trendAdjustment
	return math.Min(95, math.Max(30, confidence))
}

// determineTrendPrediction determines trend prediction
func (s *BudgetSimulator) determineTrendPrediction(
	projectedRoas, currentRoas float64,
	trendResult TrendResult,
) TrendPrediction {
	roasChange := (projectedRoas - currentRoas) / currentRoas * 100

	// Significant change threshold: 5%
	if roasChange > 5 {
		return TrendPredictionUp
	}
	if roasChange < -5 {
		return TrendPredictionDown
	}
	return TrendPredictionStagnant
}

// calculateOptimalBudget calculates optimal budget for target ROAS
func (s *BudgetSimulator) calculateOptimalBudget(
	targetRoas float64,
	data ProductHistoricalData,
	satResult SaturationResult,
) float64 {
	if data.CurrentRoas == 0 || data.CurrentSpend == 0 {
		return data.CurrentSpend
	}

	var optimalBudget float64
	if data.CurrentRoas >= targetRoas {
		// Already achieving target — can increase budget with headroom
		// Cap headroom factor to prevent unreasonable budgets
		headroomFactor := 1 + math.Min(satResult.Headroom/100*0.5, 0.5)
		optimalBudget = data.CurrentSpend * headroomFactor
	} else {
		// Need to reduce budget to improve ROAS
		roasImprovement := targetRoas / data.CurrentRoas
		budgetReduction := 1 / roasImprovement
		optimalBudget = data.CurrentSpend * budgetReduction
	}

	// Cap optimal budget to reasonable range (never more than 5x current spend)
	maxBudget := data.CurrentSpend * 5
	if optimalBudget > maxBudget {
		optimalBudget = maxBudget
	}

	// Minimum budget floor
	if optimalBudget < 10000 {
		optimalBudget = 10000
	}

	return optimalBudget
}

// generateRecommendation generates Indonesian recommendation text
func (s *BudgetSimulator) generateRecommendation(
	targetRoas, budgetPerDay, projectedRoas, optimalBudget float64,
	feasibility FeasibilityStatus,
) string {
	optBudgetStr := formatRupiah(optimalBudget)
	budgetStr := formatRupiah(budgetPerDay)
	projRoasStr := formatFloat(projectedRoas)
	targetRoasStr := formatFloat(targetRoas)

	switch feasibility {
	case FeasibilityAchievable:
		if budgetPerDay > optimalBudget {
			return "Untuk mencapai ROAS " + targetRoasStr + "x, budget optimal adalah " +
				optBudgetStr + "/hari. Dengan " + budgetStr + "/hari, target tetap tercapai dengan " +
				"projected ROAS " + projRoasStr + "x."
		}
		return "Target ROAS " + targetRoasStr + "x achievable dengan budget " + budgetStr +
			"/hari. Projected ROAS: " + projRoasStr + "x."

	case FeasibilityDifficult:
		return "Untuk mencapai ROAS " + targetRoasStr + "x, budget optimal adalah " +
			optBudgetStr + "/hari. Dengan " + budgetStr + "/hari, ROAS akan turun ke " +
			projRoasStr + "x karena diminishing returns. Alternatif: Naikkan target ke ROAS " +
			projRoasStr + "x untuk budget " + budgetStr + "/hari."

	default:
		return "Target ROAS " + targetRoasStr + "x sulit dicapai dengan kondisi saat ini. " +
			"Pertimbangkan untuk menurunkan target ke ROAS " + projRoasStr +
			"x atau optimalkan creative terlebih dahulu."
	}
}

// generateAlternatives generates alternative scenarios
func (s *BudgetSimulator) generateAlternatives(
	targetRoas, budgetPerDay float64,
	data ProductHistoricalData,
	satResult SaturationResult,
) []BudgetAlternative {
	alternatives := make([]BudgetAlternative, 0, 2)

	// Alternative 1: Lower target ROAS for current budget
	projectedRoas := s.projectRoasWithBudget(data.CurrentRoas, data.CurrentSpend, budgetPerDay, satResult)
	if projectedRoas < targetRoas {
		alternatives = append(alternatives, BudgetAlternative{
			TargetRoas:     math.Round(projectedRoas*10) / 10,
			RequiredBudget: budgetPerDay,
			ExpectedRoas:   projectedRoas,
		})
	}

	// Alternative 2: Budget needed for target ROAS
	optimalBudget := s.calculateOptimalBudget(targetRoas, data, satResult)
	if math.Abs(optimalBudget-budgetPerDay) > 10000 {
		alternatives = append(alternatives, BudgetAlternative{
			TargetRoas:     targetRoas,
			RequiredBudget: optimalBudget,
			ExpectedRoas:   targetRoas,
		})
	}

	return alternatives
}

func (s *BudgetSimulator) insufficientData() SimulationResult {
	return SimulationResult{
		Feasibility:       FeasibilityDifficult,
		ConfidencePercent: 30,
		Recommendation:    "Data historis tidak cukup untuk simulasi akurat. Minimal 7 hari data diperlukan.",
		Alternatives:      []BudgetAlternative{},
	}
}

// GetCalendarContext gets calendar context for a date range
func (s *BudgetSimulator) GetCalendarContext(startDate time.Time, days int) map[string]interface{} {
	events := s.calendar.GetUpcomingEvents(startDate, days)

	// Calculate average multiplier
	var totalMult float64
	for i := 0; i < days; i++ {
		d := startDate.AddDate(0, 0, i)
		totalMult += s.calendar.GetMultiplier(d)
	}
	avgMult := totalMult / float64(days)

	return map[string]interface{}{
		"upcoming_events":    events,
		"average_multiplier": math.Round(avgMult*100) / 100,
		"is_favorable":       avgMult >= 1.0,
	}
}

func formatRupiah(amount float64) string {
	// Simple formatting for Indonesian Rupiah
	if amount >= 1000000 {
		return "Rp " + formatFloat(amount/1000000) + " jt"
	}
	if amount >= 1000 {
		return "Rp " + intToString(int(amount/1000)) + " rb"
	}
	return "Rp " + intToString(int(amount))
}
