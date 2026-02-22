package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/services/google"
	"github.com/omni/backend/internal/services/inventory"
	lazadaService "github.com/omni/backend/internal/services/lazada"
	masterProductService "github.com/omni/backend/internal/services/master_product"
	"github.com/omni/backend/internal/services/orders"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	"github.com/omni/backend/internal/services/sync"
	tiktokService "github.com/omni/backend/internal/services/tiktok"
	"github.com/omni/backend/internal/utils"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"gorm.io/gorm"
)

// lockedTodayHandler handles the "Locked Today" auto function
// This syncs orders and aggregates locked orders for the day
// Time-based logic:
// - 00:00-14:00 (before 2pm Jakarta): unprocess + processed orders
// - 14:00-24:00 (after 2pm Jakarta): unprocess orders only
func lockedTodayHandler(ctx context.Context, tenantID string, cfg *models.AutoFunctionConfig) (string, error) {
	log.Printf("[AutoFunction] Running 'Locked Today' for tenant: %s", tenantID)

	basePath := os.Getenv("DATA_PATH")
	if basePath == "" {
		basePath = "./data"
	}

	// Get tenant DB
	db, err := config.GetTenantDBWithContext(tenantID, basePath)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant DB: %w", err)
	}

	// Get order sync service
	syncService, err := sync.GetOrderSyncService(tenantID)
	if err != nil {
		log.Printf("[AutoFunction] Order sync service not available for tenant %s: %v", tenantID, err)
		return "Order sync service not available: " + err.Error(), nil
	}

	// Time-based logic: Check current hour (Jakarta timezone UTC+7)
	now := time.Now().UTC()
	jakartaHour := (now.Hour() + 7) % 24
	includeProcessed := jakartaHour < 14 // Before 2pm include processed

	log.Printf("[AutoFunction] Locked Today - tenant: %s, jakarta_hour: %d, include_processed: %v",
		tenantID, jakartaHour, includeProcessed)

	// Sync and get unprocess orders
	_, err = syncService.SyncByCategory(ctx, sync.OrderStatusCategory("unprocess"), 7, nil)
	if err != nil {
		log.Printf("[AutoFunction] Failed to sync unprocess orders: %v", err)
	}
	unprocessOrders, _ := syncService.GetOrdersByCategory(ctx, sync.OrderStatusCategory("unprocess"), nil)

	// Sync and get processed orders (only before 2pm)
	var processedOrders []sync.Order
	if includeProcessed {
		_, err = syncService.SyncByCategory(ctx, sync.OrderStatusCategory("processed"), 7, nil)
		if err != nil {
			log.Printf("[AutoFunction] Failed to sync processed orders: %v", err)
		}
		processedOrders, _ = syncService.GetOrdersByCategory(ctx, sync.OrderStatusCategory("processed"), nil)
	}

	// Aggregate locked orders by SKU + ProductName + Variation
	lockedItems := aggregateLockedOrdersForAutoFunc(unprocessOrders, processedOrders)

	// Save to database
	lockedService := orders.NewLockedOrderService(db)
	savedItems := make([]orders.LockedOrderItem, len(lockedItems))
	for i, item := range lockedItems {
		savedItems[i] = orders.LockedOrderItem{
			SKU:           item.SKU,
			ProductName:   item.ProductName,
			VariationName: item.VariationName,
			Qty:           item.Qty,
		}
	}

	savedCount, err := lockedService.SaveLockedOrders(ctx, tenantID, savedItems)
	if err != nil {
		return "", fmt.Errorf("failed to save locked orders: %w", err)
	}

	totalQty, _ := lockedService.GetTotalQty(ctx, tenantID)

	mode := "unprocess-only"
	if includeProcessed {
		mode = "unprocess+processed"
	}

	result := fmt.Sprintf("Locked Today completed: %d SKUs, %d total qty (mode: %s)", savedCount, totalQty, mode)
	log.Printf("[AutoFunction] %s for tenant: %s", result, tenantID)

	return result, nil
}

// aggregatedItem for locked orders aggregation
type aggregatedItem struct {
	SKU           string
	ProductName   string
	VariationName string
	Qty           int
}

