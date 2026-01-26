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

// TiktokAdsHandler handles TikTok Ads analytics requests
type TiktokAdsHandler struct {
	basePath string
}

// NewTiktokAdsHandler creates a new TikTok ads handler
func NewTiktokAdsHandler(basePath string) *TiktokAdsHandler {
	return &TiktokAdsHandler{basePath: basePath}
}

// GetDashboard handles GET /api/analytics/tiktok-ads/dashboard
// Response format matches frontend useTiktokAdsAnalytics.ts DashboardSummary
func (h *TiktokAdsHandler) GetDashboard(c *gin.Context) {
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

	// Query aggregated data from tiktok_ads_creative_data
	var summary struct {
		TotalCost        float64 `gorm:"column:total_cost"`
		TotalRevenue     float64 `gorm:"column:total_revenue"`
		TotalOrders      int64   `gorm:"column:total_orders"`
		TotalImpressions int64   `gorm:"column:total_impressions"`
		TotalClicks      int64   `gorm:"column:total_clicks"`
	}

	db.WithContext(ctx).Model(&models.TiktokAdsCreativeData{}).
		Select(`
			COALESCE(SUM(cost), 0) as total_cost,
			COALESCE(SUM(gross_revenue), 0) as total_revenue,
			COALESCE(SUM(orders_sku), 0) as total_orders,
			COALESCE(SUM(impressions), 0) as total_impressions,
			COALESCE(SUM(clicks), 0) as total_clicks
		`).Scan(&summary)

	// Calculate derived metrics
	avgRoi := float64(0)
	avgCtr := float64(0)
	avgConversionRate := float64(0)

	if summary.TotalCost > 0 {
		avgRoi = summary.TotalRevenue / summary.TotalCost
	}
	if summary.TotalImpressions > 0 {
		avgCtr = float64(summary.TotalClicks) / float64(summary.TotalImpressions) * 100
	}
	if summary.TotalClicks > 0 {
		avgConversionRate = float64(summary.TotalOrders) / float64(summary.TotalClicks) * 100
	}

	// Get top products by revenue
	var topProducts []gin.H
	var products []struct {
		ProductID string  `gorm:"column:product_id"`
		Cost      float64 `gorm:"column:cost"`
		Revenue   float64 `gorm:"column:revenue"`
		Orders    int     `gorm:"column:orders"`
	}
	db.WithContext(ctx).Model(&models.TiktokAdsCreativeData{}).
		Select("product_id, COALESCE(SUM(cost), 0) as cost, COALESCE(SUM(gross_revenue), 0) as revenue, COALESCE(SUM(orders_sku), 0) as orders").
		Group("product_id").
		Order("revenue DESC").
		Limit(10).
		Find(&products)

	for _, p := range products {
		roi := float64(0)
		if p.Cost > 0 {
			roi = p.Revenue / p.Cost
		}
		topProducts = append(topProducts, gin.H{
			"product_id": p.ProductID,
			"cost":       p.Cost,
			"revenue":    p.Revenue,
			"orders":     p.Orders,
			"roi":        roi,
		})
	}

	// Get creative type comparison
	var creativeTypeStats []gin.H
	var creativeTypes []struct {
		CreativeType string  `gorm:"column:creative_type"`
		Cost         float64 `gorm:"column:cost"`
		Revenue      float64 `gorm:"column:revenue"`
		Orders       int     `gorm:"column:orders"`
	}
	db.WithContext(ctx).Model(&models.TiktokAdsCreativeData{}).
		Select("creative_type, COALESCE(SUM(cost), 0) as cost, COALESCE(SUM(gross_revenue), 0) as revenue, COALESCE(SUM(orders_sku), 0) as orders").
		Group("creative_type").
		Find(&creativeTypes)

	for _, ct := range creativeTypes {
		roi := float64(0)
		costPerOrder := float64(0)
		if ct.Cost > 0 {
			roi = ct.Revenue / ct.Cost
		}
		if ct.Orders > 0 {
			costPerOrder = ct.Cost / float64(ct.Orders)
		}
		creativeTypeStats = append(creativeTypeStats, gin.H{
			"creative_type":  ct.CreativeType,
			"cost":           ct.Cost,
			"revenue":        ct.Revenue,
			"orders":         ct.Orders,
			"roi":            roi,
			"cost_per_order": costPerOrder,
		})
	}

	// Response format matching frontend DashboardSummary interface
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_cost":               summary.TotalCost,
			"total_revenue":            summary.TotalRevenue,
			"total_orders":             summary.TotalOrders,
			"avg_roi":                  avgRoi,
			"total_impressions":        summary.TotalImpressions,
			"total_clicks":             summary.TotalClicks,
			"avg_ctr":                  avgCtr,
			"avg_conversion_rate":      avgConversionRate,
			"top_products":             topProducts,
			"creative_type_comparison": creativeTypeStats,
		},
	})
}

