package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/handlers/analytics"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/cache"
)

// RegisterTiktokAnalyticsRoutes registers TikTok analytics routes
// Maps to /api/analytics/tiktok/* (matching Node.js backend)
func RegisterTiktokAnalyticsRoutes(router *gin.RouterGroup, handler *handlers.TiktokAnalyticsHandler) {
	tiktok := router.Group("/analytics/tiktok")
	tiktok.Use(middleware.Auth())
	tiktok.Use(middleware.Tenant())
	{
		// Settings endpoints
		tiktok.GET("/settings", handler.GetSettings)
		tiktok.POST("/settings", handler.SaveSettings)

		// Sync status endpoints
		tiktok.GET("/sync-status", handler.GetSyncStatus)
		tiktok.POST("/sync", handler.SyncEscrow)
		tiktok.DELETE("/sync", handler.DeleteSyncData)

		// Analysis endpoints
		tiktok.GET("/reconciliation", handler.GetReconciliation)
		tiktok.GET("/shipping-fee", handler.GetShippingFeeAnalysis)

		// Utility: re-populate escrow items from raw_order_data already in DB
		tiktok.POST("/repopulate-items", handler.RepopulateItems)
	}
}

// RegisterShopeeAnalyticsRoutes registers Shopee analytics routes
// Maps to /api/analytics/shopee/* (matching Node.js backend)
func RegisterShopeeAnalyticsRoutes(router *gin.RouterGroup, handler *handlers.ShopeeAnalyticsHandler) {
	shopee := router.Group("/analytics/shopee")
	shopee.Use(middleware.Auth())
	shopee.Use(middleware.Tenant())
	{
		// Settings endpoints
		shopee.GET("/settings", handler.GetSettings)
		shopee.POST("/settings", handler.SaveSettings)

		// Sync status endpoints
		shopee.GET("/sync-status", handler.GetSyncStatus)
		shopee.POST("/sync", handler.SyncEscrow)
		shopee.DELETE("/sync", handler.DeleteSyncData)

		// Analysis endpoints
		shopee.GET("/reconciliation", handler.GetReconciliation)
		shopee.GET("/shipping-fee", handler.GetShippingFeeAnalysis)

		// Utility: re-populate escrow items from raw_order_income already in DB
		shopee.POST("/repopulate-items", handler.RepopulateItems)
	}
}

// RegisterAdsRoutes registers ads report routes
func RegisterAdsRoutes(router *gin.RouterGroup, handler *handlers.AdsHandler) {
	ads := router.Group("/ads")
	ads.Use(middleware.Auth())
	ads.Use(middleware.Tenant())
	{
		// Shopee Ads
		ads.POST("/shopee/upload", handler.UploadShopeeAds)
		ads.GET("/shopee", handler.GetShopeeAds)
		ads.GET("/shopee/summary", handler.GetShopeeAdsSummary)
		ads.GET("/shopee/trends", handler.GetShopeeAdsTrends)
		ads.GET("/shopee/performance", handler.GetShopeeProductPerformance)

		// TikTok Ads
		ads.POST("/tiktok/upload", handler.UploadTiktokAds)
		ads.GET("/tiktok", handler.GetTiktokAds)
		ads.GET("/tiktok/summary", handler.GetTiktokAdsSummary)
		ads.GET("/tiktok/trends", handler.GetTiktokAdsTrends)
		ads.GET("/tiktok/performance", handler.GetTiktokProductPerformance)
		ads.GET("/tiktok/predictions", handler.GetTiktokPredictions)

		// TikTok Product Name Mapping
		ads.POST("/tiktok/product-names", handler.ImportTiktokProductNames)
		ads.GET("/tiktok/product-names", handler.GetTiktokProductNames)
		ads.POST("/tiktok/product-names/backfill", handler.BackfillTiktokProductNames)
	}
}

// RegisterShopeeAdsAnalyticsRoutes registers Shopee Ads analytics routes
// Maps to /api/analytics/shopee-ads/* (matching frontend composable)
func RegisterShopeeAdsAnalyticsRoutes(router *gin.RouterGroup, basePath string) {
	handler := analytics.NewAdsHandler(basePath)
	shopeeAds := router.Group("/analytics/shopee-ads")
	shopeeAds.Use(middleware.Auth())
	shopeeAds.Use(middleware.Tenant())
	{
		shopeeAds.GET("/dashboard", handler.GetDashboard)
		shopeeAds.GET("/data", handler.GetData)
		shopeeAds.GET("/uploads", handler.GetUploads)
		// Upload handled by /api/ads/shopee/upload (AdsHandler.UploadShopeeAds)
	}
}

