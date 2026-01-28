package models

import "time"

// MLProductAnalysis represents comprehensive ML analysis for a product
// This is a computed struct (not stored in DB), aggregated from ads data
type MLProductAnalysis struct {
	TenantID     string `json:"-"` // Hidden from response
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	CreativeType string `json:"creative_type"`

	// Financial Metrics
	TotalCost    float64 `json:"total_cost"`
	TotalRevenue float64 `json:"total_revenue"`
	TotalProfit  float64 `json:"total_profit"`
	TotalOrders  int     `json:"total_orders"`
	ROAS         float64 `json:"roas"`
	PeriodCount  int     `json:"period_count"`

	// Unified Score (0-100)
	UnifiedScore    float64 `json:"unified_score"`
	ROASScore       float64 `json:"roas_score"`
	TrendScore      float64 `json:"trend_score"`
	VolatilityScore float64 `json:"volatility_score"`
	MomentumScore   float64 `json:"momentum_score"`

	// Category & Action
	Category        string  `json:"category"`     // STAR, GROWTH, STABLE, WATCH, PROBLEM
	Action          string  `json:"action"`       // SCALE_UP, MAINTAIN, REDUCE, STOP
	ActionLabel     string  `json:"action_label"` // Human readable
	BudgetChangePct float64 `json:"budget_change_pct"`

	// Alerts
	HasFatigueWarning bool    `json:"has_fatigue_warning"`
	HasChurnRisk      bool    `json:"has_churn_risk"`
	FatigueStatus     string  `json:"fatigue_status"` // FRESH, AGING, FATIGUED, DEAD
	ChurnRiskScore    float64 `json:"churn_risk_score"`

	// Confidence
	ConfidenceLevel    string  `json:"confidence_level"` // HIGH, MEDIUM, LOW
	SuccessProbability float64 `json:"success_probability"`

	// Trend Indicators
	TrendDirection string  `json:"trend_direction"` // UP, DOWN, STABLE
	TrendStrength  float64 `json:"trend_strength"`
}

// PortfolioHealth represents overall portfolio health summary
type PortfolioHealth struct {
	TenantID string `json:"-"`

	// Health Score (0-100)
	HealthScore float64   `json:"health_score"`
	HealthLabel string    `json:"health_label"` // Excellent, Good, Fair, Poor
	LastUpdated time.Time `json:"last_updated"`

	// Category Counts
	TotalProducts int `json:"total_products"`
	StarCount     int `json:"star_count"`
	GrowthCount   int `json:"growth_count"`
	StableCount   int `json:"stable_count"`
	WatchCount    int `json:"watch_count"`
	ProblemCount  int `json:"problem_count"`

	// Action Counts
	ScaleUpCount  int `json:"scale_up_count"`
	MaintainCount int `json:"maintain_count"`
	ReduceCount   int `json:"reduce_count"`
	StopCount     int `json:"stop_count"`

	// Financial Summary
	TotalCost    float64 `json:"total_cost"`
	TotalRevenue float64 `json:"total_revenue"`
	TotalProfit  float64 `json:"total_profit"`
	OverallROAS  float64 `json:"overall_roas"`

	// Alert Summary
	ActiveAlerts    int `json:"active_alerts"`
	FatigueWarnings int `json:"fatigue_warnings"`
	ChurnRisks      int `json:"churn_risks"`
}

// MLAlert represents an active alert for a product
type MLAlert struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"-"`
	ProductID   string    `json:"product_id"`
	ProductName string    `json:"product_name"`
	AlertType   string    `json:"alert_type"` // FATIGUE_WARNING, CHURN_RISK, BUDGET_REC
	Severity    string    `json:"severity"`   // HIGH, MEDIUM, LOW
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"` // ACTIVE, DISMISSED
}

// BudgetSimRequest represents a budget simulation request
type BudgetSimRequest struct {
	ProductIDs      []string `json:"product_ids"`
	BudgetChangePct float64  `json:"budget_change_pct"`
}

// BudgetSimResult represents budget simulation result
type BudgetSimResult struct {
	ExpectedRevenue  float64 `json:"expected_revenue"`
	ExpectedProfit   float64 `json:"expected_profit"`
	ExpectedROAS     float64 `json:"expected_roas"`
	ConfidenceLevel  string  `json:"confidence_level"`
	RevenueChangeAmt float64 `json:"revenue_change_amt"`
	ProfitChangeAmt  float64 `json:"profit_change_amt"`
}

// ScoreDistribution represents score distribution for charts
type ScoreDistribution struct {
	Category   string  `json:"category"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// TrendDataPointML represents trend data for ML charts
type TrendDataPointML struct {
	Period   string  `json:"period"`
	Cost     float64 `json:"cost"`
	Revenue  float64 `json:"revenue"`
	Profit   float64 `json:"profit"`
	ROAS     float64 `json:"roas"`
	Orders   int     `json:"orders"`
	AvgScore float64 `json:"avg_score"`
}
