package intelligence

import (
	"math"
)

// FatigueStatus represents creative fatigue status
type FatigueStatus string

const (
	FatigueStatusFresh            FatigueStatus = "FRESH"    // CTR decay <10%
	FatigueStatusAging            FatigueStatus = "AGING"    // CTR decay 10-25%
	FatigueStatusFatigued         FatigueStatus = "FATIGUED" // CTR decay 25-50%
	FatigueStatusDead             FatigueStatus = "DEAD"     // CTR decay >50%
	FatigueStatusInsufficientData FatigueStatus = "INSUFFICIENT_DATA"
)

// FatigueResult contains fatigue analysis results
type FatigueResult struct {
	Status        FatigueStatus `json:"status"`
	CTRDecay      float64       `json:"ctr_decay"`
	PeakCTR       float64       `json:"peak_ctr"`
	CurrentCTR    float64       `json:"current_ctr"`
	DaysSincePeak int           `json:"days_since_peak"`
	Score         float64       `json:"score"`
	Description   string        `json:"description"`
	Action        string        `json:"action"`
}

// FatigueDetector detects creative fatigue by analyzing CTR decay
type FatigueDetector struct {
	MinDataPoints int
}

// NewFatigueDetector creates a new fatigue detector
func NewFatigueDetector() *FatigueDetector {
	return &FatigueDetector{MinDataPoints: 3}
}

// Analyze performs fatigue analysis on CTR history
func (f *FatigueDetector) Analyze(ctrHistory []float64) FatigueResult {
	if len(ctrHistory) < f.MinDataPoints {
		return f.insufficientData()
	}

	// Find peak CTR
	peakCTR := float64(0)
	peakIndex := 0
	for i, ctr := range ctrHistory {
		if ctr > peakCTR {
			peakCTR = ctr
			peakIndex = i
		}
	}

	// Get current CTR (most recent)
	currentCTR := ctrHistory[len(ctrHistory)-1]

	// Calculate CTR decay
	ctrDecay := float64(0)
	if peakCTR > 0 {
		ctrDecay = ((peakCTR - currentCTR) / peakCTR) * 100
	}

	// Days since peak
	daysSincePeak := len(ctrHistory) - peakIndex - 1

	// Calculate fatigue index (0 to 1)
	fatigueIndex := f.calculateFatigueIndex(ctrHistory, ctrDecay)

	// Determine status
	status := f.mapStatus(fatigueIndex)

	// Calculate score (100 = fresh, 0 = dead)
	score := (1 - fatigueIndex) * 100

	// Determine action
	action := f.getAction(status)

	return FatigueResult{
		Status:        status,
		CTRDecay:      math.Round(ctrDecay*10) / 10,
		PeakCTR:       math.Round(peakCTR*100) / 100,
		CurrentCTR:    math.Round(currentCTR*100) / 100,
		DaysSincePeak: daysSincePeak,
		Score:         math.Round(score*10) / 10,
		Description:   f.getDescription(status, ctrDecay),
		Action:        action,
	}
}

// calculateFatigueIndex calculates fatigue index (0-1)
func (f *FatigueDetector) calculateFatigueIndex(ctrHistory []float64, ctrDecay float64) float64 {
	if len(ctrHistory) < 3 {
		return 0
	}

	// Factor 1: CTR decay percentage (0-50% maps to 0-1)
	decayFactor := math.Min(1, ctrDecay/50)

	// Factor 2: Trend direction (declining = higher fatigue)
	n := len(ctrHistory)
	firstHalf := mean(ctrHistory[:n/2])
	secondHalf := mean(ctrHistory[n/2:])
	trendFactor := float64(0)
	if firstHalf > 0 && secondHalf < firstHalf {
		trendFactor = (firstHalf - secondHalf) / firstHalf
	}

	// Factor 3: Duration (longer = more fatigue risk)
	durationFactor := math.Min(1, float64(n)/30) * 0.3

	// Combined fatigue index
	fatigueIndex := (decayFactor*0.5 + trendFactor*0.3 + durationFactor*0.2)
	return math.Min(1, math.Max(0, fatigueIndex))
}

// mapStatus maps fatigue index to status
func (f *FatigueDetector) mapStatus(fatigueIndex float64) FatigueStatus {
	if fatigueIndex < 0.1 {
		return FatigueStatusFresh
	}
	if fatigueIndex < 0.25 {
		return FatigueStatusAging
	}
	if fatigueIndex < 0.5 {
		return FatigueStatusFatigued
	}
	return FatigueStatusDead
}

// getDescription returns Indonesian description
func (f *FatigueDetector) getDescription(status FatigueStatus, decay float64) string {
	decayStr := formatFloat(decay)
	switch status {
	case FatigueStatusFresh:
		return "Creative masih fresh (decay " + decayStr + "%)"
	case FatigueStatusAging:
		return "Creative mulai aging (decay " + decayStr + "%)"
	case FatigueStatusFatigued:
		return "Creative fatigued (decay " + decayStr + "%)"
	case FatigueStatusDead:
		return "Creative mati (decay " + decayStr + "%)"
	default:
		return "Data tidak cukup"
	}
}

// getAction returns recommended action
func (f *FatigueDetector) getAction(status FatigueStatus) string {
	switch status {
	case FatigueStatusFresh:
		return "maintain"
	case FatigueStatusAging:
		return "monitor_closely"
	case FatigueStatusFatigued:
		return "prepare_new_creative"
	case FatigueStatusDead:
		return "replace_immediately"
	default:
		return "gather_more_data"
	}
}

func (f *FatigueDetector) insufficientData() FatigueResult {
	return FatigueResult{
		Status:      FatigueStatusInsufficientData,
		Score:       50,
		Description: "Data tidak cukup untuk analisis fatigue",
		Action:      "gather_more_data",
	}
}
