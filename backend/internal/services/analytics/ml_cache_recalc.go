package analytics

import (
	"context"
	"time"

	zlog "github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TriggerRecalculate starts a background ML recalculation
func (s *MLCacheService) TriggerRecalculate(
	db *gorm.DB, tenantID string,
) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	status := s.GetStatus(context.Background(), db, tenantID)
	if status.Status == "PROCESSING" {
		return "ALREADY_PROCESSING", nil
	}

	now := time.Now()
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "started_at", "error_message", "product_count"}),
	}).Create(&MLCalculationStatus{
		TenantID: tenantID, Status: "PROCESSING", StartedAt: &now,
		ErrorMessage: "", ProductCount: 0,
	})

	go s.runRecalculation(db, tenantID)
	return "PROCESSING", nil
}

// runRecalculation performs the heavy ML scoring in the background
func (s *MLCacheService) runRecalculation(db *gorm.DB, tenantID string) {
	ctx := context.Background()
	zlog.Info().Str("tenant_id", tenantID).Msg("ML recalculation started")

	svc := NewMLAnalyticsService(db, tenantID)
	products, err := svc.getProductAnalyses(ctx, "", 0, "")

	if err != nil {
		now := time.Now()
		db.Model(&MLCalculationStatus{}).Where("tenant_id = ?", tenantID).
			Updates(map[string]interface{}{"status": "ERROR", "completed_at": &now, "error_message": err.Error()})
		zlog.Error().Err(err).Str("tenant_id", tenantID).Msg("ML recalculation failed")
		return
	}

	db.Where("tenant_id = ?", tenantID).Delete(&MLScoreCache{})

	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	for _, p := range products {
		cache := MLScoreCache{
			TenantID: tenantID, ProductID: p.ProductID, OriginalProductID: p.ProductID,
			Platform: p.Platform, ProductName: p.ProductName, SKU: p.SKU,
			UnifiedScore: p.UnifiedScore, Category: p.Category, Action: p.Action,
			ActionLabel: p.ActionLabel, Recommendation: p.Recommendation,
			TotalCost: p.TotalCost, TotalRevenue: p.TotalRevenue, TotalProfit: p.TotalProfit,
			ROAS: p.ROAS, CTR: p.CTR, Clicks: p.Clicks, Impressions: p.Impressions,
			HasFatigueWarning: p.HasFatigueWarning, HasChurnRisk: p.HasChurnRisk,
			FatigueStatus: p.FatigueStatus, ChurnRiskScore: p.ChurnRiskScore,
			ROASScore: p.ROASScore, TrendScore: p.TrendScore,
			VolatilityScore: p.VolatilityScore, MomentumScore: p.MomentumScore,
			BudgetChangePct: p.BudgetChangePct, ConfidenceLevel: p.ConfidenceLevel,
			SuccessProbability: p.SuccessProbability, TrendDirection: p.TrendDirection,
			TrendStrength: p.TrendStrength, CalculatedAt: now,
		}
		p.LastUpdated = nowStr
		db.Create(&cache)
	}

	db.Model(&MLCalculationStatus{}).Where("tenant_id = ?", tenantID).
		Updates(map[string]interface{}{"status": "DONE", "completed_at": &now, "product_count": len(products)})

	zlog.Info().Str("tenant_id", tenantID).Int("product_count", len(products)).Msg("ML recalculation completed")
}
