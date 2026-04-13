package analytics

import (
	"context"
	"sync"
	"time"

	"github.com/omni/backend/internal/models"
	zlog "github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MLScoreCache represents a cached ML product score row
type MLScoreCache struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	TenantID           string    `gorm:"column:tenant_id;type:varchar(255);index;not null" json:"tenant_id"`
	ProductID          string    `gorm:"column:product_id;type:varchar(255)" json:"product_id"`
	OriginalProductID  string    `gorm:"column:original_product_id;type:varchar(255)" json:"original_product_id"`
	Platform           string    `gorm:"column:platform;type:varchar(20)" json:"platform"`
	ProductName        string    `gorm:"column:product_name;type:text" json:"product_name"`
	SKU                string    `gorm:"column:sku;type:varchar(255)" json:"sku"`
	UnifiedScore       float64   `gorm:"column:unified_score" json:"unified_score"`
	Category           string    `gorm:"column:category;type:varchar(50)" json:"category"`
	Action             string    `gorm:"column:action;type:varchar(50)" json:"action"`
	ActionLabel        string    `gorm:"column:action_label;type:varchar(100)" json:"action_label"`
	Recommendation     string    `gorm:"column:recommendation;type:text" json:"recommendation"`
	TotalCost          float64   `gorm:"column:total_cost" json:"total_cost"`
	TotalRevenue       float64   `gorm:"column:total_revenue" json:"total_revenue"`
	TotalProfit        float64   `gorm:"column:total_profit" json:"total_profit"`
	ROAS               float64   `gorm:"column:roas" json:"roas"`
	CTR                float64   `gorm:"column:ctr" json:"ctr"`
	Clicks             int       `gorm:"column:clicks" json:"clicks"`
	Impressions        int       `gorm:"column:impressions" json:"impressions"`
	HasFatigueWarning  bool      `gorm:"column:has_fatigue_warning" json:"has_fatigue_warning"`
	HasChurnRisk       bool      `gorm:"column:has_churn_risk" json:"has_churn_risk"`
	FatigueStatus      string    `gorm:"column:fatigue_status;type:varchar(50)" json:"fatigue_status"`
	ChurnRiskScore     float64   `gorm:"column:churn_risk_score" json:"churn_risk_score"`
	ROASScore          float64   `gorm:"column:roas_score" json:"roas_score"`
	TrendScore         float64   `gorm:"column:trend_score" json:"trend_score"`
	VolatilityScore    float64   `gorm:"column:volatility_score" json:"volatility_score"`
	MomentumScore      float64   `gorm:"column:momentum_score" json:"momentum_score"`
	BudgetChangePct    float64   `gorm:"column:budget_change_pct" json:"budget_change_pct"`
	ConfidenceLevel    string    `gorm:"column:confidence_level;type:varchar(20)" json:"confidence_level"`
	SuccessProbability float64   `gorm:"column:success_probability" json:"success_probability"`
	TrendDirection     string    `gorm:"column:trend_direction;type:varchar(50)" json:"trend_direction"`
	TrendStrength      float64   `gorm:"column:trend_strength" json:"trend_strength"`
	CalculatedAt       time.Time `gorm:"column:calculated_at" json:"calculated_at"`
}

func (MLScoreCache) TableName() string { return models.GetTableName("MLScoreCache") }

// MLCalculationStatus tracks background recalculation state per tenant
type MLCalculationStatus struct {
	TenantID     string     `gorm:"column:tenant_id;primaryKey;type:varchar(255)" json:"tenant_id"`
	Status       string     `gorm:"column:status;type:varchar(20);not null;default:'IDLE'" json:"status"` // IDLE | PROCESSING | DONE | ERROR
	StartedAt    *time.Time `gorm:"column:started_at" json:"started_at"`
	CompletedAt  *time.Time `gorm:"column:completed_at" json:"completed_at"`
	ErrorMessage string     `gorm:"column:error_message;type:text" json:"error_message"`
	ProductCount int        `gorm:"column:product_count" json:"product_count"`
}

func (MLCalculationStatus) TableName() string { return models.GetTableName("MLCalculationStatus") }

