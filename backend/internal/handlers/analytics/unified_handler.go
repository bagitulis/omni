package analytics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	analyticsService "github.com/omni/backend/internal/services/analytics"
	"github.com/omni/backend/internal/services/cache"
	"github.com/rs/zerolog/log"
)

// Cache TTL constants
const (
	CacheTTLAnalytics = 60 * time.Second // 1 minute for analytics
)

// UnifiedHandler handles combined analytics requests
type UnifiedHandler struct {
	basePath     string
	cacheService *analyticsService.CacheService
	appCache     cache.CacheManager
}

// NewUnifiedHandler creates a new unified handler
func NewUnifiedHandler(basePath string, appCache cache.CacheManager) *UnifiedHandler {
	return &UnifiedHandler{
		basePath:     basePath,
		cacheService: analyticsService.NewCacheService(basePath),
		appCache:     appCache,
	}
}

// GetUnifiedSummary handles GET /api/analytics/unified/summary
// Returns combined TikTok + Shopee summary
func (h *UnifiedHandler) GetUnifiedSummary(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	ctx := c.Request.Context()
	cacheKey := "analytics:unified:summary"

	// Try cache first
	if h.appCache != nil {
		if cached, found := h.appCache.Get(tenantID, cacheKey); found {
			log.Debug().
				Str("tenant_id", tenantID).
				Str("cache_key", cacheKey).
				Bool("cache_hit", true).
				Msg("Unified summary cache hit")

			c.JSON(http.StatusOK, cached)
			return
		}
		log.Debug().
			Str("tenant_id", tenantID).
			Str("cache_key", cacheKey).
			Bool("cache_hit", false).
			Msg("Unified summary cache miss")
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	// Check if MVs are populated, auto-refresh if empty
	var mvCount int64
	db.WithContext(ctx).Table("mv_tiktok_ads_summary").
		Where("tenant_id = ?", tenantID).
		Count(&mvCount)

	if mvCount == 0 {
		// Auto-refresh MVs if not populated
		if refreshResults := h.cacheService.RefreshAllMVs(ctx, db, tenantID); refreshResults == nil {
			log.Warn().Str("tenant_id", tenantID).Msg("MV refresh returned nil results")
		}
	}

	// Get TikTok summary from MV
	var tiktokSummary struct {
		TotalCost    float64 `gorm:"column:total_cost"`
		TotalRevenue float64 `gorm:"column:total_revenue"`
		TotalOrders  int64   `gorm:"column:total_orders"`
		OverallRoas  float64 `gorm:"column:overall_roas"`
	}

	if err := db.WithContext(ctx).Table("mv_tiktok_ads_summary").
		Where("tenant_id = ?", tenantID).
		First(&tiktokSummary).Error; err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID).Msg("Failed to query TikTok ads summary MV")
		// Continue with zero values — MV may not exist yet
	}

	// Get Shopee summary from MV
	var shopeeSummary struct {
		TotalCost    float64 `gorm:"column:total_cost"`
		TotalRevenue float64 `gorm:"column:total_revenue"`
		TotalOrders  int64   `gorm:"column:total_conversions"`
		OverallRoas  float64 `gorm:"column:overall_roas"`
	}

	if err := db.WithContext(ctx).Table("mv_shopee_ads_summary").
		Where("tenant_id = ?", tenantID).
		First(&shopeeSummary).Error; err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID).Msg("Failed to query Shopee ads summary MV")
		// Continue with zero values — MV may not exist yet
	}

	// Calculate combined metrics
	totalCost := tiktokSummary.TotalCost + shopeeSummary.TotalCost
	totalRevenue := tiktokSummary.TotalRevenue + shopeeSummary.TotalRevenue
	totalOrders := tiktokSummary.TotalOrders + shopeeSummary.TotalOrders

	combinedRoas := float64(0)
	if totalCost > 0 {
		combinedRoas = totalRevenue / totalCost
	}

	result := gin.H{
		"success": true,
		"data": gin.H{
			"combined": gin.H{
				"total_cost":    totalCost,
				"total_revenue": totalRevenue,
				"total_orders":  totalOrders,
				"avg_roas":      combinedRoas,
			},
			"tiktok": gin.H{
				"total_cost":    tiktokSummary.TotalCost,
				"total_revenue": tiktokSummary.TotalRevenue,
				"total_orders":  tiktokSummary.TotalOrders,
				"avg_roas":      tiktokSummary.OverallRoas,
			},
			"shopee": gin.H{
				"total_cost":    shopeeSummary.TotalCost,
				"total_revenue": shopeeSummary.TotalRevenue,
				"total_orders":  shopeeSummary.TotalOrders,
				"avg_roas":      shopeeSummary.OverallRoas,
			},
		},
	}

	// Cache the result (graceful - don't fail if cache fails)
	if h.appCache != nil {
		if err := h.appCache.Set(tenantID, cacheKey, result, CacheTTLAnalytics); err != nil {
			log.Warn().
				Err(err).
				Str("tenant_id", tenantID).
				Str("cache_key", cacheKey).
				Msg("Failed to cache unified summary")
		}
	}

	c.JSON(http.StatusOK, result)
}

