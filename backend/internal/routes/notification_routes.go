package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/middleware"
)

// RegisterNotificationRoutes registers notification API routes (protected)
func RegisterNotificationRoutes(router *gin.RouterGroup, handler *handlers.NotificationHandler) {
	// SSE stream uses ticket-based auth (EventSource cannot send Authorization header)
	// It is authenticated inside the handler via ValidateSSETicket(c.Query("ticket")).
	router.GET("/notifications/stream", handler.StreamNotifications)

	notifs := router.Group("/notifications")
	notifs.Use(middleware.Auth())
	notifs.Use(middleware.Tenant())
	{
		// V2 primary routes — per-user reads, filters, bulk, counts, snooze.
		notifs.GET("", handler.ListV2)
		notifs.POST("", handler.CreateNotification)
		notifs.GET("/counts", handler.GetCounts)
		notifs.PATCH("/:id/read", handler.MarkReadV2)
		notifs.PATCH("/read-all", handler.MarkAllReadV2)
		notifs.POST("/bulk/read", handler.BulkMarkRead)
		notifs.POST("/bulk/delete", handler.BulkDelete)
		notifs.POST("/:id/snooze", handler.Snooze)
		notifs.DELETE("/:id", handler.DeleteNotification)
		notifs.DELETE("", handler.DeleteAllNotifications)
		notifs.GET("/settings", handler.GetSettings)
		notifs.PUT("/settings", handler.UpdateSettings)
		notifs.GET("/:id/detail", handler.GetNotificationDetail)

		// Legacy alias — kept for any pre-V2 clients on unread-count.
		notifs.GET("/unread-count", handler.GetUnreadCount)
	}
}
