package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/middleware"
)

// RegisterMarketplaceSyncHistoryRoutes registers marketplace sync history routes
func RegisterMarketplaceSyncHistoryRoutes(router *gin.RouterGroup, basePath string) {
	handler := handlers.NewMarketplaceSyncHistoryHandler(basePath)

	syncHistory := router.Group("/marketplace-sync-history")
	syncHistory.Use(middleware.Auth())
	syncHistory.Use(middleware.Tenant())
	{
		syncHistory.GET("", handler.List)
		syncHistory.POST("", handler.Create)
	}
}
