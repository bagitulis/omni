package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	lazadaHandler "github.com/omni/backend/internal/handlers/lazada"
	tiktokHandler "github.com/omni/backend/internal/handlers/tiktok"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/webhooks"
)

// RegisterPlatformAuthRoutes registers platform auth extended routes
func RegisterPlatformAuthRoutes(router *gin.RouterGroup, handler *handlers.PlatformAuthHandler) {
	auth := router.Group("/platform-auth")
	auth.Use(middleware.Auth())
	auth.Use(middleware.Tenant())
	{
		auth.GET("/urls", handler.GetOAuthURLs)
		auth.GET("/status", handler.GetStatus)
		// NOTE: /logs is registered in routes.go via oauthHandler.GetOAuthLogs
		auth.GET("/shopee/disconnect", handler.DisconnectShopee)
		auth.GET("/lazada/disconnect", handler.DisconnectLazada)
		auth.POST("/check-all", handler.CheckAllConnections)
		auth.GET("/tiktok/shops", handler.GetTiktokShops)
		auth.GET("/tiktok/active-shop", handler.GetActiveTiktokShop)
	}
}

// RegisterWebhookExtendedRoutes registers webhook extended routes
func RegisterWebhookExtendedRoutes(
	router *gin.RouterGroup,
	shopeeProcessor *webhooks.ShopeeWebhookProcessor,
	lazadaProcessor *webhooks.LazadaWebhookProcessor,
	tiktokProcessor *webhooks.TiktokWebhookProcessor,
	basePath string,
) {
	handler := handlers.NewWebhookExtendedHandler(shopeeProcessor, lazadaProcessor, tiktokProcessor, basePath)

	// Public tenant-specific webhooks (no auth)
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/:tenantId/shopee", handler.ShopeeWebhookTenant)
		webhooks.POST("/:tenantId/lazada", handler.LazadaWebhookTenant)
		webhooks.POST("/:tenantId/tiktok", handler.TiktokWebhookTenant)
	}

	// Protected webhook endpoints
	protected := router.Group("/webhooks")
	protected.Use(middleware.Auth())
	protected.Use(middleware.Tenant())
	{
		protected.POST("/test", handler.TestWebhook)
		protected.GET("/config", handler.GetWebhookConfig)
		protected.POST("/config", handler.SaveWebhookConfig)
	}
}

// RegisterTiktokProductExtendedRoutes registers TikTok product extended routes
func RegisterTiktokProductExtendedRoutes(router *gin.RouterGroup, basePath string) {
	createHandler := tiktokHandler.NewProductCreateHandler(basePath)
	imageHandler := tiktokHandler.NewProductImageHandler(basePath)
	complianceHandler := tiktokHandler.NewProductComplianceHandler(basePath)
	searchHandler := tiktokHandler.NewProductSearchHandler(basePath)

	products := router.Group("/tiktok/products")
	products.Use(middleware.Auth())
	products.Use(middleware.Tenant())
	{
		// Sync from API (POST /api/tiktok/products/search - mimics Node.js)
		products.POST("/search", searchHandler.SearchProducts)

		// Draft operations
		products.POST("/draft", createHandler.SaveDraft)
		products.POST("/publish/:draftId", createHandler.PublishDraft)

		// Database operations
		products.GET("/db", createHandler.GetProductsFromDB)
		products.GET("/db/:productId", createHandler.GetProductFromDB)
		products.GET("/search", createHandler.SearchProducts) // GET for DB search

		// Category/attribute metadata
		products.GET("/categories", createHandler.GetCategories)
		products.GET("/categories/:categoryId/attributes", createHandler.GetAttributes)
		products.GET("/categories/:categoryId/rules", createHandler.GetRules)
		products.GET("/brands", createHandler.GetBrands)
		products.GET("/delivery-options", createHandler.GetDeliveryOptions)
		products.GET("/warehouses", createHandler.GetWarehouses)

		// Image operations
		products.POST("/upload-image", imageHandler.UploadImage)
		products.GET("/image-upload-tasks", imageHandler.GetUploadTasks)

		// Compliance operations
		products.GET("/:productId/compliance", complianceHandler.GetCompliance)
		products.POST("/:productId/compliance", complianceHandler.UpdateCompliance)
		products.GET("/global-products", complianceHandler.GetGlobalProducts)
		products.POST("/publish-global", complianceHandler.PublishGlobal)
	}
}

