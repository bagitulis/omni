package analytics

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
)

// allowedShopeeAdsOrderColumns defines valid orderBy columns to prevent SQL injection
var allowedShopeeAdsOrderColumns = map[string]bool{
	"id": true, "product_id": true, "product_name": true, "bidding_mode": true,
	"cost": true, "revenue": true, "direct_revenue": true, "conversions": true,
	"roas": true, "direct_roas": true, "impressions": true, "clicks": true,
	"ctr": true, "conversion_rate": true, "period_start": true, "period_end": true,
	"period_label": true,
}

// AdsHandler handles Shopee Ads analytics requests
type AdsHandler struct {
	basePath string
}

// NewAdsHandler creates a new ads handler
func NewAdsHandler(basePath string) *AdsHandler {
	return &AdsHandler{basePath: basePath}
}

// GetData handles GET /api/analytics/shopee-ads/data
// Returns paginated product data matching frontend ProductData interface
func (h *AdsHandler) GetData(c *gin.Context) {
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

	// Parse query params
	limit, offset := 100, 0
	orderBy, orderDir := "revenue", "desc"

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	if ob := c.Query("orderBy"); ob != "" {
		if allowedShopeeAdsOrderColumns[ob] {
			orderBy = ob
		}
	}
	if od := c.Query("orderDir"); od == "asc" || od == "desc" {
		orderDir = od
	}

	// Build query with tenant filter
	query := db.WithContext(ctx).Model(&models.ShopeeAdsProductData{}).
		Where("tenant_id = ?", tenantID)

	// Apply filters
	if productId := c.Query("product_id"); productId != "" {
		query = query.Where("product_id = ?", productId)
	}
	if biddingMode := c.Query("bidding_mode"); biddingMode != "" {
		query = query.Where("bidding_mode = ?", biddingMode)
	}

	// Count total
	var total int64
	if result := query.Count(&total); result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to count records"))
		return
	}

	// Get data with pagination
	var products []models.ShopeeAdsProductData
	if result := query.Order(orderBy + " " + orderDir).
		Offset(offset).
		Limit(limit).
		Find(&products); result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to query records"))
		return
	}

	// Transform to response format
	data := make([]gin.H, 0, len(products))
	for _, p := range products {
		data = append(data, gin.H{
			"id":              p.ID,
			"product_id":      p.ProductID,
			"product_name":    p.ProductName,
			"bidding_mode":    p.BiddingMode,
			"cost":            p.Cost,
			"revenue":         p.Revenue,
			"direct_revenue":  p.DirectRevenue,
			"conversions":     p.Conversions,
			"roas":            p.ROAS,
			"direct_roas":     p.DirectROAS,
			"impressions":     p.Impressions,
			"clicks":          p.Clicks,
			"ctr":             p.CTR,
			"conversion_rate": p.ConversionRate,
			"period_start":    p.PeriodStart,
			"period_end":      p.PeriodEnd,
			"period_label":    p.PeriodLabel,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetUploads handles GET /api/analytics/shopee-ads/uploads
func (h *AdsHandler) GetUploads(c *gin.Context) {
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
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var uploads []models.ShopeeAdsUploadBatch
	db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&uploads)

	data := make([]gin.H, 0, len(uploads))
	for _, u := range uploads {
		data = append(data, gin.H{
			"id":           u.ID,
			"filename":     u.FileName,
			"period_label": u.PeriodLabel,
			"record_count": u.TotalRows,
			"uploaded_at":  u.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}
