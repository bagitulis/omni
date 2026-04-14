package intelligence

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

// FunnelBreakdown contains the projected funnel metrics per day
type FunnelBreakdown struct {
	ProjectedClicks      float64 `json:"projected_clicks"`
	ProjectedOrders      float64 `json:"projected_orders"`
	ProjectedRevenue     float64 `json:"projected_revenue"`
	ProjectedCPC         float64 `json:"projected_cpc"`
	ProjectedCPO         float64 `json:"projected_cpo"`
	ProjectedCVR         float64 `json:"projected_cvr"`
	ProjectedAOV         float64 `json:"projected_aov"`
	ProjectedImpressions float64 `json:"projected_impressions"`
	ProjectedCTR         float64 `json:"projected_ctr"`
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
	// Core metrics
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

	// Period projection (NEW)
	TotalBudget          float64 `json:"total_budget"`
	TotalProjectedRevenue float64 `json:"total_projected_revenue"`
	TotalProjectedOrders float64 `json:"total_projected_orders"`

	// Funnel breakdown per day (NEW)
	Funnel *FunnelBreakdown `json:"funnel,omitempty"`

	// Calendar context (NEW)
	CalendarMultiplier float64                  `json:"calendar_multiplier"`
	CalendarEvents     []map[string]interface{} `json:"calendar_events"`

	// Scale-up analysis (NEW — for over-performing products)
	MaxSafeBudget float64 `json:"max_safe_budget"`
	ScaleFactor   float64 `json:"scale_factor"`
	OptimalLabel  string  `json:"optimal_label"` // "Max Safe Scale-Up" or "Optimal Budget"
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
	// Funnel averages (NEW)
	AvgCTR float64
	AvgCVR float64
	AvgCPC float64
}
