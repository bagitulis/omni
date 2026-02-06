package app

import (
	"context"
	"os"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/autofunction"
	"github.com/omni/backend/internal/services/google"
	"github.com/omni/backend/internal/services/jobs"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ExtendedHandlers contains additional handlers not in core App
// Handlers requiring complex dependencies (monitoring, security, autofunction)
// are initialized separately when their services are available
type ExtendedHandlers struct {
	// Utility handlers
	CSRFHandler  *handlers.CSRFHandler
	DocsHandler  *handlers.DocsHandler
	RouteHandler *handlers.RouteHandler

	// Business logic handlers
	OrderSyncHandler    *handlers.OrderSyncHandler
	OrderManagerHandler *handlers.OrderManagerHandler
	LockedOrderHandler  *handlers.LockedOrderHandler
	JobQueueHandler     *handlers.JobQueueHandler

	// Auto Function handler
	AutoFunctionHandler *handlers.AutoFunctionHandler

	// Product handlers
	ProductMasterHandler *handlers.ProductMasterHandler
	ProductCreateHandler *handlers.ProductCreateHandler
	ProductCloneHandler  *handlers.ProductCloneHandler
	PriceHandler         *handlers.PriceHandler
	StockHandler         *handlers.StockHandler
	SKUCheckHandler      *handlers.SKUCheckHandler

	// Inventory & Settings
	InventoryHandler            *handlers.InventoryHandler
	SkuBatchCheckHandler        *handlers.SkuBatchCheckHandler
	SettingsHandler             *handlers.SettingsHandler
	RouteConfigHandler          *handlers.RouteConfigHandler
	RouteExecutionConfigHandler *handlers.RouteExecutionConfigHandler

	// Integration handlers
	SpreadsheetRegistryHandler *handlers.SpreadsheetRegistryHandler
	FilterPreferenceHandler    *handlers.FilterPreferenceHandler
	WholesaleHandler           *handlers.WholesaleHandler

	// Ads handlers
	AdsHandler             *handlers.AdsHandler
	TiktokAnalyticsHandler *handlers.TiktokAnalyticsHandler
	ShopeeAnalyticsHandler *handlers.ShopeeAnalyticsHandler

	// Monitoring handler
	MonitoringHandler *handlers.MonitoringHandler

	// API Client factories for platform routes
	ShopeeAPIClientFactory func(tenantID string) shopeeService.APIClient
}

// InitExtendedHandlers initializes extended handlers that use *gorm.DB
// Handlers requiring special services (Monitoring, Security, AutoFunction, GoogleSheets)
// should be initialized separately after their services are created
func (a *App) InitExtendedHandlers(db *gorm.DB, googleAuth *google.AuthService) *ExtendedHandlers {
	routeBasePath := os.Getenv("ROUTE_BASE_PATH")
	if routeBasePath == "" {
		routeBasePath = "./internal"
	}

	basePath := os.Getenv("DATA_PATH")
	if basePath == "" {
		basePath = "./data"
	}

	// Initialize auto function services
	executor := autofunction.NewExecutor(db)
	scheduler := autofunction.NewScheduler(db, executor)

	// Register default auto function handlers (names MUST match database values)
	registerDefaultAutoFunctionHandlers(executor)

	// Create and START multi-tenant scheduler
	// This iterates all tenant schemas every 60 seconds to check/execute due functions
	// Similar to Node.js AutoFunctionScheduler.start()
	multiTenantScheduler := autofunction.NewMultiTenantScheduler(db, executor, basePath)
	multiTenantScheduler.Start()

	// Initialize and start background job executor for long-running operations
	// This processes escrow sync jobs, order sync jobs, etc. that run in background
	startBackgroundJobExecutor(db, basePath)

	// Create Shopee API client factory function
	shopeeAPIClientFactory := createShopeeAPIClientFactory(db, basePath)

	return &ExtendedHandlers{
		// Utility handlers
		CSRFHandler:  handlers.NewCSRFHandler(),
		DocsHandler:  handlers.NewDocsHandler(),
		RouteHandler: handlers.NewRouteHandler(routeBasePath),

		// Business logic handlers
		OrderSyncHandler:    handlers.NewOrderSyncHandler(),
		OrderManagerHandler: handlers.NewOrderManagerHandler(basePath),
		LockedOrderHandler:  handlers.NewLockedOrderHandler(db),
		JobQueueHandler:     handlers.NewJobQueueHandler(db),

		// Auto Function handler
		AutoFunctionHandler: handlers.NewAutoFunctionHandler(db, scheduler, executor),

		// Product handlers
		ProductMasterHandler: handlers.NewProductMasterHandler(db),
		ProductCreateHandler: handlers.NewProductCreateHandler(db),
		ProductCloneHandler:  handlers.NewProductCloneHandler(db),
		PriceHandler:         handlers.NewPriceHandler(db),
		StockHandler:         handlers.NewStockHandler(db),
		SKUCheckHandler:      handlers.NewSKUCheckHandler(db),

		// Inventory & Settings - pass Google Auth to inventory handler
		InventoryHandler:            handlers.NewInventoryHandlerWithGoogle(db, googleAuth),
		SkuBatchCheckHandler:        handlers.NewSkuBatchCheckHandler(db),
		SettingsHandler:             handlers.NewSettingsHandler(db),
		RouteConfigHandler:          handlers.NewRouteConfigHandler(db),
		RouteExecutionConfigHandler: handlers.NewRouteExecutionConfigHandler(db),

		// Integration handlers
		SpreadsheetRegistryHandler: handlers.NewSpreadsheetRegistryHandler(db),
		FilterPreferenceHandler:    handlers.NewFilterPreferenceHandler(db),
		WholesaleHandler:           handlers.NewWholesaleHandler(db),

		// Ads handlers
		AdsHandler:             handlers.NewAdsHandler(db),
		TiktokAnalyticsHandler: handlers.NewTiktokAnalyticsHandler(),
		ShopeeAnalyticsHandler: handlers.NewShopeeAnalyticsHandler(),

		// Monitoring handler
		MonitoringHandler: handlers.NewSimpleMonitoringHandler(),

		// API Client factories
		ShopeeAPIClientFactory: shopeeAPIClientFactory,
	}
}

// registerDefaultAutoFunctionHandlers registers handlers for built-in auto functions
// Handler names MUST match database values (snake_case format)
func registerDefaultAutoFunctionHandlers(executor *autofunction.Executor) {
	// locked_today - locks orders at end of day
	executor.RegisterHandler("locked_today", lockedTodayHandler)

	// auto_update_token - refreshes platform OAuth tokens
	executor.RegisterHandler("auto_update_token", autoUpdateTokenHandler)

	// sync_from_sheets - syncs inventory from Google Sheets
	executor.RegisterHandler("sync_from_sheets", syncFromSheetsHandler)
}

// startBackgroundJobExecutor initializes and starts the multi-tenant job executor
// for processing long-running background jobs like escrow sync
func startBackgroundJobExecutor(systemDB *gorm.DB, basePath string) {
	jobExecutor := jobs.NewMultiTenantExecutor(systemDB, basePath)

	// Create escrow sync handler
	escrowHandler := jobs.NewEscrowSyncHandler(systemDB)

	// Register handlers for escrow sync job types
	jobExecutor.RegisterHandler(models.JobTypeShopeeEscrowSync, escrowHandler.HandleShopeeEscrowSync)
	jobExecutor.RegisterHandler(models.JobTypeTiktokEscrowSync, escrowHandler.HandleTiktokEscrowSync)

	// Start the executor
	jobExecutor.Start()
}

// createShopeeAPIClientFactory creates a factory function that returns Shopee API clients
// for a given tenant. It loads credentials from both tenant DB and system DB.
func createShopeeAPIClientFactory(systemDB *gorm.DB, basePath string) func(tenantID string) shopeeService.APIClient {
	return func(tenantID string) shopeeService.APIClient {
		// Get tenant database
		tenantDB, err := config.GetTenantDB(tenantID, basePath)
		if err != nil {
			log.Error().Err(err).Str("tenant_id", tenantID).Msg("Failed to get tenant DB for Shopee client")
			return nil
		}

		// Get tenant credentials (ShopID, AccessToken)
		credRepo := repositories.NewPlatformCredentialsRepository(tenantDB)
		tenantCreds, err := credRepo.GetShopeeCredentials(context.Background())
		if err != nil {
			log.Error().Err(err).Str("tenant_id", tenantID).Msg("Failed to get Shopee tenant credentials")
			return nil
		}

		// Get global credentials (PartnerID, PartnerKey)
		configRepo := repositories.NewGlobalConfigRepository(systemDB)
		globalCreds, err := configRepo.GetShopeeCredentials(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("Failed to get Shopee global credentials")
			return nil
		}

		// Create and configure client
		client := shopeePkg.NewClient(globalCreds.PartnerID, globalCreds.PartnerKey, true)
		client.SetShopCredentials(tenantCreds.ShopIDInt, tenantCreds.AccessToken)

		return client
	}
}
