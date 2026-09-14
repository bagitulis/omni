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
	"github.com/omni/backend/internal/extensions"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/handlers/lazada"
	"github.com/omni/backend/internal/handlers/shopee"
	"github.com/omni/backend/internal/handlers/tiktok"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/routes"
	googleService "github.com/omni/backend/internal/services/google"
	"gorm.io/gorm"
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

	// Create server context for background services
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()
	handlers.StartSSECleanup(serverCtx)
	middleware.StartCSRFCleanup(serverCtx)

	// Initialize extended handlers with Google Auth
	extHandlers := application.InitExtendedHandlers(serverCtx, application.SystemDB, googleAuthService)

	router := gin.Default()
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	extHandlers.RouteHandler.SetEngine(router)

	// Serve uploaded images (product thumbnails, etc.)
	router.Static("/uploads", "uploads")

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
	routes.RegisterDeveloperRoutes(api, application.DeveloperHandler)
	routes.RegisterAuditRoutes(api, application.AuditHandler)
	routes.RegisterCaptchaRoutes(api, application.CaptchaHandler)
	routes.RegisterOAuthRoutes(api, application.OAuthHandler)
	routes.RegisterTokenRoutes(api, application.TokenHandler)
	routes.RegisterAnalyticsRoutes(api, application.AnalyticsHandler)
	routes.RegisterWebhookRoutes(api, application.WebhookHandler)
	routes.RegisterWebhookAdminRoutes(api, application.WebhookHandler)

	// ====== Platform Auth Routes ======
	routes.RegisterPlatformAuthRoutes(api, application.PlatformAuthHandler)
	routes.RegisterCredentialRoutes(api, application.PlatformAuthHandler)
	routes.RegisterCredentialCallbackRoute(api, application.PlatformAuthHandler)

	// ====== Extended Routes ======
	// Utility routes
	routes.RegisterCSRFRoutes(api, extHandlers.CSRFHandler)
	routes.RegisterDocsRoutes(api, extHandlers.DocsHandler)

	// Business routes
	routes.RegisterOrderSyncRoutes(api, extHandlers.OrderSyncHandler)
	routes.RegisterOrderManagerRoutes(api, extHandlers.OrderManagerHandler)
	routes.RegisterLockedOrderRoutes(api, extHandlers.LockedOrderHandler)
	routes.RegisterJobQueueRoutes(api, extHandlers.JobQueueHandler)

	// Shipping files routes
	shippingFilesHandler := handlers.NewShippingFilesHandler(cfg.DatabasePath)
	routes.RegisterShippingFilesRoutes(api, shippingFilesHandler)

	// Auto Function routes (Script Monitor)
	routes.RegisterAutoFunctionRoutes(api, extHandlers.AutoFunctionHandler)

	// Product routes
	routes.RegisterProductMasterRoutes(api, extHandlers.ProductMasterHandler)
	routes.RegisterProductCreateRoutes(api, extHandlers.ProductCreateHandler)
	routes.RegisterProductCloneRoutes(api, extHandlers.ProductCloneHandler)
	// PriceRoutes and StockRoutes removed (dead code - frontend uses /api/inventory/update-price and /api/inventory/update-stock)
	routes.RegisterSKUCheckRoutes(api, extHandlers.SKUCheckHandler)

	// Master Product routes (unified product management)
	routes.RegisterMasterProductRoutes(api, cfg.DatabasePath)
	routes.RegisterMasterProductImportRoutes(api, cfg.DatabasePath)
	routes.RegisterMasterProductSyncRoutes(api, cfg.DatabasePath)

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

	// Image Gallery routes

	// Image Gallery routes
	imageHandler := handlers.NewImageHandler(application.SystemDB)
	routes.RegisterImageRoutes(api, imageHandler)

	// Marketplace Sync History routes
	routes.RegisterMarketplaceSyncHistoryRoutes(api, cfg.DatabasePath)

	// ====== Google Routes ======
	// Use the already initialized googleAuthService
	googleQuotaService := googleService.NewQuotaService()
	googleHandlers := routes.NewGoogleHandlers(googleAuthService, googleQuotaService, application.SystemDB)
	routes.RegisterGoogleRoutes(api, googleHandlers)

	// ====== Monitoring Routes ======
	routes.RegisterMonitoringRoutes(api, extHandlers.MonitoringHandler)

	// ====== Notification Routes (Facebook-style persistent) ======
	routes.RegisterNotificationRoutes(api, extHandlers.NotificationHandler)

	// ====== Extensions (browser automation) ======
	// The hub owns its own event loop; the context is cancelled on shutdown so
	// live extension sockets are closed rather than leaking goroutines.
	extHub := extensions.NewHub(nil)
	extCtx, extCancel := context.WithCancel(serverCtx)
	go extHub.Run(extCtx)
	defer extCancel()

	tenantDBFor := func(tenantID string) (*gorm.DB, error) {
		return config.GetTenantDBWithContext(tenantID, cfg.DatabasePath)
	}

	extService := extensions.NewService(tenantDBFor, extHub)

	// A confirming browser holds no JWT — the pairing code is its only handle —
	// so the code is resolved across tenants. Codes are unique by construction
	// (40 bits of CSPRNG, single-use, 5-minute TTL).
	extService.SetPairingCodeResolver(func(ctx context.Context, code string) (string, error) {
		return extensions.ResolveTenantForPairingCode(ctx, code, tenantDBFor, config.GetRealTenants)
	})
	extService.SetTokenResolver(func(ctx context.Context, token string) (string, *models.Extension, error) {
		return extensions.ResolveExtensionByToken(ctx, token, tenantDBFor, config.GetRealTenants)
	})

	extHandler := handlers.NewExtensionHandler(extService)
	extWSConfig := handlers.BuildExtensionWSConfig(extService, extHub, []string{})
	routes.RegisterExtensionRoutes(api, extHandler, handlers.NewExtensionWSHandler(extWSConfig))

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
	s.GET("/db/products/list", sDBProd.GetProductList)
	s.GET("/db/products/base", sDBProd.GetProductBase)
	s.GET("/db/products/model", sDBProd.GetProductModel)
	s.GET("/db/products/base/:itemId", sDBProd.GetProductBaseByID)
	s.GET("/db/products/models/:itemId", sDBProd.GetProductModelsByID)
	s.GET("/db/products/variations/:itemId", sDBProd.GetProductVariationsByID)
	s.GET("/db/products/full/:itemId", sDBProd.GetProductFull)
	s.GET("/db/products/search", sDBProd.SearchProducts)
	s.GET("/db/products/status/:status", sDBProd.GetProductsByStatus)
	s.GET("/db/stats", sDBProd.GetDBStats)
	s.GET("/db/sync/unprocessed", sDBProd.GetUnprocessedItems)
	s.GET("/db/sync/no-models", sDBProd.GetItemsWithoutModels)
	s.GET("/db/sync/logs", sDBProd.GetSyncLogs)

	// Lazada
	l := protected.Group("/lazada")
	l.GET("/orders", lOrder.GetOrders)
	l.GET("/orders/:orderId", lOrder.GetOrderByID)
	l.POST("/orders/ship", lOrder.ShipOrder)
	l.POST("/orders/cancel", lOrder.CancelOrder)
	l.POST("/orders/document", lOrder.GetDocument)
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
	t.POST("/sync/orders", tSync.SyncOrders)
	t.POST("/sync/products", tSync.SyncProducts)
	t.GET("/db/products", tDBProd.GetDBProducts)
	t.GET("/db/products/master", tDBProd.GetMasterProducts)
	t.GET("/db/products/list", tDBProd.GetProductList)
	t.GET("/db/products/status/:status", tDBProd.GetProductsByStatus)
	t.GET("/db/products/:productId", tDBProd.GetProductByID)
	t.DELETE("/db/products/:productId", tDBProd.DeleteProduct)
	t.GET("/db/search", tDBProd.SearchProducts)
	t.GET("/db/statistics", tDBProd.GetStatistics)

	// ====== Shipping Routes ======
	// Shopee Shipping routes (arrange pickup, get label, tracking)
	routes.RegisterShopeeShippingRoutes(api, extHandlers.ShopeeAPIClientFactory)
	routes.RegisterShopeeShippingFeeRoutes(api, extHandlers.ShopeeAPIClientFactory, googleAuthService)

	// Shopee Wallet & Escrow routes
	routes.RegisterShopeeWalletRoutes(api, extHandlers.ShopeeAPIClientFactory, cfg.DatabasePath)
	routes.RegisterShopeeWalletReportRoutes(api, extHandlers.ShopeeAPIClientFactory, googleAuthService)
	routes.RegisterShopeeEscrowRoutes(api, extHandlers.ShopeeAPIClientFactory)

	// TikTok Shipping routes (shipping document/label)
	routes.RegisterTiktokShippingRoutes(api, cfg.DatabasePath)

	// ====== Extended Platform Routes ======
	// Lazada Product Extended routes (db products, categories, attributes)
	routes.RegisterLazadaProductExtendedRoutes(api, cfg.DatabasePath)

	// TikTok Product Extended routes (search, draft, db, categories, compliance, image)
	routes.RegisterTiktokProductExtendedRoutes(api, cfg.DatabasePath)

	// NOTE: ProductMetadataHandler removed — all 14 endpoints were 501 skeletons with 0 frontend callers.
	// Platform-specific endpoints already serve this functionality:
	//   - Shopee: /api/products/categories/:platform (product_create.go Layer 1)
	//   - Lazada: /api/lazada/products/categories, /attributes (product_extended_handler.go Layer 3)
	//   - TikTok: /api/tiktok/products/categories, /attributes, /warehouses (product_create_handler.go Layer 3)

	// Webhook Extended routes (tenant-specific webhooks, test, config)
	routes.RegisterWebhookExtendedRoutes(api, application.ShopeeProcessor, application.LazadaProcessor, application.TiktokProcessor, cfg.DatabasePath)

	// ====== Shopee/TikTok Analytics Report Routes ======
	routes.RegisterShopeeAnalyticsRoutes(api, application.ShopeeAnalyticsHandler)
	routes.RegisterTiktokAnalyticsRoutes(api, application.TiktokAnalyticsHandler)

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

	// Signal all background services to stop via context cancellation
	serverCancel()
	handlers.StopSSECleanup()
	middleware.StopCSRFCleanup()

	// Stop job executor gracefully
	if extHandlers.JobExecutor != nil {
		log.Println("Stopping job executor...")
		extHandlers.JobExecutor.Stop()
		log.Println("Job executor stopped")
	}

	// Create context with timeout for HTTP server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Shutdown HTTP server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close database connections
	config.CloseAllDBs()

	log.Println("Server exited gracefully")
}
