package intelligence

import "math"

// generateRecommendation generates contextual Indonesian recommendation.
func (s *BudgetSimulator) generateRecommendation(
	targetRoas, budgetPerDay, projectedRoas, optimalBudget, maxSafeBudget,
	currentRoas float64, feasibility FeasibilityStatus,
) string {
	budgetStr := formatRupiah(budgetPerDay)
	projStr := formatFloat(projectedRoas)
	targetStr := formatFloat(targetRoas)
	currentStr := formatFloat(currentRoas)
	maxStr := formatRupiah(maxSafeBudget)

	if currentRoas > targetRoas*1.5 {
		return "Produk STAR! ROAS saat ini (" + currentStr + "x) jauh di atas target (" +
			targetStr + "x). Dengan budget " + budgetStr + "/hari, projected ROAS = " +
			projStr + "x. Scale-up aman hingga " + maxStr + "/hari, ROAS tetap >= " +
			targetStr + "x."
	}

	if currentRoas > targetRoas {
		return "ROAS saat ini (" + currentStr + "x) sudah di atas target (" +
			targetStr + "x). Budget bisa ditingkatkan hingga " + maxStr +
			"/hari sebelum mendekati batas target. Projected ROAS: " + projStr + "x."
	}

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

	projectedRoas, _, _ := s.projectRoasWithBudget(
		data.CurrentRoas, data.CurrentSpend, budgetPerDay, satResult, data,
	)
	if projectedRoas < targetRoas {
		alternatives = append(alternatives, BudgetAlternative{
			TargetRoas:     math.Round(projectedRoas*10) / 10,
			RequiredBudget: budgetPerDay,
			ExpectedRoas:   math.Round(projectedRoas*100) / 100,
		})
	}

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

func formatRupiah(amount float64) string {
	if amount >= 1000000 {
		return "Rp " + formatFloat(amount/1000000) + " jt"
	}
	if amount >= 1000 {
		return "Rp " + intToString(int(amount/1000)) + " rb"
	}
	return "Rp " + intToString(int(amount))
}
