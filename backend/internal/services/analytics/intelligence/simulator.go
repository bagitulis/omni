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

// SimulationMode indicates which math model was used
type SimulationMode string

const (
	SimulationModeFunnel SimulationMode = "FUNNEL"
	SimulationModeYield  SimulationMode = "YIELD"
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
	SimulationMode    SimulationMode      `json:"simulation_mode"`
	MLCategory        string              `json:"ml_category"`
	CurrentDailySpend float64             `json:"current_daily_spend"`
}

// ProductHistoricalData contains historical data for a product
type ProductHistoricalData struct {
	ProductID        string
	ProductName      string
	SpendHistory     []float64
	RevenueHistory   []float64
	RoasHistory      []float64
	CurrentRoas      float64
	CurrentSpend     float64 // Daily average spend
	DaysOfData       int
	TotalClicks      int
	TotalOrders      int
	TotalImpressions int
	MLCategory       string // from ML cache: STAR, GROWTH, STABLE, WATCH, PROBLEM
	FatigueStatus    string // FRESH, AGING, FATIGUED, DEAD
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

	// Calculate projected ROAS with new budget (dual-mode)
	projectedRoas, simMode := s.projectRoasWithBudget(
		data.CurrentRoas, data.CurrentSpend, req.BudgetPerDay,
		satResult, data,
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
		req.TargetRoas, req.BudgetPerDay, projectedRoas,
		optimalBudget, feasibility,
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
		SimulationMode:    simMode,
		MLCategory:        data.MLCategory,
		CurrentDailySpend: math.Round(data.CurrentSpend),
	}
}

// getElasticityForCategory returns the power-law elasticity exponent
// based on ML product category and creative fatigue status.
// More negative = steeper diminishing returns when scaling up.
func getElasticityForCategory(category, fatigueStatus string) float64 {
	base := -0.12 // Default: moderate diminishing returns
	switch category {
	case "STAR":
		base = -0.06 // Strong performers scale well
	case "GROWTH":
		base = -0.10 // Growing products — reasonable headroom
	case "STABLE":
		base = -0.15 // Stable but limited upside
	case "WATCH":
		base = -0.22 // Already struggling
	case "PROBLEM":
		base = -0.35 // Scaling will worsen losses
	}
	// Fatigue penalty — fatigued creatives have steeper drop-off
	switch fatigueStatus {
	case "AGING":
		base -= 0.03
	case "FATIGUED":
		base -= 0.08
	case "DEAD":
		base -= 0.15
	}
	return base
}

// getSaturationPenalty adjusts elasticity based on saturation analysis
func getSaturationPenalty(satResult SaturationResult) float64 {
	switch satResult.Status {
	case SaturationHighElasticity:
		return 1.0 // No penalty
	case SaturationModerate:
		return 1.1 // Slightly steeper
	case SaturationApproachingSaturation:
		return 1.3
	case SaturationSaturated:
		return 1.6
	case SaturationOverSaturated:
		return 2.0 // Very steep
	default:
		return 1.0
	}
}

// projectRoasWithBudget projects ROAS using dual-mode power-law model.
// Mode A (Funnel): uses CPC/CVR/CTR when click+order data is available.
// Mode B (Yield): uses raw power-law on ROAS when funnel data is incomplete.
func (s *BudgetSimulator) projectRoasWithBudget(
	currentRoas, currentSpend, newBudget float64,
	satResult SaturationResult, data ProductHistoricalData,
) (float64, SimulationMode) {
	if currentSpend <= 0 {
		return currentRoas, SimulationModeYield
	}

	scale := newBudget / currentSpend
	if scale <= 0 {
		return currentRoas, SimulationModeYield
	}

	// Edge case: identical budget
	if math.Abs(scale-1.0) < 0.001 {
		return currentRoas, SimulationModeYield
	}

	// Determine elasticity from ML context
	elasticity := getElasticityForCategory(data.MLCategory, data.FatigueStatus)
	elasticity *= getSaturationPenalty(satResult)

	// Dual-mode selection
	if data.TotalClicks > 0 && data.TotalOrders > 0 && data.TotalImpressions > 0 {
		roas := s.projectFunnel(currentRoas, currentSpend, scale, elasticity, data)
		return roas, SimulationModeFunnel
	}
	roas := s.projectYield(currentRoas, scale, elasticity)
	return roas, SimulationModeYield
}

