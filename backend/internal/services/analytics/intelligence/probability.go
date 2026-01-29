package intelligence

import (
	"math"
	"math/rand"
	"sort"
)

// ConfidenceLevel represents confidence level
type ConfidenceLevel string

const (
	ConfidenceHigh   ConfidenceLevel = "HIGH"
	ConfidenceMedium ConfidenceLevel = "MEDIUM"
	ConfidenceLow    ConfidenceLevel = "LOW"
)

// ProbabilityResult contains probability analysis results
type ProbabilityResult struct {
	SuccessProbability float64         `json:"success_probability"`
	FailureProbability float64         `json:"failure_probability"`
	ConfidenceLevel    ConfidenceLevel `json:"confidence_level"`
	CILower            float64         `json:"ci_lower"`
	CIUpper            float64         `json:"ci_upper"`
	ExpectedRoas       float64         `json:"expected_roas"`
	Description        string          `json:"description"`
}

// ProbabilityEngine calculates success probability
type ProbabilityEngine struct {
	MinDataPoints int
}

// NewProbabilityEngine creates a new probability engine
func NewProbabilityEngine() *ProbabilityEngine {
	return &ProbabilityEngine{MinDataPoints: 3}
}

// Analyze calculates success probability based on ROAS history
func (p *ProbabilityEngine) Analyze(roasHistory []float64, threshold float64) ProbabilityResult {
	if len(roasHistory) < p.MinDataPoints {
		return p.insufficientData()
	}

	// Default threshold
	if threshold == 0 {
		threshold = 2.0
	}

	// Calculate success probability (P(ROAS > threshold))
	successProb := p.calculateSuccessProbability(roasHistory, threshold)

	// Calculate failure probability (P(ROAS < 1.0))
	failureProb := 1 - p.calculateSuccessProbability(roasHistory, 1.0)

	// Calculate confidence interval using bootstrap
	ciLower, ciUpper := p.bootstrapConfidenceInterval(roasHistory, 0.95)

	// Calculate expected ROAS
	expectedRoas := mean(roasHistory)

	// Determine confidence level based on sample size
	confidenceLevel := p.getConfidenceLevel(len(roasHistory))

	return ProbabilityResult{
		SuccessProbability: math.Round(successProb*1000) / 10,
		FailureProbability: math.Round(failureProb*1000) / 10,
		ConfidenceLevel:    confidenceLevel,
		CILower:            math.Round(ciLower*100) / 100,
		CIUpper:            math.Round(ciUpper*100) / 100,
		ExpectedRoas:       math.Round(expectedRoas*100) / 100,
		Description:        p.getDescription(successProb, confidenceLevel),
	}
}

// calculateSuccessProbability calculates P(ROAS > threshold)
func (p *ProbabilityEngine) calculateSuccessProbability(values []float64, threshold float64) float64 {
	if len(values) == 0 {
		return 0.5
	}

	// Count values above threshold
	successCount := 0
	for _, v := range values {
		if v >= threshold {
			successCount++
		}
	}

	// Simple empirical probability with Laplace smoothing
	return float64(successCount+1) / float64(len(values)+2)
}

// bootstrapConfidenceInterval calculates CI using bootstrap method
func (p *ProbabilityEngine) bootstrapConfidenceInterval(values []float64, confidence float64) (lower, upper float64) {
	if len(values) < 3 {
		m := mean(values)
		return m * 0.8, m * 1.2
	}

	nBootstrap := 1000
	bootstrapMeans := make([]float64, nBootstrap)

	for i := 0; i < nBootstrap; i++ {
		// Sample with replacement
		sample := make([]float64, len(values))
		for j := range sample {
			sample[j] = values[rand.Intn(len(values))]
		}
		bootstrapMeans[i] = mean(sample)
	}

	// Sort for percentile calculation
	sort.Float64s(bootstrapMeans)

	alpha := (1 - confidence) / 2
	lowerIdx := int(alpha * float64(nBootstrap))
	upperIdx := int((1 - alpha) * float64(nBootstrap))

	if lowerIdx < 0 {
		lowerIdx = 0
	}
	if upperIdx >= nBootstrap {
		upperIdx = nBootstrap - 1
	}

	return bootstrapMeans[lowerIdx], bootstrapMeans[upperIdx]
}

// getConfidenceLevel determines confidence based on sample size
func (p *ProbabilityEngine) getConfidenceLevel(n int) ConfidenceLevel {
	if n >= 14 {
		return ConfidenceHigh
	}
	if n >= 7 {
		return ConfidenceMedium
	}
	return ConfidenceLow
}

// getDescription returns Indonesian description
func (p *ProbabilityEngine) getDescription(successProb float64, confidence ConfidenceLevel) string {
	probPct := successProb * 100
	confStr := string(confidence)

	if successProb >= 0.8 {
		return "Sangat likely sukses (" + formatFloat(probPct) + "%) - " + confStr + " confidence"
	}
	if successProb >= 0.6 {
		return "Likely sukses (" + formatFloat(probPct) + "%) - " + confStr + " confidence"
	}
	if successProb >= 0.4 {
		return "Moderate chance (" + formatFloat(probPct) + "%) - " + confStr + " confidence"
	}
	return "Likely gagal (" + formatFloat(probPct) + "%) - " + confStr + " confidence"
}

func (p *ProbabilityEngine) insufficientData() ProbabilityResult {
	return ProbabilityResult{
		SuccessProbability: 50,
		FailureProbability: 50,
		ConfidenceLevel:    ConfidenceLow,
		Description:        "Data tidak cukup untuk analisis probabilitas",
	}
}

// MonteCarloProjection projects future values using Monte Carlo
type MonteCarloProjection struct {
	Simulations int
}

// NewMonteCarloProjection creates a new projection engine
func NewMonteCarloProjection() *MonteCarloProjection {
	return &MonteCarloProjection{Simulations: 1000}
}

// ProjectionResult contains projection results
type ProjectionResult struct {
	ExpectedValue float64   `json:"expected_value"`
	CILower       float64   `json:"ci_lower"`
	CIUpper       float64   `json:"ci_upper"`
	Scenarios     []float64 `json:"scenarios"`
}

// Project projects future values
func (m *MonteCarloProjection) Project(historicalValues []float64, nDays int) ProjectionResult {
	if len(historicalValues) < 3 {
		return ProjectionResult{}
	}

	// Calculate mean and std from historical data
	mu := mean(historicalValues)
	sigma := stdDev(historicalValues)

	// Run simulations
	scenarios := make([]float64, m.Simulations)
	for i := 0; i < m.Simulations; i++ {
		// Simple random walk projection
		value := mu
		for d := 0; d < nDays; d++ {
			value += rand.NormFloat64() * sigma / math.Sqrt(float64(nDays))
		}
		scenarios[i] = value
	}

	// Sort for percentiles
	sort.Float64s(scenarios)

	return ProjectionResult{
		ExpectedValue: mean(scenarios),
		CILower:       scenarios[int(0.025*float64(m.Simulations))],
		CIUpper:       scenarios[int(0.975*float64(m.Simulations))],
		Scenarios:     scenarios,
	}
}

// Helper functions
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stdDev(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	m := mean(values)
	var sumSq float64
	for _, v := range values {
		sumSq += (v - m) * (v - m)
	}
	return math.Sqrt(sumSq / float64(len(values)-1))
}
