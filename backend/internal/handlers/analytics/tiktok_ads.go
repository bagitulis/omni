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

// allowedTiktokAdsOrderColumns defines valid orderBy columns to prevent SQL injection
var allowedTiktokAdsOrderColumns = map[string]bool{
	"id": true, "campaign_id": true, "campaign_name": true, "product_id": true,
	"creative_type": true, "video_title": true, "cost": true, "orders_sku": true,
	"gross_revenue": true, "roi": true, "impressions": true, "clicks": true,
	"ctr": true, "conversion_rate": true, "period_start": true, "period_end": true,
}

// TiktokAdsHandler handles TikTok Ads analytics requests
type TiktokAdsHandler struct {
	basePath string
}

// NewTiktokAdsHandler creates a new TikTok ads handler
func NewTiktokAdsHandler(basePath string) *TiktokAdsHandler {
	return &TiktokAdsHandler{basePath: basePath}
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
	limit, offset := 100, 0
	orderBy, orderDir := "gross_revenue", "desc"

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
		switch ob {
		case "revenue":
			orderBy = "gross_revenue"
		case "roi":
			orderBy = "roi"
		default:
			if allowedTiktokAdsOrderColumns[ob] {
				orderBy = ob
			}
		}
	}
	if od := c.Query("orderDir"); od == "asc" || od == "desc" {
		orderDir = od
	}

	// Build query with tenant filter
	query := db.WithContext(ctx).Model(&models.TiktokAdsCreativeData{}).
		Where("tenant_id = ?", tenantID)

	// Apply filters
	if productId := c.Query("product_id"); productId != "" {
		query = query.Where("product_id = ?", productId)
	}
	if creativeType := c.Query("creative_type"); creativeType != "" {
		query = query.Where("creative_type = ?", creativeType)
	}

	// Count total
	var total int64
	if result := query.Count(&total); result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to count records"))
		return
	}

	// Get data with pagination
	var creatives []models.TiktokAdsCreativeData
	if result := query.Order(orderBy + " " + orderDir).
		Offset(offset).
		Limit(limit).
		Find(&creatives); result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to query records"))
		return
	}

	// Transform to response format
	data := make([]gin.H, 0, len(creatives))
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
	db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&uploads)

	data := make([]gin.H, 0, len(uploads))
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

	// TODO: Implement full upload logic with CSV parsing and MV refresh
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error":   "TikTok ads upload is not yet implemented",
	})
}