// aggregateLockedOrdersForAutoFunc aggregates orders by SKU + ProductName + Variation
func aggregateLockedOrdersForAutoFunc(unprocessOrders, processedOrders []sync.Order) []aggregatedItem {
	aggregated := make(map[string]*aggregatedItem)

	processOrder := func(order sync.Order) {
		for _, item := range order.Items {
			key := fmt.Sprintf("%s|%s|%s", item.SKU, item.ProductName, item.VariationName)
			if existing, ok := aggregated[key]; ok {
				existing.Qty += item.Quantity
			} else {
				aggregated[key] = &aggregatedItem{
					SKU:           item.SKU,
					ProductName:   item.ProductName,
					VariationName: item.VariationName,
					Qty:           item.Quantity,
				}
			}
		}
	}

	for _, order := range unprocessOrders {
		processOrder(order)
	}
	for _, order := range processedOrders {
		processOrder(order)
	}

	result := make([]aggregatedItem, 0, len(aggregated))
	for _, item := range aggregated {
		result = append(result, *item)
	}
	return result
}

// autoUpdateTokenHandler handles the "Auto Update Token" auto function
// This refreshes OAuth tokens for all platforms before they expire
func autoUpdateTokenHandler(ctx context.Context, tenantID string, cfg *models.AutoFunctionConfig) (string, error) {
	log.Printf("[AutoFunction] Running 'Auto Update Token' for tenant: %s", tenantID)

	basePath := os.Getenv("DATA_PATH")
	if basePath == "" {
		basePath = "./data"
	}

	// Get system DB for global config
	systemDB, err := config.GetSystemDB(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to get system DB: %w", err)
	}

	// Initialize encryption service
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		return "", fmt.Errorf("ENCRYPTION_KEY not set")
	}
	encryption, err := utils.NewEncryptionService(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create encryption service: %w", err)
	}

	// Create token manager
	globalConfigRepo := repositories.NewGlobalConfigRepository(systemDB)
	tokenManager := services.NewTokenManager(globalConfigRepo, encryption, basePath)

	// Refresh expired tokens
	results, err := tokenManager.RefreshExpiredTokens(ctx, tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to refresh tokens: %w", err)
	}

	// Count successes and failures
	successCount := 0
	failCount := 0
	for platform, success := range results {
		if success {
			successCount++
			log.Printf("[AutoFunction] %s token refreshed for tenant: %s", platform, tenantID)
		} else {
			failCount++
			log.Printf("[AutoFunction] %s token refresh failed for tenant: %s", platform, tenantID)
		}
	}

	// If no tokens needed refresh, check status
	if len(results) == 0 {
		statuses, _ := tokenManager.GetAllTokenStatus(ctx, tenantID)
		validCount := 0
		for _, status := range statuses {
			if status != nil && status.IsValid && !status.NeedsRefresh {
				validCount++
			}
		}
		return fmt.Sprintf("No tokens needed refresh. %d tokens are valid.", validCount), nil
	}

	result := fmt.Sprintf("Token refresh completed: %d succeeded, %d failed", successCount, failCount)
	log.Printf("[AutoFunction] %s for tenant: %s", result, tenantID)

	return result, nil
}

