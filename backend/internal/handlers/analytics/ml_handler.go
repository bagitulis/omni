package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	analyticsService "github.com/omni/backend/internal/services/analytics"
	"github.com/omni/backend/internal/services/cache"
)

// MLHandler handles ML analytics endpoints
type MLHandler struct {
	appCache     cache.CacheManager
	cacheService *analyticsService.MLCacheService
}

// NewMLHandler creates a new ML analytics handler
func NewMLHandler(appCache cache.CacheManager) *MLHandler {
	return &MLHandler{
		appCache:     appCache,
		cacheService: analyticsService.NewMLCacheService(),
	}
}

// getService creates ML analytics service from context.
// Uses middleware.GetTenantID for standardized tenant ID retrieval.
func (h *MLHandler) getService(c *gin.Context) (*analyticsService.MLAnalyticsService, string, error) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		return nil, "", config.ErrMissingTenantID
	}

	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		return nil, tenantID, err
	}
	return analyticsService.NewMLAnalyticsService(tenantDB, tenantID), tenantID, nil
}

// GetPortfolioHealth returns portfolio health summary from cache
// GET /api/analytics/ml/portfolio-health
func (h *MLHandler) GetPortfolioHealth(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant ID"})
		return
	}

	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to connect to tenant database"})
		return
	}

	// Ensure cache tables exist
	h.cacheService.EnsureTables(tenantDB)

	// Read from cache
	health, hasCache := h.cacheService.GetCachedHealth(c.Request.Context(), tenantDB, tenantID)
	if !hasCache {
		// No cached data — frontend should show empty state with Analyze button
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"has_cache": false,
			"data":      nil,
		})
		return
	}

	result := dto.PortfolioHealthResponse{
		Data: *health,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"has_cache": true,
		"data":      result.Data,
	})
}

// GetProducts returns paginated product analyses from cache
// GET /api/analytics/ml/products
func (h *MLHandler) GetProducts(c *gin.Context) {
	var params dto.MLProductsQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		params = dto.DefaultMLProductsQuery()
	}

	if params.Limit <= 0 || params.Limit > 100 {
		params.Limit = 20
	}
	if params.SortBy == "" {
		params.SortBy = "unified_score"
	}
	if params.SortDir == "" {
		params.SortDir = "desc"
	}

	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant ID"})
		return
	}

	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to connect to tenant database"})
		return
	}

	h.cacheService.EnsureTables(tenantDB)

	// Read from cache (fast) instead of real-time computation
	cachedProducts, total, err := h.cacheService.GetCachedProducts(
		c.Request.Context(), tenantDB, tenantID,
		params.SortBy, params.SortDir,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get products",
		})
		return
	}

	// Convert cache rows to API response format
	products := make([]models.MLProductAnalysis, 0, len(cachedProducts))
	for _, cp := range cachedProducts {
		products = append(products, cacheToProductAnalysis(cp))
	}

	c.JSON(http.StatusOK, dto.MLProductsResponse{
		Success: true,
		Data:    products,
		Meta: dto.MLPaginationMeta{
			Total:   total,
			Limit:   int(total),
			HasMore: false,
		},
	})
}

// GetProductDetail returns detailed analysis for a single product from cache
// GET /api/analytics/ml/product/:id
func (h *MLHandler) GetProductDetail(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Product ID is required",
		})
		return
	}

	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant ID"})
		return
	}

	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to connect to tenant database"})
		return
	}

	cached, err := h.cacheService.GetCachedProductByID(c.Request.Context(), tenantDB, tenantID, productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Product not found",
		})
		return
	}

	product := cacheToProductAnalysis(*cached)

	c.JSON(http.StatusOK, dto.MLProductDetailResponse{
		Success: true,
		Data:    product,
	})
}

// cacheToProductAnalysis converts a cache row to MLProductAnalysis for API response
func cacheToProductAnalysis(cp analyticsService.MLScoreCache) models.MLProductAnalysis {
	return models.MLProductAnalysis{
		ProductID:          cp.ProductID,
		ProductName:        cp.ProductName,
		SKU:                cp.SKU,
		Platform:           cp.Platform,
		TotalCost:          cp.TotalCost,
		TotalRevenue:       cp.TotalRevenue,
		TotalProfit:        cp.TotalProfit,
		ROAS:               cp.ROAS,
		CTR:                cp.CTR,
		Clicks:             cp.Clicks,
		Impressions:        cp.Impressions,
		UnifiedScore:       cp.UnifiedScore,
		ROASScore:          cp.ROASScore,
		TrendScore:         cp.TrendScore,
		VolatilityScore:    cp.VolatilityScore,
		MomentumScore:      cp.MomentumScore,
		Category:           cp.Category,
		Action:             cp.Action,
		ActionLabel:        cp.ActionLabel,
		Recommendation:     cp.Recommendation,
		BudgetChangePct:    cp.BudgetChangePct,
		HasFatigueWarning:  cp.HasFatigueWarning,
		HasChurnRisk:       cp.HasChurnRisk,
		FatigueStatus:      cp.FatigueStatus,
		ChurnRiskScore:     cp.ChurnRiskScore,
		ConfidenceLevel:    cp.ConfidenceLevel,
		SuccessProbability: cp.SuccessProbability,
		TrendDirection:     cp.TrendDirection,
		TrendStrength:      cp.TrendStrength,
		LastUpdated:        cp.CalculatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// Recalculate triggers a background ML recalculation
// POST /api/analytics/ml/recalculate
func (h *MLHandler) Recalculate(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant ID"})
		return
	}

	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to connect to tenant database"})
		return
	}

	h.cacheService.EnsureTables(tenantDB)

	status, triggerErr := h.cacheService.TriggerRecalculate(tenantDB, tenantID)
	if triggerErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": triggerErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status":  status,
	})
}

// RecalculateStatus returns the current recalculation status
// GET /api/analytics/ml/recalculate/status
func (h *MLHandler) RecalculateStatus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant ID"})
		return
	}

	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to connect to tenant database"})
		return
	}

	status := h.cacheService.GetStatus(c.Request.Context(), tenantDB, tenantID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}
