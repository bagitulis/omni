package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UploadShopeeAds handles POST /api/ads/shopee/upload
func (h *AdsHandler) UploadShopeeAds(c *gin.Context) {
	svc, _, err := h.newShopeeService(c)
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

	result, err := svc.ParseCSV(c.Request.Context(), data, header.Filename)
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

// GetShopeeAds handles GET /api/ads/shopee
func (h *AdsHandler) GetShopeeAds(c *gin.Context) {
	svc, _, err := h.newShopeeService(c)
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

// GetShopeeAdsSummary handles GET /api/ads/shopee/summary
func (h *AdsHandler) GetShopeeAdsSummary(c *gin.Context) {
	svc, _, err := h.newShopeeService(c)
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

// GetShopeeAdsTrends handles GET /api/ads/shopee/trends
func (h *AdsHandler) GetShopeeAdsTrends(c *gin.Context) {
	svc, _, err := h.newShopeeService(c)
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

// GetShopeeProductPerformance handles GET /api/ads/shopee/performance
func (h *AdsHandler) GetShopeeProductPerformance(c *gin.Context) {
	svc, _, err := h.newShopeeService(c)
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
