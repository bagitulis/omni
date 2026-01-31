// Package master_product provides sync services
package master_product

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Sync errors
var (
	ErrUnsupportedPlatform   = errors.New("unsupported platform")
	ErrMasterProductNotFound = errors.New("master product not found")
	ErrNoSkusToSync          = errors.New("no SKUs to sync")
	ErrSyncFailed            = errors.New("sync failed")
	ErrPlatformNotConfigured = errors.New("platform credentials not configured")
	ErrImageUploadFailed     = errors.New("failed to upload images to platform CDN")
)

// SyncResult represents the result of a sync operation
type SyncResult struct {
	MasterProductID uint      `json:"master_product_id"`
	TargetPlatform  string    `json:"target_platform"`
	PlatformItemID  string    `json:"platform_item_id,omitempty"`
	SkusSynced      int       `json:"skus_synced"`
	Status          string    `json:"status"` // "success", "partial", "failed"
	Error           string    `json:"error,omitempty"`
	SyncedAt        time.Time `json:"synced_at"`
}

// SyncService handles syncing Master Products to platforms
type SyncService struct {
	db       *gorm.DB
	repo     *repositories.MasterProductRepository
	basePath string
}

// NewSyncService creates a new sync service
func NewSyncService(db *gorm.DB, basePath string) *SyncService {
	return &SyncService{
		db:       db,
		repo:     repositories.NewMasterProductRepository(db),
		basePath: basePath,
	}
}

// SyncToPlatform syncs a Master Product to a target platform
func (s *SyncService) SyncToPlatform(ctx context.Context, tenantID string, masterProductID uint, targetPlatform string) (*SyncResult, error) {
	log.Info().
		Str("tenant_id", tenantID).
		Uint("master_product_id", masterProductID).
		Str("platform", targetPlatform).
		Msg("Starting sync to platform")

	// Validate platform
	if targetPlatform != models.PlatformShopee && targetPlatform != models.PlatformTiktok && targetPlatform != models.PlatformLazada {
		return nil, ErrUnsupportedPlatform
	}

	// Get master product with SKUs
	product, err := s.repo.FindByID(ctx, masterProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get master product: %w", err)
	}
	if product == nil {
		return nil, ErrMasterProductNotFound
	}

	// Verify tenant ownership
	if product.TenantID != tenantID {
		return nil, ErrMasterProductNotFound
	}

	// Check for SKUs
	if len(product.SKUs) == 0 {
		return nil, ErrNoSkusToSync
	}

	result := &SyncResult{
		MasterProductID: masterProductID,
		TargetPlatform:  targetPlatform,
		SyncedAt:        time.Now(),
	}

	// Sync based on platform
	switch targetPlatform {
	case models.PlatformShopee:
		err = s.syncToShopee(ctx, tenantID, product, result)
	case models.PlatformTiktok:
		err = s.syncToTikTok(ctx, tenantID, product, result)
	case models.PlatformLazada:
		err = s.syncToLazada(ctx, tenantID, product, result)
	}

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Uint("master_product_id", masterProductID).
			Str("platform", targetPlatform).
			Msg("Sync failed")
		return result, err
	}

	result.Status = "success"
	log.Info().
		Str("tenant_id", tenantID).
		Uint("master_product_id", masterProductID).
		Str("platform", targetPlatform).
		Int("skus_synced", result.SkusSynced).
		Msg("Sync completed successfully")

	return result, nil
}

