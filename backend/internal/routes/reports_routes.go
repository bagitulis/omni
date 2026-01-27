package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services"
	"gorm.io/gorm"
)

// RegisterReportRoutes registers report generation routes
func RegisterReportRoutes(router *gin.RouterGroup, db *gorm.DB) {
	reportService := services.NewReportService(db)
	handler := handlers.NewReportHandler(reportService)

	reports := router.Group("/reports")
	reports.Use(middleware.Auth())
	reports.Use(middleware.Tenant())
	{
		// Shopee Ads Reports
		reports.GET("/shopee/ads", handler.GetShopeeAdsReport)
		reports.GET("/shopee/ads/latest", handler.GetLatestShopeeAdsReport)
		reports.GET("/shopee/ads/:date", handler.GetShopeeAdsReportByDate)

		// TikTok Ads Reports
		reports.GET("/tiktok/ads", handler.GetTiktokAdsReport)
		reports.GET("/tiktok/ads/latest", handler.GetLatestTiktokAdsReport)
	}
}
