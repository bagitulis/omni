package app

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
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
	// Acquire per-tenant lock to prevent concurrent sync with sync_products (full sync)
	mu := getTenantSyncMutex(tenantID)
	if !mu.TryLock() {
		return "Skipped: another sync is already running for this tenant", nil
	}
	defer mu.Unlock()

	log.Info().Msgf("[AutoFunction] Running 'Sync Products (Inventory Only)' for tenant: %s", tenantID)

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

	// 1. Validate inventory key_column_name is SKU-based
	keyColName := getInventoryKeyColumnName(ctx, db, tenantID)
	if keyColName != "" && keyColName != "SKU" && keyColName != "Kode SKU" && keyColName != "sku" {
		log.Info().Msgf("[AutoFunction] WARNING: inventory key_column_name=%q (expected 'SKU'). SKU matching may not work correctly for tenant: %s", keyColName, tenantID)
	}

	// 2. Get all unique SKUs from inventory
	inventorySkus := getInventorySkus(ctx, db, tenantID)
	if len(inventorySkus) == 0 {
		return "No inventory SKUs found — nothing to sync", nil
	}
	log.Info().Msgf("[AutoFunction] Found %d inventory SKUs (key_column=%q) for tenant: %s", len(inventorySkus), keyColName, tenantID)

	// 3. Check if platform SKU tables have data (require at least one full sync first)
	shopeeSkuCount, tiktokSkuCount, lazadaSkuCount := getPlatformSkuCounts(ctx, db, tenantID)
	if shopeeSkuCount == 0 && tiktokSkuCount == 0 && lazadaSkuCount == 0 {
		return "No platform SKU data found — run a full product sync from the Products page first", nil
	}
	log.Info().Msgf("[AutoFunction] Platform SKU counts: shopee=%d, tiktok=%d, lazada=%d", shopeeSkuCount, tiktokSkuCount, lazadaSkuCount)

	_ = systemDB
	credService := services.NewCredentialService(basePath)
	results := make([]string, 0, 3)

	// 4. Sync each platform using only inventory-matched item IDs
	if shopeeSkuCount > 0 {
		shopeeResult := syncShopeeByInventory(ctx, tenantID, db, inventorySkus, credService)
		results = append(results, shopeeResult)
	} else {
		results = append(results, "Shopee: skipped (no SKU data)")
	}

	if tiktokSkuCount > 0 {
		tiktokResult := syncTiktokByInventory(ctx, tenantID, db, inventorySkus, credService)
		results = append(results, tiktokResult)
	} else {
		results = append(results, "TikTok: skipped (no SKU data)")
	}

	if lazadaSkuCount > 0 {
		lazadaResult := syncLazadaByInventory(ctx, tenantID, db, inventorySkus, credService)
		results = append(results, lazadaResult)
	} else {
		results = append(results, "Lazada: skipped (no SKU data)")
	}

	// Auto-fix platform links after sync
	mappedCount, linkErrors, linkErr := autoFixPlatformLinks(ctx, db, tenantID)
	if linkErr != nil {
		results = append(results, fmt.Sprintf("Auto-link FAILED: %v", linkErr))
	} else if mappedCount > 0 || linkErrors > 0 {
		results = append(results, fmt.Sprintf("Auto-link: %d mapped, %d errors", mappedCount, linkErrors))
	}

	resultMsg := fmt.Sprintf("Inventory sync completed: %v", results)
	log.Info().Msgf("[AutoFunction] %s for tenant: %s", resultMsg, tenantID)
	return resultMsg, nil
}

// getInventoryKeyColumnName returns the key_column_name used in inventory records.
func getInventoryKeyColumnName(ctx context.Context, db *gorm.DB, tenantID string) string {
	var keyColName string
	db.WithContext(ctx).
		Model(&models.InventoryRecord{}).
		Where("tenant_id = ?", tenantID).
		Limit(1).
		Pluck("key_column_name", &keyColName)
	return keyColName
}

// getInventorySkus returns all unique SKU values from the inventory table.
func getInventorySkus(ctx context.Context, db *gorm.DB, tenantID string) []string {
	var skus []string
	if err := db.WithContext(ctx).
		Model(&models.InventoryRecord{}).
		Where("tenant_id = ? AND key_value != ''", tenantID).
		Distinct("key_value").
		Pluck("key_value", &skus).Error; err != nil {
		log.Info().Msgf("[AutoFunction] Failed to load inventory SKUs: %v", err)
		return nil
	}
	return skus
}

