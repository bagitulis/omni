package handlers

// UpdateAnalyticsSettingsRequest represents update settings request
type UpdateAnalyticsSettingsRequest struct {
	Platform          string  `json:"platform" binding:"required"`
	PriceColumn       string  `json:"price_column"`
	FormulaDeduction  float64 `json:"formula_deduction"`
	FormulaMultiplier float64 `json:"formula_multiplier"`
}
