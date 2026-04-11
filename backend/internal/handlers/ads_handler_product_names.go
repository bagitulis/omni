package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ImportTiktokProductNames handles POST /api/ads/tiktok/product-names
// Accepts a JSON file with product ID -> name mapping
func (h *AdsHandler) ImportTiktokProductNames(c *gin.Context) {
	svc, _, err := h.newTiktokService(c)
	if svc == nil {
		return
	}
	if err != nil {
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "JSON file required"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to read file"})
		return
	}

	inserted, updated, err := svc.ImportProductNames(c.Request.Context(), data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Auto-backfill existing data after import
	backfilled, _ := svc.BackfillProductNames(c.Request.Context())

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Product names imported successfully",
		"inserted":   inserted,
		"updated":    updated,
		"backfilled": backfilled,
	})
}

// GetTiktokProductNames handles GET /api/ads/tiktok/product-names
func (h *AdsHandler) GetTiktokProductNames(c *gin.Context) {
	svc, _, err := h.newTiktokService(c)
	if svc == nil || err != nil {
		return
	}

	names, err := svc.GetProductNames(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    names,
		"total":   len(names),
	})
}

// BackfillTiktokProductNames handles POST /api/ads/tiktok/product-names/backfill
func (h *AdsHandler) BackfillTiktokProductNames(c *gin.Context) {
	svc, _, err := h.newTiktokService(c)
	if svc == nil || err != nil {
		return
	}

	affected, err := svc.BackfillProductNames(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Backfill completed",
		"affected": affected,
	})
}