// RegisterTiktokAdsAnalyticsRoutes registers TikTok Ads analytics routes
// Maps to /api/analytics/tiktok-ads/* (matching frontend composable)
func RegisterTiktokAdsAnalyticsRoutes(router *gin.RouterGroup, basePath string) {
	handler := analytics.NewTiktokAdsHandler(basePath)
	tiktokAds := router.Group("/analytics/tiktok-ads")
	tiktokAds.Use(middleware.Auth())
	tiktokAds.Use(middleware.Tenant())
	{
		tiktokAds.GET("/dashboard", handler.GetDashboard)
		tiktokAds.GET("/data", handler.GetData)
		tiktokAds.GET("/uploads", handler.GetUploads)
		// Upload handled by /api/ads/tiktok/upload (AdsHandler.UploadTiktokAds)
	}
}

// RegisterMLAnalyticsRoutes registers ML analytics routes
// Maps to /api/analytics/ml/* (new ML-powered analytics)
func RegisterMLAnalyticsRoutes(router *gin.RouterGroup, appCache cache.CacheManager) {
	handler := analytics.NewMLHandler(appCache)
	ml := router.Group("/analytics/ml")
	ml.Use(middleware.Auth())
	ml.Use(middleware.Tenant())
	{
		// Portfolio health (lightweight summary)
		ml.GET("/portfolio-health", handler.GetPortfolioHealth)

		// Products with scores (paginated)
		ml.GET("/products", handler.GetProducts)

		// Single product detail
		ml.GET("/product/:id", handler.GetProductDetail)

		// Alerts
		ml.GET("/alerts", handler.GetAlerts)

		// Budget simulation
		ml.POST("/budget-sim", handler.SimulateBudget)

		// Score distribution for charts
		ml.GET("/distribution", handler.GetScoreDistribution)
	}
}

// RegisterSimulationRoutes registers budget simulation routes
// Maps to /api/analytics/simulation/* and /api/analytics/products/*
func RegisterSimulationRoutes(router *gin.RouterGroup, basePath string) {
	handler := analytics.NewSimulationHandler(basePath)

	// Simulation routes
	sim := router.Group("/analytics/simulation")
	sim.Use(middleware.Auth())
	sim.Use(middleware.Tenant())
	{
		sim.POST("/calculate", handler.Simulate)
	}

	// Products from ads routes
	products := router.Group("/analytics/products")
	products.Use(middleware.Auth())
	products.Use(middleware.Tenant())
	{
		products.GET("/from-ads", handler.GetProductsFromAds)
	}

	// Intelligence routes
	intel := router.Group("/analytics/intelligence")
	intel.Use(middleware.Auth())
	intel.Use(middleware.Tenant())
	{
		intel.GET("/calendar", handler.GetCalendarEvents)
	}
}

// RegisterUnifiedAnalyticsRoutes registers unified analytics routes
// Maps to /api/analytics/unified/* and /api/analytics/cache/*
func RegisterUnifiedAnalyticsRoutes(router *gin.RouterGroup, basePath string, appCache cache.CacheManager) {
	handler := analytics.NewUnifiedHandler(basePath, appCache)

	// Unified summary routes
	unified := router.Group("/analytics/unified")
	unified.Use(middleware.Auth())
	unified.Use(middleware.Tenant())
	{
		unified.GET("/summary", handler.GetUnifiedSummary)
		unified.GET("/kpi", handler.GetUnifiedKPI)
	}

	// Product classification routes
	products := router.Group("/analytics/products")
	products.Use(middleware.Auth())
	products.Use(middleware.Tenant())
	{
		products.GET("/classified", handler.GetClassifiedProducts)
		products.GET("/top", handler.GetTopProducts)
	}

	// Cache management routes
	cache := router.Group("/analytics/cache")
	cache.Use(middleware.Auth())
	cache.Use(middleware.Tenant())
	{
		cache.POST("/refresh", handler.RefreshCache)
		cache.GET("/status", handler.GetCacheStatus)
	}
}
