package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	analyticsService "github.com/omni/backend/internal/services/analytics"
)

// UnifiedHandler handles combined analytics requests
type UnifiedHandler struct {
	basePath     string
	cacheService *analyticsService.CacheService
}

// NewUnifiedHandler creates a new unified handler
func NewUnifiedHandler(basePath string) *UnifiedHandler {
	return &UnifiedHandler{
		basePath:     basePath,
		cacheService: analyticsService.NewCacheService(basePath),
	}
}

// GetUnifiedSummary handles GET /api/analytics/unified/summary
// Returns combined TikTok + Shopee summary
func (h *UnifiedHandler) GetUnifiedSummary(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()

	// Get TikTok summary from MV
	var tiktokSummary struct {
		TotalCost    float64 `gorm:"column:total_cost"`
		TotalRevenue float64 `gorm:"column:total_revenue"`
		TotalOrders  int64   `gorm:"column:total_orders"`
		OverallRoas  float64 `gorm:"column:overall_roas"`
	}

	db.WithContext(ctx).Table("mv_tiktok_ads_summary").
		Where("tenant_id = ?", tenantID).
		First(&tiktokSummary)

	// Get Shopee summary from MV
	var shopeeSummary struct {
		TotalCost    float64 `gorm:"column:total_cost"`
		TotalRevenue float64 `gorm:"column:total_revenue"`
		TotalOrders  int64   `gorm:"column:total_conversions"`
		OverallRoas  float64 `gorm:"column:overall_roas"`
	}

	db.WithContext(ctx).Table("mv_shopee_ads_summary").
		Where("tenant_id = ?", tenantID).
		First(&shopeeSummary)

	// Calculate combined metrics
	totalCost := tiktokSummary.TotalCost + shopeeSummary.TotalCost
	totalRevenue := tiktokSummary.TotalRevenue + shopeeSummary.TotalRevenue
	totalOrders := tiktokSummary.TotalOrders + shopeeSummary.TotalOrders

	combinedRoas := float64(0)
	if totalCost > 0 {
		combinedRoas = totalRevenue / totalCost
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"combined": gin.H{
				"total_cost":    totalCost,
				"total_revenue": totalRevenue,
				"total_orders":  totalOrders,
				"overall_roas":  combinedRoas,
			},
			"tiktok": gin.H{
				"total_cost":    tiktokSummary.TotalCost,
				"total_revenue": tiktokSummary.TotalRevenue,
				"total_orders":  tiktokSummary.TotalOrders,
				"overall_roas":  tiktokSummary.OverallRoas,
			},
			"shopee": gin.H{
				"total_cost":    shopeeSummary.TotalCost,
				"total_revenue": shopeeSummary.TotalRevenue,
				"total_orders":  shopeeSummary.TotalOrders,
				"overall_roas":  shopeeSummary.OverallRoas,
			},
		},
	})
}

// GetUnifiedKPI handles GET /api/analytics/unified/kpi
// Returns KPI cards data
func (h *UnifiedHandler) GetUnifiedKPI(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()

	// Get portfolio health from MV
	var portfolioHealth struct {
		TotalProducts int64   `gorm:"column:total_products"`
		AvgRoas       float64 `gorm:"column:avg_roas"`
		ScaleUpCount  int64   `gorm:"column:scale_up_count"`
		MaintainCount int64   `gorm:"column:maintain_count"`
		ReduceCount   int64   `gorm:"column:reduce_count"`
		StopCount     int64   `gorm:"column:stop_count"`
	}

	db.WithContext(ctx).Table("mv_ml_portfolio_summary").
		Where("tenant_id = ?", tenantID).
		First(&portfolioHealth)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_products": portfolioHealth.TotalProducts,
			"avg_roas":       portfolioHealth.AvgRoas,
			"actions": gin.H{
				"scale_up": portfolioHealth.ScaleUpCount,
				"maintain": portfolioHealth.MaintainCount,
				"reduce":   portfolioHealth.ReduceCount,
				"stop":     portfolioHealth.StopCount,
			},
		},
	})
}

// GetClassifiedProducts handles GET /api/analytics/products/classified
// Returns products grouped by recommended action
func (h *UnifiedHandler) GetClassifiedProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()

	// Get products from MV with scores
	var products []struct {
		ProductID      string  `gorm:"column:product_id"`
		ProductName    string  `gorm:"column:product_name"`
		TotalCost      float64 `gorm:"column:total_cost"`
		TotalRevenue   float64 `gorm:"column:total_revenue"`
		Roas           float64 `gorm:"column:roas"`
		CompositeScore float64 `gorm:"column:composite_score"`
		Action         string  `gorm:"column:action"`
	}

	db.WithContext(ctx).Table("mv_ml_product_analysis").
		Where("tenant_id = ?", tenantID).
		Order("composite_score DESC").
		Find(&products)

	// Group by action
	scaleUp := make([]gin.H, 0)
	maintain := make([]gin.H, 0)
	reduce := make([]gin.H, 0)
	stop := make([]gin.H, 0)

	for _, p := range products {
		item := gin.H{
			"product_id":      p.ProductID,
			"product_name":    p.ProductName,
			"total_cost":      p.TotalCost,
			"total_revenue":   p.TotalRevenue,
			"roas":            p.Roas,
			"composite_score": p.CompositeScore,
		}

		switch p.Action {
		case "SCALE_UP_AGGRESSIVE", "SCALE_UP_MODERATE":
			scaleUp = append(scaleUp, item)
		case "MAINTAIN":
			maintain = append(maintain, item)
		case "REDUCE_BUDGET":
			reduce = append(reduce, item)
		case "STOP_IMMEDIATELY":
			stop = append(stop, item)
		default:
			maintain = append(maintain, item)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"scale_up": scaleUp,
			"maintain": maintain,
			"reduce":   reduce,
			"stop":     stop,
		},
		"counts": gin.H{
			"scale_up": len(scaleUp),
			"maintain": len(maintain),
			"reduce":   len(reduce),
			"stop":     len(stop),
		},
	})
}

// RefreshCache handles POST /api/analytics/cache/refresh
func (h *UnifiedHandler) RefreshCache(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()

	// Refresh all MVs
	results := h.cacheService.RefreshAllMVs(ctx, db, tenantID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}

// GetCacheStatus handles GET /api/analytics/cache/status
func (h *UnifiedHandler) GetCacheStatus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()

	// Get cache status
	status := h.cacheService.GetCacheStatus(ctx, db, tenantID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// GetTopProducts handles GET /api/analytics/products/top
func (h *UnifiedHandler) GetTopProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()
	limit := 10

	// Get top products by revenue from both platforms
	var tiktokTop []struct {
		ProductID   string  `gorm:"column:product_id"`
		ProductName string  `gorm:"column:product_name"`
		Revenue     float64 `gorm:"column:total_revenue"`
		Cost        float64 `gorm:"column:total_cost"`
		Roas        float64 `gorm:"column:roas"`
	}

	db.WithContext(ctx).Table("mv_ml_product_analysis").
		Where("tenant_id = ?", tenantID).
		Order("total_revenue DESC").
		Limit(limit).
		Find(&tiktokTop)

	// Format response
	products := make([]gin.H, 0, len(tiktokTop))
	for _, p := range tiktokTop {
		products = append(products, gin.H{
			"product_id":   p.ProductID,
			"product_name": p.ProductName,
			"revenue":      p.Revenue,
			"cost":         p.Cost,
			"roas":         p.Roas,
			"source":       "tiktok",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    products,
	})
}

// Ensure models are used
var _ = models.TiktokAdsCreativeData{}
