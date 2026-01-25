package ads

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
)

// GetData retrieves ads data with filtering
func (s *ShopeeAdsService) GetData(ctx context.Context, periodLabel string) ([]models.ShopeeAdsProductData, error) {
	var data []models.ShopeeAdsProductData
	query := s.db.WithContext(ctx).Where("tenant_id = ?", s.tenantID)
	if periodLabel != "" {
		query = query.Where("period_label = ?", periodLabel)
	}
	err := query.Order("cost DESC").Find(&data).Error
	return data, err
}

// GetBatches retrieves all upload batches
func (s *ShopeeAdsService) GetBatches(ctx context.Context) ([]models.ShopeeAdsUploadBatch, error) {
	var batches []models.ShopeeAdsUploadBatch
	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", s.tenantID).
		Order("uploaded_at DESC").
		Find(&batches).Error
	return batches, err
}

// GetSummary calculates summary statistics
func (s *ShopeeAdsService) GetSummary(ctx context.Context, periodLabel string) (*AdsSummary, error) {
	var data []models.ShopeeAdsProductData
	query := s.db.WithContext(ctx).Where("tenant_id = ?", s.tenantID)
	if periodLabel != "" {
		query = query.Where("period_label = ?", periodLabel)
	}
	if err := query.Find(&data).Error; err != nil {
		return nil, err
	}

	return calculateShopeeAdsSummary(data), nil
}

// GetTrends retrieves trend data for charts within a date range
func (s *ShopeeAdsService) GetTrends(ctx context.Context, startDate, endDate time.Time) ([]TrendDataPoint, error) {
	// We aggregate by Batch (Period)
	// Assuming PeriodStart is stored as YYYY-MM-DD string in DB
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
	// Group by PeriodLabel (Batch)
	err := s.db.WithContext(ctx).
		Model(&models.ShopeeAdsProductData{}).
		Select("period_label, period_start, SUM(cost) as total_cost, SUM(revenue) as total_rev, SUM(conversions) as total_conv").
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

// GetProductPerformance retrieves detailed product performance with scores
func (s *ShopeeAdsService) GetProductPerformance(ctx context.Context, startDate, endDate time.Time) ([]ProductPerformance, error) {
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
		UnitsSold   int     `gorm:"column:units_sold"`
	}

	var results []AggResult
	err := s.db.WithContext(ctx).
		Model(&models.ShopeeAdsProductData{}).
		Select("product_id, MAX(product_name) as product_name, SUM(cost) as cost, SUM(revenue) as revenue, SUM(impressions) as impressions, SUM(clicks) as clicks, SUM(conversions) as conversions, SUM(units_sold) as units_sold").
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
			Sold:        r.UnitsSold,
			Score:       score,
		})
	}

	return perf, nil
}

// HistoricalValues holds time series data for a product
type HistoricalValues struct {
	ROIValues    []float64
	ProfitValues []float64
}

// getHistoricalDataByProduct retrieves historical ROI and profit per product
func (s *ShopeeAdsService) getHistoricalDataByProduct(ctx context.Context, startStr, endStr string) map[string]HistoricalValues {
	type PeriodData struct {
		ProductID   string  `gorm:"column:product_id"`
		PeriodStart string  `gorm:"column:period_start"`
		Cost        float64 `gorm:"column:cost"`
		Revenue     float64 `gorm:"column:revenue"`
	}

	var periodData []PeriodData
	s.db.WithContext(ctx).
		Model(&models.ShopeeAdsProductData{}).
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
