package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/middleware"
)

// RegisterRouteConfigRoutes registers route config routes
// Using /routes-config (plural) to match Node.js pattern
func RegisterRouteConfigRoutes(router *gin.RouterGroup, handler *handlers.RouteConfigHandler) {
	routeConfig := router.Group("/routes-config")
	routeConfig.Use(middleware.Auth())
	routeConfig.Use(middleware.Tenant())
	{
		routeConfig.GET("", handler.List)
		routeConfig.GET("/all", handler.ListAll)
		routeConfig.GET("/categories", handler.GetCategories)
		routeConfig.GET("/by-category/:category", handler.GetByCategory)
		routeConfig.GET("/:id", handler.Get)
		routeConfig.POST("", handler.Create)
		routeConfig.PATCH("/:id", handler.Update)
		routeConfig.DELETE("/:id", handler.Delete)
		routeConfig.POST("/bulk-update", handler.BulkUpdate)
		routeConfig.POST("/apply-preset/:preset", handler.ApplyPreset)
		routeConfig.POST("/reset", handler.Reset)
	}
}

// RegisterRouteExecutionConfigRoutes registers route execution config routes
// Path: /api/route-execution-config
func RegisterRouteExecutionConfigRoutes(router *gin.RouterGroup, handler *handlers.RouteExecutionConfigHandler) {
	execConfig := router.Group("/route-execution-config")
	execConfig.Use(middleware.Auth())
	execConfig.Use(middleware.Tenant())
	{
		execConfig.GET("", handler.List)
		execConfig.GET("/:routeKey", handler.Get)
		execConfig.GET("/:routeKey/mode", handler.GetMode)
		execConfig.POST("", handler.Create)
		execConfig.PUT("/:routeKey", handler.Update)
		execConfig.POST("/:routeKey/toggle", handler.Toggle)
		execConfig.DELETE("/:routeKey", handler.Delete)
	}
}

// RegisterMonitoringRoutes registers monitoring routes
func RegisterMonitoringRoutes(router *gin.RouterGroup, handler *handlers.MonitoringHandler) {
	monitoring := router.Group("/monitoring")
	monitoring.Use(middleware.Auth())
	{
		monitoring.GET("/metrics", handler.GetMetrics)
		monitoring.GET("/health/detailed", handler.GetDetailedHealthMonitoring)
	}
}

// RegisterSecurityRoutes registers security routes
func RegisterSecurityRoutes(router *gin.RouterGroup, handler *handlers.SecurityHandler) {
	security := router.Group("/security")
	security.Use(middleware.Auth())
	security.Use(middleware.Tenant())
	{
		security.GET("/alerts", handler.GetAlerts)
		security.POST("/report", handler.ReportIssue)
		security.GET("/stats", handler.GetStats)
	}
}

// RegisterFilterPreferenceRoutes registers filter preference routes
func RegisterFilterPreferenceRoutes(router *gin.RouterGroup, handler *handlers.FilterPreferenceHandler) {
	filterPref := router.Group("/filter-preferences")
	filterPref.Use(middleware.Auth())
	filterPref.Use(middleware.Tenant())
	{
		filterPref.GET("", handler.Get)
		filterPref.POST("", handler.Save)
		filterPref.DELETE("", handler.Delete) // Uses query params: ?platform=xxx&tab=yyy
	}
}

// RegisterCSRFRoutes registers CSRF token routes (public)
func RegisterCSRFRoutes(router *gin.RouterGroup, handler *handlers.CSRFHandler) {
	router.GET("/csrf-token", handler.GetCSRFToken)
}

// RegisterDocsRoutes registers documentation routes (public)
func RegisterDocsRoutes(router *gin.RouterGroup, handler *handlers.DocsHandler) {
	router.GET("/docs", handler.GetAPIInfo)
	router.GET("/docs/swagger.json", handler.GetSwaggerJSON)
}

