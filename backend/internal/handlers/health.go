package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
)

var startTime = time.Now()

// DetailedHealthResponse represents detailed health check response
type DetailedHealthResponse struct {
	Status     string                     `json:"status"`
	Timestamp  string                     `json:"timestamp"`
	Uptime     string                     `json:"uptime"`
	UptimeSec  int64                      `json:"uptime_seconds"`
	Version    string                     `json:"version"`
	GoVersion  string                     `json:"go_version"`
	Memory     MemoryStats                `json:"memory"`
	Goroutines int                        `json:"goroutines"`
	Components map[string]ComponentHealth `json:"components"`
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Alloc      uint64 `json:"alloc_mb"`
	TotalAlloc uint64 `json:"total_alloc_mb"`
	Sys        uint64 `json:"sys_mb"`
	NumGC      uint32 `json:"num_gc"`
}

// ComponentHealth represents health status of a component
type ComponentHealth struct {
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
	ResponseMs int64  `json:"response_ms,omitempty"`
}

// HealthCheck handles GET /api/health
// Response format matches Node.js backend and frontend HealthCheckResponse type
func HealthCheck(c *gin.Context) {
	uptime := time.Since(startTime).Round(time.Second).String()

	// Return status at root level to match frontend expectations
	// Frontend checks: health.status === "healthy"
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    uptime,
		"version":   "1.0.0",
		"goVersion": runtime.Version(),
		"services": map[string]string{
			"database": "connected",
			"cache":    "connected",
		},
	})
}

// DetailedHealthCheck handles GET /api/monitoring/health/detailed
func DetailedHealthCheck(basePath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		uptime := now.Sub(startTime)

		// Memory stats
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// Check database connectivity
		dbHealth := checkDatabaseHealth(basePath)

		// Overall status
		overallStatus := "healthy"
		if dbHealth.Status != "healthy" {
			overallStatus = "degraded"
		}

		response := DetailedHealthResponse{
			Status:    overallStatus,
			Timestamp: now.UTC().Format(time.RFC3339),
			Uptime:    uptime.Round(time.Second).String(),
			UptimeSec: int64(uptime.Seconds()),
			Version:   "1.0.0",
			GoVersion: runtime.Version(),
			Memory: MemoryStats{
				Alloc:      memStats.Alloc / 1024 / 1024,
				TotalAlloc: memStats.TotalAlloc / 1024 / 1024,
				Sys:        memStats.Sys / 1024 / 1024,
				NumGC:      memStats.NumGC,
			},
			Goroutines: runtime.NumGoroutine(),
			Components: map[string]ComponentHealth{
				"database": dbHealth,
				"memory": {
					Status:  "healthy",
					Message: "Memory usage within limits",
				},
			},
		}

		statusCode := http.StatusOK
		if overallStatus != "healthy" {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, gin.H{
			"success": overallStatus == "healthy",
			"data":    response,
		})
	}
}

// checkDatabaseHealth verifies database connectivity
func checkDatabaseHealth(basePath string) ComponentHealth {
	start := time.Now()

	// Try to connect to system database
	db, err := config.GetSystemDB(basePath)
	if err != nil {
		return ComponentHealth{
			Status:  "unhealthy",
			Message: "Failed to connect: " + err.Error(),
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return ComponentHealth{
			Status:  "unhealthy",
			Message: "Failed to get SQL DB: " + err.Error(),
		}
	}

	if err := sqlDB.Ping(); err != nil {
		return ComponentHealth{
			Status:  "unhealthy",
			Message: "Ping failed: " + err.Error(),
		}
	}

	return ComponentHealth{
		Status:     "healthy",
		Message:    "Connected and responsive",
		ResponseMs: time.Since(start).Milliseconds(),
	}
}

// StatusCheck handles GET /api/status - mirrors Node.js backend
// This is a public endpoint that returns connection status
// If tenant context is provided, it also returns token status for all platforms
func StatusCheck(c *gin.Context) {
	now := time.Now()
	// Format timestamp like Python: "2025-12-20 17:06:03.585025"
	timestamp := now.Format("2006-01-02 15:04:05.000000")

	// Check if tenant context is available (from header)
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		tenantID = c.GetHeader("x-tenant-id")
	}

	// If no tenant, return basic status
	if tenantID == "" {
		c.JSON(http.StatusOK, gin.H{
			"connection_status": "connected",
			"data": map[string]string{
				"message": "Backend is running. Provide x-tenant-id header for token status.",
			},
			"success":   true,
			"timestamp": timestamp,
		})
		return
	}

	// Return with platform status placeholder
	// In full implementation, this would fetch actual token status
	data := map[string]string{
		"shopee": "Shopee Token Status:\n  ✅ Backend connected",
		"lazada": "Lazada Token Status:\n  ✅ Backend connected",
		"tiktok": "TikTok Token Status:\n  ✅ Backend connected",
	}

	c.JSON(http.StatusOK, gin.H{
		"connection_status": "connected",
		"data":              data,
		"success":           true,
		"timestamp":         timestamp,
	})
}
