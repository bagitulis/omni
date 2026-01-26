package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
)

// AdsHandler handles Shopee Ads analytics requests
type AdsHandler struct {
	basePath string
}

// NewAdsHandler creates a new ads handler
func NewAdsHandler(basePath string) *AdsHandler {
	return &AdsHandler{basePath: basePath}
}

// GetDashboard handles GET /api/analytics/shopee-ads/dashboard
// Response format matches frontend useShopeeAdsAnalytics.ts DashboardSummary
func (h *AdsHandler) GetDashboard(c *gin.Context) {
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

	// Query aggregated data from shopee_ads_product_data
	var summary struct {
		TotalCost          float64 `gorm:"column:total_cost"`
		TotalRevenue       float64 `gorm:"column:total_revenue"`
		TotalDirectRevenue float64 `gorm:"column:total_direct_revenue"`
		TotalOrders        int64   `gorm:"column:total_orders"`
		TotalImpressions   int64   `gorm:"column:total_impressions"`
		TotalClicks        int64   `gorm:"column:total_clicks"`
	}

	db.WithContext(ctx).Model(&models.ShopeeAdsProductData{}).
		Select(`
			COALESCE(SUM(cost), 0) as total_cost,
			COALESCE(SUM(revenue), 0) as total_revenue,
			COALESCE(SUM(direct_revenue), 0) as total_direct_revenue,
			COALESCE(SUM(conversions), 0) as total_orders,
			COALESCE(SUM(impressions), 0) as total_impressions,
			COALESCE(SUM(clicks), 0) as total_clicks
		`).Scan(&summary)

	// Calculate derived metrics
	avgRoas := float64(0)
	avgDirectRoas := float64(0)
	avgCtr := float64(0)
	avgConversionRate := float64(0)

	if summary.TotalCost > 0 {
		avgRoas = summary.TotalRevenue / summary.TotalCost
		avgDirectRoas = summary.TotalDirectRevenue / summary.TotalCost
	}
	if summary.TotalImpressions > 0 {
		avgCtr = float64(summary.TotalClicks) / float64(summary.TotalImpressions) * 100
	}
	if summary.TotalClicks > 0 {
		avgConversionRate = float64(summary.TotalOrders) / float64(summary.TotalClicks) * 100
	}

	// Get top products by revenue
	var topProducts []gin.H
	var products []models.ShopeeAdsProductData
	db.WithContext(ctx).Model(&models.ShopeeAdsProductData{}).
		Select("product_id, product_name, COALESCE(SUM(cost), 0) as cost, COALESCE(SUM(revenue), 0) as revenue, COALESCE(SUM(conversions), 0) as orders").
		Group("product_id, product_name").
		Order("revenue DESC").
		Limit(10).
		Find(&products)

	for _, p := range products {
		roas := float64(0)
		if p.Cost > 0 {
			roas = p.Revenue / p.Cost
		}
		topProducts = append(topProducts, gin.H{
			"product_id":   p.ProductID,
			"product_name": p.ProductName,
			"cost":         p.Cost,
			"revenue":      p.Revenue,
			"orders":       p.Conversions,
			"roas":         roas,
		})
	}

	// Get bidding mode comparison
	var biddingModeStats []gin.H
	var biddingModes []struct {
		BiddingMode  string  `gorm:"column:bidding_mode"`
		Cost         float64 `gorm:"column:cost"`
		Revenue      float64 `gorm:"column:revenue"`
		Orders       int     `gorm:"column:orders"`
		ProductCount int     `gorm:"column:product_count"`
	}
	db.WithContext(ctx).Model(&models.ShopeeAdsProductData{}).
		Select("bidding_mode, COALESCE(SUM(cost), 0) as cost, COALESCE(SUM(revenue), 0) as revenue, COALESCE(SUM(conversions), 0) as orders, COUNT(DISTINCT product_id) as product_count").
		Group("bidding_mode").
		Find(&biddingModes)

	for _, bm := range biddingModes {
		roas := float64(0)
		costPerOrder := float64(0)
		if bm.Cost > 0 {
			roas = bm.Revenue / bm.Cost
		}
		if bm.Orders > 0 {
			costPerOrder = bm.Cost / float64(bm.Orders)
		}
		biddingModeStats = append(biddingModeStats, gin.H{
			"bidding_mode":   bm.BiddingMode,
			"cost":           bm.Cost,
			"revenue":        bm.Revenue,
			"orders":         bm.Orders,
			"roas":           roas,
			"cost_per_order": costPerOrder,
			"product_count":  bm.ProductCount,
		})
	}

	// Response format matching frontend DashboardSummary interface
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_cost":              summary.TotalCost,
			"total_revenue":           summary.TotalRevenue,
			"total_direct_revenue":    summary.TotalDirectRevenue,
			"total_orders":            summary.TotalOrders,
			"avg_roas":                avgRoas,
			"avg_direct_roas":         avgDirectRoas,
			"total_impressions":       summary.TotalImpressions,
			"total_clicks":            summary.TotalClicks,
			"avg_ctr":                 avgCtr,
			"avg_conversion_rate":     avgConversionRate,
			"top_products":            topProducts,
			"bidding_mode_comparison": biddingModeStats,
		},
	})
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
	limit := 100
	offset := 0
	orderBy := "revenue"
	orderDir := "desc"

	if l := c.Query("limit"); l != "" {
		if parsed, err := parseInt(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := parseInt(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	if ob := c.Query("orderBy"); ob != "" {
		orderBy = ob
	}
	if od := c.Query("orderDir"); od != "" {
		orderDir = od
	}

	// Build query
	query := db.WithContext(ctx).Model(&models.ShopeeAdsProductData{})

	// Apply filters
	if productId := c.Query("productId"); productId != "" {
		query = query.Where("product_id = ?", productId)
	}
	if biddingMode := c.Query("biddingMode"); biddingMode != "" {
		query = query.Where("bidding_mode = ?", biddingMode)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Get data with pagination
	var products []models.ShopeeAdsProductData
	query.Order(orderBy + " " + orderDir).
		Offset(offset).
		Limit(limit).
		Find(&products)

	// Transform to response format
	var data []gin.H
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
		if parsed, err := parseInt(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	var uploads []models.ShopeeAdsUploadBatch
	db.WithContext(ctx).Order("uploaded_at DESC").Limit(limit).Find(&uploads)

	var data []gin.H
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

// Upload handles POST /api/analytics/shopee-ads/upload
func (h *AdsHandler) Upload(c *gin.Context) {
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

// parseInt helper function
func parseInt(s string) (int, error) {
	var result int
	_, err := parseIntScan(s, &result)
	return result, err
}

func parseIntScan(s string, result *int) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	*result = n
	return n, nil
}
