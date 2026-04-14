package intelligence

import "math"

// projectRoasWithBudget projects ROAS using dual-mode power-law model.
// Mode A (Funnel): uses CPC/CVR/CTR when click+order data is available.
// Mode B (Yield): uses raw power-law on ROAS when funnel data is incomplete.
// Returns projected ROAS, simulation mode, and optional funnel breakdown.
func (s *BudgetSimulator) projectRoasWithBudget(
	currentRoas, currentSpend, newBudget float64,
	satResult SaturationResult, data ProductHistoricalData,
) (float64, SimulationMode, *FunnelBreakdown) {
	if currentSpend <= 0 {
		return currentRoas, SimulationModeYield, nil
	}

	scale := newBudget / currentSpend
	if scale <= 0 {
		return currentRoas, SimulationModeYield, nil
	}

	// Edge case: identical budget
	if math.Abs(scale-1.0) < 0.001 {
		return currentRoas, SimulationModeYield, nil
	}

	// Determine elasticity from ML context
	elasticity := getElasticityForCategory(data.MLCategory, data.FatigueStatus)
	elasticity *= getSaturationPenalty(satResult)

	// Dual-mode selection
	if data.TotalClicks > 0 && data.TotalOrders > 0 && data.TotalImpressions > 0 {
		roas, fb := s.projectFunnel(currentRoas, currentSpend, scale, elasticity, data)
		return roas, SimulationModeFunnel, fb
	}
	roas := s.projectYield(currentRoas, scale, elasticity)
	return roas, SimulationModeYield, nil
}

// projectFunnel simulates the full advertising funnel:
// Budget → Clicks (CPC) → Orders (CVR) → Revenue (AOV)
// Each stage degrades with a power-law penalty as budget scales up.
// Returns projected ROAS and a full FunnelBreakdown.
func (s *BudgetSimulator) projectFunnel(
	currentRoas, currentSpend, scale, elasticity float64,
	data ProductHistoricalData,
) (float64, *FunnelBreakdown) {
	totalSpend := currentSpend * float64(data.DaysOfData)

	baseCPC := totalSpend / float64(data.TotalClicks)
	baseCVR := float64(data.TotalOrders) / float64(data.TotalClicks)
	baseAOV := (totalSpend * currentRoas) / float64(data.TotalOrders)
	baseCTR := float64(data.TotalClicks) / float64(data.TotalImpressions)

	if baseCPC <= 0 || baseCVR <= 0 || baseAOV <= 0 {
		roas := s.projectYield(currentRoas, scale, elasticity)
		return roas, nil
	}

	// Apply scale penalties (power-law)
	adjCPC := baseCPC * math.Pow(scale, 0.08)
	adjCVR := baseCVR * math.Pow(scale, elasticity*1.2)
	adjCTR := baseCTR * math.Pow(scale, -0.03)

	// Guards
	if adjCVR <= 0 {
		adjCVR = baseCVR * 0.01
	}
	if adjCVR > 1.0 {
		adjCVR = baseCVR
	}
	if adjCTR <= 0 {
		adjCTR = baseCTR * 0.1
	}

	// Project daily metrics
	newBudget := currentSpend * scale
	dailyClicks := newBudget / adjCPC
	dailyOrders := dailyClicks * adjCVR
	dailyRevenue := dailyOrders * baseAOV

	// Derive impressions from projected CTR
	dailyImpressions := 0.0
	if adjCTR > 0 {
		dailyImpressions = dailyClicks / adjCTR
	}

	cpo := 0.0
	if dailyOrders > 0 {
		cpo = newBudget / dailyOrders
	}

	fb := &FunnelBreakdown{
		ProjectedClicks:      math.Round(dailyClicks),
		ProjectedOrders:      math.Round(dailyOrders*10) / 10,
		ProjectedRevenue:     math.Round(dailyRevenue),
		ProjectedCPC:         math.Round(adjCPC),
		ProjectedCPO:         math.Round(cpo),
		ProjectedCVR:         math.Round(adjCVR*10000) / 100,
		ProjectedAOV:         math.Round(baseAOV),
		ProjectedImpressions: math.Round(dailyImpressions),
		ProjectedCTR:         math.Round(adjCTR*10000) / 100,
	}

	if newBudget > 0 {
		return dailyRevenue / newBudget, fb
	}
	return currentRoas, fb
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

// getElasticityForCategory returns the power-law elasticity exponent
// based on ML product category and creative fatigue status.
func getElasticityForCategory(category, fatigueStatus string) float64 {
	base := -0.12
	switch category {
	case "STAR":
		base = -0.06
	case "GROWTH":
		base = -0.10
	case "STABLE":
		base = -0.15
	case "WATCH":
		base = -0.22
	case "PROBLEM":
		base = -0.35
	}
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
		return 1.0
	case SaturationModerate:
		return 1.1
	case SaturationApproachingSaturation:
		return 1.3
	case SaturationSaturated:
		return 1.6
	case SaturationOverSaturated:
		return 2.0
	default:
		return 1.0
	}
}