// MLCacheService handles ML score caching and background recalculation
type MLCacheService struct {
	mu sync.Mutex
}

// NewMLCacheService creates a new ML cache service
func NewMLCacheService() *MLCacheService {
	return &MLCacheService{}
}

// EnsureTables creates the cache tables if they don't exist
func (s *MLCacheService) EnsureTables(db *gorm.DB) {
	_ = db.AutoMigrate(&MLScoreCache{}, &MLCalculationStatus{})
}

// GetCachedProducts retrieves cached ML product scores
func (s *MLCacheService) GetCachedProducts(
	ctx context.Context, db *gorm.DB, tenantID string,
	limit int, sortBy, sortDir string,
) ([]MLScoreCache, int64, error) {
	var products []MLScoreCache
	var total int64

	db.WithContext(ctx).Model(&MLScoreCache{}).
		Where("tenant_id = ?", tenantID).
		Count(&total)

	query := db.WithContext(ctx).
		Where("tenant_id = ?", tenantID)

	// Apply sorting
	orderClause := "unified_score DESC"
	if sortBy != "" {
		dir := "DESC"
		if sortDir == "asc" {
			dir = "ASC"
		}
		orderClause = sortBy + " " + dir
	}
	query = query.Order(orderClause)

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&products).Error
	return products, total, err
}

// GetCachedHealth builds portfolio health from cached scores
func (s *MLCacheService) GetCachedHealth(
	ctx context.Context, db *gorm.DB, tenantID string,
) (*models.PortfolioHealth, bool) {
	var count int64
	db.WithContext(ctx).Model(&MLScoreCache{}).
		Where("tenant_id = ?", tenantID).Count(&count)

	if count == 0 {
		return nil, false
	}

	health := &models.PortfolioHealth{
		TenantID:    tenantID,
		LastUpdated: time.Now(),
	}

	var products []MLScoreCache
	db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&products)

	var totalScore float64
	for _, p := range products {
		health.TotalProducts++
		health.TotalCost += p.TotalCost
		health.TotalRevenue += p.TotalRevenue
		health.TotalProfit += p.TotalProfit
		totalScore += p.UnifiedScore

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

		if p.HasFatigueWarning {
			health.FatigueWarnings++
			health.ActiveAlerts++
		}
		if p.HasChurnRisk {
			health.ChurnRisks++
			health.ActiveAlerts++
		}
	}

	if health.TotalCost > 0 {
		health.OverallROAS = health.TotalRevenue / health.TotalCost
	}
	if health.TotalProducts > 0 {
		health.HealthScore = roundTo(totalScore/float64(health.TotalProducts), 1)
	}
	health.HealthLabel = getHealthLabel(health.HealthScore)

	return health, true
}

// GetStatus retrieves calculation status for a tenant
func (s *MLCacheService) GetStatus(
	ctx context.Context, db *gorm.DB, tenantID string,
) MLCalculationStatus {
	var status MLCalculationStatus
	err := db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&status).Error
	if err != nil {
		return MLCalculationStatus{TenantID: tenantID, Status: "IDLE"}
	}
	return status
}

// TriggerRecalculate starts a background ML recalculation
func (s *MLCacheService) TriggerRecalculate(
	db *gorm.DB, tenantID string,
) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already processing
	status := s.GetStatus(context.Background(), db, tenantID)
	if status.Status == "PROCESSING" {
		return "ALREADY_PROCESSING", nil
	}

	// Set status to PROCESSING
	now := time.Now()
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "started_at", "error_message", "product_count"}),
	}).Create(&MLCalculationStatus{
		TenantID:     tenantID,
		Status:       "PROCESSING",
		StartedAt:    &now,
		ErrorMessage: "",
		ProductCount: 0,
	})

	// Launch background goroutine
	go s.runRecalculation(db, tenantID)

	return "PROCESSING", nil
}

