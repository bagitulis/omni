package routes

import (
	"github.com/gin-gonic/gin"
	masterProductHandler "github.com/omni/backend/internal/handlers/master_product"
	"github.com/omni/backend/internal/middleware"
	"gorm.io/gorm"
)

// RegisterMasterProductRoutes registers Master Product API routes
// Path: /api/master-products
func RegisterMasterProductRoutes(router *gin.RouterGroup, db *gorm.DB) {
	handler := masterProductHandler.NewHandler(db)

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
	}
}

// RegisterMasterProductImportRoutes registers import-related routes
// Path: /api/master-products/import, /api/master-products/mapping
func RegisterMasterProductImportRoutes(router *gin.RouterGroup, db *gorm.DB, basePath string) {
	importHandler := masterProductHandler.NewImportHandler(db, basePath)

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
		}

		// Mapping routes
		mappingGroup := masterProducts.Group("/mapping")
		{
			// Auto-map SKU by seller_sku
			// POST /api/master-products/mapping/auto
			mappingGroup.POST("/auto", importHandler.AutoMap)

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
func RegisterMasterProductSyncRoutes(router *gin.RouterGroup, db *gorm.DB, basePath string) {
	syncHandler := masterProductHandler.NewSyncHandler(db, basePath)

	masterProducts := router.Group("/master-products")
	masterProducts.Use(middleware.Auth())
	masterProducts.Use(middleware.Tenant())
	{
		// Sync to platform
		// POST /api/master-products/:id/sync
		masterProducts.POST("/:id/sync", syncHandler.Sync)

		// Get sync status
		// GET /api/master-products/:id/sync-status
		masterProducts.GET("/:id/sync-status", syncHandler.GetSyncStatus)
	}
}
