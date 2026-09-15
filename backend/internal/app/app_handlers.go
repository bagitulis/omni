package app

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/notify"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/services/analytics"
	"github.com/omni/backend/internal/services/autofunction"
	"github.com/omni/backend/internal/services/google"
	"github.com/omni/backend/internal/services/jobs"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// notifyFanout is the process-wide fanout for notification events. Every
// request creates a fresh Bus (tenant-scoped) but they all share this fanout
// so subscribers on any goroutine see events from every emitter. In-process
// is fine for single-replica; swap to notify.NewRedisFanout() when scaling
// beyond one API replica.
var notifyFanout = notify.NewInProcessFanout()

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
	// PriceHandler and StockHandler removed (dead code - frontend uses /api/inventory/update-price instead)
	SKUCheckHandler *handlers.SKUCheckHandler

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
	// Notification handler
	NotificationHandler *handlers.NotificationHandler

	// Monitoring handler
	MonitoringHandler *handlers.MonitoringHandler
	// API Client factories for platform routes
	ShopeeAPIClientFactory func(tenantID string) shopeeService.APIClient
	// Background services
	JobExecutor *jobs.MultiTenantExecutor
}

// InitExtendedHandlers initializes extended handlers that use *gorm.DB
// Handlers requiring special services (Monitoring, Security, AutoFunction, GoogleSheets)
// should be initialized separately after their services are created
func (a *App) InitExtendedHandlers(ctx context.Context, db *gorm.DB, googleAuth *google.AuthService) *ExtendedHandlers {
	routeBasePath := os.Getenv("ROUTE_BASE_PATH")
	if routeBasePath == "" {
		routeBasePath = "./internal"
	}

	basePath := os.Getenv("UPLOAD_PATH")
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
	jobExecutor := startBackgroundJobExecutor(ctx, db, basePath)

	// Startup sync: run product sync for all tenants after a short delay
	// This catches up on any changes that happened while server was offline
	go runStartupSync(ctx, db, executor, basePath)

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
		// Notification handler — factory encapsulates tenant DB + service construction.
		// V2: attaches a notify.Bus so Emit/EmitWithKey go through sanitizer +
		// dedup UPSERT + fanout. Fanout impl defaults to in-process; swap to
		// notify.NewRedisFanout() once multi-replica is enabled.
		NotificationHandler: handlers.NewNotificationHandler(func(c *gin.Context) (*services.NotificationService, error) {
			db, err := handlers.GetTenantDBFromContext(c, db)
			if err != nil {
				return nil, err
			}
			tenantID := c.GetString("tenant_id")
			repo := repositories.NewNotificationRepository(db)
			bus := notify.NewBus(repo, notifyFanout)
			return services.NewNotificationService(repo).WithTenant(tenantID).WithBus(bus), nil
		}),

		// Monitoring handler
		MonitoringHandler: handlers.NewMonitoringHandler(db, nil),
		// API Client factories
		ShopeeAPIClientFactory: shopeeAPIClientFactory,

		// Background services
		JobExecutor: jobExecutor,
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

	// sync_products - full product sync from all platforms (Shopee, TikTok, Lazada)
	executor.RegisterHandler("sync_products", syncProductsHandler)

	// sync_products_inventory - syncs only products whose SKUs exist in inventory
	executor.RegisterHandler("sync_products_inventory", syncProductsInventoryHandler)

	// retry_failed_syncs - retries failed marketplace sync operations from last 24h
	executor.RegisterHandler("retry_failed_syncs", retryFailedSyncsHandler)

	// price_drift_detection - detects SKUs with price drift
	executor.RegisterHandler("price_drift_detection", priceDriftDetectionHandler)
}

// startBackgroundJobExecutor initializes and starts the multi-tenant job executor
// for processing long-running background jobs
func startBackgroundJobExecutor(ctx context.Context, systemDB *gorm.DB, basePath string) *jobs.MultiTenantExecutor {
	jobExecutor := jobs.NewMultiTenantExecutor(ctx, systemDB, basePath)

	// Register escrow sync handlers
	// EscrowSyncHandler uses factory functions because tenantDB is per-request
	escrowHandler := jobs.NewEscrowSyncHandler(systemDB)
	escrowHandler.SetShopeeSyncServiceFactory(func(systemDB, tenantDB *gorm.DB, tenantID string) jobs.EscrowSyncService {
		return analytics.NewShopeeEscrowSyncService(systemDB, tenantDB, tenantID)
	})
	escrowHandler.SetTiktokSyncServiceFactory(func(systemDB, tenantDB *gorm.DB, tenantID string) jobs.EscrowSyncService {
		return analytics.NewTiktokEscrowSyncService(systemDB, tenantDB, tenantID)
	})
	jobExecutor.RegisterHandler(models.JobTypeShopeeEscrowSync, escrowHandler.HandleShopeeEscrowSync)
	jobExecutor.RegisterHandler(models.JobTypeTiktokEscrowSync, escrowHandler.HandleTiktokEscrowSync)

	// Start the executor
	jobExecutor.Start()
	return jobExecutor
}

// createShopeeAPIClientFactory creates a factory function that returns Shopee API clients
// for a given tenant. It loads credentials through CredentialService canonical-first lookup.
func createShopeeAPIClientFactory(systemDB *gorm.DB, basePath string) func(tenantID string) shopeeService.APIClient {
	return func(tenantID string) shopeeService.APIClient {
		_ = systemDB
		credService := services.NewCredentialService(basePath)
		creds, err := credService.GetPlatformCredentials(tenantID, "shopee")
		if err != nil {
			log.Error().Err(err).Str("tenant_id", tenantID).Msg("Failed to get Shopee credentials")
			return nil
		}

		// Create and configure client
		client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
		client.SetShopCredentials(creds.ShopID, creds.AccessToken)

		return client
	}
}
