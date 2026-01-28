package analytics

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto"
	analyticsService "github.com/omni/backend/internal/services/analytics"
)

// MLHandler handles ML analytics endpoints
type MLHandler struct{}

// NewMLHandler creates a new ML analytics handler
func NewMLHandler() *MLHandler {
	return &MLHandler{}
}

// getService creates ML analytics service from context
func (h *MLHandler) getService(c *gin.Context) (*analyticsService.MLAnalyticsService, error) {
	// Try multiple context key variations
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		tenantID = c.GetString("tenantID")
	}
	if tenantID == "" {
		tenantID = c.GetString("tenantId")
	}

	log.Printf("[MLHandler] tenant_id from context: %s", tenantID)

	if tenantID == "" {
		log.Printf("[MLHandler] ERROR: No tenant_id in context")
		return nil, config.ErrMissingTenantID
	}

	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		log.Printf("[MLHandler] ERROR: Failed to get tenant DB: %v", err)
		return nil, err
	}
	return analyticsService.NewMLAnalyticsService(tenantDB, tenantID), nil
}

// GetPortfolioHealth returns portfolio health summary
// GET /api/analytics/ml/portfolio-health
func (h *MLHandler) GetPortfolioHealth(c *gin.Context) {
	platform := c.DefaultQuery("platform", "tiktok")
	svc, err := h.getService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get database connection",
		})
		return
	}

	health, err := svc.GetPortfolioHealth(c.Request.Context(), platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get portfolio health",
		})
		return
	}

	c.JSON(http.StatusOK, dto.PortfolioHealthResponse{
		Success: true,
		Data:    *health,
	})
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

	svc, err := h.getService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get database connection",
		})
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

	svc, err := h.getService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get database connection",
		})
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

// GetAlerts returns active alerts
// GET /api/analytics/ml/alerts
func (h *MLHandler) GetAlerts(c *gin.Context) {
	svc, err := h.getService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get database connection",
		})
		return
	}

	alerts, err := svc.GetAlerts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get alerts",
		})
		return
	}

	// Count by severity
	highCount, mediumCount, lowCount := 0, 0, 0
	for _, a := range alerts {
		switch a.Severity {
		case "HIGH":
			highCount++
		case "MEDIUM":
			mediumCount++
		case "LOW":
			lowCount++
		}
	}

	c.JSON(http.StatusOK, dto.MLAlertsResponse{
		Success: true,
		Data:    alerts,
		Meta: dto.AlertsMeta{
			TotalActive:    len(alerts),
			HighPriority:   highCount,
			MediumPriority: mediumCount,
			LowPriority:    lowCount,
		},
	})
}

// SimulateBudget simulates budget changes
// POST /api/analytics/ml/budget-sim
func (h *MLHandler) SimulateBudget(c *gin.Context) {
	var req dto.BudgetSimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	if len(req.ProductIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "At least one product ID is required",
		})
		return
	}

	svc, err := h.getService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get database connection",
		})
		return
	}

	result, err := svc.SimulateBudget(c.Request.Context(), req.ProductIDs, req.BudgetChangePct)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to simulate budget",
		})
		return
	}

	c.JSON(http.StatusOK, dto.BudgetSimResponse{
		Success: true,
		Data:    *result,
	})
}

// GetScoreDistribution returns score distribution for charts
// GET /api/analytics/ml/distribution
func (h *MLHandler) GetScoreDistribution(c *gin.Context) {
	svc, err := h.getService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get database connection",
		})
		return
	}

	distribution, err := svc.GetScoreDistribution(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get distribution",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ScoreDistributionResponse{
		Success: true,
		Data:    distribution,
	})
}
