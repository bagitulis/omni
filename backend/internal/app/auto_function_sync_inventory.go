package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	lazadaService "github.com/omni/backend/internal/services/lazada"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	tiktokService "github.com/omni/backend/internal/services/tiktok"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"gorm.io/gorm"
)

// syncProductsInventoryHandler syncs ONLY products whose SKUs exist in inventory.
// Unlike sync_products (full sync), this filters by inventory and uses SyncProductsByIDs
// to avoid fetching all products from each platform, saving API calls and time.
func syncProductsInventoryHandler(ctx context.Context, tenantID string, cfg *models.AutoFunctionConfig) (string, error) {
	log.Printf("[AutoFunction] Running 'Sync Products (Inventory Only)' for tenant: %s", tenantID)

	basePath := os.Getenv("DATA_PATH")
	if basePath == "" {
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

	// 1. Get all unique SKUs from inventory
	inventorySkus := getInventorySkus(ctx, db, tenantID)
	if len(inventorySkus) == 0 {
		return "No inventory SKUs found — nothing to sync", nil
	}
	log.Printf("[AutoFunction] Found %d inventory SKUs for tenant: %s", len(inventorySkus), tenantID)

	credRepo := repositories.NewPlatformCredentialsRepository(db)
	globalConfigRepo := repositories.NewGlobalConfigRepository(systemDB)
	results := make([]string, 0, 3)

	// 2. Sync each platform using only inventory-matched item IDs
	shopeeResult := syncShopeeByInventory(ctx, tenantID, db, inventorySkus, credRepo, globalConfigRepo)
	results = append(results, shopeeResult)

	tiktokResult := syncTiktokByInventory(ctx, tenantID, db, inventorySkus, credRepo, globalConfigRepo)
	results = append(results, tiktokResult)

	lazadaResult := syncLazadaByInventory(ctx, tenantID, db, inventorySkus, credRepo, globalConfigRepo)
	results = append(results, lazadaResult)

	// Auto-fix platform links after sync
	autoFixPlatformLinks(ctx, db, tenantID)

	resultMsg := fmt.Sprintf("Inventory sync completed: %v", results)
	log.Printf("[AutoFunction] %s for tenant: %s", resultMsg, tenantID)
	return resultMsg, nil
}

// getInventorySkus returns all unique SKU values from the inventory table.
func getInventorySkus(ctx context.Context, db *gorm.DB, tenantID string) []string {
	var skus []string
	if err := db.WithContext(ctx).
		Model(&models.InventoryRecord{}).
		Where("tenant_id = ? AND key_value != ''", tenantID).
		Distinct("key_value").
		Pluck("key_value", &skus).Error; err != nil {
		log.Printf("[AutoFunction] Failed to load inventory SKUs: %v", err)
		return nil
	}
	return skus
}

// syncShopeeByInventory finds Shopee items matching inventory SKUs and syncs them.
func syncShopeeByInventory(
	ctx context.Context, tenantID string, db *gorm.DB,
	inventorySkus []string,
	credRepo *repositories.PlatformCredentialsRepository,
	globalConfigRepo *repositories.GlobalConfigRepository,
) string {
	// Find Shopee item IDs where seller_sku matches inventory
	var itemIDs []int64
	if err := db.WithContext(ctx).
		Table("shopee_skus").
		Where("tenant_id = ? AND seller_sku IN ?", tenantID, inventorySkus).
		Distinct("item_id").
		Pluck("item_id", &itemIDs).Error; err != nil {
		return fmt.Sprintf("Shopee: error finding items (%v)", err)
	}
	if len(itemIDs) == 0 {
		return "Shopee: 0 items matched inventory"
	}

	tenantCreds, err := credRepo.GetShopeeCredentials(ctx)
	if err != nil {
		return fmt.Sprintf("Shopee: no credentials (%v)", err)
	}
	globalCreds, err := globalConfigRepo.GetShopeeCredentials(ctx)
	if err != nil {
		return fmt.Sprintf("Shopee: no global config (%v)", err)
	}

	client := shopeePkg.NewClient(globalCreds.PartnerID, globalCreds.PartnerKey, true)
	client.SetShopCredentials(tenantCreds.ShopIDInt, tenantCreds.AccessToken)
	svc := shopeeService.NewProductSyncService(client, db, tenantID)

	count, syncErr := svc.SyncProductsByIDs(ctx, itemIDs)
	if syncErr != nil {
		return fmt.Sprintf("Shopee: error (%v)", syncErr)
	}
	return fmt.Sprintf("Shopee: %d/%d items synced", count, len(itemIDs))
}

// syncTiktokByInventory finds TikTok products matching inventory SKUs and syncs them.
func syncTiktokByInventory(
	ctx context.Context, tenantID string, db *gorm.DB,
	inventorySkus []string,
	credRepo *repositories.PlatformCredentialsRepository,
	globalConfigRepo *repositories.GlobalConfigRepository,
) string {
	// TikTok: seller_sku → product_id (FK to tiktok_products.id) → tiktok_products.product_id (string)
	var productDBIDs []uint
	if err := db.WithContext(ctx).
		Table("tiktok_skus").
		Where("tenant_id = ? AND seller_sku IN ?", tenantID, inventorySkus).
		Distinct("product_id").
		Pluck("product_id", &productDBIDs).Error; err != nil {
		return fmt.Sprintf("TikTok: error finding items (%v)", err)
	}
	if len(productDBIDs) == 0 {
		return "TikTok: 0 items matched inventory"
	}

	// Resolve DB IDs to TikTok product_id strings
	var productIDs []string
	if err := db.WithContext(ctx).
		Table("tiktok_products").
		Where("id IN ? AND tenant_id = ?", productDBIDs, tenantID).
		Pluck("product_id", &productIDs).Error; err != nil {
		return fmt.Sprintf("TikTok: error resolving product IDs (%v)", err)
	}
	if len(productIDs) == 0 {
		return "TikTok: 0 products resolved"
	}

	tenantCreds, err := credRepo.GetTiktokCredentials(ctx)
	if err != nil {
		return fmt.Sprintf("TikTok: no credentials (%v)", err)
	}
	appKey := tenantCreds.AppKey
	appSecret := tenantCreds.AppSecret
	if appKey == "" || appSecret == "" {
		globalCreds, err := globalConfigRepo.GetTiktokCredentials(ctx)
		if err != nil {
			return fmt.Sprintf("TikTok: no global config (%v)", err)
		}
		appKey = globalCreds.AppKey
		appSecret = globalCreds.AppSecret
	}

	client := tiktokPkg.NewClient(appKey, appSecret)
	client.SetCredentials(tenantCreds.AccessToken, tenantCreds.ShopCipher)
	svc := tiktokService.NewSyncServiceWithTenant(client, db, tenantID)

	count, syncErr := svc.SyncProductsByIDs(ctx, productIDs)
	if syncErr != nil {
		return fmt.Sprintf("TikTok: error (%v)", syncErr)
	}
	return fmt.Sprintf("TikTok: %d/%d products synced", count, len(productIDs))
}

// syncLazadaByInventory finds Lazada items matching inventory SKUs and syncs them.
func syncLazadaByInventory(
	ctx context.Context, tenantID string, db *gorm.DB,
	inventorySkus []string,
	credRepo *repositories.PlatformCredentialsRepository,
	globalConfigRepo *repositories.GlobalConfigRepository,
) string {
	// Lazada: seller_sku OR shop_sku → item_id
	var itemIDStrs []string
	if err := db.WithContext(ctx).
		Table("lazada_skus").
		Where("tenant_id = ? AND (seller_sku IN ? OR shop_sku IN ?)", tenantID, inventorySkus, inventorySkus).
		Distinct("item_id").
		Pluck("item_id", &itemIDStrs).Error; err != nil {
		return fmt.Sprintf("Lazada: error finding items (%v)", err)
	}
	if len(itemIDStrs) == 0 {
		return "Lazada: 0 items matched inventory"
	}

	// Convert string item_ids to int64
	var itemIDs []int64
	for _, s := range itemIDStrs {
		id, err := strconv.ParseInt(s, 10, 64)
		if err == nil && id > 0 {
			itemIDs = append(itemIDs, id)
		}
	}
	if len(itemIDs) == 0 {
		return "Lazada: 0 valid item IDs"
	}

	tenantCreds, err := credRepo.GetLazadaCredentials(ctx)
	if err != nil {
		return fmt.Sprintf("Lazada: no credentials (%v)", err)
	}
	globalCreds, err := globalConfigRepo.GetLazadaCredentials(ctx)
	if err != nil {
		return fmt.Sprintf("Lazada: no global config (%v)", err)
	}

	client := lazadaPkg.NewClient(globalCreds.AppKey, globalCreds.AppSecret, "ID")
	client.SetAccessToken(tenantCreds.AccessToken)
	svc := lazadaService.NewSyncServiceWithTenant(client, db, tenantID)

	count, syncErr := svc.SyncProductsByIDs(ctx, itemIDs)
	if syncErr != nil {
		return fmt.Sprintf("Lazada: error (%v)", syncErr)
	}
	return fmt.Sprintf("Lazada: %d/%d items synced", count, len(itemIDs))
}
