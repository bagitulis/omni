package dto

import (
	"time"

	"github.com/omni/backend/internal/models"
)

// PortfolioHealthResponse is the API response for portfolio health
type PortfolioHealthResponse struct {
	Success bool                   `json:"success"`
	Data    models.PortfolioHealth `json:"data"`
}

// MLProductsResponse is the paginated response for ML product analysis
type MLProductsResponse struct {
	Success bool                       `json:"success"`
	Data    []models.MLProductAnalysis `json:"data"`
	Meta    MLPaginationMeta           `json:"meta"`
}

// MLPaginationMeta contains cursor-based pagination info
type MLPaginationMeta struct {
	Total      int64   `json:"total"`
	Limit      int     `json:"limit"`
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor,omitempty"`
}

// MLProductDetailResponse is the response for a single product detail
type MLProductDetailResponse struct {
	Success bool                     `json:"success"`
	Data    models.MLProductAnalysis `json:"data"`
}

// MLAlertsResponse is the response for active alerts
type MLAlertsResponse struct {
	Success bool             `json:"success"`
	Data    []models.MLAlert `json:"data"`
	Meta    AlertsMeta       `json:"meta"`
}

// AlertsMeta contains alert summary
type AlertsMeta struct {
	TotalActive    int `json:"total_active"`
	HighPriority   int `json:"high_priority"`
	MediumPriority int `json:"medium_priority"`
	LowPriority    int `json:"low_priority"`
}

// BudgetSimResponse is the response for budget simulation
type BudgetSimResponse struct {
	Success bool                   `json:"success"`
	Data    models.BudgetSimResult `json:"data"`
}

// BudgetSimRequest is the request body for budget simulation
type BudgetSimRequest struct {
	ProductIDs      []string `json:"product_ids" binding:"required"`
	BudgetChangePct float64  `json:"budget_change_pct" binding:"required"`
}

// ScoreDistributionResponse is the response for score distribution chart
type ScoreDistributionResponse struct {
	Success bool                       `json:"success"`
	Data    []models.ScoreDistribution `json:"data"`
}

// MLTrendResponse is the response for ML trend chart data
type MLTrendResponse struct {
	Success bool                      `json:"success"`
	Data    []models.TrendDataPointML `json:"data"`
}

// MLProductsQueryParams represents query parameters for products list
type MLProductsQueryParams struct {
	Limit    int    `form:"limit"`
	Cursor   string `form:"cursor"`
	SortBy   string `form:"sort_by"`
	SortDir  string `form:"sort_dir"`
	Category string `form:"category"`
	Action   string `form:"action"`
	Platform string `form:"platform"`
}

// DefaultMLProductsQuery returns default query params
func DefaultMLProductsQuery() MLProductsQueryParams {
	return MLProductsQueryParams{
		Limit:   20,
		SortBy:  "unified_score",
		SortDir: "desc",
	}
}

// MLTrendQueryParams represents query parameters for trend data
type MLTrendQueryParams struct {
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
	Granularity string `form:"granularity"` // daily, weekly, monthly
	Platform    string `form:"platform"`
}

// BatchActionRequest for applying recommendations to multiple products
type BatchActionRequest struct {
	ProductIDs []string `json:"product_ids" binding:"required"`
	Action     string   `json:"action" binding:"required"`
}

// BatchActionResponse for batch action results
type BatchActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Updated int    `json:"updated"`
	Failed  int    `json:"failed"`
}

// AlertDismissRequest for dismissing an alert
type AlertDismissRequest struct {
	AlertID string `json:"alert_id" binding:"required"`
}

// MLDashboardSummary combines multiple data for dashboard
type MLDashboardSummary struct {
	Health       models.PortfolioHealth     `json:"health"`
	TopProducts  []models.MLProductAnalysis `json:"top_products"`
	Alerts       []models.MLAlert           `json:"alerts"`
	Distribution []models.ScoreDistribution `json:"distribution"`
	UpdatedAt    time.Time                  `json:"updated_at"`
}

// MLDashboardResponse is the response for dashboard summary
type MLDashboardResponse struct {
	Success bool               `json:"success"`
	Data    MLDashboardSummary `json:"data"`
}