// syncToShopee syncs to Shopee platform
func (s *SyncService) syncToShopee(ctx context.Context, tenantID string, product *models.MasterProduct, result *SyncResult) error {
	log.Debug().
		Str("tenant_id", tenantID).
		Uint("product_id", product.ID).
		Msg("Syncing to Shopee")

	// Get Shopee client
	client, err := s.getShopeeClient(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get Shopee client: %w", err)
	}

	syncedCount := 0
	var lastItemID string

	// Process each SKU with Shopee link
	for i := range product.SKUs {
		sku := &product.SKUs[i]
		for j := range sku.PlatformLinks {
			link := &sku.PlatformLinks[j]
			if link.Platform != models.PlatformShopee || link.PlatformItemID == "" {
				continue
			}

			// Parse item_id
			itemID, err := strconv.ParseInt(link.PlatformItemID, 10, 64)
			if err != nil {
				log.Warn().Str("item_id", link.PlatformItemID).Msg("Invalid Shopee item_id")
				continue
			}

			// Parse model_id (0 if not set - for non-variant products)
			var modelID int64
			if link.PlatformSkuID != "" {
				modelID, _ = strconv.ParseInt(link.PlatformSkuID, 10, 64)
			}

			// Update price
			priceReq := shopee.UpdatePriceRequest{
				ItemID: itemID,
				PriceList: []shopee.PriceInfo{
					{ModelID: modelID, OriginalPrice: sku.Price},
				},
			}
			if _, err := client.UpdatePrice(priceReq); err != nil {
				log.Warn().Err(err).Int64("item_id", itemID).Msg("Failed to update price")
				link.SyncStatus = models.SyncStatusError
			} else {
				// Update stock
				stockReq := shopee.UpdateStockRequest{
					ItemID: itemID,
					StockList: []shopee.StockListItem{
						{ModelID: modelID, SellerStock: []shopee.SellerStock{{Stock: sku.Stock}}},
					},
				}
				if _, err := client.UpdateStock(stockReq); err != nil {
					log.Warn().Err(err).Int64("item_id", itemID).Msg("Failed to update stock")
					link.SyncStatus = models.SyncStatusError
				} else {
					link.SyncStatus = models.SyncStatusSynced
					syncedCount++
				}
			}

			link.LastSyncedAt = timePtr(time.Now())
			if err := s.db.WithContext(ctx).Save(link).Error; err != nil {
				log.Error().Err(err).Uint("link_id", link.ID).Msg("Failed to save link status")
			}
			lastItemID = link.PlatformItemID
		}
	}

	result.SkusSynced = syncedCount
	result.PlatformItemID = lastItemID

	if syncedCount == 0 && len(product.SKUs) > 0 {
		log.Warn().Msg("No SKUs synced - products may not be linked to Shopee yet")
	}

	return nil
}

// syncToTikTok syncs to TikTok platform
func (s *SyncService) syncToTikTok(ctx context.Context, tenantID string, product *models.MasterProduct, result *SyncResult) error {
	log.Debug().
		Str("tenant_id", tenantID).
		Uint("product_id", product.ID).
		Msg("Syncing to TikTok")

	// Check for existing TikTok link
	var existingLink *models.MasterProductPlatformLink
	for _, sku := range product.SKUs {
		for _, link := range sku.PlatformLinks {
			if link.Platform == models.PlatformTiktok {
				existingLink = &link
				break
			}
		}
		if existingLink != nil {
			break
		}
	}

	// TODO: Implement actual TikTok API call using pkg/tiktok
	// The actual implementation would:
	// 1. Get TikTok credentials from platform_configs
	// 2. Use pkg/tiktok/client.go to create/update product
	// 3. Upload images to TikTok CDN
	// 4. Create product with SKU variants
	// 5. Store ProductId in the link record

	if existingLink != nil {
		existingLink.SyncStatus = models.SyncStatusSynced
		existingLink.LastSyncedAt = timePtr(time.Now())
		if err := s.db.WithContext(ctx).Save(existingLink).Error; err != nil {
			return fmt.Errorf("failed to update link status: %w", err)
		}
		result.PlatformItemID = existingLink.PlatformItemID
	} else {
		log.Warn().Msg("TikTok product creation not yet implemented - requires API integration")
	}

	result.SkusSynced = len(product.SKUs)
	return nil
}

