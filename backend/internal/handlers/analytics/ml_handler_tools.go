package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
)

// GetAlerts returns active alerts
// GET /api/analytics/ml/alerts
func (h *MLHandler) GetAlerts(c *gin.Context) {
	svc, err := h.getService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get database connection",
		})
		return
	}

	alerts, err := svc.GetAlerts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get alerts",
		})
		return
	}

	highCount, mediumCount, lowCount := countAlertsBySeverity(alerts)

	c.JSON(http.StatusOK, dto.MLAlertsResponse{
		Success: true,
		Data:    alerts,
		Meta: dto.AlertsMeta{
			TotalActive:    len(alerts),
			HighPriority:   highCount,
			MediumPriority: mediumCount,
			LowPriority:    lowCount,
		},
	})
}

// countAlertsBySeverity counts alerts by severity level
func countAlertsBySeverity(alerts []models.MLAlert) (high, medium, low int) {
	for _, a := range alerts {
		switch a.Severity {
		case "HIGH":
			high++
		case "MEDIUM":
			medium++
		case "LOW":
			low++
		}
	}
	return
}

// SimulateBudget simulates budget changes
// POST /api/analytics/ml/budget-sim
func (h *MLHandler) SimulateBudget(c *gin.Context) {
	var req dto.BudgetSimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	if len(req.ProductIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "At least one product ID is required",
		})
		return
	}

	svc, err := h.getService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get database connection",
		})
		return
	}

	result, err := svc.SimulateBudget(c.Request.Context(), req.ProductIDs, req.BudgetChangePct)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to simulate budget",
		})
		return
	}

	c.JSON(http.StatusOK, dto.BudgetSimResponse{
		Success: true,
		Data:    *result,
	})
}

// GetScoreDistribution returns score distribution for charts
// GET /api/analytics/ml/distribution
func (h *MLHandler) GetScoreDistribution(c *gin.Context) {
	svc, err := h.getService(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get database connection",
		})
		return
	}

	distribution, err := svc.GetScoreDistribution(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get distribution",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ScoreDistributionResponse{
		Success: true,
		Data:    distribution,
	})
}
