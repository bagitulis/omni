package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/middleware"
)

// RegisterShopeeAnalyticsRoutes registers Shopee analytics report routes
// Path: /api/analytics/shopee
func RegisterShopeeAnalyticsRoutes(router *gin.RouterGroup, handler *handlers.ShopeeAnalyticsHandler) {
	shopee := router.Group("/analytics/shopee")
	shopee.Use(middleware.Auth(), middleware.Tenant())
	{
		shopee.GET("/settings", handler.GetSettings)
		shopee.POST("/settings", handler.SaveSettings)
		shopee.GET("/sync-status", handler.GetSyncStatus)
		shopee.POST("/sync", handler.SyncEscrow)
		shopee.DELETE("/sync", handler.DeleteSyncData)
		shopee.GET("/reconciliation", handler.GetReconciliation)
		shopee.GET("/shipping-fee", handler.GetShippingFeeAnalysis)
		shopee.POST("/repopulate-items", handler.RepopulateItems)
	}
}

// RegisterTiktokAnalyticsRoutes registers TikTok analytics report routes
// Path: /api/analytics/tiktok
func RegisterTiktokAnalyticsRoutes(router *gin.RouterGroup, handler *handlers.TiktokAnalyticsHandler) {
	tiktok := router.Group("/analytics/tiktok")
	tiktok.Use(middleware.Auth(), middleware.Tenant())
	{
		tiktok.GET("/settings", handler.GetSettings)
		tiktok.POST("/settings", handler.SaveSettings)
		tiktok.GET("/sync-status", handler.GetSyncStatus)
		tiktok.POST("/sync", handler.SyncEscrow)
		tiktok.DELETE("/sync", handler.DeleteSyncData)
		tiktok.GET("/reconciliation", handler.GetReconciliation)
		tiktok.GET("/shipping-fee", handler.GetShippingFeeAnalysis)
		tiktok.POST("/repopulate-items", handler.RepopulateItems)
	}
}
