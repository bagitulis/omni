package analytics

import (
	"context"
	"math"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// MLAnalyticsService handles ML analytics calculations
type MLAnalyticsService struct {
	db       *gorm.DB
	tenantID string
}

// NewMLAnalyticsService creates a new ML analytics service
func NewMLAnalyticsService(db *gorm.DB, tenantID string) *MLAnalyticsService {
	return &MLAnalyticsService{db: db, tenantID: tenantID}
}

// GetPortfolioHealth calculates overall portfolio health from TikTok ads data
func (s *MLAnalyticsService) GetPortfolioHealth(ctx context.Context, platform string) (*models.PortfolioHealth, error) {
	products, err := s.getProductAnalyses(ctx, platform, 0, "")
	if err != nil {
		return nil, err
	}

	health := &models.PortfolioHealth{
		TenantID:    s.tenantID,
		LastUpdated: time.Now(),
	}

	if len(products) == 0 {
		health.HealthLabel = "No Data"
		return health, nil
	}

	var totalScore float64
	for _, p := range products {
		health.TotalProducts++
		health.TotalCost += p.TotalCost
		health.TotalRevenue += p.TotalRevenue
		health.TotalProfit += p.TotalProfit
		totalScore += p.UnifiedScore

		// Count by category
		switch p.Category {
		case "STAR":
			health.StarCount++
		case "GROWTH":
			health.GrowthCount++
		case "STABLE":
			health.StableCount++
		case "WATCH":
			health.WatchCount++
		case "PROBLEM":
			health.ProblemCount++
		}

		// Count by action
		switch p.Action {
		case "SCALE_UP":
			health.ScaleUpCount++
		case "MAINTAIN":
			health.MaintainCount++
		case "REDUCE":
			health.ReduceCount++
		case "STOP":
			health.StopCount++
		}

		// Count alerts
		if p.HasFatigueWarning {
			health.FatigueWarnings++
			health.ActiveAlerts++
		}
		if p.HasChurnRisk {
			health.ChurnRisks++
			health.ActiveAlerts++
		}
	}

	// Calculate overall ROAS
	if health.TotalCost > 0 {
		health.OverallROAS = health.TotalRevenue / health.TotalCost
	}

	// Calculate health score (average of unified scores)
	health.HealthScore = roundTo(totalScore/float64(len(products)), 1)
	health.HealthLabel = getHealthLabel(health.HealthScore)

	return health, nil
}

// GetProducts retrieves paginated product analyses with cursor pagination
func (s *MLAnalyticsService) GetProducts(
	ctx context.Context,
	platform string,
	limit int,
	cursor string,
	sortBy string,
	sortDir string,
	categoryFilter string,
	actionFilter string,
) ([]models.MLProductAnalysis, int64, bool, *string, error) {

	products, err := s.getProductAnalyses(ctx, platform, limit+1, cursor)
	if err != nil {
		return nil, 0, false, nil, err
	}

	// Apply filters
	filtered := s.applyFilters(products, categoryFilter, actionFilter)

	// Sort
	s.sortProducts(filtered, sortBy, sortDir)

	// Pagination
	hasMore := len(filtered) > limit
	if hasMore {
		filtered = filtered[:limit]
	}

	var nextCursor *string
	if hasMore && len(filtered) > 0 {
		lastID := filtered[len(filtered)-1].ProductID
		nextCursor = &lastID
	}

	// Count total (from DB directly for accuracy)
	var total int64
	s.db.WithContext(ctx).
		Model(&models.TiktokAdsCreativeData{}).
		Where("tenant_id = ?", s.tenantID).
		Select("COUNT(DISTINCT product_id)").
		Scan(&total)

	return filtered, total, hasMore, nextCursor, nil
}

// GetProductDetail retrieves detailed analysis for a single product
func (s *MLAnalyticsService) GetProductDetail(ctx context.Context, productID string) (*models.MLProductAnalysis, error) {
	products, err := s.getProductAnalyses(ctx, "", 0, "")
	if err != nil {
		return nil, err
	}

	for _, p := range products {
		if p.ProductID == productID {
			return &p, nil
		}
	}

	return nil, nil
}

// GetAlerts retrieves active alerts based on product analyses
func (s *MLAnalyticsService) GetAlerts(ctx context.Context) ([]models.MLAlert, error) {
	products, err := s.getProductAnalyses(ctx, "", 0, "")
	if err != nil {
		return nil, err
	}

	var alerts []models.MLAlert
	now := time.Now()

	for _, p := range products {
		if p.HasFatigueWarning {
			alerts = append(alerts, models.MLAlert{
				ID:          p.ProductID + "_fatigue",
				TenantID:    s.tenantID,
				ProductID:   p.ProductID,
				ProductName: p.ProductName,
				AlertType:   "FATIGUE_WARNING",
				Severity:    "HIGH",
				Message:     "Creative fatigue detected: " + p.FatigueStatus,
				CreatedAt:   now,
				Status:      "ACTIVE",
			})
		}
		if p.HasChurnRisk {
			severity := "MEDIUM"
			if p.ChurnRiskScore > 70 {
				severity = "HIGH"
			}
			alerts = append(alerts, models.MLAlert{
				ID:          p.ProductID + "_churn",
				TenantID:    s.tenantID,
				ProductID:   p.ProductID,
				ProductName: p.ProductName,
				AlertType:   "CHURN_RISK",
				Severity:    severity,
				Message:     "High churn risk score: " + formatPercent(p.ChurnRiskScore),
				CreatedAt:   now,
				Status:      "ACTIVE",
			})
		}
	}

	return alerts, nil
}

// SimulateBudget simulates budget changes and predicts outcomes
func (s *MLAnalyticsService) SimulateBudget(
	ctx context.Context,
	productIDs []string,
	budgetChangePct float64,
) (*models.BudgetSimResult, error) {

	products, err := s.getProductAnalyses(ctx, "", 0, "")
	if err != nil {
		return nil, err
	}

	// Filter to selected products
	selected := make([]models.MLProductAnalysis, 0)
	for _, p := range products {
		for _, id := range productIDs {
			if p.ProductID == id {
				selected = append(selected, p)
				break
			}
		}
	}

	if len(selected) == 0 {
		return &models.BudgetSimResult{
			ConfidenceLevel: "LOW",
		}, nil
	}

	// Aggregate totals from selected products
	var totalCost, totalRevenue, totalProfit float64
	for _, p := range selected {
		totalCost += p.TotalCost
		totalRevenue += p.TotalRevenue
		totalProfit += p.TotalProfit
	}

	// Power-law diminishing returns model
	// Revenue scales sub-linearly as budget increases
	budgetMultiplier := 1 + (budgetChangePct / 100)
	if budgetMultiplier <= 0 {
		budgetMultiplier = 0.01 // Protect against negative/zero
	}
	elasticity := 0.82 // Revenue elasticity to spend (< 1.0 = diminishing returns)
	revenueMultiplier := math.Pow(budgetMultiplier, elasticity)

	expectedRevenue := totalRevenue * revenueMultiplier
	expectedCost := totalCost * budgetMultiplier
	expectedProfit := expectedRevenue - expectedCost

	var expectedROAS float64
	if expectedCost > 0 {
		expectedROAS = expectedRevenue / expectedCost
	}

	confidence := "MEDIUM"
	if len(selected) >= 5 {
		confidence = "HIGH"
	} else if len(selected) == 1 {
		confidence = "LOW"
	}

	return &models.BudgetSimResult{
		ExpectedRevenue:  roundTo(expectedRevenue, 0),
		ExpectedProfit:   roundTo(expectedProfit, 0),
		ExpectedROAS:     roundTo(expectedROAS, 2),
		ConfidenceLevel:  confidence,
		RevenueChangeAmt: roundTo(expectedRevenue-totalRevenue, 0),
		ProfitChangeAmt:  roundTo(expectedProfit-totalProfit, 0),
	}, nil
}

// GetScoreDistribution returns distribution of products by category
func (s *MLAnalyticsService) GetScoreDistribution(ctx context.Context) ([]models.ScoreDistribution, error) {
	products, err := s.getProductAnalyses(ctx, "", 0, "")
	if err != nil {
		return nil, err
	}

	counts := map[string]int{"STAR": 0, "GROWTH": 0, "STABLE": 0, "WATCH": 0, "PROBLEM": 0}

	for _, p := range products {
		counts[p.Category]++
	}

	total := len(products)
	result := make([]models.ScoreDistribution, 0, 5)
	for _, cat := range []string{"STAR", "GROWTH", "STABLE", "WATCH", "PROBLEM"} {
		pct := 0.0
		if total > 0 {
			pct = float64(counts[cat]) / float64(total) * 100
		}
		result = append(result, models.ScoreDistribution{
			Category:   cat,
			Count:      counts[cat],
			Percentage: roundTo(pct, 1),
		})
	}

	return result, nil
}
