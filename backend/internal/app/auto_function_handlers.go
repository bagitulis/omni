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
	"github.com/omni/backend/internal/services/orders"
	"github.com/omni/backend/internal/services/sync"
	"github.com/omni/backend/internal/utils"
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
