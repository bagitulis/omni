package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/app"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/handlers/lazada"
	"github.com/omni/backend/internal/handlers/shopee"
	"github.com/omni/backend/internal/handlers/tiktok"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/routes"
	googleService "github.com/omni/backend/internal/services/google"
)

func main() {
	cfg := config.Load()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize application dependencies
	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.Close()

	// ====== Initialize Google Services EARLY (needed for InventoryHandler) ======
	saPath := googleService.GetDefaultServiceAccountPath()
	saLoader := googleService.NewServiceAccountLoader(saPath)
	googleCredentials, err := saLoader.LoadFirstAvailable()
	if err != nil {
		log.Printf("Warning: Failed to load Google service account credentials: %v", err)
		log.Printf("Google Sheets features will not work. Path searched: %s", saPath)
		// Continue with nil credentials - will fail gracefully when used
		googleCredentials = nil
	}
	googleAuthService := googleService.NewAuthService(googleCredentials)

	// Initialize extended handlers with Google Auth
	extHandlers := application.InitExtendedHandlers(application.SystemDB, googleAuthService)

	router := gin.Default()
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// Public
	router.GET("/api/health", handlers.HealthCheck)
	// /api/status with real token data (uses TokenHandler for actual token status)
	router.GET("/api/status", application.TokenHandler.GetStatusWithTokens)

	// API Group
	api := router.Group("/api")

	// ====== Core Routes (from App) ======
	routes.RegisterAuthRoutes(api, application.AuthHandler)
	routes.RegisterProtectedAuthRoutes(api, application.AuthHandler)
	routes.RegisterUserRoutes(api, application.UserHandler)
	routes.RegisterAuditRoutes(api, application.AuditHandler)
	routes.RegisterCaptchaRoutes(api, application.CaptchaHandler)
	routes.RegisterOAuthRoutes(api, application.OAuthHandler)
	routes.RegisterTokenRoutes(api, application.TokenHandler)
	routes.RegisterAnalyticsRoutes(api, application.AnalyticsHandler)
	routes.RegisterWebhookRoutes(api, application.WebhookHandler)
	routes.RegisterWebhookAdminRoutes(api, application.WebhookHandler)

	// ====== Platform Auth Routes ======
	routes.RegisterPlatformAuthRoutes(api, application.PlatformAuthHandler)

	// ====== Extended Routes ======
	// Utility routes
	routes.RegisterCSRFRoutes(api, extHandlers.CSRFHandler)
	routes.RegisterDocsRoutes(api, extHandlers.DocsHandler)

	// Business routes
	routes.RegisterOrderSyncRoutes(api, extHandlers.OrderSyncHandler)
	routes.RegisterOrderManagerRoutes(api, extHandlers.OrderManagerHandler)
	routes.RegisterLockedOrderRoutes(api, extHandlers.LockedOrderHandler)
	routes.RegisterJobQueueRoutes(api, extHandlers.JobQueueHandler)

	// Auto Function routes (Script Monitor)
	routes.RegisterAutoFunctionRoutes(api, extHandlers.AutoFunctionHandler)

	// Product routes
	routes.RegisterProductMasterRoutes(api, extHandlers.ProductMasterHandler)
	routes.RegisterProductCreateRoutes(api, extHandlers.ProductCreateHandler)
	routes.RegisterProductCloneRoutes(api, extHandlers.ProductCloneHandler)
	routes.RegisterPriceRoutes(api, extHandlers.PriceHandler)
	routes.RegisterStockRoutes(api, extHandlers.StockHandler)
	routes.RegisterSKUCheckRoutes(api, extHandlers.SKUCheckHandler)

	// Master Product routes (unified product management)
	routes.RegisterMasterProductRoutes(api, application.SystemDB)
	routes.RegisterMasterProductImportRoutes(api, application.SystemDB, cfg.DatabasePath)
	routes.RegisterMasterProductSyncRoutes(api, application.SystemDB, cfg.DatabasePath)

	// Settings routes
	routes.RegisterSettingsRoutes(api, extHandlers.SettingsHandler)
	routes.RegisterRouteConfigRoutes(api, extHandlers.RouteConfigHandler)
	routes.RegisterRouteExecutionConfigRoutes(api, extHandlers.RouteExecutionConfigHandler)
	routes.RegisterRouteMappingRoutes(api, extHandlers.RouteHandler)

	// Inventory routes - simple routes for basic operations + Google Sheets sync
	// Stock/Price update routes are registered separately at /api/stock and /api/price
	routes.RegisterInventorySimpleRoutes(api, extHandlers.InventoryHandler)
	routes.RegisterSkuBatchCheckRoutes(api, extHandlers.SkuBatchCheckHandler)

	// Integration routes
	routes.RegisterSpreadsheetRegistryRoutes(api, extHandlers.SpreadsheetRegistryHandler)
	routes.RegisterFilterPreferenceRoutes(api, extHandlers.FilterPreferenceHandler)
	routes.RegisterWholesaleRoutes(api, extHandlers.WholesaleHandler)
	routes.RegisterWholesaleExtendedRoutes(api, cfg.DatabasePath)

	// Ads routes
	routes.RegisterAdsRoutes(api, extHandlers.AdsHandler)
	routes.RegisterTiktokAnalyticsRoutes(api, extHandlers.TiktokAnalyticsHandler)
	routes.RegisterShopeeAnalyticsRoutes(api, extHandlers.ShopeeAnalyticsHandler)

	// Ads Analytics routes (for frontend composables)
	routes.RegisterShopeeAdsAnalyticsRoutes(api, cfg.DatabasePath)
	routes.RegisterTiktokAdsAnalyticsRoutes(api, cfg.DatabasePath)

	// ML Analytics routes (new ML-powered analytics)
	routes.RegisterMLAnalyticsRoutes(api, application.CacheService)

	// Budget Simulation & Intelligence routes
	routes.RegisterSimulationRoutes(api, cfg.DatabasePath)

	// Unified Analytics & Cache Management routes
	routes.RegisterUnifiedAnalyticsRoutes(api, cfg.DatabasePath, application.CacheService)

	// ML Report routes
	routes.RegisterMLReportRoutes(api, cfg.DatabasePath)

	// ====== Google Routes ======
	// Use the already initialized googleAuthService
	googleQuotaService := googleService.NewQuotaService()
	googleHandlers := routes.NewGoogleHandlers(googleAuthService, googleQuotaService, application.SystemDB)
	routes.RegisterGoogleRoutes(api, googleHandlers)

	// ====== Monitoring Routes ======
	routes.RegisterMonitoringRoutes(api, extHandlers.MonitoringHandler)

	// Platform-specific handlers (existing)
	sOrder := shopee.NewOrderHandler(cfg.DatabasePath)
	sProd := shopee.NewProductHandler(cfg.DatabasePath)
	sSync := shopee.NewSyncHandlerWithCache(cfg.DatabasePath, application.CacheService)
	sDBProd := shopee.NewDBProductHandler(cfg.DatabasePath)
	lOrder := lazada.NewOrderHandler(cfg.DatabasePath)
	lProd := lazada.NewProductHandler(cfg.DatabasePath)
	lSync := lazada.NewSyncHandlerWithCache(cfg.DatabasePath, application.CacheService)
	lDBProd := lazada.NewDBProductHandler(cfg.DatabasePath)
	tOrder := tiktok.NewOrderHandler(cfg.DatabasePath)
	tProd := tiktok.NewProductHandler(cfg.DatabasePath)
	tSync := tiktok.NewSyncHandlerWithCache(cfg.DatabasePath, application.CacheService)
	tDBProd := tiktok.NewDBProductHandler(cfg.DatabasePath)

	// Protected platform routes
	protected := api.Group("")
	protected.Use(middleware.Auth())
	protected.Use(middleware.Tenant())

	// Shopee
	s := protected.Group("/shopee")
	s.GET("/orders", sOrder.GetOrders)
	s.GET("/orders/:orderSn", sOrder.GetOrderByID)
	s.POST("/orders/ship", sOrder.ShipOrder)
	s.POST("/orders/cancel", sOrder.CancelOrder)
	s.GET("/products", sProd.GetProducts)
	s.GET("/products/:itemId", sProd.GetProductByID)
	s.POST("/products", sProd.CreateProduct)
	s.PUT("/products/:itemId", sProd.UpdateProduct)
	s.DELETE("/products/:itemId", sProd.DeleteProduct)
	s.POST("/sync/orders", sSync.SyncOrders)
	s.POST("/sync/products", sSync.SyncProducts)
	s.GET("/db/products", sDBProd.GetDBProducts)
	s.GET("/db/products/master", sDBProd.GetMasterProducts)

	// Lazada
	l := protected.Group("/lazada")
	l.GET("/orders", lOrder.GetOrders)
	l.GET("/orders/:orderId", lOrder.GetOrderByID)
	l.POST("/orders/ship", lOrder.ShipOrder)
	l.POST("/orders/cancel", lOrder.CancelOrder)
	l.GET("/products", lProd.GetProducts)
	l.GET("/products/:itemId", lProd.GetProductByID)
	l.POST("/products", lProd.CreateProduct)
	l.PUT("/products/:itemId", lProd.UpdateProduct)
	l.DELETE("/products/:itemId", lProd.DeleteProduct)
	l.POST("/sync/orders", lSync.SyncOrders)
	l.POST("/sync/products", lSync.SyncProducts)
	l.GET("/db/products", lDBProd.GetDBProducts)
	l.GET("/db/products/master", lDBProd.GetMasterProducts)

	// TikTok
	tSearch := tiktok.NewProductSearchHandler(cfg.DatabasePath)
	t := protected.Group("/tiktok")
	t.GET("/orders", tOrder.GetOrders)
	t.GET("/orders/:orderId", tOrder.GetOrderByID)
	t.POST("/orders/ship", tOrder.ShipOrder)
	t.POST("/orders/cancel", tOrder.CancelOrder)
	t.GET("/products", tProd.GetProducts)
	t.GET("/products/:productId", tProd.GetProductByID)
	t.POST("/products", tProd.CreateProduct)
	t.PUT("/products/:productId", tProd.UpdateProduct)
	t.DELETE("/products/:productId", tProd.DeleteProduct)
	t.POST("/products/search", tSearch.SearchProducts) // Sync from TikTok API
	t.POST("/sync/orders", tSync.SyncOrders)
	t.POST("/sync/products", tSync.SyncProducts)
	t.GET("/db/products", tDBProd.GetDBProducts)
	t.GET("/db/products/master", tDBProd.GetMasterProducts)

	// Note: Analytics routes are registered via routes.RegisterShopeeAnalyticsRoutes,
	// routes.RegisterTiktokAnalyticsRoutes, and routes.RegisterAdsRoutes above.

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	// Create HTTP server with graceful shutdown support
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Go Backend starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close database connections
	config.CloseAllDBs()

	log.Println("Server exited gracefully")
}
