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
		notifs.GET("", handler.ListNotifications)
		notifs.POST("", handler.CreateNotification)
		notifs.GET("/unread-count", handler.GetUnreadCount)
		notifs.PATCH("/:id/read", handler.MarkAsRead)
		notifs.PATCH("/read-all", handler.MarkAllAsRead)
		notifs.DELETE("/:id", handler.DeleteNotification)
		notifs.DELETE("", handler.DeleteAllNotifications)
		notifs.GET("/settings", handler.GetSettings)
		notifs.PUT("/settings", handler.UpdateSettings)
		notifs.GET("/:id/detail", handler.GetNotificationDetail)
	}
}
