package intelligence

import (
	"math"
	"time"
)

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
	projectedRoas, simMode, funnel := s.projectRoasWithBudget(
		data.CurrentRoas, data.CurrentSpend, req.BudgetPerDay,
		satResult, data,
	)

	// Calendar context for the simulation period
	calendarCtx := s.GetCalendarContext(time.Now(), req.PeriodDays)
	avgMult := calendarCtx["average_multiplier"].(float64)
	calendarEvents, _ := calendarCtx["upcoming_events"].([]map[string]interface{})
	if calendarEvents == nil {
		calendarEvents = []map[string]interface{}{}
	}

	// Apply calendar seasonal adjustment to projected ROAS
	projectedRoas *= avgMult

	// Determine feasibility
	feasibility := s.determineFeasibility(req.TargetRoas, projectedRoas, probResult)

	// Calculate confidence
	confidence := s.calculateConfidence(probResult, trendResult, len(data.RoasHistory))

	// Determine trend prediction
	trend := s.determineTrendPrediction(projectedRoas, data.CurrentRoas, trendResult)

	// Calculate optimal budget + scale-up analysis
	optimalBudget, maxSafeBudget, optimalLabel := s.calculateOptimalBudget(
		req.TargetRoas, data, satResult,
	)
	scaleFactor := 0.0
	if data.CurrentSpend > 0 {
		scaleFactor = math.Round(maxSafeBudget/data.CurrentSpend*10) / 10
	}

	// Generate recommendation
	recommendation := s.generateRecommendation(
		req.TargetRoas, req.BudgetPerDay, projectedRoas,
		optimalBudget, maxSafeBudget, data.CurrentRoas, feasibility,
	)

	// Generate alternatives
	alternatives := s.generateAlternatives(req.TargetRoas, req.BudgetPerDay, data, satResult)

	// Period projections
	periodDays := float64(req.PeriodDays)
	totalBudget := req.BudgetPerDay * periodDays
	dailyRevenue := req.BudgetPerDay * projectedRoas
	totalRevenue := dailyRevenue * periodDays
	dailyOrders := 0.0
	if funnel != nil {
		dailyOrders = funnel.ProjectedOrders
	} else if data.CurrentSpend > 0 && data.TotalOrders > 0 {
		baseDaily := float64(data.TotalOrders) / float64(data.DaysOfData)
		scale := req.BudgetPerDay / data.CurrentSpend
		dailyOrders = baseDaily * scale * 0.85 // conservative scaling
	}
	totalOrders := dailyOrders * periodDays

	return SimulationResult{
		Feasibility:           feasibility,
		ConfidencePercent:     confidence,
		CurrentRoas:           math.Round(data.CurrentRoas*100) / 100,
		ProjectedRoas:         math.Round(projectedRoas*100) / 100,
		TrendPrediction:       trend,
		OptimalBudget:         math.Round(optimalBudget),
		Recommendation:        recommendation,
		Alternatives:          alternatives,
		SimulationMode:        simMode,
		MLCategory:            data.MLCategory,
		CurrentDailySpend:     math.Round(data.CurrentSpend),
		TotalBudget:           math.Round(totalBudget),
		TotalProjectedRevenue: math.Round(totalRevenue),
		TotalProjectedOrders:  math.Round(totalOrders*10) / 10,
		Funnel:                funnel,
		CalendarMultiplier:    avgMult,
		CalendarEvents:        calendarEvents,
		MaxSafeBudget:         math.Round(maxSafeBudget),
		ScaleFactor:           scaleFactor,
		OptimalLabel:          optimalLabel,
	}
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

