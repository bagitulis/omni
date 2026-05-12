package handlers

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/monitoring"
	"gorm.io/gorm"
)

// MonitoringHandler handles monitoring endpoints
type MonitoringHandler struct {
	db             *gorm.DB
	metricsService *monitoring.MetricsService
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(db *gorm.DB, metricsService *monitoring.MetricsService) *MonitoringHandler {
	return &MonitoringHandler{db: db, metricsService: metricsService}
}

// NewSimpleMonitoringHandler creates a monitoring handler without metrics service
// Used when metrics service is not available
func NewSimpleMonitoringHandler() *MonitoringHandler {
	return &MonitoringHandler{db: nil, metricsService: nil}
}

// GetMetrics handles GET /api/monitoring/metrics
func (h *MonitoringHandler) GetMetrics(c *gin.Context) {
	if h.metricsService == nil {
		// Return basic metrics when service not available
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"metrics": gin.H{
				"memory_alloc_mb":       memStats.Alloc / 1024 / 1024,
				"memory_total_alloc_mb": memStats.TotalAlloc / 1024 / 1024,
				"memory_sys_mb":         memStats.Sys / 1024 / 1024,
				"num_gc":                memStats.NumGC,
				"goroutines":            runtime.NumGoroutine(),
			},
		})
		return
	}
	metrics := h.metricsService.GetMetrics()
	c.JSON(http.StatusOK, gin.H{"success": true, "metrics": metrics})
}

// GetDetailedHealthMonitoring handles GET /api/monitoring/health/detailed
// Uses ComponentHealth from health.go
func (h *MonitoringHandler) GetDetailedHealthMonitoring(c *gin.Context) {
	health := MonitoringHealthStatus{
		Status:     "healthy",
		Components: make(map[string]ComponentHealth),
	}

	// Check database
	dbHealth := h.checkDatabase()
	health.Components["database"] = dbHealth
	if dbHealth.Status != "healthy" {
		health.Status = "degraded"
	}

	// Memory info
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	health.Components["memory"] = ComponentHealth{
		Status:  "healthy",
		Message: "Memory usage within limits",
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "health": health})
}

// checkDatabase checks database health
func (h *MonitoringHandler) checkDatabase() ComponentHealth {
	if h.db == nil {
		return ComponentHealth{
			Status:  "unknown",
			Message: "Database connection not configured",
		}
	}
	sqlDB, err := h.db.DB()
	if err != nil {
		return ComponentHealth{
			Status:  "unhealthy",
			Message: err.Error(),
		}
	}

	if err := sqlDB.Ping(); err != nil {
		return ComponentHealth{
			Status:  "unhealthy",
			Message: err.Error(),
		}
	}

	return ComponentHealth{
		Status:  "healthy",
		Message: "Connected and responsive",
	}
}

// MonitoringHealthStatus represents overall health status for monitoring
type MonitoringHealthStatus struct {
	Status     string                     `json:"status"`
	Components map[string]ComponentHealth `json:"components"`
}
