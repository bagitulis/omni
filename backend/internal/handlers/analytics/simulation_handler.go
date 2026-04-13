package analytics

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/analytics/intelligence"
	"gorm.io/gorm"
)

// SimulationHandler handles budget simulation requests
type SimulationHandler struct {
	basePath  string
	simulator *intelligence.BudgetSimulator
}

// NewSimulationHandler creates a new simulation handler
func NewSimulationHandler(basePath string) *SimulationHandler {
	return &SimulationHandler{
		basePath:  basePath,
		simulator: intelligence.NewBudgetSimulator(),
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

	// Default period
	if req.PeriodDays == 0 {
		req.PeriodDays = 7
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()

	// Get product historical data from ads database
	historicalData := h.getProductHistoricalData(ctx, db, tenantID, req.ProductID)

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

// periodAgg holds per-period aggregated spend and revenue
type periodAgg struct {
	PeriodLabel string  `gorm:"column:period_label"`
	Spend       float64 `gorm:"column:spend"`
	Revenue     float64 `gorm:"column:revenue"`
}

// buildHistorical converts period aggregations into ProductHistoricalData
func buildHistorical(productID string, periods []periodAgg) intelligence.ProductHistoricalData {
	result := intelligence.ProductHistoricalData{
		ProductID:  productID,
		DaysOfData: len(periods),
	}

	var totalSpend, totalRevenue float64
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
	}

	// Use weighted average ROAS (total_revenue / total_spend) across ALL periods
	// This matches the avg_roas displayed in the product selector dropdown
	if totalSpend > 0 {
		result.CurrentRoas = totalRevenue / totalSpend
	}

	// CurrentSpend = average spend per period (more realistic for daily budget comparison)
	if n := len(periods); n > 0 {
		result.CurrentSpend = totalSpend / float64(n)
	}

	return result
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
			COALESCE(SUM(cost), 0) as spend,
			COALESCE(SUM(gross_revenue), 0) as revenue
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
			COALESCE(SUM(cost), 0) as spend,
			COALESCE(SUM(revenue), 0) as revenue
		`).
		Group("period_label").
		Order("period_label ASC").
		Scan(&periods)

	return buildHistorical(productID, periods)
}

// GetProductsFromAds handles GET /api/analytics/products/from-ads
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

	// Get products from TikTok ads
	var tiktokProducts []struct {
		ProductID   string  `gorm:"column:product_id"`
		ProductName string  `gorm:"column:product_name"`
		TotalCost   float64 `gorm:"column:total_cost"`
		TotalRev    float64 `gorm:"column:total_revenue"`
		AvgRoas     float64 `gorm:"column:avg_roas"`
		Source      string  `gorm:"column:source"`
	}

	db.WithContext(ctx).Model(&models.TiktokAdsCreativeData{}).
		Where("tenant_id = ? AND product_id != '' AND product_id != '-1'", tenantID).
		Select(`
			product_id,
			MAX(product_name) as product_name,
			COALESCE(SUM(cost), 0) as total_cost,
			COALESCE(SUM(gross_revenue), 0) as total_revenue,
			CASE WHEN SUM(cost) > 0 THEN SUM(gross_revenue) / SUM(cost) ELSE 0 END as avg_roas,
			'tiktok' as source
		`).
		Group("product_id").
		Having("SUM(cost) > 0").
		Order("total_revenue DESC").
		Limit(100).
		Find(&tiktokProducts)

	// Get products from Shopee ads
	var shopeeProducts []struct {
		ProductID   string  `gorm:"column:product_id"`
		ProductName string  `gorm:"column:product_name"`
		TotalCost   float64 `gorm:"column:total_cost"`
		TotalRev    float64 `gorm:"column:total_revenue"`
		AvgRoas     float64 `gorm:"column:avg_roas"`
		Source      string  `gorm:"column:source"`
	}

	db.WithContext(ctx).Model(&models.ShopeeAdsProductData{}).
		Where("tenant_id = ?", tenantID).
		Select(`
			product_id,
			MAX(product_name) as product_name,
			COALESCE(SUM(cost), 0) as total_cost,
			COALESCE(SUM(revenue), 0) as total_revenue,
			CASE WHEN SUM(cost) > 0 THEN SUM(revenue) / SUM(cost) ELSE 0 END as avg_roas,
			'shopee' as source
		`).
		Group("product_id").
		Having("SUM(cost) > 0").
		Order("total_revenue DESC").
		Limit(100).
		Find(&shopeeProducts)

	// Combine and format results
	products := make([]gin.H, 0, len(tiktokProducts)+len(shopeeProducts))

	for _, p := range tiktokProducts {
		products = append(products, gin.H{
			"product_id":    p.ProductID,
			"product_name":  p.ProductName,
			"total_cost":    p.TotalCost,
			"total_revenue": p.TotalRev,
			"avg_roas":      p.AvgRoas,
			"source":        "tiktok",
		})
	}

	for _, p := range shopeeProducts {
		products = append(products, gin.H{
			"product_id":    p.ProductID,
			"product_name":  p.ProductName,
			"total_cost":    p.TotalCost,
			"total_revenue": p.TotalRev,
			"avg_roas":      p.AvgRoas,
			"source":        "shopee",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    products,
		"count":   len(products),
	})
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

// Helper functions - removed unused parseInt and now functions
