package analytics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/middleware"
	analyticsService "github.com/omni/backend/internal/services/analytics"
	"github.com/omni/backend/internal/services/cache"
	zlog "github.com/rs/zerolog/log"
)

// MLHandler handles ML analytics endpoints
type MLHandler struct {
	appCache cache.CacheManager
}

// NewMLHandler creates a new ML analytics handler
func NewMLHandler(appCache cache.CacheManager) *MLHandler {
	return &MLHandler{
		appCache: appCache,
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
		zlog.Error().Err(err).Str("tenant_id", tenantID).Msg("Failed to get tenant DB for ML analytics")
		return nil, tenantID, err
	}
	return analyticsService.NewMLAnalyticsService(tenantDB, tenantID), tenantID, nil
}

// GetPortfolioHealth returns portfolio health summary
// GET /api/analytics/ml/portfolio-health
func (h *MLHandler) GetPortfolioHealth(c *gin.Context) {
	platform := c.DefaultQuery("platform", "tiktok")

	// Use standardized tenant ID retrieval
	tenantID := middleware.GetTenantID(c)

	cacheKey := "analytics:ml:portfolio-health:" + platform

	// Try cache first
	if h.appCache != nil && tenantID != "" {
		if cached, found := h.appCache.Get(tenantID, cacheKey); found {
			zlog.Debug().
				Str("tenant_id", tenantID).
				Str("cache_key", cacheKey).
				Bool("cache_hit", true).
				Msg("ML portfolio health cache hit")

			c.JSON(http.StatusOK, cached)
			return
		}
		zlog.Debug().
			Str("tenant_id", tenantID).
			Str("cache_key", cacheKey).
			Bool("cache_hit", false).
			Msg("ML portfolio health cache miss")
	}

	svc, svcTenantID, err := h.getService(c)
	if err != nil {
		if err == config.ErrMissingTenantID {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant ID"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to connect to tenant database"})
		}
		return
	}
	if tenantID == "" {
		tenantID = svcTenantID
	}

	health, err := svc.GetPortfolioHealth(c.Request.Context(), platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get portfolio health",
		})
		return
	}

	result := dto.PortfolioHealthResponse{
		Success: true,
		Data:    *health,
	}

	// Cache the result (60 seconds)
	if h.appCache != nil && tenantID != "" {
		if err := h.appCache.Set(tenantID, cacheKey, result, 60*time.Second); err != nil {
			zlog.Warn().
				Err(err).
				Str("tenant_id", tenantID).
				Str("cache_key", cacheKey).
				Msg("Failed to cache ML portfolio health")
		}
	}

	c.JSON(http.StatusOK, result)
}

// GetProducts returns paginated product analyses
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

	svc, _, err := h.getService(c)
	if err != nil {
		if err == config.ErrMissingTenantID {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant ID"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to connect to tenant database"})
		}
		return
	}

	products, total, hasMore, nextCursor, err := svc.GetProducts(
		c.Request.Context(),
		params.Platform,
		params.Limit,
		params.Cursor,
		params.SortBy,
		params.SortDir,
		params.Category,
		params.Action,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get products",
		})
		return
	}

	c.JSON(http.StatusOK, dto.MLProductsResponse{
		Success: true,
		Data:    products,
		Meta: dto.MLPaginationMeta{
			Total:      total,
			Limit:      params.Limit,
			HasMore:    hasMore,
			NextCursor: nextCursor,
		},
	})
}

// GetProductDetail returns detailed analysis for a single product
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

	svc, _, err := h.getService(c)
	if err != nil {
		if err == config.ErrMissingTenantID {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant ID"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to connect to tenant database"})
		}
		return
	}

	product, err := svc.GetProductDetail(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get product detail",
		})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Product not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.MLProductDetailResponse{
		Success: true,
		Data:    *product,
	})
}
