package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/middleware"
)

// RegisterJobQueueRoutes registers job queue routes
// Extended to match Node.js pattern with all endpoints
func RegisterJobQueueRoutes(router *gin.RouterGroup, handler *handlers.JobQueueHandler) {
	jobs := router.Group("/jobs")
	jobs.Use(middleware.Auth())
	jobs.Use(middleware.Tenant())
	{
		// Core CRUD endpoints
		jobs.POST("", handler.AddJob)
		jobs.GET("", handler.ListJobs)
		jobs.GET("/:id", handler.GetJob)
		jobs.DELETE("/:id", handler.CancelJob)

		// Extended endpoints to match Node.js
		jobs.POST("/enqueue", handler.EnqueueJob)
		jobs.GET("/status", handler.GetStatus)
		jobs.GET("/status/:jobId", handler.GetJobStatus)
		jobs.GET("/queue", handler.GetQueue)
		jobs.GET("/monitor", handler.GetMonitor)
		jobs.GET("/history", handler.GetHistory)
		jobs.GET("/history-paginated", handler.GetHistoryPaginated)
		jobs.GET("/history-job-types", handler.GetHistoryJobTypes)
		jobs.DELETE("/history", handler.ClearHistory)
		jobs.POST("/cancel/:jobId", handler.CancelJobByJobId)
		jobs.POST("/force-cancel/:jobId", handler.ForceCancelJob)
		jobs.GET("/check-timeout", handler.CheckTimeout)
		jobs.GET("/stats", handler.GetStats)
	}
}

// RegisterAutoFunctionRoutes registers auto function routes under /jobs/auto-functions
// to be consistent with Node.js backend pattern
func RegisterAutoFunctionRoutes(router *gin.RouterGroup, handler *handlers.AutoFunctionHandler) {
	// Mount under /jobs/auto-functions to match Node.js pattern
	autoFunc := router.Group("/jobs/auto-functions")
	autoFunc.Use(middleware.Auth())
	autoFunc.Use(middleware.Tenant())
	{
		// List all configs
		autoFunc.GET("", handler.List)

		// Get history (must be before :name to avoid conflict)
		autoFunc.GET("/history", handler.GetHistory)

		// Get single config by name
		autoFunc.GET("/:name", handler.GetByName)

		// Create new config
		autoFunc.POST("", handler.Create)

		// Update config by name (create or update)
		autoFunc.PUT("/:name", handler.Update)

		// Delete config by name
		autoFunc.DELETE("/:name", handler.Delete)

		// Actions by name - use :name consistently
		autoFunc.POST("/:name/enable", handler.Enable)
		autoFunc.POST("/:name/disable", handler.Disable)
		autoFunc.POST("/:name/run", handler.Run)
		autoFunc.POST("/:name/cancel-scheduled", handler.CancelScheduledByName)
	}
}
