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
	historicalData := h.getProductHistoricalData(ctx, db, tenantID, req.ProductID)
	h.enrichWithMLData(ctx, db, tenantID, req.ProductID, &historicalData)

	simReq := intelligence.SimulationRequest{
		ProductID: req.ProductID, TargetRoas: req.TargetRoas,
		BudgetPerDay: req.BudgetPerDay, PeriodDays: req.PeriodDays,
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": h.simulator.Simulate(simReq, historicalData)})
}

// enrichWithMLData looks up ML cache to inject category and fatigue info
func (h *SimulationHandler) enrichWithMLData(
	ctx context.Context, db *gorm.DB, tenantID, productID string, data *intelligence.ProductHistoricalData,
) {
	for _, prefix := range []string{"tiktok_", "shopee_"} {
		cached, err := h.cacheService.GetCachedProductByID(ctx, db, tenantID, prefix+productID)
		if err == nil && cached != nil {
			data.MLCategory = cached.Category
			data.FatigueStatus = cached.FatigueStatus
			return
		}
	}
	data.MLCategory = "STABLE"
	data.FatigueStatus = "FRESH"
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
	var cached []analyticsService.MLScoreCache
	db.WithContext(ctx).Where("tenant_id = ? AND total_cost > 0", tenantID).
		Order("total_revenue DESC").Limit(100).Find(&cached)

	products := make([]gin.H, 0, len(cached))
	for _, p := range cached {
		rawID := p.OriginalProductID
		if rawID == "" {
			rawID = p.ProductID
		}
		simProductID := rawID
		if len(rawID) > 7 && rawID[:7] == "tiktok_" {
			simProductID = rawID[7:]
		} else if len(rawID) > 7 && rawID[:7] == "shopee_" {
			simProductID = rawID[7:]
		}

		var dailySpend float64
		if p.TotalCost > 0 {
			dailySpend = math.Round(p.TotalCost / 180.0)
		}

		products = append(products, gin.H{
			"product_id": simProductID, "product_name": p.ProductName,
			"total_cost": p.TotalCost, "total_revenue": p.TotalRevenue,
			"avg_roas": p.ROAS, "source": p.Platform,
			"current_daily_spend": dailySpend, "ml_category": p.Category,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": products, "count": len(products)})
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
		"data":    gin.H{"events": events, "days_ahead": daysAhead},
	})
}
