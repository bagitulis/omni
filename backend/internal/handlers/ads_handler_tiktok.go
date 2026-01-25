package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UploadTiktokAds handles POST /api/ads/tiktok/upload
func (h *AdsHandler) UploadTiktokAds(c *gin.Context) {
	svc, _, err := h.newTiktokService(c)
	if svc == nil {
		return // Error already sent
	}
	if err != nil {
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File required"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	result, err := svc.ParseExcel(c.Request.Context(), data, header.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := svc.SaveBatch(c.Request.Context(), header.Filename, result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Upload successful",
		"totalRows": len(result.Data),
		"period":    result.Period,
	})
}

// GetTiktokAds handles GET /api/ads/tiktok
func (h *AdsHandler) GetTiktokAds(c *gin.Context) {
	svc, _, err := h.newTiktokService(c)
	if svc == nil || err != nil {
		return
	}

	periodLabel := c.Query("period")
	data, err := svc.GetData(c.Request.Context(), periodLabel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// GetTiktokAdsSummary handles GET /api/ads/tiktok/summary
func (h *AdsHandler) GetTiktokAdsSummary(c *gin.Context) {
	svc, _, err := h.newTiktokService(c)
	if svc == nil || err != nil {
		return
	}

	periodLabel := c.Query("period")
	summary, err := svc.GetSummary(c.Request.Context(), periodLabel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": summary})
}

// GetTiktokAdsTrends handles GET /api/ads/tiktok/trends
func (h *AdsHandler) GetTiktokAdsTrends(c *gin.Context) {
	svc, _, err := h.newTiktokService(c)
	if svc == nil || err != nil {
		return
	}

	startDate, endDate := parseDateRange(c)
	trends, err := svc.GetTrends(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": trends})
}

// GetTiktokProductPerformance handles GET /api/ads/tiktok/performance
func (h *AdsHandler) GetTiktokProductPerformance(c *gin.Context) {
	svc, _, err := h.newTiktokService(c)
	if svc == nil || err != nil {
		return
	}

	startDate, endDate := parseDateRange(c)
	perf, err := svc.GetProductPerformance(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": perf})
}

// GetTiktokPredictions handles GET /api/ads/tiktok/predictions
func (h *AdsHandler) GetTiktokPredictions(c *gin.Context) {
	svc, _, err := h.newTiktokService(c)
	if svc == nil || err != nil {
		return
	}

	predictions, err := svc.GetPredictions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": predictions})
}