// runRecalculation performs the heavy ML scoring in the background
func (s *MLCacheService) runRecalculation(db *gorm.DB, tenantID string) {
	ctx := context.Background()

	zlog.Info().Str("tenant_id", tenantID).Msg("ML recalculation started")

	// Use the existing MLAnalyticsService to do the heavy computation
	svc := NewMLAnalyticsService(db, tenantID)
	products, err := svc.getProductAnalyses(ctx, "", 0, "")

	if err != nil {
		now := time.Now()
		db.Model(&MLCalculationStatus{}).
			Where("tenant_id = ?", tenantID).
			Updates(map[string]interface{}{
				"status":        "ERROR",
				"completed_at":  &now,
				"error_message": err.Error(),
			})
		zlog.Error().Err(err).Str("tenant_id", tenantID).Msg("ML recalculation failed")
		return
	}

	// Delete old cache for this tenant
	db.Where("tenant_id = ?", tenantID).Delete(&MLScoreCache{})

	// Insert new scores
	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	for _, p := range products {
		cache := MLScoreCache{
			TenantID:           tenantID,
			ProductID:          p.ProductID,
			OriginalProductID:  p.ProductID,
			Platform:           p.Platform,
			ProductName:        p.ProductName,
			SKU:                p.SKU,
			UnifiedScore:       p.UnifiedScore,
			Category:           p.Category,
			Action:             p.Action,
			ActionLabel:        p.ActionLabel,
			Recommendation:     p.Recommendation,
			TotalCost:          p.TotalCost,
			TotalRevenue:       p.TotalRevenue,
			TotalProfit:        p.TotalProfit,
			ROAS:               p.ROAS,
			CTR:                p.CTR,
			Clicks:             p.Clicks,
			Impressions:        p.Impressions,
			HasFatigueWarning:  p.HasFatigueWarning,
			HasChurnRisk:       p.HasChurnRisk,
			FatigueStatus:      p.FatigueStatus,
			ChurnRiskScore:     p.ChurnRiskScore,
			ROASScore:          p.ROASScore,
			TrendScore:         p.TrendScore,
			VolatilityScore:    p.VolatilityScore,
			MomentumScore:      p.MomentumScore,
			BudgetChangePct:    p.BudgetChangePct,
			ConfidenceLevel:    p.ConfidenceLevel,
			SuccessProbability: p.SuccessProbability,
			TrendDirection:     p.TrendDirection,
			TrendStrength:      p.TrendStrength,
			CalculatedAt:       now,
		}
		// Set last_updated on the product for JSON response
		p.LastUpdated = nowStr
		db.Create(&cache)
	}

	// Update status
	db.Model(&MLCalculationStatus{}).
		Where("tenant_id = ?", tenantID).
		Updates(map[string]interface{}{
			"status":        "DONE",
			"completed_at":  &now,
			"product_count": len(products),
		})

	zlog.Info().
		Str("tenant_id", tenantID).
		Int("product_count", len(products)).
		Msg("ML recalculation completed")
}

// GetCachedProductByID retrieves a single cached product by product_id
func (s *MLCacheService) GetCachedProductByID(
	ctx context.Context, db *gorm.DB, tenantID, productID string,
) (*MLScoreCache, error) {
	var product MLScoreCache
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND product_id = ?", tenantID, productID).
		First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetCachedAlerts derives alerts from cached product scores
func (s *MLCacheService) GetCachedAlerts(
	ctx context.Context, db *gorm.DB, tenantID string,
) []MLScoreCache {
	var products []MLScoreCache
	db.WithContext(ctx).
		Where("tenant_id = ? AND (has_fatigue_warning = ? OR has_churn_risk = ?)", tenantID, true, true).
		Order("churn_risk_score DESC").
		Find(&products)
	return products
}

// GetCachedDistribution derives score distribution from cached products
func (s *MLCacheService) GetCachedDistribution(
	ctx context.Context, db *gorm.DB, tenantID string,
) map[string]int {
	type CategoryCount struct {
		Category string `gorm:"column:category"`
		Count    int    `gorm:"column:count"`
	}
	var counts []CategoryCount
	db.WithContext(ctx).
		Model(&MLScoreCache{}).
		Select("category, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("category").
		Scan(&counts)

	result := map[string]int{"STAR": 0, "GROWTH": 0, "STABLE": 0, "WATCH": 0, "PROBLEM": 0}
	for _, c := range counts {
		result[c.Category] = c.Count
	}
	return result
}

