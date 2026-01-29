package intelligence

import (
	"math"
)

// SaturationStatus represents budget saturation status
type SaturationStatus string

const (
	SaturationHighElasticity        SaturationStatus = "HIGH_ELASTICITY"
	SaturationModerate              SaturationStatus = "MODERATE"
	SaturationApproachingSaturation SaturationStatus = "APPROACHING_SATURATION"
	SaturationSaturated             SaturationStatus = "SATURATED"
	SaturationOverSaturated         SaturationStatus = "OVER_SATURATED"
	SaturationInsufficientData      SaturationStatus = "INSUFFICIENT_DATA"
)

// SaturationResult contains saturation analysis results
type SaturationResult struct {
	Status          SaturationStatus `json:"status"`
	MarginalRoas    float64          `json:"marginal_roas"`
	SaturationPoint float64          `json:"saturation_point"`
	CurrentSpend    float64          `json:"current_spend"`
	Headroom        float64          `json:"headroom"`
	Score           float64          `json:"score"`
	Description     string           `json:"description"`
}

// SaturationModel models diminishing returns for budget spending
type SaturationModel struct {
	MaxBudget int
}

// NewSaturationModel creates a new saturation model
func NewSaturationModel() *SaturationModel {
	return &SaturationModel{MaxBudget: 2000000}
}

// Analyze performs saturation analysis on spend/revenue history
func (s *SaturationModel) Analyze(spendHistory, revenueHistory []float64) SaturationResult {
	if len(spendHistory) < 5 || len(revenueHistory) < 5 {
		return s.insufficientData()
	}

	// Ensure same length
	minLen := min(len(spendHistory), len(revenueHistory))
	spend := spendHistory[:minLen]
	revenue := revenueHistory[:minLen]

	// Calculate current spend
	currentSpend := spend[len(spend)-1]

	// Calculate marginal ROAS
	marginalRoas := s.calculateMarginalRoas(spend, revenue)

	// Estimate saturation point using diminishing returns
	saturationPoint := s.estimateSaturationPoint(spend, revenue)

	// Calculate headroom
	headroom := float64(0)
	if saturationPoint > 0 && currentSpend > 0 {
		headroom = ((saturationPoint - currentSpend) / currentSpend) * 100
	}

	// Determine status
	status := s.determineStatus(marginalRoas, headroom)

	// Calculate score
	score := s.calculateScore(status)

	return SaturationResult{
		Status:          status,
		MarginalRoas:    math.Round(marginalRoas*100) / 100,
		SaturationPoint: math.Round(saturationPoint),
		CurrentSpend:    math.Round(currentSpend),
		Headroom:        math.Round(math.Max(0, headroom)*10) / 10,
		Score:           score,
		Description:     s.getDescription(status, marginalRoas, headroom),
	}
}

// calculateMarginalRoas calculates the marginal ROAS
func (s *SaturationModel) calculateMarginalRoas(spend, revenue []float64) float64 {
	if len(spend) < 2 {
		return 1.0
	}

	n := len(spend)
	deltaSpend := spend[n-1] - spend[n-2]
	deltaRevenue := revenue[n-1] - revenue[n-2]

	if deltaSpend == 0 {
		return 1.0
	}

	return deltaRevenue / deltaSpend
}

