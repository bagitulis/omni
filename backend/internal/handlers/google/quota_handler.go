// Package google handles Google API related endpoints
package google

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/google"
)

// QuotaHandler handles Google API quota endpoints
type QuotaHandler struct {
	quotaService *google.QuotaService
}

// NewQuotaHandler creates a new quota handler
func NewQuotaHandler(quotaService *google.QuotaService) *QuotaHandler {
	return &QuotaHandler{quotaService: quotaService}
}

// GetStatus handles GET /api/google/quota/status
// Returns current quota usage status (public endpoint)
func (h *QuotaHandler) GetStatus(c *gin.Context) {
	status := h.quotaService.GetStatus()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// GetStats handles GET /api/google/quota/stats
// Returns detailed quota statistics (public endpoint)
func (h *QuotaHandler) GetStats(c *gin.Context) {
	stats := h.quotaService.GetDetailedStats()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