// getPlatformSkuCounts returns count of SKUs per platform for a tenant.
func getPlatformSkuCounts(ctx context.Context, db *gorm.DB, tenantID string) (shopee, tiktok, lazada int64) {
	db.WithContext(ctx).Table("shopee_skus").Where("tenant_id = ?", tenantID).Count(&shopee)
	db.WithContext(ctx).Table("tiktok_skus").Where("tenant_id = ?", tenantID).Count(&tiktok)
	db.WithContext(ctx).Table("lazada_skus").Where("tenant_id = ?", tenantID).Count(&lazada)
	return
}

// syncShopeeByInventory finds Shopee items matching inventory SKUs and syncs them.
func syncShopeeByInventory(
	ctx context.Context, tenantID string, db *gorm.DB,
	inventorySkus []string,
	credService *services.CredentialService,
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
	log.Info().Msgf("[AutoFunction] Shopee: %d items matched inventory SKUs", len(itemIDs))

	creds, err := credService.GetPlatformCredentials(tenantID, "shopee")
	if err != nil {
		return fmt.Sprintf("Shopee: no credentials (%v)", err)
	}

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)
	svc := shopeeService.NewProductSyncService(client, db, tenantID)

	count, syncErr := svc.SyncProductsByIDs(ctx, itemIDs)
	if syncErr != nil {
		return fmt.Sprintf("Shopee: API error syncing %d items — %v", len(itemIDs), syncErr)
	}
	return fmt.Sprintf("Shopee: %d/%d items synced", count, len(itemIDs))
}

// syncTiktokByInventory finds TikTok products matching inventory SKUs and syncs them.
func syncTiktokByInventory(
	ctx context.Context, tenantID string, db *gorm.DB,
	inventorySkus []string,
	credService *services.CredentialService,
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
	log.Info().Msgf("[AutoFunction] TikTok: %d products matched inventory SKUs (from %d DB IDs)", len(productIDs), len(productDBIDs))

	creds, err := credService.GetPlatformCredentials(tenantID, "tiktok")
	if err != nil {
		return fmt.Sprintf("TikTok: no credentials (%v)", err)
	}

	client := tiktokPkg.NewClient(creds.AppKey, creds.AppSecret)
	client.SetCredentials(creds.AccessToken, creds.ShopCipher)
	svc := tiktokService.NewSyncServiceWithTenant(client, db, tenantID)

	count, syncErr := svc.SyncProductsByIDs(ctx, productIDs)
	if syncErr != nil {
		return fmt.Sprintf("TikTok: API error syncing %d products — %v", len(productIDs), syncErr)
	}
	return fmt.Sprintf("TikTok: %d/%d products synced", count, len(productIDs))
}

// syncLazadaByInventory finds Lazada items matching inventory SKUs and syncs them.
func syncLazadaByInventory(
	ctx context.Context, tenantID string, db *gorm.DB,
	inventorySkus []string,
	credService *services.CredentialService,
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

	// Convert string item_ids to int64, log skipped conversions
	var itemIDs []int64
	skipped := 0
	for _, s := range itemIDStrs {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil || id <= 0 {
			log.Info().Msgf("[AutoFunction] Lazada: skipping invalid item_id=%q (err=%v)", s, err)
			skipped++
			continue
		}
		itemIDs = append(itemIDs, id)
	}
	if len(itemIDs) == 0 {
		return fmt.Sprintf("Lazada: 0 valid item IDs (all %d skipped)", skipped)
	}
	log.Info().Msgf("[AutoFunction] Lazada: %d items matched inventory SKUs (skipped=%d)", len(itemIDs), skipped)

	creds, err := credService.GetPlatformCredentials(tenantID, "lazada")
	if err != nil {
		return fmt.Sprintf("Lazada: no credentials (%v)", err)
	}
	region := creds.Region
	if region == "" {
		region = "ID"
	}

	client := lazadaPkg.NewClient(creds.AppKey, creds.AppSecret, region)
	client.SetAccessToken(creds.AccessToken)
	svc := lazadaService.NewSyncServiceWithTenant(client, db, tenantID)

	count, syncErr := svc.SyncProductsByIDs(ctx, itemIDs)
	if syncErr != nil {
		return fmt.Sprintf("Lazada: API error syncing %d items — %v", len(itemIDs), syncErr)
	}
	return fmt.Sprintf("Lazada: %d/%d items synced", count, len(itemIDs))
}
