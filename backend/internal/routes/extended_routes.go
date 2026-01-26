package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/handlers/analytics"
	inventoryHandler "github.com/omni/backend/internal/handlers/inventory"
	shopeeHandler "github.com/omni/backend/internal/handlers/shopee"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/google"
	"github.com/omni/backend/internal/services/inventory"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	"gorm.io/gorm"
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
		shopeeAds.POST("/upload", handler.Upload)
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
		tiktokAds.POST("/upload", handler.Upload)
	}
}

// RegisterShopeeWalletRoutes registers Shopee wallet routes
func RegisterShopeeWalletRoutes(router *gin.RouterGroup, getAPIClient func(tenantID string) shopeeService.APIClient) {
	handler := shopeeHandler.NewWalletHandler(getAPIClient)
	wallet := router.Group("/shopee/wallet")
	wallet.Use(middleware.Auth())
	wallet.Use(middleware.Tenant())
	{
		wallet.GET("/balance", handler.GetBalance)
		wallet.GET("/transactions", handler.GetTransactions)
		wallet.GET("/income", handler.GetNetIncome)
	}
}

// RegisterShopeeWalletReportRoutes registers Shopee wallet report routes
func RegisterShopeeWalletReportRoutes(router *gin.RouterGroup, getAPIClient func(tenantID string) shopeeService.APIClient, googleAuth *google.AuthService) {
	handler := shopeeHandler.NewWalletReportHandler(getAPIClient, googleAuth)
	wallet := router.Group("/shopee/wallet")
	wallet.Use(middleware.Auth())
	wallet.Use(middleware.Tenant())
	{
		wallet.POST("/report", handler.GetWalletReport)
		wallet.POST("/export", handler.ExportWallet)
		wallet.POST("/export-to-sheets", handler.ExportToSheets)
	}
}

// RegisterShopeeEscrowRoutes registers Shopee escrow routes
func RegisterShopeeEscrowRoutes(router *gin.RouterGroup, getAPIClient func(tenantID string) shopeeService.APIClient) {
	handler := shopeeHandler.NewEscrowHandler(getAPIClient)
	wallet := router.Group("/shopee/wallet")
	wallet.Use(middleware.Auth())
	wallet.Use(middleware.Tenant())
	{
		wallet.POST("/escrow-detail", handler.GetEscrowDetail)
		wallet.POST("/escrow-detail-batch", handler.GetEscrowDetailBatch)
	}
}

// RegisterShopeeShippingRoutes registers Shopee shipping routes
func RegisterShopeeShippingRoutes(router *gin.RouterGroup, getAPIClient func(tenantID string) shopeeService.APIClient) {
	handler := shopeeHandler.NewShippingHandler(getAPIClient)
	shipping := router.Group("/shopee/shipping")
	shipping.Use(middleware.Auth())
	shipping.Use(middleware.Tenant())
	{
		shipping.GET("/options", handler.GetOptions)
		shipping.POST("/arrange", handler.ArrangeShipment)
		shipping.GET("/tracking/:orderSn", handler.GetTracking)
		shipping.GET("/info/:orderSn", handler.GetShipment)
	}
}

// RegisterShopeeShippingFeeRoutes registers Shopee shipping fee routes
func RegisterShopeeShippingFeeRoutes(router *gin.RouterGroup, getAPIClient func(tenantID string) shopeeService.APIClient, googleAuth *google.AuthService) {
	handler := shopeeHandler.NewShippingFeeHandler(getAPIClient, googleAuth)
	shipping := router.Group("/shopee/shipping")
	shipping.Use(middleware.Auth())
	shipping.Use(middleware.Tenant())
	{
		shipping.POST("/process-fee", handler.ProcessShippingFee)
		shipping.POST("/export-fee", handler.ExportShippingFee)
		shipping.POST("/export-to-sheets", handler.ExportToSheets)
	}
}