// GetData handles GET /api/analytics/tiktok-ads/data
// Returns paginated creative data matching frontend CreativeData interface
func (h *TiktokAdsHandler) GetData(c *gin.Context) {
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
	limit := 100
	offset := 0
	orderBy := "gross_revenue"
	orderDir := "desc"

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
		// Map frontend field names to DB column names
		switch ob {
		case "revenue":
			orderBy = "gross_revenue"
		case "roi":
			orderBy = "roi"
		default:
			orderBy = ob
		}
	}
	if od := c.Query("orderDir"); od != "" {
		orderDir = od
	}

	// Build query
	query := db.WithContext(ctx).Model(&models.TiktokAdsCreativeData{})

	// Apply filters
	if productId := c.Query("productId"); productId != "" {
		query = query.Where("product_id = ?", productId)
	}
	if creativeType := c.Query("creativeType"); creativeType != "" {
		query = query.Where("creative_type = ?", creativeType)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Get data with pagination
	var creatives []models.TiktokAdsCreativeData
	query.Order(orderBy + " " + orderDir).
		Offset(offset).
		Limit(limit).
		Find(&creatives)

	// Transform to response format
	var data []gin.H
	for _, cr := range creatives {
		data = append(data, gin.H{
			"id":              cr.ID,
			"campaign_id":     cr.CampaignID,
			"campaign_name":   cr.CampaignName,
			"product_id":      cr.ProductID,
			"creative_type":   cr.CreativeType,
			"video_title":     cr.VideoTitle,
			"cost":            cr.Cost,
			"orders_sku":      cr.OrdersSKU,
			"gross_revenue":   cr.GrossRevenue,
			"roi":             cr.ROI,
			"impressions":     cr.Impressions,
			"clicks":          cr.Clicks,
			"ctr":             cr.CTR,
			"conversion_rate": cr.ConversionRate,
			"period_start":    cr.PeriodStart,
			"period_end":      cr.PeriodEnd,
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

// GetUploads handles GET /api/analytics/tiktok-ads/uploads
func (h *TiktokAdsHandler) GetUploads(c *gin.Context) {
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

	var uploads []models.TiktokAdsUploadBatch
	db.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&uploads)

	var data []gin.H
	for _, u := range uploads {
		data = append(data, gin.H{
			"id":            u.ID,
			"file_name":     u.FileName,
			"period_start":  u.PeriodStart,
			"period_end":    u.PeriodEnd,
			"total_rows":    u.TotalRows,
			"inserted_rows": u.InsertedRows,
			"skipped_rows":  u.SkippedRows,
			"status":        u.Status,
			"created_at":    u.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// Upload handles POST /api/analytics/tiktok-ads/upload
func (h *TiktokAdsHandler) Upload(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	// TODO: Implement full upload logic
	// For now, return placeholder response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Upload functionality - implementation in progress",
		"data": gin.H{
			"batch_id":      "",
			"file_name":     "",
			"total_rows":    0,
			"inserted_rows": 0,
			"skipped_rows":  0,
			"updated_rows":  0,
			"errors":        []string{},
		},
	})
}
