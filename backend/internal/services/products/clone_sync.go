// Package products provides product cloning - sync operations after clone
package products

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

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

// =============================================================================
// Product Sync Methods (Auto-sync after clone)
// =============================================================================

// triggerProductSync syncs products from target platform API to update database
// This is called automatically after a successful clone to update product manager
func (s *CloneService) triggerProductSync(ctx context.Context, platform string) string {
	if s.credService == nil || s.dbPath == "" {
		return "sync skipped - no credential service"
	}

	// Get system DB for global config
	systemDB, err := config.GetSystemDB(s.dbPath)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", s.tenantID).Msg("[Clone] Failed to get system DB for sync")
		return fmt.Sprintf("sync failed - system DB error: %v", err)
	}

	switch platform {
	case "shopee":
		return s.syncShopeeProducts(ctx, systemDB)
	case "lazada":
		return s.syncLazadaProducts(ctx, systemDB)
	case "tiktok":
		return s.syncTiktokProducts(ctx, systemDB)
	default:
		return fmt.Sprintf("sync skipped - unknown platform: %s", platform)
	}
}

// =============================================================================
// Platform-specific Sync Methods
// =============================================================================

// syncShopeeProducts syncs products from Shopee API
func (s *CloneService) syncShopeeProducts(ctx context.Context, systemDB *gorm.DB) string {
	tenantCreds, globalCreds, err := s.getShopeeCredentials(ctx, systemDB)
	if err != nil {
		return err.Error()
	}

	// Create client and sync
	client := shopeePkg.NewClient(globalCreds.PartnerID, globalCreds.PartnerKey, true)
	client.SetShopCredentials(tenantCreds.ShopIDInt, tenantCreds.AccessToken)

	syncSvc := shopeeService.NewSyncServiceWithTenant(client, s.db, s.tenantID)
	count, err := syncSvc.SyncProducts(ctx)
	if err != nil {
		return fmt.Sprintf("sync failed: %v", err)
	}

	log.Info().Str("tenant_id", s.tenantID).Int("count", count).Msg("[Sync] Shopee products synced")
	return fmt.Sprintf("synced %d products from Shopee", count)
}

// syncLazadaProducts syncs products from Lazada API
func (s *CloneService) syncLazadaProducts(ctx context.Context, systemDB *gorm.DB) string {
	tenantCreds, globalCreds, err := s.getLazadaCredentials(ctx, systemDB)
	if err != nil {
		return err.Error()
	}

	// Use region from tenant credentials (fallback to ID for Indonesia)
	region := tenantCreds.Region
	if region == "" {
		region = "id"
	}

	// Create client and sync
	client := lazadaPkg.NewClient(globalCreds.AppKey, globalCreds.AppSecret, region)
	client.SetAccessToken(tenantCreds.AccessToken)

	syncSvc := lazadaService.NewSyncServiceWithTenant(client, s.db, s.tenantID)
	count, err := syncSvc.SyncProducts(ctx)
	if err != nil {
		return fmt.Sprintf("sync failed: %v", err)
	}

	log.Info().Str("tenant_id", s.tenantID).Int("count", count).Msg("[Sync] Lazada products synced")
	return fmt.Sprintf("synced %d products from Lazada", count)
}

// syncTiktokProducts syncs products from TikTok API
func (s *CloneService) syncTiktokProducts(ctx context.Context, systemDB *gorm.DB) string {
	tenantCreds, globalCreds, err := s.getTiktokCredentials(ctx, systemDB)
	if err != nil {
		return err.Error()
	}

	// Create client and sync
	client := tiktokPkg.NewClient(globalCreds.AppKey, globalCreds.AppSecret)
	client.SetCredentials(tenantCreds.AccessToken, tenantCreds.ShopCipher)

	syncSvc := tiktokService.NewSyncServiceWithTenant(client, s.db, s.tenantID)
	count, err := syncSvc.SyncProducts(ctx)
	if err != nil {
		return fmt.Sprintf("sync failed: %v", err)
	}

	log.Info().Str("tenant_id", s.tenantID).Int("count", count).Msg("[Sync] TikTok products synced")
	return fmt.Sprintf("synced %d products from TikTok", count)
}

// =============================================================================
// Credential Helper Methods - DRY principle
// =============================================================================

// getShopeeCredentials retrieves both tenant and global Shopee credentials
func (s *CloneService) getShopeeCredentials(ctx context.Context, systemDB *gorm.DB) (*models.TenantPlatformCredentials, *models.PlatformCredentials, error) {
	credRepo := repositories.NewPlatformCredentialsRepository(s.db)
	tenantCreds, err := credRepo.GetShopeeCredentials(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("sync failed - no Shopee credentials: %v", err)
	}
	if tenantCreds.ShopIDInt == 0 || tenantCreds.AccessToken == "" {
		return nil, nil, fmt.Errorf("sync skipped - Shopee not configured")
	}

	configRepo := repositories.NewGlobalConfigRepository(systemDB)
	globalCreds, err := configRepo.GetShopeeCredentials(ctx)
	if err != nil || globalCreds.PartnerID == 0 {
		return nil, nil, fmt.Errorf("sync failed - no Shopee global credentials")
	}

	return tenantCreds, globalCreds, nil
}

// getLazadaCredentials retrieves both tenant and global Lazada credentials
func (s *CloneService) getLazadaCredentials(ctx context.Context, systemDB *gorm.DB) (*models.TenantPlatformCredentials, *models.PlatformCredentials, error) {
	credRepo := repositories.NewPlatformCredentialsRepository(s.db)
	tenantCreds, err := credRepo.GetLazadaCredentials(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("sync failed - no Lazada credentials: %v", err)
	}
	if tenantCreds.AccessToken == "" {
		return nil, nil, fmt.Errorf("sync skipped - Lazada not configured")
	}

	configRepo := repositories.NewGlobalConfigRepository(systemDB)
	globalCreds, err := configRepo.GetLazadaCredentials(ctx)
	if err != nil || globalCreds.AppKey == "" {
		return nil, nil, fmt.Errorf("sync failed - no Lazada global credentials")
	}

	return tenantCreds, globalCreds, nil
}

// getTiktokCredentials retrieves both tenant and global TikTok credentials
func (s *CloneService) getTiktokCredentials(ctx context.Context, systemDB *gorm.DB) (*models.TenantPlatformCredentials, *models.PlatformCredentials, error) {
	credRepo := repositories.NewPlatformCredentialsRepository(s.db)
	tenantCreds, err := credRepo.GetTiktokCredentials(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("sync failed - no TikTok credentials: %v", err)
	}
	if tenantCreds.ShopCipher == "" || tenantCreds.AccessToken == "" {
		return nil, nil, fmt.Errorf("sync skipped - TikTok not configured")
	}

	configRepo := repositories.NewGlobalConfigRepository(systemDB)
	globalCreds, err := configRepo.GetTiktokCredentials(ctx)
	if err != nil || globalCreds.AppKey == "" {
		return nil, nil, fmt.Errorf("sync failed - no TikTok global credentials")
	}

	return tenantCreds, globalCreds, nil
}