// RegisterInventoryRoutes registers inventory management routes
func RegisterInventoryRoutes(router *gin.RouterGroup, db *gorm.DB, sheetsClient inventory.SheetWriterClient) {
	dataHandler := inventoryHandler.NewDataHandler(db)
	configHandler := inventoryHandler.NewConfigHandler(db)
	syncHandler := inventoryHandler.NewSyncHandler(db, nil) // Pass nil for SheetsClient, or use type assertion if needed
	columnsHandler := inventoryHandler.NewColumnsHandler(db)
	stockHandler := inventoryHandler.NewStockHandler(db)
	priceHandler := inventoryHandler.NewPriceHandler(db)
	sheetHandler := inventoryHandler.NewSheetHandler(db, sheetsClient)

	inv := router.Group("/inventory")
	inv.Use(middleware.Auth())
	inv.Use(middleware.Tenant())
	{
		// Data endpoints
		inv.GET("/data", dataHandler.List)
		inv.GET("/data/:sku", dataHandler.GetBySKU)
		inv.PUT("/data/:sku", dataHandler.Update)
		inv.DELETE("/data/:sku", dataHandler.Delete)
		inv.GET("/categories", dataHandler.GetCategories)

		// Config endpoints
		inv.GET("/config", configHandler.GetSettings)
		inv.PUT("/config", configHandler.UpdateSettings)

		// Sync endpoints
		inv.POST("/sync", syncHandler.TriggerSync)
		inv.GET("/sync/history", syncHandler.GetSyncHistory)
		inv.POST("/sync/partial", sheetHandler.PartialSync)

		// Column configuration
		inv.GET("/columns/available", columnsHandler.GetAvailableColumns)
		inv.GET("/columns/selected", columnsHandler.GetSelectedColumns)
		inv.POST("/columns/selected", columnsHandler.SaveSelectedColumns)

		// Stock updates
		inv.POST("/update-stock", stockHandler.UpdateStock)
		inv.POST("/update-stock-batch", stockHandler.UpdateStockBatch)
		inv.POST("/lookup-platform-ids", stockHandler.LookupPlatformIds)

		// Price updates
		inv.POST("/update-price", priceHandler.UpdatePrice)
		inv.POST("/update-price-batch", priceHandler.UpdatePriceBatch)

		// Sheet export/import
		inv.GET("/export", sheetHandler.Export)
		inv.POST("/export-to-sheet", sheetHandler.ExportToSheet)
		inv.POST("/import-from-sheet", sheetHandler.ImportFromSheet)
		inv.POST("/sync-from-sheet", sheetHandler.SyncFromSheet)
	}
}

// RegisterStockRoutes registers stock management routes
func RegisterStockRoutes(router *gin.RouterGroup, handler *handlers.StockHandler) {
	stock := router.Group("/stock")
	stock.Use(middleware.Auth())
	stock.Use(middleware.Tenant())
	{
		stock.GET("", handler.List)
		stock.GET("/alerts", handler.GetAlerts)
		stock.GET("/:sku", handler.GetBySKU)
		stock.PUT("/:sku", handler.Update)
		stock.POST("/bulk", handler.BulkUpdate)
	}
}

// RegisterPriceRoutes registers price management routes
func RegisterPriceRoutes(router *gin.RouterGroup, handler *handlers.PriceHandler) {
	price := router.Group("/price")
	price.Use(middleware.Auth())
	price.Use(middleware.Tenant())
	{
		price.GET("", handler.List)
		price.GET("/:sku", handler.GetBySKU)
		price.PUT("/:sku", handler.Update)
		price.POST("/bulk", handler.BulkUpdate)
	}
}

// RegisterProductCloneRoutes registers product clone routes
func RegisterProductCloneRoutes(router *gin.RouterGroup, handler *handlers.ProductCloneHandler) {
	clone := router.Group("/products/clone")
	clone.Use(middleware.Auth())
	clone.Use(middleware.Tenant())
	{
		clone.POST("", handler.Clone)
		clone.GET("/status/:id", handler.GetStatus)
		clone.POST("/batch", handler.BatchClone)
	}

	// Clone data endpoints (used by frontend for clone modal)
	cloneData := router.Group("/clone")
	cloneData.Use(middleware.Auth())
	cloneData.Use(middleware.Tenant())
	{
		cloneData.GET("/product-data", handler.GetProductData)
		cloneData.GET("/available-targets", handler.GetAvailableTargets)
	}
}