// calculateOptimalBudget finds the daily budget for target ROAS.
// For over-performing products (current > target): finds max safe scale-up budget.
// For under-performing products: finds minimum budget to achieve target.
// Returns: (optimalBudget, maxSafeBudget, label)
func (s *BudgetSimulator) calculateOptimalBudget(
	targetRoas float64,
	data ProductHistoricalData,
	satResult SaturationResult,
) (float64, float64, string) {
	if data.CurrentRoas == 0 || data.CurrentSpend == 0 {
		return data.CurrentSpend, data.CurrentSpend, "Optimal Budget"
	}

	isOverPerforming := data.CurrentRoas >= targetRoas

	// Cap at 5x current spend for realism
	scaleCap := 5.0
	low := data.CurrentSpend * 0.05
	high := data.CurrentSpend * scaleCap

	// Binary search
	for i := 0; i < 50; i++ {
		mid := (low + high) / 2
		projectedRoas, _, _ := s.projectRoasWithBudget(
			data.CurrentRoas, data.CurrentSpend, mid,
			satResult, data,
		)

		if math.Abs(projectedRoas-targetRoas) < 0.01 {
			break
		}

		if projectedRoas > targetRoas {
			low = mid
		} else {
			high = mid
		}
	}

	optimal := (low + high) / 2
	if optimal < 10000 {
		optimal = 10000
	}

	if isOverPerforming {
		// Max safe budget = budget ceiling where ROAS still >= target
		maxSafe := optimal
		if maxSafe > data.CurrentSpend*scaleCap {
			maxSafe = data.CurrentSpend * scaleCap
		}
		return maxSafe, maxSafe, "Max Safe Scale-Up Budget"
	}

	// Under-performing: optimal is the budget needed to hit target
	return optimal, optimal, "Optimal Budget"
}

// generateRecommendation generates contextual Indonesian recommendation.
// 5 cases: strong over-performing, slight over-performing, achievable,
// difficult, and not achievable.
func (s *BudgetSimulator) generateRecommendation(
	targetRoas, budgetPerDay, projectedRoas, optimalBudget, maxSafeBudget,
	currentRoas float64, feasibility FeasibilityStatus,
) string {
	budgetStr := formatRupiah(budgetPerDay)
	projStr := formatFloat(projectedRoas)
	targetStr := formatFloat(targetRoas)
	currentStr := formatFloat(currentRoas)
	maxStr := formatRupiah(maxSafeBudget)

	// Case 1: Strongly over-performing (current ROAS > 1.5x target)
	if currentRoas > targetRoas*1.5 {
		return "Produk STAR! ROAS saat ini (" + currentStr + "x) jauh di atas target (" +
			targetStr + "x). Dengan budget " + budgetStr + "/hari, projected ROAS = " +
			projStr + "x. Scale-up aman hingga " + maxStr + "/hari, ROAS tetap >= " +
			targetStr + "x."
	}

	// Case 2: Slightly over-performing (current ROAS > target)
	if currentRoas > targetRoas {
		return "ROAS saat ini (" + currentStr + "x) sudah di atas target (" +
			targetStr + "x). Budget bisa ditingkatkan hingga " + maxStr +
			"/hari sebelum mendekati batas target. Projected ROAS: " + projStr + "x."
	}

	// Case 3-5: Under-performing
	switch feasibility {
	case FeasibilityAchievable:
		return "Target ROAS " + targetStr + "x achievable dengan budget " + budgetStr +
			"/hari. Projected ROAS: " + projStr + "x."

	case FeasibilityDifficult:
		gapPct := ((targetRoas - projectedRoas) / targetRoas) * 100
		if gapPct < 20 {
			return "Hampir mencapai target! Gap hanya " + formatFloat(gapPct) +
				"%. Tingkatkan budget ke " + formatRupiah(optimalBudget) +
				"/hari atau optimalkan creative untuk mencapai ROAS " + targetStr + "x."
		}
		return "Target ROAS " + targetStr + "x sulit dengan budget " + budgetStr +
			"/hari (projected: " + projStr + "x). Pertimbangkan menaikkan budget" +
			" atau menurunkan target."

	default:
		return "Target ROAS " + targetStr + "x tidak realistis dengan kondisi saat ini (" +
			currentStr + "x). Turunkan target ke " + projStr +
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
	projectedRoas, _, _ := s.projectRoasWithBudget(
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
	optimalBudget, _, _ := s.calculateOptimalBudget(targetRoas, data, satResult)
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
