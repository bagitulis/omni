package analytics

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	analyticsService "github.com/omni/backend/internal/services/analytics"
	"github.com/omni/backend/internal/services/analytics/intelligence"
	"gorm.io/gorm"
)

// SimulationHandler handles budget simulation requests
type SimulationHandler struct {
	basePath     string
	simulator    *intelligence.BudgetSimulator
	cacheService *analyticsService.MLCacheService
}

// NewSimulationHandler creates a new simulation handler
func NewSimulationHandler(basePath string) *SimulationHandler {
	return &SimulationHandler{
		basePath:     basePath,
		simulator:    intelligence.NewBudgetSimulator(),
		cacheService: analyticsService.NewMLCacheService(),
	}
}

// SimulateRequest represents the simulation request body
type SimulateRequest struct {
	ProductID    string  `json:"product_id" binding:"required"`
	TargetRoas   float64 `json:"target_roas" binding:"required,gt=0"`
	BudgetPerDay float64 `json:"budget_per_day" binding:"required,gt=0"`
	PeriodDays   int     `json:"period_days"`
}

// Simulate handles POST /api/analytics/simulation/calculate
func (h *SimulationHandler) Simulate(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	var req SimulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	if req.PeriodDays == 0 {
		req.PeriodDays = 7
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()

	// Get product historical data from ads database (daily-normalized)
	historicalData := h.getProductHistoricalData(ctx, db, tenantID, req.ProductID)

	// Enrich with ML category from cache
	h.enrichWithMLData(ctx, db, tenantID, req.ProductID, &historicalData)

	// Run simulation
	simReq := intelligence.SimulationRequest{
		ProductID:    req.ProductID,
		TargetRoas:   req.TargetRoas,
		BudgetPerDay: req.BudgetPerDay,
		PeriodDays:   req.PeriodDays,
	}

	result := h.simulator.Simulate(simReq, historicalData)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// enrichWithMLData looks up ML cache to inject category and fatigue info
func (h *SimulationHandler) enrichWithMLData(
	ctx context.Context, db *gorm.DB,
	tenantID, productID string,
	data *intelligence.ProductHistoricalData,
) {
	// ML cache stores IDs with platform prefix: "tiktok_123" or "shopee_123"
	for _, prefix := range []string{"tiktok_", "shopee_"} {
		prefixedID := prefix + productID
		cached, err := h.cacheService.GetCachedProductByID(ctx, db, tenantID, prefixedID)
		if err == nil && cached != nil {
			data.MLCategory = cached.Category
			data.FatigueStatus = cached.FatigueStatus
			return
		}
	}
	// Default if not found in cache
	data.MLCategory = "STABLE"
	data.FatigueStatus = "FRESH"
}

// periodAgg holds per-period aggregated data with actual date ranges
type periodAgg struct {
	PeriodLabel string    `gorm:"column:period_label"`
	PeriodStart time.Time `gorm:"column:period_start"`
	PeriodEnd   time.Time `gorm:"column:period_end"`
	Spend       float64   `gorm:"column:spend"`
	Revenue     float64   `gorm:"column:revenue"`
	Clicks      int       `gorm:"column:clicks"`
	Orders      int       `gorm:"column:orders"`
	Impressions int       `gorm:"column:impressions"`
}

// buildHistorical converts period aggregations into ProductHistoricalData
// with daily normalization based on actual period_start and period_end dates.
func buildHistorical(productID string, periods []periodAgg) intelligence.ProductHistoricalData {
	result := intelligence.ProductHistoricalData{
		ProductID: productID,
	}

	if len(periods) == 0 {
		return result
	}

	var totalSpend, totalRevenue float64
	var totalClicks, totalOrders, totalImpressions int
	var totalActualDays float64

	for _, p := range periods {
		result.SpendHistory = append(result.SpendHistory, p.Spend)
		result.RevenueHistory = append(result.RevenueHistory, p.Revenue)
		roas := 0.0
		if p.Spend > 0 {
			roas = p.Revenue / p.Spend
		}
		result.RoasHistory = append(result.RoasHistory, roas)

		totalSpend += p.Spend
		totalRevenue += p.Revenue
		totalClicks += p.Clicks
		totalOrders += p.Orders
		totalImpressions += p.Impressions

		// Calculate actual days for this period from timestamps
		periodDays := p.PeriodEnd.Sub(p.PeriodStart).Hours() / 24.0
		if periodDays <= 0 {
			periodDays = 7 // Fallback if timestamps are invalid
		}
		totalActualDays += periodDays
	}

	// Fallback if totalActualDays is still 0
	if totalActualDays <= 0 {
		totalActualDays = float64(len(periods)) * 7
	}

	// Set actual day count (rounded)
	result.DaysOfData = int(math.Round(totalActualDays))

	// CurrentRoas: weighted average across ALL periods
	if totalSpend > 0 {
		result.CurrentRoas = totalRevenue / totalSpend
	}

	// CurrentSpend: truly DAILY average (total spend / actual calendar days)
	result.CurrentSpend = totalSpend / totalActualDays

	// Funnel metrics
	result.TotalClicks = totalClicks
	result.TotalOrders = totalOrders
	result.TotalImpressions = totalImpressions

	return result
}

// getProductHistoricalData retrieves historical data for a product
func (h *SimulationHandler) getProductHistoricalData(
	ctx context.Context,
	db *gorm.DB,
	tenantID, productID string,
) intelligence.ProductHistoricalData {
	// Try TikTok ads first
	tiktokData := h.getTiktokProductData(ctx, db, tenantID, productID)
	if len(tiktokData.RoasHistory) >= 3 {
		return tiktokData
	}

	// Try Shopee ads
	shopeeData := h.getShopeeProductData(ctx, db, tenantID, productID)
	if len(shopeeData.RoasHistory) >= 3 {
		return shopeeData
	}

	// Return whichever has more data
	if len(tiktokData.RoasHistory) > len(shopeeData.RoasHistory) {
		return tiktokData
	}
	return shopeeData
}

// getTiktokProductData gets TikTok product historical data per period
func (h *SimulationHandler) getTiktokProductData(
	ctx context.Context,
	db *gorm.DB,
	tenantID, productID string,
) intelligence.ProductHistoricalData {
	var periods []periodAgg

	db.WithContext(ctx).Model(&models.TiktokAdsCreativeData{}).
		Where("tenant_id = ? AND product_id = ?", tenantID, productID).
		Select(`
			period_label,
			MIN(period_start) as period_start,
			MAX(period_end) as period_end,
			COALESCE(SUM(cost), 0) as spend,
			COALESCE(SUM(gross_revenue), 0) as revenue,
			COALESCE(SUM(clicks), 0) as clicks,
			COALESCE(SUM(orders_sku), 0) as orders,
			COALESCE(SUM(impressions), 0) as impressions
		`).
		Group("period_label").
		Order("period_label ASC").
		Scan(&periods)

	return buildHistorical(productID, periods)
}

// getShopeeProductData gets Shopee product historical data per period
func (h *SimulationHandler) getShopeeProductData(
	ctx context.Context,
	db *gorm.DB,
	tenantID, productID string,
) intelligence.ProductHistoricalData {
	var periods []periodAgg

	db.WithContext(ctx).Model(&models.ShopeeAdsProductData{}).
		Where("tenant_id = ? AND product_id = ?", tenantID, productID).
		Select(`
			period_label,
			MIN(period_start) as period_start,
			MAX(period_end) as period_end,
			COALESCE(SUM(cost), 0) as spend,
			COALESCE(SUM(revenue), 0) as revenue,
			COALESCE(SUM(clicks), 0) as clicks,
			COALESCE(SUM(conversions), 0) as orders,
			COALESCE(SUM(impressions), 0) as impressions
		`).
		Group("period_label").
		Order("period_label ASC").
		Scan(&periods)

	return buildHistorical(productID, periods)
}

// GetProductsFromAds handles GET /api/analytics/products/from-ads
// Uses ML score cache for instant response instead of raw table aggregation.
func (h *SimulationHandler) GetProductsFromAds(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()

	// Query ML score cache (fast: ~2K rows vs 500K+ raw rows)
	var cached []analyticsService.MLScoreCache
	db.WithContext(ctx).
		Where("tenant_id = ? AND total_cost > 0", tenantID).
		Order("total_revenue DESC").
		Limit(100).
		Find(&cached)

	products := make([]gin.H, 0, len(cached))
	for _, p := range cached {
		// Extract original product_id (strip platform prefix for simulator)
		rawID := p.OriginalProductID
		if rawID == "" {
			rawID = p.ProductID
		}
		// Strip "tiktok_" or "shopee_" prefix for the simulation endpoint
		simProductID := rawID
		if len(rawID) > 7 && rawID[:7] == "tiktok_" {
			simProductID = rawID[7:]
		} else if len(rawID) > 7 && rawID[:7] == "shopee_" {
			simProductID = rawID[7:]
		}

		// Compute daily spend from MV date range if available
		var dailySpend float64
		if p.TotalCost > 0 {
			dailySpend = h.getDailySpendFromMV(ctx, db, tenantID, simProductID, p.Platform, p.TotalCost)
		}

		products = append(products, gin.H{
			"product_id":          simProductID,
			"product_name":        p.ProductName,
			"total_cost":          p.TotalCost,
			"total_revenue":       p.TotalRevenue,
			"avg_roas":            p.ROAS,
			"source":              p.Platform,
			"current_daily_spend": dailySpend,
			"ml_category":         p.Category,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    products,
		"count":   len(products),
	})
}

// getDailySpendFromMV calculates daily spend using the materialized view date range.
// Falls back to estimating from period count if MV is unavailable.
func (h *SimulationHandler) getDailySpendFromMV(
	ctx context.Context, db *gorm.DB,
	tenantID, productID, platform string, totalCost float64,
) float64 {
	var result struct {
		FirstPeriod time.Time `gorm:"column:first_period"`
		LastPeriod  time.Time `gorm:"column:last_period"`
	}

	mvTable := "mv_ml_product_analysis"
	err := db.WithContext(ctx).
		Table(mvTable).
		Where("tenant_id = ? AND product_id = ?", tenantID, productID).
		Select("first_period, last_period").
		Limit(1).
		Scan(&result).Error

	if err == nil && !result.FirstPeriod.IsZero() && !result.LastPeriod.IsZero() {
		days := result.LastPeriod.Sub(result.FirstPeriod).Hours() / 24.0
		if days > 0 {
			return math.Round(totalCost / days)
		}
	}

	// Fallback: estimate from total cost / 30 days
	return math.Round(totalCost / 30.0)
}

// calcDailySpend computes daily average spend from total cost and date range
func calcDailySpend(totalCost float64, minStart, maxEnd time.Time) float64 {
	days := maxEnd.Sub(minStart).Hours() / 24.0
	if days <= 0 {
		return 0
	}
	return math.Round(totalCost / days)
}

// GetCalendarEvents handles GET /api/analytics/intelligence/calendar
func (h *SimulationHandler) GetCalendarEvents(c *gin.Context) {
	daysAhead := 30

	if days := c.Query("days"); days != "" {
		if parsed, err := strconv.Atoi(days); err == nil && parsed > 0 && parsed <= 90 {
			daysAhead = parsed
		}
	}

	calendar := intelligence.NewIndonesianCalendar()
	events := calendar.GetUpcomingEvents(time.Now(), daysAhead)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"events":     events,
			"days_ahead": daysAhead,
		},
	})
}