// syncFromSheetsHandler handles the "Sync From Sheets" auto function
// This syncs inventory data from Google Sheets
func syncFromSheetsHandler(ctx context.Context, tenantID string, cfg *models.AutoFunctionConfig) (string, error) {
	log.Printf("[AutoFunction] Running 'Sync From Sheets' for tenant: %s", tenantID)

	basePath := os.Getenv("DATA_PATH")
	if basePath == "" {
		basePath = "./data"
	}

	// Get tenant DB
	db, err := config.GetTenantDBWithContext(tenantID, basePath)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant DB: %w", err)
	}

	// Get spreadsheet config: try inventory_settings first, then google_sheets_settings (same as manual sync)
	invService := inventory.NewInventoryService(db, tenantID)
	settings, err := invService.GetSettings(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get inventory settings: %w", err)
	}

	spreadsheetID := ""
	sheetName := ""

	// Source 1: inventory_settings
	if settings != nil {
		spreadsheetID = settings.SpreadsheetID
		sheetName = settings.SheetName
	}

	// Source 2: google_sheets_settings (fallback, same as manual sync handler)
	if spreadsheetID == "" || sheetName == "" {
		var gsSettings models.GoogleSheetsSettings
		if err := db.First(&gsSettings).Error; err == nil {
			if spreadsheetID == "" && gsSettings.InventorySpreadsheetID != "" {
				spreadsheetID = gsSettings.InventorySpreadsheetID
			}
			if sheetName == "" && gsSettings.InventorySheetName != "" {
				sheetName = gsSettings.InventorySheetName
			}
		}
	}

	if spreadsheetID == "" || sheetName == "" {
		return "", fmt.Errorf("inventory spreadsheet not configured: spreadsheet_id=%q, sheet_name=%q. Please configure in Settings > Google Sheets", spreadsheetID, sheetName)
	}

	// Load Google service account credentials
	saPath := google.GetDefaultServiceAccountPath()
	saLoader := google.NewServiceAccountLoader(saPath)
	credentials, err := saLoader.LoadFirstAvailable()
	if err != nil {
		return "", fmt.Errorf("failed to load Google service account: %w", err)
	}

	// Create Google Auth and Sheets services
	googleAuth := google.NewAuthService(credentials)
	sheetsService := google.NewSheetsService(googleAuth, tenantID)

	// Create sync service
	syncService := inventory.NewSyncService(db, tenantID, sheetsService)

	// Perform sync
	result, err := syncService.SyncFromSheets(ctx, spreadsheetID, sheetName)
	if err != nil {
		return "", fmt.Errorf("sync from sheets failed: %w", err)
	}

	if result.Status == "ERROR" {
		return "", fmt.Errorf("sync from sheets error: %s", result.Message)
	}

	resultMsg := fmt.Sprintf("Sync completed: %d total, %d new, %d updated, %d unchanged, %d failed (duration: %dms)",
		result.TotalRecords, result.NewRecords, result.UpdatedRecords,
		result.UnchangedRecords, result.FailedRecords, result.Duration)

	log.Printf("[AutoFunction] %s for tenant: %s", resultMsg, tenantID)

	return resultMsg, nil
}

// syncProductsHandler handles the "Sync Products" auto function.
// Syncs products from all 3 platforms (Shopee, TikTok, Lazada) to staging tables.
// Each platform sync is independent — failures are logged but don't block others.
func syncProductsHandler(ctx context.Context, tenantID string, cfg *models.AutoFunctionConfig) (string, error) {
	log.Printf("[AutoFunction] Running 'Sync Products' for tenant: %s", tenantID)

	basePath := os.Getenv("DATA_PATH")
	if basePath == ""  {
		basePath = "./data"
	}

	db, err := config.GetTenantDBWithContext(tenantID, basePath)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant DB: %w", err)
	}

	systemDB, err := config.GetSystemDB(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to get system DB: %w", err)
	}

	results := make([]string, 0, 3)

	// Shopee product sync
	shopeeCount, shopeeErr := syncPlatformProducts(ctx, "shopee", tenantID, db, systemDB, basePath)
	if shopeeErr != nil {
		log.Printf("[AutoFunction] Shopee sync failed for %s: %v", tenantID, shopeeErr)
		results = append(results, fmt.Sprintf("Shopee: error (%v)", shopeeErr))
	} else {
		results = append(results, fmt.Sprintf("Shopee: %d products", shopeeCount))
	}

	// TikTok product sync
	tiktokCount, tiktokErr := syncPlatformProducts(ctx, "tiktok", tenantID, db, systemDB, basePath)
	if tiktokErr != nil {
		log.Printf("[AutoFunction] TikTok sync failed for %s: %v", tenantID, tiktokErr)
		results = append(results, fmt.Sprintf("TikTok: error (%v)", tiktokErr))
	} else {
		results = append(results, fmt.Sprintf("TikTok: %d products", tiktokCount))
	}

	// Lazada product sync
	lazadaCount, lazadaErr := syncPlatformProducts(ctx, "lazada", tenantID, db, systemDB, basePath)
	if lazadaErr != nil {
		log.Printf("[AutoFunction] Lazada sync failed for %s: %v", tenantID, lazadaErr)
		results = append(results, fmt.Sprintf("Lazada: error (%v)", lazadaErr))
	} else {
		results = append(results, fmt.Sprintf("Lazada: %d products", lazadaCount))
	}

	resultMsg := fmt.Sprintf("Product sync completed: %s", fmt.Sprintf("%v", results))
	log.Printf("[AutoFunction] %s for tenant: %s", resultMsg, tenantID)

	// Auto-link: refresh platform links for all master SKUs after sync
	autoLinkAllSkus(ctx, db, tenantID)

	// Return error only if ALL platforms failed
	if shopeeErr != nil && tiktokErr != nil && lazadaErr != nil {
		return resultMsg, fmt.Errorf("all platform syncs failed")
	}

	return resultMsg, nil
}

