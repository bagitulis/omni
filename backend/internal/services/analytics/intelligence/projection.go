package intelligence

import (
	"math"
	"math/rand"
	"sort"
	"time"
)

// ForecastResult contains forecasting results
type ForecastResult struct {
	ForecastDays    int               `json:"forecast_days"`
	Predictions     []DailyPrediction `json:"predictions"`
	SummaryStats    ForecastSummary   `json:"summary_stats"`
	ConfidenceLevel ConfidenceLevel   `json:"confidence_level"`
}

// DailyPrediction contains single day prediction
type DailyPrediction struct {
	Date          string  `json:"date"`
	ExpectedValue float64 `json:"expected_value"`
	LowerBound    float64 `json:"lower_bound"`
	UpperBound    float64 `json:"upper_bound"`
	Trend         string  `json:"trend"`
}

// ForecastSummary contains summary statistics
type ForecastSummary struct {
	TotalExpectedRevenue float64 `json:"total_expected_revenue"`
	TotalExpectedSpend   float64 `json:"total_expected_spend"`
	ExpectedRoas         float64 `json:"expected_roas"`
	BestCaseRoas         float64 `json:"best_case_roas"`
	WorstCaseRoas        float64 `json:"worst_case_roas"`
}

// ProjectionService handles forecasting and projections
type ProjectionService struct {
	NumSimulations int
	calendar       *IndonesianCalendar
}

// NewProjectionService creates a new projection service
func NewProjectionService() *ProjectionService {
	return &ProjectionService{
		NumSimulations: 1000,
		calendar:       NewIndonesianCalendar(),
	}
}

// ForecastRevenue forecasts revenue for N days
func (p *ProjectionService) ForecastRevenue(
	historicalRevenue []float64,
	historicalSpend []float64,
	forecastDays int,
	startDate time.Time,
) ForecastResult {
	if len(historicalRevenue) < 5 {
		return p.insufficientData(forecastDays)
	}

	// Calculate historical stats
	revMean := mean(historicalRevenue)
	revStd := stdDev(historicalRevenue)
	spendMean := mean(historicalSpend)
	_ = stdDev(historicalSpend) // For future use in spend variance

	// Generate daily predictions
	predictions := make([]DailyPrediction, forecastDays)
	totalExpectedRev := float64(0)
	totalExpectedSpend := float64(0)

	for i := 0; i < forecastDays; i++ {
		date := startDate.AddDate(0, 0, i)

		// Get calendar multiplier
		multiplier := p.calendar.GetMultiplier(date)

		// Base prediction with calendar adjustment
		expectedRev := revMean * multiplier
		expectedSpend := spendMean

		// Monte Carlo for confidence interval
		ciLower, ciUpper := p.monteCarloCI(revMean, revStd, multiplier)

		// Determine trend based on calendar
		trend := "STABLE"
		if multiplier > 1.1 {
			trend = "UP"
		} else if multiplier < 0.9 {
			trend = "DOWN"
		}

		predictions[i] = DailyPrediction{
			Date:          date.Format("2006-01-02"),
			ExpectedValue: math.Round(expectedRev*100) / 100,
			LowerBound:    math.Round(ciLower*100) / 100,
			UpperBound:    math.Round(ciUpper*100) / 100,
			Trend:         trend,
		}

		totalExpectedRev += expectedRev
		totalExpectedSpend += expectedSpend
	}

	// Calculate summary stats
	expectedRoas := float64(0)
	bestCaseRoas := float64(0)
	worstCaseRoas := float64(0)

	if totalExpectedSpend > 0 {
		expectedRoas = totalExpectedRev / totalExpectedSpend

		// Best/worst case based on confidence intervals
		totalUpperBound := float64(0)
		totalLowerBound := float64(0)
		for _, pred := range predictions {
			totalUpperBound += pred.UpperBound
			totalLowerBound += pred.LowerBound
		}
		bestCaseRoas = totalUpperBound / totalExpectedSpend
		worstCaseRoas = totalLowerBound / totalExpectedSpend
	}

	// Determine confidence level
	confidenceLevel := p.getConfidenceLevel(len(historicalRevenue))

	return ForecastResult{
		ForecastDays: forecastDays,
		Predictions:  predictions,
		SummaryStats: ForecastSummary{
			TotalExpectedRevenue: math.Round(totalExpectedRev),
			TotalExpectedSpend:   math.Round(totalExpectedSpend),
			ExpectedRoas:         math.Round(expectedRoas*100) / 100,
			BestCaseRoas:         math.Round(bestCaseRoas*100) / 100,
			WorstCaseRoas:        math.Round(worstCaseRoas*100) / 100,
		},
		ConfidenceLevel: confidenceLevel,
	}
}

// monteCarloCI calculates confidence interval using Monte Carlo
func (p *ProjectionService) monteCarloCI(mean, std, multiplier float64) (lower, upper float64) {
	if std == 0 {
		return mean * multiplier * 0.9, mean * multiplier * 1.1
	}

	samples := make([]float64, p.NumSimulations)
	for i := 0; i < p.NumSimulations; i++ {
		// Normal distribution sampling
		sample := mean + rand.NormFloat64()*std
		samples[i] = sample * multiplier
	}

	sort.Float64s(samples)

	lowerIdx := int(0.025 * float64(p.NumSimulations))
	upperIdx := int(0.975 * float64(p.NumSimulations))

	return samples[lowerIdx], samples[upperIdx]
}

// getConfidenceLevel determines confidence based on data points
func (p *ProjectionService) getConfidenceLevel(dataPoints int) ConfidenceLevel {
	if dataPoints >= 30 {
		return ConfidenceHigh
	}
	if dataPoints >= 14 {
		return ConfidenceMedium
	}
	return ConfidenceLow
}

func (p *ProjectionService) insufficientData(days int) ForecastResult {
	return ForecastResult{
		ForecastDays:    days,
		Predictions:     []DailyPrediction{},
		ConfidenceLevel: ConfidenceLow,
	}
}

// TrendForecast forecasts trend direction
type TrendForecast struct {
	Direction      string  `json:"direction"`
	Probability    float64 `json:"probability"`
	ExpectedChange float64 `json:"expected_change"`
}

// ForecastTrend predicts trend for next period
func (p *ProjectionService) ForecastTrend(values []float64) TrendForecast {
	if len(values) < 5 {
		return TrendForecast{Direction: "STABLE", Probability: 50}
	}

	// Calculate recent momentum
	n := len(values)
	recentAvg := mean(values[n-3:])
	olderAvg := mean(values[:n-3])

	change := float64(0)
	if olderAvg > 0 {
		change = ((recentAvg - olderAvg) / olderAvg) * 100
	}

	// Determine direction and probability
	var direction string
	var probability float64

	if change > 10 {
		direction = "STRONG_UP"
		probability = 80
	} else if change > 5 {
		direction = "UP"
		probability = 70
	} else if change > -5 {
		direction = "STABLE"
		probability = 60
	} else if change > -10 {
		direction = "DOWN"
		probability = 70
	} else {
		direction = "STRONG_DOWN"
		probability = 80
	}

	return TrendForecast{
		Direction:      direction,
		Probability:    probability,
		ExpectedChange: math.Round(change*10) / 10,
	}
}
