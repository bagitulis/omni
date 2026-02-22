package app

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	lazadaService "github.com/omni/backend/internal/services/lazada"
	masterProductService "github.com/omni/backend/internal/services/master_product"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	tiktokService "github.com/omni/backend/internal/services/tiktok"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"gorm.io/gorm"
)

// syncProductsHandler handles the "Sync Products" auto function.
// Syncs products from all 3 platforms (Shopee, TikTok, Lazada) to staging tables.
// Each platform sync is independent — failures are logged but don't block others.
func syncProductsHandler(ctx context.Context, tenantID string, cfg *models.AutoFunctionConfig) (string, error) {
	log.Printf("[AutoFunction] Running 'Sync Products' for tenant: %s", tenantID)

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

	// Auto-fix: bidirectional platform link scan (master↔staging) for all platforms
	autoFixPlatformLinks(ctx, db, tenantID)

	// Return error only if ALL platforms failed
	if shopeeErr != nil && tiktokErr != nil && lazadaErr != nil {
		return resultMsg, fmt.Errorf("all platform syncs failed")
	}

	return resultMsg, nil
}

// singlePlatformSyncHandler creates an auto-function handler for a single platform.
func singlePlatformSyncHandler(platform string) func(context.Context, string, *models.AutoFunctionConfig) (string, error) {
	return func(ctx context.Context, tenantID string, cfg *models.AutoFunctionConfig) (string, error) {
		log.Printf("[AutoFunction] Running 'Sync %s' for tenant: %s", platform, tenantID)

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

		count, syncErr := syncPlatformProducts(ctx, platform, tenantID, db, systemDB, basePath)
		if syncErr != nil {
			return fmt.Sprintf("%s sync failed: %v", platform, syncErr), syncErr
		}

		autoFixPlatformLinks(ctx, db, tenantID)

		resultMsg := fmt.Sprintf("%s: %d products synced", platform, count)
		log.Printf("[AutoFunction] %s for tenant: %s", resultMsg, tenantID)
		return resultMsg, nil
	}
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

// autoFixPlatformLinks runs bidirectional auto-link for all platforms.
// Direction 1: master SKUs → find in staging tables (forward scan)
// Direction 2: staging SKUs → find in master (reverse scan)
// This ensures links are created regardless of which side has the SKU.
func autoFixPlatformLinks(ctx context.Context, db *gorm.DB, tenantID string) {
	mapper := masterProductService.NewSkuMapper(db, tenantID)

	// Collect all unique seller_skus from both master and all staging tables
	allSkus := collectAllSellerSkus(ctx, db, tenantID)
	if len(allSkus) == 0 {
		return
	}

	result, err := mapper.AutoMapAndLinkBySkus(ctx, allSkus)
	if err != nil {
		log.Printf("[AutoFunction] Auto-fix link failed for %s: %v", tenantID, err)
		return
	}
	log.Printf("[AutoFunction] Auto-fix links for %s: %d mapped, %d skipped, %d total SKUs scanned",
		tenantID, result.MappedCount, result.SkippedCount, len(allSkus))
}

// collectAllSellerSkus gathers unique seller_skus from master + all platform staging tables.
// This enables bidirectional matching: a staging SKU can find its master product, and vice versa.
func collectAllSellerSkus(ctx context.Context, db *gorm.DB, tenantID string) []string {
	seen := make(map[string]struct{})

	// 1. Master SKUs
	var masterSkus []string
	if err := db.WithContext(ctx).
		Model(&models.MasterProductSku{}).
		Where("tenant_id = ? AND seller_sku != ''", tenantID).
		Pluck("seller_sku", &masterSkus).Error; err != nil {
		log.Printf("[AutoFunction] Failed to load master SKUs: %v", err)
	}
	for _, sku := range masterSkus {
		seen[sku] = struct{}{}
	}

	// 2. Shopee staging SKUs
	var shopeeSkus []string
	if err := db.WithContext(ctx).
		Table("shopee_skus").
		Where("tenant_id = ? AND seller_sku != ''", tenantID).
		Pluck("seller_sku", &shopeeSkus).Error; err != nil {
		log.Printf("[AutoFunction] Failed to load Shopee SKUs: %v", err)
	}
	for _, sku := range shopeeSkus {
		seen[sku] = struct{}{}
	}

	// 3. TikTok staging SKUs
	var tiktokSkus []string
	if err := db.WithContext(ctx).
		Table("tiktok_skus").
		Where("tenant_id = ? AND seller_sku != ''", tenantID).
		Pluck("seller_sku", &tiktokSkus).Error; err != nil {
		log.Printf("[AutoFunction] Failed to load TikTok SKUs: %v", err)
	}
	for _, sku := range tiktokSkus {
		seen[sku] = struct{}{}
	}

	// 4. Lazada staging SKUs (uses shop_sku as seller_sku equivalent)
	var lazadaSkus []string
	if err := db.WithContext(ctx).
		Table("lazada_skus").
		Where("tenant_id = ? AND shop_sku != ''", tenantID).
		Pluck("shop_sku", &lazadaSkus).Error; err != nil {
		log.Printf("[AutoFunction] Failed to load Lazada SKUs: %v", err)
	}
	for _, sku := range lazadaSkus {
		seen[sku] = struct{}{}
	}

	// Convert to slice
	result := make([]string, 0, len(seen))
	for sku := range seen {
		result = append(result, sku)
	}
	return result
}
