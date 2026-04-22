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

	trendResult := s.trend.Analyze(data.RoasHistory)
	satResult := s.saturation.Analyze(data.SpendHistory, data.RevenueHistory)
	probResult := s.probability.Analyze(data.RoasHistory, req.TargetRoas)

	projectedRoas, simMode, funnel := s.projectRoasWithBudget(
		data.CurrentRoas, data.CurrentSpend, req.BudgetPerDay, satResult, data,
	)

	calendarCtx := s.GetCalendarContext(time.Now(), req.PeriodDays)
	avgMult := calendarCtx["average_multiplier"].(float64)
	calendarEvents, _ := calendarCtx["upcoming_events"].([]map[string]interface{})
	if calendarEvents == nil {
		calendarEvents = []map[string]interface{}{}
	}

	projectedRoas *= avgMult
	feasibility := s.determineFeasibility(req.TargetRoas, projectedRoas, probResult)
	confidence := s.calculateConfidence(probResult, trendResult, len(data.RoasHistory))
	trend := s.determineTrendPrediction(projectedRoas, data.CurrentRoas, trendResult)

	optimalBudget, maxSafeBudget, optimalLabel := s.calculateOptimalBudget(req.TargetRoas, data, satResult)
	scaleFactor := 0.0
	if data.CurrentSpend > 0 {
		scaleFactor = math.Round(maxSafeBudget/data.CurrentSpend*10) / 10
	}

	recommendation := s.generateRecommendation(
		req.TargetRoas, req.BudgetPerDay, projectedRoas, optimalBudget, maxSafeBudget, data.CurrentRoas, feasibility,
	)
	alternatives := s.generateAlternatives(req.TargetRoas, req.BudgetPerDay, data, satResult)

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
		dailyOrders = baseDaily * scale * 0.85
	}

	return SimulationResult{
		Feasibility: feasibility, ConfidencePercent: confidence,
		CurrentRoas: math.Round(data.CurrentRoas*100) / 100,
		ProjectedRoas: math.Round(projectedRoas*100) / 100,
		TrendPrediction: trend, OptimalBudget: math.Round(optimalBudget),
		Recommendation: recommendation, Alternatives: alternatives,
		SimulationMode: simMode, MLCategory: data.MLCategory,
		CurrentDailySpend: math.Round(data.CurrentSpend),
		TotalBudget: math.Round(totalBudget),
		TotalProjectedRevenue: math.Round(totalRevenue),
		TotalProjectedOrders: math.Round(dailyOrders*periodDays*10) / 10,
		Funnel: funnel, CalendarMultiplier: avgMult, CalendarEvents: calendarEvents,
		MaxSafeBudget: math.Round(maxSafeBudget), ScaleFactor: scaleFactor,
		OptimalLabel: optimalLabel,
	}
}

// determineFeasibility determines if target ROAS is achievable
func (s *BudgetSimulator) determineFeasibility(
	targetRoas, projectedRoas float64, probResult ProbabilityResult,
) FeasibilityStatus {
	if projectedRoas >= targetRoas {
		if probResult.SuccessProbability >= 50 {
			return FeasibilityAchievable
		}
		return FeasibilityDifficult
	}

	roasGapPct := ((targetRoas - projectedRoas) / targetRoas) * 100
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
	probResult ProbabilityResult, trendResult TrendResult, dataPoints int,
) float64 {
	var dataConfidence float64
	switch {
	case dataPoints >= 30:
		dataConfidence = 90
	case dataPoints >= 14:
		dataConfidence = 75
	case dataPoints >= 7:
		dataConfidence = 60
	default:
		dataConfidence = 40
	}
	return math.Min(95, math.Max(30, dataConfidence+(trendResult.Score-50)/5))
}

// determineTrendPrediction determines trend prediction
func (s *BudgetSimulator) determineTrendPrediction(
	projectedRoas, currentRoas float64, trendResult TrendResult,
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
func (s *BudgetSimulator) calculateOptimalBudget(
	targetRoas float64, data ProductHistoricalData, satResult SaturationResult,
) (float64, float64, string) {
	if data.CurrentRoas == 0 || data.CurrentSpend == 0 {
		return data.CurrentSpend, data.CurrentSpend, "Optimal Budget"
	}

	isOverPerforming := data.CurrentRoas >= targetRoas
	scaleCap := 5.0
	low := data.CurrentSpend * 0.05
	high := data.CurrentSpend * scaleCap

	for i := 0; i < 50; i++ {
		mid := (low + high) / 2
		projectedRoas, _, _ := s.projectRoasWithBudget(data.CurrentRoas, data.CurrentSpend, mid, satResult, data)
		if math.Abs(projectedRoas-targetRoas) < 0.01 {
			break
		}
		if projectedRoas > targetRoas {
			low = mid
		} else {
			high = mid
		}
	}

	optimal := math.Max((low+high)/2, 10000)

	if isOverPerforming {
		maxSafe := math.Min(optimal, data.CurrentSpend*scaleCap)
		return maxSafe, maxSafe, "Max Safe Scale-Up Budget"
	}
	return optimal, optimal, "Optimal Budget"
}

// GetCalendarContext gets calendar context for a date range
func (s *BudgetSimulator) GetCalendarContext(startDate time.Time, days int) map[string]interface{} {
	events := s.calendar.GetUpcomingEvents(startDate, days)

	var totalMult float64
	for i := 0; i < days; i++ {
		totalMult += s.calendar.GetMultiplier(startDate.AddDate(0, 0, i))
	}
	avgMult := totalMult / float64(days)

	return map[string]interface{}{
		"upcoming_events":    events,
		"average_multiplier": math.Round(avgMult*100) / 100,
		"is_favorable":       avgMult >= 1.0,
	}
}