// GetUnifiedKPI handles GET /api/analytics/unified/kpi
// Returns KPI cards data with action counts calculated from ROAS thresholds
func (h *UnifiedHandler) GetUnifiedKPI(c *gin.Context) {
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

	// Check if MVs are populated, auto-refresh if empty
	var mvCount int64
	db.WithContext(ctx).Table("mv_ml_portfolio_summary").
		Where("tenant_id = ?", tenantID).
		Count(&mvCount)

	if mvCount == 0 {
		// Auto-refresh MVs if not populated
		if refreshResults := h.cacheService.RefreshAllMVs(ctx, db, tenantID); refreshResults == nil {
			log.Warn().Str("tenant_id", tenantID).Msg("KPI MV refresh returned nil results")
		}
	}

	// Get basic portfolio metrics
	var portfolioHealth struct {
		TotalProducts int64   `gorm:"column:total_products"`
		OverallRoas   float64 `gorm:"column:overall_roas"`
	}

	if err := db.WithContext(ctx).Table("mv_ml_portfolio_summary").
		Where("tenant_id = ?", tenantID).
		First(&portfolioHealth).Error; err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID).Msg("Failed to query portfolio summary MV")
	}

	// Calculate action counts from product data based on ROAS thresholds
	var actionCounts struct {
		ScaleUpCount  int64 `gorm:"column:scale_up_count"`
		MaintainCount int64 `gorm:"column:maintain_count"`
		ReduceCount   int64 `gorm:"column:reduce_count"`
		StopCount     int64 `gorm:"column:stop_count"`
	}

	db.WithContext(ctx).Table("mv_ml_product_analysis").
		Where("tenant_id = ? AND total_cost > 0", tenantID).
		Select(`
			COUNT(CASE WHEN roas >= 5 THEN 1 END) as scale_up_count,
			COUNT(CASE WHEN roas >= 2 AND roas < 5 THEN 1 END) as maintain_count,
			COUNT(CASE WHEN roas >= 1 AND roas < 2 THEN 1 END) as reduce_count,
			COUNT(CASE WHEN roas < 1 THEN 1 END) as stop_count
		`).
		Scan(&actionCounts)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_products": portfolioHealth.TotalProducts,
			"avg_roas":       portfolioHealth.OverallRoas,
			"actions": gin.H{
				"scale_up": actionCounts.ScaleUpCount,
				"maintain": actionCounts.MaintainCount,
				"reduce":   actionCounts.ReduceCount,
				"stop":     actionCounts.StopCount,
			},
		},
	})
}

// RefreshCache handles POST /api/analytics/cache/refresh
func (h *UnifiedHandler) RefreshCache(c *gin.Context) {
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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
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

// Ensure models are used
var _ = models.TiktokAdsCreativeData{}
