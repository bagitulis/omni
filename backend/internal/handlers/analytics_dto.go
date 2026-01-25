package handlers

// UpdateAnalyticsSettingsRequest represents update settings request
type UpdateAnalyticsSettingsRequest struct {
	Platform          string  `json:"platform" binding:"required"`
	PriceColumn       string  `json:"priceColumn"`
	FormulaDeduction  float64 `json:"formulaDeduction"`
	FormulaMultiplier float64 `json:"formulaMultiplier"`
}