// estimateSaturationPoint estimates where diminishing returns plateau
func (s *SaturationModel) estimateSaturationPoint(spend, revenue []float64) float64 {
	if len(spend) < 3 {
		return float64(s.MaxBudget)
	}

	// Calculate average ROAS
	var totalSpend, totalRevenue float64
	for i := range spend {
		totalSpend += spend[i]
		totalRevenue += revenue[i]
	}

	avgRoas := float64(1)
	if totalSpend > 0 {
		avgRoas = totalRevenue / totalSpend
	}

	// Calculate marginal ROAS trend
	marginals := s.calculateMarginalTrend(spend, revenue)

	// If marginal ROAS is declining, estimate saturation
	if len(marginals) >= 2 {
		lastMarginal := marginals[len(marginals)-1]
		avgMarginal := float64(0)
		for _, m := range marginals {
			avgMarginal += m
		}
		avgMarginal /= float64(len(marginals))

		// If last marginal is below 1.0, we're past saturation
		if lastMarginal < 1.0 {
			return spend[len(spend)-1] * 0.8 // Current spend is over
		}

		// Estimate based on rate of decline
		if avgMarginal > lastMarginal && lastMarginal > 0 {
			declineRate := (avgMarginal - lastMarginal) / avgMarginal
			currentSpend := spend[len(spend)-1]
			// Estimate where marginal ROAS hits 1.0
			if declineRate > 0 {
				return currentSpend * (1 + (lastMarginal-1.0)/(declineRate*avgRoas))
			}
		}
	}

	// Default: estimate based on current spend and ROAS
	currentSpend := spend[len(spend)-1]
	return currentSpend * 1.5 // 50% headroom by default
}

// calculateMarginalTrend calculates marginal returns over time
func (s *SaturationModel) calculateMarginalTrend(spend, revenue []float64) []float64 {
	if len(spend) < 2 {
		return nil
	}

	var marginals []float64
	for i := 1; i < len(spend); i++ {
		deltaSpend := spend[i] - spend[i-1]
		deltaRevenue := revenue[i] - revenue[i-1]

		if deltaSpend != 0 {
			marginals = append(marginals, deltaRevenue/deltaSpend)
		}
	}

	return marginals
}

// determineStatus determines saturation status
func (s *SaturationModel) determineStatus(marginalRoas, headroom float64) SaturationStatus {
	if headroom <= 0 {
		return SaturationOverSaturated
	}
	if headroom < 10 {
		return SaturationSaturated
	}
	if headroom < 30 {
		return SaturationApproachingSaturation
	}
	if marginalRoas >= 2 {
		return SaturationHighElasticity
	}
	return SaturationModerate
}

// calculateScore calculates saturation score (0-100, higher = more room to grow)
func (s *SaturationModel) calculateScore(status SaturationStatus) float64 {
	scores := map[SaturationStatus]float64{
		SaturationHighElasticity:        90,
		SaturationModerate:              70,
		SaturationApproachingSaturation: 50,
		SaturationSaturated:             30,
		SaturationOverSaturated:         10,
		SaturationInsufficientData:      50,
	}
	return scores[status]
}

// getDescription returns Indonesian description
func (s *SaturationModel) getDescription(status SaturationStatus, marginalRoas, headroom float64) string {
	switch status {
	case SaturationHighElasticity:
		return "High elasticity! Budget dapat ditambah " + formatPercent(headroom)
	case SaturationModerate:
		return "Moderate response. Headroom " + formatPercent(headroom)
	case SaturationApproachingSaturation:
		return "Mendekati saturasi. Hanya " + formatPercent(headroom) + " headroom tersisa"
	case SaturationSaturated:
		return "Saturated! Marginal ROAS hanya " + formatFloat(marginalRoas) + "x"
	case SaturationOverSaturated:
		return "Over-saturated! Kurangi budget"
	default:
		return "Data tidak cukup untuk analisis saturasi"
	}
}

func (s *SaturationModel) insufficientData() SaturationResult {
	return SaturationResult{
		Status:      SaturationInsufficientData,
		Score:       50,
		Description: "Data tidak cukup untuk analisis saturasi",
	}
}

func formatPercent(v float64) string {
	return formatFloat(v) + "%"
}

func formatFloat(v float64) string {
	// Simple float formatting
	intPart := int(v)
	decPart := int((v - float64(intPart)) * 10)
	if decPart < 0 {
		decPart = -decPart
	}
	result := ""
	if intPart < 0 {
		result = "-"
		intPart = -intPart
	}
	return result + intToString(intPart) + "." + string(rune(decPart+48))
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune(n%10+48)) + result
		n /= 10
	}
	return result
}
