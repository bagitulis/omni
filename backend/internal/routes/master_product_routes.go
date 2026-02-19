package routes

import (
	"github.com/gin-gonic/gin"
	masterProductHandler "github.com/omni/backend/internal/handlers/master_product"
	"github.com/omni/backend/internal/middleware"
)

// RegisterMasterProductRoutes registers Master Product API routes
// Path: /api/master-products
func RegisterMasterProductRoutes(router *gin.RouterGroup, basePath string) {
	handler := masterProductHandler.NewHandler(basePath)

	masterProducts := router.Group("/master-products")
	masterProducts.Use(middleware.Auth())
	masterProducts.Use(middleware.Tenant())
	{
		// List products with pagination
		// GET /api/master-products?page=1&limit=20&status=active&search=keyword
		masterProducts.GET("", handler.List)

		// Get single product with SKUs
		// GET /api/master-products/:id
		masterProducts.GET("/:id", handler.GetByID)

		// Create new product
		// POST /api/master-products
		masterProducts.POST("", handler.Create)

		// Update product
		// PUT /api/master-products/:id
		masterProducts.PUT("/:id", handler.Update)

		// Delete product
		// DELETE /api/master-products/:id
		masterProducts.DELETE("/:id", handler.Delete)

		// Batch update SKUs (price/stock)
		// PUT /api/master-products/skus/batch
		masterProducts.PUT("/skus/batch", handler.BatchUpdateSkus)

		// Update single SKU (price/stock)
		// PUT /api/master-products/skus/:id
		masterProducts.PUT("/skus/:id", handler.UpdateSku)
	}
}

// RegisterMasterProductImportRoutes registers import-related routes
// Path: /api/master-products/import, /api/master-products/mapping
func RegisterMasterProductImportRoutes(router *gin.RouterGroup, basePath string) {
	importHandler := masterProductHandler.NewImportHandler(basePath)
	stagingImportHandler := masterProductHandler.NewStagingImportHandler(basePath)

	masterProducts := router.Group("/master-products")
	masterProducts.Use(middleware.Auth())
	masterProducts.Use(middleware.Tenant())
	{
		// Import routes
		importGroup := masterProducts.Group("/import")
		{
			// Preview import
			// GET /api/master-products/import/preview?platform=shopee&item_id=123
			importGroup.GET("/preview", importHandler.Preview)

			// Execute import
			// POST /api/master-products/import
			importGroup.POST("", importHandler.Import)

			// Import from staging DB (no live API calls)
			// POST /api/master-products/import/from-staging/shopee
			importGroup.POST("/from-staging/shopee", stagingImportHandler.ImportFromShopee)
			// POST /api/master-products/import/from-staging/tiktok
			importGroup.POST("/from-staging/tiktok", stagingImportHandler.ImportFromTiktok)
			// POST /api/master-products/import/from-staging/lazada
			importGroup.POST("/from-staging/lazada", stagingImportHandler.ImportFromLazada)
		}

		// Mapping routes
		mappingGroup := masterProducts.Group("/mapping")
		{
			// Auto-map SKU by seller_sku
			// POST /api/master-products/mapping/auto
			mappingGroup.POST("/auto", importHandler.AutoMap)

			// Auto-map and link SKUs in batch
			// POST /api/master-products/mapping/auto-link
			mappingGroup.POST("/auto-link", importHandler.AutoMapBatch)

			// Manual link
			// POST /api/master-products/mapping/link
			mappingGroup.POST("/link", importHandler.ManualLink)

			// Unlink
			// DELETE /api/master-products/mapping/link
			mappingGroup.DELETE("/link", importHandler.Unlink)
		}

		// Get mapping status for a product
		// GET /api/master-products/:id/mapping
		masterProducts.GET("/:id/mapping", importHandler.GetMappingStatus)
	}
}

// RegisterMasterProductSyncRoutes registers sync-related routes
// Path: /api/master-products/:id/sync, /api/master-products/:id/sync-status
func RegisterMasterProductSyncRoutes(router *gin.RouterGroup, basePath string) {
	syncHandler := masterProductHandler.NewSyncHandler(basePath)

	masterProducts := router.Group("/master-products")
	masterProducts.Use(middleware.Auth())
	masterProducts.Use(middleware.Tenant())
	{
		// Backfill master product images from platform cache
		// POST /api/master-products/images/backfill
		masterProducts.POST("/images/backfill", syncHandler.BackfillImages)

		// Refresh a single master product images from linked platform cache
		// POST /api/master-products/:id/images/refresh
		masterProducts.POST("/:id/images/refresh", syncHandler.RefreshProductImages)

		// Sync to platform
		// POST /api/master-products/:id/sync
		masterProducts.POST("/:id/sync", syncHandler.Sync)

		// Get sync status
		// GET /api/master-products/:id/sync-status
		masterProducts.GET("/:id/sync-status", syncHandler.GetSyncStatus)
	}
}
