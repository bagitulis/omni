package ml

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/ml"
)

// ReportHandler handles ML report requests
type ReportHandler struct {
	basePath string
}

// NewReportHandler creates a new report handler
func NewReportHandler(basePath string) *ReportHandler {
	return &ReportHandler{basePath: basePath}
}

// Generate handles POST /api/ml/reports/generate
func (h *ReportHandler) Generate(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	platform := c.PostForm("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, response.Error("platform is required"))
		return
	}

	if platform != "shopee" && platform != "tiktok" {
		c.JSON(http.StatusBadRequest, response.Error("platform must be shopee or tiktok"))
		return
	}

	db, err := handlers.GetTenantDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	service := ml.NewReportService(db, "./output", "./notebooks")

	// Generate report (async in background recommended, but for now sync)
	report, err := service.GenerateReport(c.Request.Context(), tenantID, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":           report.ID,
			"platform":     report.Platform,
			"report_type":  report.ReportType,
			"file_name":    report.FileName,
			"file_size":    report.FileSize,
			"period_label": report.PeriodLabel,
			"created_at":   report.CreatedAt,
		},
	})
}

// List handles GET /api/ml/reports/:platform/list
func (h *ReportHandler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	platform := c.Param("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, response.Error("platform is required"))
		return
	}

	db, err := handlers.GetTenantDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	service := ml.NewReportService(db, "./output", "./notebooks")
	reports, err := service.ListReports(c.Request.Context(), tenantID, platform, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var data []gin.H
	for _, r := range reports {
		data = append(data, gin.H{
			"id":           r.ID,
			"platform":     r.Platform,
			"report_type":  r.ReportType,
			"file_name":    r.FileName,
			"file_size":    r.FileSize,
			"period_label": r.PeriodLabel,
			"status":       r.Status,
			"created_at":   r.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetLatest handles GET /api/ml/reports/:platform/latest
func (h *ReportHandler) GetLatest(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	platform := c.Param("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, response.Error("platform is required"))
		return
	}

	db, err := handlers.GetTenantDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	service := ml.NewReportService(db, "./output", "./notebooks")
	report, err := service.GetLatestReport(c.Request.Context(), tenantID, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if report == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "No reports found",
		})
		return
	}

	// Read HTML content
	htmlContent, err := service.ReadReportHTML(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to read report: " + err.Error(),
		})
		return
	}

	// Return HTML directly
	c.Data(http.StatusOK, "text/html; charset=utf-8", htmlContent)
}

// GetByFilename handles GET /api/ml/reports/:platform/:filename
func (h *ReportHandler) GetByFilename(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	platform := c.Param("platform")
	filename := c.Param("filename")

	if platform == "" || filename == "" {
		c.JSON(http.StatusBadRequest, response.Error("platform and filename are required"))
		return
	}

	db, err := handlers.GetTenantDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Database connection failed"})
		return
	}

	service := ml.NewReportService(db, "./output", "./notebooks")
	report, err := service.GetReportByFilename(c.Request.Context(), tenantID, platform, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if report == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Report not found",
		})
		return
	}

	// Read HTML content
	htmlContent, err := service.ReadReportHTML(report)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to read report: " + err.Error(),
		})
		return
	}

	// Return HTML directly
	c.Data(http.StatusOK, "text/html; charset=utf-8", htmlContent)
}
