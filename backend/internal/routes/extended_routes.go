package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	inventoryHandler "github.com/omni/backend/internal/handlers/inventory"
	shopeeHandler "github.com/omni/backend/internal/handlers/shopee"
	tiktokHandler "github.com/omni/backend/internal/handlers/tiktok"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/google"
	"github.com/omni/backend/internal/services/inventory"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	"gorm.io/gorm"
)

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
		shipping.GET("/label/:orderSn", handler.GetShippingLabel)
		shipping.GET("/download/:orderSn", handler.DownloadShippingLabel)
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
		cloneData.GET("/preview", handler.Preview)
	}
}

// RegisterTiktokShippingRoutes registers TikTok shipping routes
func RegisterTiktokShippingRoutes(router *gin.RouterGroup, basePath string) {
	handler := tiktokHandler.NewShippingHandler(basePath)
	shipping := router.Group("/tiktok/shipping")
	shipping.Use(middleware.Auth())
	shipping.Use(middleware.Tenant())
	{
		// Shipping arrangement
		shipping.POST("/arrange", handler.ArrangeShipment)

		// Shipping documents/labels
		shipping.GET("/document/:packageId", handler.GetShippingDocument)
		shipping.GET("/document/order/:orderId", handler.GetShippingDocumentByOrder)

		// Download shipping label to local file
		shipping.GET("/download/order/:orderId", handler.DownloadShippingDocumentByOrder)
		shipping.POST("/download/batch", handler.BatchDownloadShippingDocuments)

		// Handover time slots for pickup scheduling
		shipping.GET("/timeslots/:orderOrPackageId", handler.GetHandoverTimeSlots)

		// Order detail with package info
		shipping.GET("/order/:orderId", handler.GetOrderDetail)
	}
}
