package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
)

// GetAlerts returns active alerts from cache
// GET /api/analytics/ml/alerts
func (h *MLHandler) GetAlerts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant ID"})
		return
	}

	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to connect to tenant database"})
		return
	}

	h.cacheService.EnsureTables(tenantDB)

	// Read alerts from cache
	cachedAlerts := h.cacheService.GetCachedAlerts(c.Request.Context(), tenantDB, tenantID)

	// Convert to alert format
	var alerts []models.MLAlert
	for _, cp := range cachedAlerts {
		if cp.HasFatigueWarning {
			alerts = append(alerts, models.MLAlert{
				ID:          cp.ProductID + "_fatigue",
				TenantID:    tenantID,
				ProductID:   cp.ProductID,
				ProductName: cp.ProductName,
				AlertType:   "FATIGUE_WARNING",
				Severity:    "HIGH",
				Message:     "Creative fatigue detected: " + cp.FatigueStatus,
				CreatedAt:   cp.CalculatedAt,
				Status:      "ACTIVE",
			})
		}
		if cp.HasChurnRisk {
			severity := "MEDIUM"
			if cp.ChurnRiskScore > 70 {
				severity = "HIGH"
			}
			alerts = append(alerts, models.MLAlert{
				ID:          cp.ProductID + "_churn",
				TenantID:    tenantID,
				ProductID:   cp.ProductID,
				ProductName: cp.ProductName,
				AlertType:   "CHURN_RISK",
				Severity:    severity,
				Message:     "High churn risk detected",
				CreatedAt:   cp.CalculatedAt,
				Status:      "ACTIVE",
			})
		}
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

// SimulateBudget simulates budget changes using cached data
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

	svc, _, err := h.getService(c)
	if err != nil {
		if err == config.ErrMissingTenantID {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant ID"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to connect to tenant database"})
		}
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

// GetScoreDistribution returns score distribution from cache
// GET /api/analytics/ml/distribution
func (h *MLHandler) GetScoreDistribution(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant ID"})
		return
	}

	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to connect to tenant database"})
		return
	}

	h.cacheService.EnsureTables(tenantDB)

	// Read from cache
	counts := h.cacheService.GetCachedDistribution(c.Request.Context(), tenantDB, tenantID)

	// Calculate total
	total := 0
	for _, count := range counts {
		total += count
	}

	// Build response
	var distribution []models.ScoreDistribution
	for _, cat := range []string{"STAR", "GROWTH", "STABLE", "WATCH", "PROBLEM"} {
		pct := 0.0
		if total > 0 {
			pct = float64(counts[cat]) / float64(total) * 100
		}
		distribution = append(distribution, models.ScoreDistribution{
			Category:   cat,
			Count:      counts[cat],
			Percentage: pct,
		})
	}

	c.JSON(http.StatusOK, dto.ScoreDistributionResponse{
		Success: true,
		Data:    distribution,
	})
}