// syncPlatformProducts syncs products for a single platform using existing sync services.
func syncPlatformProducts(ctx context.Context, platform, tenantID string, db, systemDB *gorm.DB, basePath string) (int, error) {
	credRepo := repositories.NewPlatformCredentialsRepository(db)
	globalConfigRepo := repositories.NewGlobalConfigRepository(systemDB)

	switch platform {
	case "shopee":
		tenantCreds, err := credRepo.GetShopeeCredentials(ctx)
		if err != nil {
			return 0, fmt.Errorf("no shopee credentials: %w", err)
		}
		globalCreds, err := globalConfigRepo.GetShopeeCredentials(ctx)
		if err != nil {
			return 0, fmt.Errorf("no shopee global config: %w", err)
		}
		client := shopeePkg.NewClient(globalCreds.PartnerID, globalCreds.PartnerKey, true)
		client.SetShopCredentials(tenantCreds.ShopIDInt, tenantCreds.AccessToken)
		svc := shopeeService.NewProductSyncService(client, db, tenantID)
		return svc.SyncProducts(ctx)

	case "tiktok":
		tenantCreds, err := credRepo.GetTiktokCredentials(ctx)
		if err != nil {
			return 0, fmt.Errorf("no tiktok credentials: %w", err)
		}
		appKey := tenantCreds.AppKey
		appSecret := tenantCreds.AppSecret
		if appKey == "" || appSecret == "" {
			globalCreds, err := globalConfigRepo.GetTiktokCredentials(ctx)
			if err != nil {
				return 0, fmt.Errorf("no tiktok global config: %w", err)
			}
			appKey = globalCreds.AppKey
			appSecret = globalCreds.AppSecret
		}
		client := tiktokPkg.NewClient(appKey, appSecret)
		client.SetCredentials(tenantCreds.AccessToken, tenantCreds.ShopCipher)
		svc := tiktokService.NewSyncServiceWithTenant(client, db, tenantID)
		return svc.SyncProducts(ctx)

	case "lazada":
		tenantCreds, err := credRepo.GetLazadaCredentials(ctx)
		if err != nil {
			return 0, fmt.Errorf("no lazada credentials: %w", err)
		}
		globalCreds, err := globalConfigRepo.GetLazadaCredentials(ctx)
		if err != nil {
			return 0, fmt.Errorf("no lazada global config: %w", err)
		}
		client := lazadaPkg.NewClient(globalCreds.AppKey, globalCreds.AppSecret, "ID")
		client.SetAccessToken(tenantCreds.AccessToken)
		svc := lazadaService.NewSyncServiceWithTenant(client, db, tenantID)
		return svc.SyncProducts(ctx)

	default:
		return 0, fmt.Errorf("unknown platform: %s", platform)
	}
}

// autoLinkAllSkus refreshes platform links for all master SKUs in a tenant.
// Called after product sync to ensure newly synced products get linked.
func autoLinkAllSkus(ctx context.Context, db *gorm.DB, tenantID string) {
	var sellerSkus []string
	err := db.WithContext(ctx).
		Model(&models.MasterProductSku{}).
		Where("tenant_id = ?", tenantID).
		Pluck("seller_sku", &sellerSkus).Error
	if err != nil {
		log.Printf("[AutoFunction] Failed to load master SKUs for auto-link: %v", err)
		return
	}
	if len(sellerSkus) == 0 {
		return
	}

	mapper := masterProductService.NewSkuMapper(db, tenantID)
	result, err := mapper.AutoMapAndLinkBySkus(ctx, sellerSkus)
	if err != nil {
		log.Printf("[AutoFunction] Auto-link failed for %s: %v", tenantID, err)
		return
	}
	log.Printf("[AutoFunction] Auto-link for %s: %d mapped, %d skipped", tenantID, result.MappedCount, result.SkippedCount)
}