// projectFunnel simulates the full advertising funnel:
// Budget → Impressions → Clicks (CTR) → Orders (CVR) → Revenue (AOV)
// Each stage degrades with a power-law penalty as budget scales up.
func (s *BudgetSimulator) projectFunnel(
	currentRoas, currentSpend, scale, elasticity float64,
	data ProductHistoricalData,
) float64 {
	totalSpend := currentSpend * float64(data.DaysOfData)

	baseCPC := totalSpend / float64(data.TotalClicks)
	baseCVR := float64(data.TotalOrders) / float64(data.TotalClicks)
	baseAOV := (totalSpend * currentRoas) / float64(data.TotalOrders)

	if baseCPC <= 0 || baseCVR <= 0 || baseAOV <= 0 {
		return s.projectYield(currentRoas, scale, elasticity)
	}

	// Apply scale penalties (power-law)
	// CPC rises as budget increases (auction competition)
	adjCPC := baseCPC * math.Pow(scale, 0.08)
	// CVR drops as budget increases (audience dilution)
	adjCVR := baseCVR * math.Pow(scale, elasticity*1.2)

	// Guard: prevent CVR from going negative or absurdly high
	if adjCVR <= 0 {
		adjCVR = baseCVR * 0.01
	}
	if adjCVR > 1.0 {
		adjCVR = baseCVR // cap at baseline
	}

	// Project daily metrics
	newBudget := currentSpend * scale
	dailyClicks := newBudget / adjCPC
	dailyOrders := dailyClicks * adjCVR
	dailyRevenue := dailyOrders * baseAOV

	if newBudget > 0 {
		return dailyRevenue / newBudget
	}
	return currentRoas
}

// projectYield uses a simple power-law model on ROAS directly.
// Used as fallback when funnel data (clicks/orders) is incomplete.
func (s *BudgetSimulator) projectYield(currentRoas, scale, elasticity float64) float64 {
	if scale <= 0 {
		return currentRoas
	}
	projected := currentRoas * math.Pow(scale, elasticity)
	// Safety floor: never project below 5% of current ROAS
	if projected < currentRoas*0.05 {
		projected = currentRoas * 0.05
	}
	return projected
}

// determineFeasibility determines if target ROAS is achievable
func (s *BudgetSimulator) determineFeasibility(
	targetRoas, projectedRoas float64,
	probResult ProbabilityResult,
) FeasibilityStatus {
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

	trendAdjustment := (trendResult.Score - 50) / 5
	confidence := dataConfidence + trendAdjustment
	return math.Min(95, math.Max(30, confidence))
}

// determineTrendPrediction determines trend prediction
func (s *BudgetSimulator) determineTrendPrediction(
	projectedRoas, currentRoas float64,
	trendResult TrendResult,
) TrendPrediction {
	roasChange := (projectedRoas - currentRoas) / currentRoas * 100
	if roasChange > 5 {
		return TrendPredictionUp
	}
	if roasChange < -5 {
		return TrendPredictionDown
	}
	return TrendPredictionStagnant
}

// calculateOptimalBudget finds the daily budget that achieves target ROAS
// using binary search over the power-law projection.
func (s *BudgetSimulator) calculateOptimalBudget(
	targetRoas float64,
	data ProductHistoricalData,
	satResult SaturationResult,
) float64 {
	if data.CurrentRoas == 0 || data.CurrentSpend == 0 {
		return data.CurrentSpend
	}

	// If current ROAS is already below target and scaling down won't help enough
	if data.CurrentRoas < targetRoas*0.3 {
		return math.Max(10000, data.CurrentSpend*0.5)
	}

	// Binary search for optimal budget
	low := data.CurrentSpend * 0.05  // 5% of current
	high := data.CurrentSpend * 10.0 // 10x of current

	for i := 0; i < 50; i++ {
		mid := (low + high) / 2
		projectedRoas, _ := s.projectRoasWithBudget(
			data.CurrentRoas, data.CurrentSpend, mid,
			satResult, data,
		)

		if math.Abs(projectedRoas-targetRoas) < 0.01 {
			break
		}

		// Higher budget → lower ROAS (diminishing returns)
		if projectedRoas > targetRoas {
			low = mid // Can increase budget more
		} else {
			high = mid // Need to decrease budget
		}
	}

	optimal := (low + high) / 2

	// Cap to reasonable range
	if optimal < 10000 {
		optimal = 10000
	}
	if optimal > data.CurrentSpend*10 {
		optimal = data.CurrentSpend * 10
	}

	return optimal
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
	projectedRoas, _ := s.projectRoasWithBudget(
		data.CurrentRoas, data.CurrentSpend, budgetPerDay,
		satResult, data,
	)
	if projectedRoas < targetRoas {
		alternatives = append(alternatives, BudgetAlternative{
			TargetRoas:     math.Round(projectedRoas*10) / 10,
			RequiredBudget: budgetPerDay,
			ExpectedRoas:   math.Round(projectedRoas*100) / 100,
		})
	}

	// Alternative 2: Budget needed for target ROAS
	optimalBudget := s.calculateOptimalBudget(targetRoas, data, satResult)
	if math.Abs(optimalBudget-budgetPerDay) > 10000 {
		alternatives = append(alternatives, BudgetAlternative{
			TargetRoas:     targetRoas,
			RequiredBudget: math.Round(optimalBudget),
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
	if amount >= 1000000 {
		return "Rp " + formatFloat(amount/1000000) + " jt"
	}
	if amount >= 1000 {
		return "Rp " + intToString(int(amount/1000)) + " rb"
	}
	return "Rp " + intToString(int(amount))
}
