package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services"
)

// ReportHandler handles report generation endpoints
type ReportHandler struct {
	reportService *services.ReportService
}

// NewReportHandler creates a new report handler
func NewReportHandler(reportService *services.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// GetShopeeAdsReport handles GET /api/reports/shopee/ads
func (h *ReportHandler) GetShopeeAdsReport(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if startDate == "" || endDate == "" {
		now := time.Now()
		endDate = now.Format("2006-01-02")
		startDate = now.AddDate(0, 0, -30).Format("2006-01-02")
	}

	result, err := h.reportService.GetShopeeAdsReport(c.Request.Context(), tenantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetLatestShopeeAdsReport handles GET /api/reports/shopee/ads/latest
func (h *ReportHandler) GetLatestShopeeAdsReport(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	result, err := h.reportService.GetLatestShopeeAdsReport(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetShopeeAdsReportByDate handles GET /api/reports/shopee/ads/:date
func (h *ReportHandler) GetShopeeAdsReportByDate(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	date := c.Param("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Date required"})
		return
	}

	result, err := h.reportService.GetShopeeAdsReportByDate(c.Request.Context(), tenantID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetTiktokAdsReport handles GET /api/reports/tiktok/ads
func (h *ReportHandler) GetTiktokAdsReport(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if startDate == "" || endDate == "" {
		now := time.Now()
		endDate = now.Format("2006-01-02")
		startDate = now.AddDate(0, 0, -30).Format("2006-01-02")
	}

	result, err := h.reportService.GetTiktokAdsReport(c.Request.Context(), tenantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetLatestTiktokAdsReport handles GET /api/reports/tiktok/ads/latest
func (h *ReportHandler) GetLatestTiktokAdsReport(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	result, err := h.reportService.GetLatestTiktokAdsReport(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