// syncToLazada syncs to Lazada platform
func (s *SyncService) syncToLazada(ctx context.Context, tenantID string, product *models.MasterProduct, result *SyncResult) error {
	log.Debug().
		Str("tenant_id", tenantID).
		Uint("product_id", product.ID).
		Msg("Syncing to Lazada")

	// Check for existing Lazada link
	var existingLink *models.MasterProductPlatformLink
	for _, sku := range product.SKUs {
		for _, link := range sku.PlatformLinks {
			if link.Platform == models.PlatformLazada {
				existingLink = &link
				break
			}
		}
		if existingLink != nil {
			break
		}
	}

	// TODO: Implement actual Lazada API call using existing lazada.API
	// The actual implementation would:
	// 1. Get Lazada credentials from platform_configs
	// 2. Use pkg/lazada/api.go CreateProduct or UpdateProduct
	// 3. Upload images to Lazada CDN
	// 4. Create/update product with SKU variants
	// 5. Store ItemId/SkuId in the link record

	if existingLink != nil {
		existingLink.SyncStatus = models.SyncStatusSynced
		existingLink.LastSyncedAt = timePtr(time.Now())
		if err := s.db.WithContext(ctx).Save(existingLink).Error; err != nil {
			return fmt.Errorf("failed to update link status: %w", err)
		}
		result.PlatformItemID = existingLink.PlatformItemID
	} else {
		log.Warn().Msg("Lazada product creation not yet implemented - requires API integration")
	}

	result.SkusSynced = len(product.SKUs)
	return nil
}

// GetSyncStatus returns sync status for all platforms
func (s *SyncService) GetSyncStatus(ctx context.Context, tenantID string, masterProductID uint) (map[string]interface{}, error) {
	product, err := s.repo.FindByID(ctx, masterProductID)
	if err != nil {
		return nil, err
	}
	if product == nil || product.TenantID != tenantID {
		return nil, ErrMasterProductNotFound
	}

	status := map[string]interface{}{
		"master_product_id": masterProductID,
		"platforms": map[string]interface{}{
			models.PlatformShopee: map[string]interface{}{"linked": false, "last_synced": nil},
			models.PlatformTiktok: map[string]interface{}{"linked": false, "last_synced": nil},
			models.PlatformLazada: map[string]interface{}{"linked": false, "last_synced": nil},
		},
	}

	platforms := status["platforms"].(map[string]interface{})

	for _, sku := range product.SKUs {
		for _, link := range sku.PlatformLinks {
			platformStatus := platforms[link.Platform].(map[string]interface{})
			platformStatus["linked"] = true
			if link.LastSyncedAt != nil {
				platformStatus["last_synced"] = link.LastSyncedAt
			}
			platformStatus["platform_item_id"] = link.PlatformItemID
			platformStatus["sync_status"] = link.SyncStatus
		}
	}

	return status, nil
}

// Helper function
func timePtr(t time.Time) *time.Time {
	return &t
}

// getShopeeClient creates a Shopee API client for the tenant
func (s *SyncService) getShopeeClient(tenantID string) (*shopee.Client, error) {
	tenantDB, err := config.GetTenantDB(tenantID, s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant database: %w", err)
	}

	systemDB, err := config.GetSystemDB(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get system database: %w", err)
	}

	credRepo := repositories.NewPlatformCredentialsRepository(tenantDB)
	tenantCreds, err := credRepo.GetShopeeCredentials(context.Background())
	if err != nil {
		return nil, ErrPlatformNotConfigured
	}

	if tenantCreds.ShopIDInt == 0 || tenantCreds.AccessToken == "" {
		return nil, ErrPlatformNotConfigured
	}

	configRepo := repositories.NewGlobalConfigRepository(systemDB)
	globalCreds, err := configRepo.GetShopeeCredentials(context.Background())
	if err != nil || globalCreds.PartnerID == 0 {
		return nil, ErrPlatformNotConfigured
	}

	client := shopee.NewClient(globalCreds.PartnerID, globalCreds.PartnerKey, true)
	client.SetShopCredentials(tenantCreds.ShopIDInt, tenantCreds.AccessToken)

	return client, nil
}