// RegisterOrderSyncRoutes registers order sync routes
func RegisterOrderSyncRoutes(router *gin.RouterGroup, handler *handlers.OrderSyncHandler) {
	orders := router.Group("/orders")
	orders.Use(middleware.Auth())
	orders.Use(middleware.Tenant())
	{
		orders.POST("/sync/:category", handler.SyncByCategory)
		orders.POST("/sync/platform/:platform", handler.SyncPlatformOrders)
		orders.GET("/category/:category", handler.GetOrdersByCategory)
		orders.GET("/details/:platform", handler.GetOrderDetails)
	}
}

// RegisterOrderManagerRoutes registers order manager routes for frontend compatibility
func RegisterOrderManagerRoutes(router *gin.RouterGroup, handler *handlers.OrderManagerHandler) {
	orders := router.Group("/orders")
	orders.Use(middleware.Auth())
	orders.Use(middleware.Tenant())
	{
		// Category-specific GET endpoints (frontend expects these)
		orders.GET("/unpaid", handler.GetUnpaidOrders)
		orders.GET("/unprocess", handler.GetUnprocessOrders)
		orders.GET("/processed", handler.GetProcessedOrders)

		// Special endpoints - support both GET and POST for locked-today
		orders.GET("/locked-today", handler.GetSavedLockedOrders)
		orders.POST("/locked-today", handler.GetLockedTodayOrders)

		// Order Today - GET retrieves saved, POST syncs fresh data
		orders.GET("/today", handler.GetOrdersToday)
		orders.POST("/today", handler.SyncOrdersToday)

		orders.POST("/sync-all", handler.SyncAll)
		orders.POST("/bulk-print-labels", handler.BulkPrintLabels)
	}
}

// RegisterLockedOrderRoutes registers locked order routes
func RegisterLockedOrderRoutes(router *gin.RouterGroup, handler *handlers.LockedOrderHandler) {
	locked := router.Group("/locked-orders")
	locked.Use(middleware.Auth())
	locked.Use(middleware.Tenant())
	{
		locked.GET("", handler.GetLockedOrders)
		locked.POST("", handler.SaveLockedOrders)
		locked.DELETE("", handler.ClearLockedOrders)
	}
}

// RegisterProductMasterRoutes registers product master routes
func RegisterProductMasterRoutes(router *gin.RouterGroup, handler *handlers.ProductMasterHandler) {
	products := router.Group("/products/master")
	products.Use(middleware.Auth())
	products.Use(middleware.Tenant())
	{
		products.GET("", handler.GetMasterProductList)
		products.GET("/stats", handler.GetMasterProductStats)
	}

	// Platform-specific product detail
	platformProducts := router.Group("/products")
	platformProducts.Use(middleware.Auth())
	platformProducts.Use(middleware.Tenant())
	{
		platformProducts.GET("/:platform/:itemId", handler.GetProductByID)
	}
}

// RegisterRouteMappingRoutes registers route mapping routes
func RegisterRouteMappingRoutes(router *gin.RouterGroup, handler *handlers.RouteHandler) {
	routeMap := router.Group("/routes")
	routeMap.Use(middleware.Auth())
	{
		routeMap.GET("", handler.GetAllRoutes)
		routeMap.GET("/tag/:tag", handler.GetRoutesByTag)
		routeMap.GET("/analyze", handler.AnalyzeRoutes)
		routeMap.GET("/scan", handler.ScanRoutes)
		routeMap.GET("/middleware", handler.ScanMiddleware)
	}
}

// RegisterProductCreateRoutes registers product create routes
func RegisterProductCreateRoutes(router *gin.RouterGroup, handler *handlers.ProductCreateHandler) {
	create := router.Group("/products/create")
	create.Use(middleware.Auth())
	create.Use(middleware.Tenant())
	{
		create.POST("/shopee", handler.CreateOnShopee)
		create.POST("/lazada", handler.CreateOnLazada)
		create.POST("/tiktok", handler.CreateOnTiktok)
	}

	// Categories
	categories := router.Group("/products/categories")
	categories.Use(middleware.Auth())
	categories.Use(middleware.Tenant())
	{
		categories.GET("/:platform", handler.GetCategories)
	}
}