// RegisterWholesaleExtendedRoutes registers wholesale extended routes
func RegisterWholesaleExtendedRoutes(router *gin.RouterGroup, basePath string) {
	// Pass nil for DB - handler will get DB from context
	handler := handlers.NewWholesaleExtendedHandler(basePath, nil)
	batchHandler := handlers.NewWholesaleBatchHandler(basePath, nil)

	wholesale := router.Group("/wholesale")
	wholesale.Use(middleware.Auth())
	wholesale.Use(middleware.Tenant())
	{
		// Shopee wholesale
		wholesale.DELETE("/shopee/:itemId", handler.DeleteWholesale)
		wholesale.PUT("/shopee/:itemId", handler.UpdateWholesale)
		wholesale.GET("/shopee/:itemId", handler.GetWholesaleInfo)
		wholesale.GET("/shopee/:itemId/info", handler.GetWholesaleInfo)
		wholesale.POST("/shopee/batch-delete", batchHandler.BatchDeleteByItemIds)
		wholesale.POST("/shopee/batch-delete-skus", batchHandler.BatchDeleteBySkus)
		wholesale.POST("/shopee/batch-add", handler.BatchAdd)
		wholesale.GET("/shopee/lookup/:sku", handler.LookupItemId)
		wholesale.POST("/shopee/preview", handler.Preview)
		wholesale.POST("/shopee/import", handler.ImportWholesale)
		wholesale.POST("/shopee/batch-mpq", handler.BatchSetMpq)
		wholesale.POST("/shopee/batch-wholesale-reset", handler.BatchWholesaleReset)
		wholesale.POST("/shopee/batch-reset", handler.BatchWholesaleReset)

		// NEW: Batch update by SKUs
		wholesale.POST("/shopee/batch-update-skus", batchHandler.BatchUpdateBySkus)

		// TikTok wholesale
		wholesale.POST("/tiktok/batch-mpq", handler.BatchSetTiktokMpq)
		wholesale.POST("/tiktok/:productId", handler.SetTiktokWholesale)
	}
}

// RegisterLazadaProductExtendedRoutes registers Lazada product extended routes
func RegisterLazadaProductExtendedRoutes(router *gin.RouterGroup, basePath string) {
	handler := lazadaHandler.NewProductExtendedHandler(basePath)

	products := router.Group("/lazada/products")
	products.Use(middleware.Auth())
	products.Use(middleware.Tenant())
	{
		products.GET("/db", handler.GetProductsFromDB)
		products.GET("/db/:itemId", handler.GetProductFromDB)
		products.GET("/categories", handler.GetCategories)
		products.GET("/attributes/:categoryId", handler.GetAttributes)
	}
}

// RegisterLazadaSyncRoutes registers Lazada sync routes
func RegisterLazadaSyncRoutes(router *gin.RouterGroup, basePath string) {
	handler := lazadaHandler.NewSyncHandler(basePath)

	sync := router.Group("/lazada/sync")
	sync.Use(middleware.Auth())
	sync.Use(middleware.Tenant())
	{
		sync.POST("/orders", handler.SyncOrders)
		sync.POST("/products", handler.SyncProducts)
	}
}

// RegisterProductMetadataRoutes registers product metadata routes
func RegisterProductMetadataRoutes(router *gin.RouterGroup, basePath string) {
	handler := handlers.NewProductMetadataHandler(basePath)

	create := router.Group("/products/create")
	create.Use(middleware.Auth())
	create.Use(middleware.Tenant())
	{
		// Shopee metadata
		create.GET("/shopee/categories", handler.GetShopeeCategories)
		create.GET("/shopee/attributes/:catId", handler.GetShopeeAttributes)
		create.GET("/shopee/brands/:catId", handler.GetShopeeBrands)
		create.GET("/shopee/logistics", handler.GetShopeeLogistics)

		// Lazada metadata
		create.GET("/lazada/categories", handler.GetLazadaCategories)
		create.GET("/lazada/attributes/:catId", handler.GetLazadaAttributes)
		create.GET("/lazada/brands/:catId", handler.GetLazadaBrands)

		// TikTok metadata
		create.GET("/tiktok/categories", handler.GetTiktokCategories)
		create.GET("/tiktok/attributes/:catId", handler.GetTiktokAttributes)
		create.GET("/tiktok/brands", handler.GetTiktokBrands)
		create.GET("/tiktok/warehouses", handler.GetTiktokWarehouses)

		// Common operations
		create.POST("/upload-image", handler.UploadImage)
		create.POST("/validate", handler.ValidateProduct)
		create.GET("/templates", handler.GetTemplates)
	}
}
