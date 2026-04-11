package ads

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
)

// GetData retrieves TikTok ads data
func (s *TiktokAdsService) GetData(ctx context.Context, periodLabel string) ([]models.TiktokAdsCreativeData, error) {
	var data []models.TiktokAdsCreativeData
	query := s.db.WithContext(ctx).Where("tenant_id = ?", s.tenantID)
	if periodLabel != "" {
		query = query.Where("period_label = ?", periodLabel)
	}
	err := query.Order("cost DESC").Find(&data).Error
	return data, err
}

// GetSummary retrieves product summary
func (s *TiktokAdsService) GetSummary(ctx context.Context, periodLabel string) ([]models.TiktokAdsProductSummary, error) {
	var summary []models.TiktokAdsProductSummary
	query := s.db.WithContext(ctx).Where("tenant_id = ?", s.tenantID)
	if periodLabel != "" {
		query = query.Where("period_label = ?", periodLabel)
	}
	err := query.Order("total_revenue DESC").Find(&summary).Error
	return summary, err
}

// GetTrends retrieves trend data for TikTok charts
func (s *TiktokAdsService) GetTrends(ctx context.Context, startDate, endDate time.Time) ([]TrendDataPoint, error) {
	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	type Result struct {
		PeriodLabel string  `gorm:"column:period_label"`
		PeriodStart string  `gorm:"column:period_start"`
		TotalCost   float64 `gorm:"column:total_cost"`
		TotalRev    float64 `gorm:"column:total_rev"`
		TotalConv   int     `gorm:"column:total_conv"`
	}

	var results []Result
	err := s.db.WithContext(ctx).
		Model(&models.TiktokAdsCreativeData{}).
		Select("period_label, period_start, SUM(cost) as total_cost, SUM(gross_revenue) as total_rev, SUM(orders_sku) as total_conv").
		Where("tenant_id = ? AND period_start >= ? AND period_end <= ?", s.tenantID, startStr, endStr).
		Group("period_label, period_start").
		Order("period_start ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	var points []TrendDataPoint
	for _, r := range results {
		roas := 0.0
		if r.TotalCost > 0 {
			roas = r.TotalRev / r.TotalCost
		}
		points = append(points, TrendDataPoint{
			Label:       r.PeriodLabel,
			PeriodStart: r.PeriodStart,
			Spend:       r.TotalCost,
			GMV:         r.TotalRev,
			ROAS:        roas,
			OrderCount:  r.TotalConv,
		})
	}
	return points, nil
}

// GetProductPerformance retrieves aggregated product performance for TikTok
func (s *TiktokAdsService) GetProductPerformance(ctx context.Context, startDate, endDate time.Time) ([]ProductPerformance, error) {
	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	type AggResult struct {
		ProductID   string  `gorm:"column:product_id"`
		ProductName string  `gorm:"column:product_name"`
		Cost        float64 `gorm:"column:cost"`
		Revenue     float64 `gorm:"column:revenue"`
		Impressions int     `gorm:"column:impressions"`
		Clicks      int     `gorm:"column:clicks"`
		Conversions int     `gorm:"column:conversions"`
	}

	var results []AggResult
	err := s.db.WithContext(ctx).
		Model(&models.TiktokAdsCreativeData{}).
		Select("product_id, MAX(product_name) as product_name, SUM(cost) as cost, SUM(revenue) as revenue, SUM(impressions) as impressions, SUM(clicks) as clicks, SUM(conversions) as conversions").
		Where("tenant_id = ? AND period_start >= ? AND period_end <= ?", s.tenantID, startStr, endStr).
		Group("product_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Get historical data for trend analysis
	historicalData := s.getHistoricalDataByProduct(ctx, startStr, endStr)

	var perf []ProductPerformance
	cfg := DefaultScoringConfig()

	for _, r := range results {
		if r.ProductID == "" {
			continue
		}

		roas := 0.0
		if r.Cost > 0 {
			roas = r.Revenue / r.Cost
		}

		ctr := 0.0
		if r.Impressions > 0 {
			ctr = (float64(r.Clicks) / float64(r.Impressions)) * 100
		}

		profit := r.Revenue - r.Cost

		// Get historical values for this product
		var score ScoreResult
		if hist, ok := historicalData[r.ProductID]; ok && len(hist.ROIValues) >= 2 {
			score = CalculateFullScore(cfg, roas, profit, hist.ROIValues, hist.ProfitValues)
		} else {
			score = CalculateScores(cfg, roas, profit, nil, nil, nil)
		}

		perf = append(perf, ProductPerformance{
			ProductID:   r.ProductID,
			ProductName: r.ProductName,
			Spend:       r.Cost,
			GMV:         r.Revenue,
			ROAS:        roas,
			Impressions: r.Impressions,
			Clicks:      r.Clicks,
			CTR:         ctr,
			Conversions: r.Conversions,
			Score:       score,
		})
	}

	return perf, nil
}

// getHistoricalDataByProduct retrieves historical ROI and profit per product
func (s *TiktokAdsService) getHistoricalDataByProduct(ctx context.Context, startStr, endStr string) map[string]HistoricalValues {
	type PeriodData struct {
		ProductID   string  `gorm:"column:product_id"`
		PeriodStart string  `gorm:"column:period_start"`
		Cost        float64 `gorm:"column:cost"`
		Revenue     float64 `gorm:"column:revenue"`
	}

	var periodData []PeriodData
	s.db.WithContext(ctx).
		Model(&models.TiktokAdsCreativeData{}).
		Select("product_id, period_start, SUM(cost) as cost, SUM(revenue) as revenue").
		Where("tenant_id = ? AND period_start >= ? AND period_end <= ?", s.tenantID, startStr, endStr).
		Group("product_id, period_start").
		Order("period_start ASC").
		Scan(&periodData)

	result := make(map[string]HistoricalValues)

	for _, pd := range periodData {
		roi := 0.0
		if pd.Cost > 0 {
			roi = pd.Revenue / pd.Cost
		}
		profit := pd.Revenue - pd.Cost

		hist := result[pd.ProductID]
		hist.ROIValues = append(hist.ROIValues, roi)
		hist.ProfitValues = append(hist.ProfitValues, profit)
		result[pd.ProductID] = hist
	}

	return result
}

// GetPredictions retrieves ML predictions
func (s *TiktokAdsService) GetPredictions(ctx context.Context) ([]models.TiktokAdsMLPrediction, error) {
	var predictions []models.TiktokAdsMLPrediction
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND valid_until > ?", s.tenantID, time.Now()).
		Order("predicted_roi DESC").
		Find(&predictions).Error
	return predictions, err
}

// GetDataWithCursor retrieves TikTok ads data with cursor-based pagination
func (s *TiktokAdsService) GetDataWithCursor(ctx context.Context, periodLabel string, cursor *int, limit int) (*CursorPaginationResult, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100 // Default limit
	}

	query := s.db.WithContext(ctx).Where("tenant_id = ?", s.tenantID)
	if periodLabel != "" {
		query = query.Where("period_label = ?", periodLabel)
	}

	// Get total count
	var total int64
	if err := query.Model(&models.TiktokAdsCreativeData{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply cursor
	if cursor != nil {
		query = query.Where("id > ?", *cursor)
	}

	var data []models.TiktokAdsCreativeData
	if err := query.Order("id ASC").Limit(limit + 1).Find(&data).Error; err != nil {
		return nil, err
	}

	hasMore := len(data) > limit
	if hasMore {
		data = data[:limit] // Trim to limit
	}

	var nextCursor *string
	if hasMore && len(data) > 0 {
		lastID := data[len(data)-1].ID
		cursorStr := encodeIntCursor(int(lastID))
		nextCursor = &cursorStr
	}

	return &CursorPaginationResult{
		Data:       data,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Total:      total,
	}, nil
}
