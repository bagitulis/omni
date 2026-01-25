package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/middleware"
)

// RegisterSkuBatchCheckRoutes registers SKU batch check routes
// Path: /api/inventory (batch-check-sku, batch-save-platform-status, platform-status)
func RegisterSkuBatchCheckRoutes(router *gin.RouterGroup, handler *handlers.SkuBatchCheckHandler) {
	inventory := router.Group("/inventory")
	inventory.Use(middleware.Auth())
	inventory.Use(middleware.Tenant())
	{
		// SKU batch check (check SKUs across all platforms)
		inventory.POST("/batch-check-sku", handler.BatchCheckSku)

		// Save platform status (cache check results)
		inventory.POST("/batch-save-platform-status", handler.BatchSavePlatformStatus)

		// Get cached platform status
		inventory.GET("/platform-status", handler.GetPlatformStatus)
	}
}

// RegisterInventorySimpleRoutes registers simple inventory routes
// Path: /api/inventory (basic endpoints)
// NOTE: Use RegisterInventoryRoutes in extended_routes.go for full inventory functionality
func RegisterInventorySimpleRoutes(router *gin.RouterGroup, handler *handlers.InventoryHandler) {
	inventory := router.Group("/inventory")
	inventory.Use(middleware.Auth())
	inventory.Use(middleware.Tenant())
	{
		inventory.GET("/config", handler.GetConfig)
		inventory.GET("/stats", handler.GetStats)
		inventory.GET("/list", handler.GetList)

		// Column management
		columns := inventory.Group("/columns")
		columns.GET("/available", handler.GetAvailableColumns)
		columns.GET("/selected", handler.GetSelectedColumns)
		columns.POST("/selected", handler.UpdateSelectedColumns)

		// Sync endpoints (Node.js compatibility)
		sync := inventory.Group("/sync")
		sync.POST("/from-sheets", handler.SyncFromSheets)
		sync.POST("/to-sheets", handler.SyncToSheets)

		// Sheet operations
		inventory.POST("/sync-status", handler.GetSyncStatus)
		inventory.POST("/export-to-sheet", handler.ExportToSheet)
		inventory.POST("/import-from-sheet", handler.ImportFromSheet)

		// Stock updates
		inventory.POST("/update-stock", handler.UpdateStock)
		inventory.POST("/update-stock-batch", handler.UpdateStockBatch)

		// Price updates
		inventory.POST("/update-price", handler.UpdatePrice)
		inventory.POST("/update-price-batch", handler.UpdatePriceBatch)
	}
}

// RegisterSettingsRoutes registers settings routes
func RegisterSettingsRoutes(router *gin.RouterGroup, handler *handlers.SettingsHandler) {
	settings := router.Group("/settings")
	settings.Use(middleware.Auth())
	settings.Use(middleware.Tenant())
	{
		settings.GET("/inventory", handler.GetInventorySettings)
		settings.PUT("/inventory", handler.UpdateInventorySettings)
		settings.GET("/google-sheets", handler.GetGoogleSheetsSettings)
		settings.PUT("/google-sheets", handler.UpdateGoogleSheetsSettings)
	}
}

// RegisterSpreadsheetRegistryRoutes registers spreadsheet registry routes
func RegisterSpreadsheetRegistryRoutes(router *gin.RouterGroup, handler *handlers.SpreadsheetRegistryHandler) {
	spreadsheets := router.Group("/spreadsheets")
	spreadsheets.Use(middleware.Auth())
	spreadsheets.Use(middleware.Tenant())
	{
		spreadsheets.GET("", handler.List)
		spreadsheets.GET("/:id", handler.Get)
		spreadsheets.POST("", handler.Register)
		spreadsheets.PUT("/:id", handler.Update)
		spreadsheets.DELETE("/:id", handler.Delete)
		spreadsheets.POST("/:id/synced", handler.MarkSynced)
	}
}
