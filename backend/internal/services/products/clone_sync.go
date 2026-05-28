// Package products provides product cloning - sync operations after clone
package products

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	lazadaService "github.com/omni/backend/internal/services/lazada"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	tiktokService "github.com/omni/backend/internal/services/tiktok"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
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

	switch platform {
	case "shopee":
		return s.syncShopeeProducts(ctx)
	case "lazada":
		return s.syncLazadaProducts(ctx)
	case "tiktok":
		return s.syncTiktokProducts(ctx)
	default:
		return fmt.Sprintf("sync skipped - unknown platform: %s", platform)
	}
}

// =============================================================================
// Platform-specific Sync Methods
// =============================================================================

// syncShopeeProducts syncs products from Shopee API
func (s *CloneService) syncShopeeProducts(ctx context.Context) string {
	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "shopee")
	if err != nil {
		return err.Error()
	}

	// Create client and sync
	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)

	syncSvc := shopeeService.NewSyncServiceWithTenant(client, s.db, s.tenantID)
	count, err := syncSvc.SyncProducts(ctx)
	if err != nil {
		return fmt.Sprintf("sync failed: %v", err)
	}

	log.Info().Str("tenant_id", s.tenantID).Int("count", count).Msg("[Sync] Shopee products synced")
	return fmt.Sprintf("synced %d products from Shopee", count)
}

// syncLazadaProducts syncs products from Lazada API
func (s *CloneService) syncLazadaProducts(ctx context.Context) string {
	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "lazada")
	if err != nil {
		return err.Error()
	}

	// Use region from tenant credentials (fallback to ID for Indonesia)
	region := creds.Region
	if region == "" {
		region = "id"
	}

	// Create client and sync
	client := lazadaPkg.NewClient(creds.AppKey, creds.AppSecret, region)
	client.SetAccessToken(creds.AccessToken)

	syncSvc := lazadaService.NewSyncServiceWithTenant(client, s.db, s.tenantID)
	count, err := syncSvc.SyncProducts(ctx)
	if err != nil {
		return fmt.Sprintf("sync failed: %v", err)
	}

	log.Info().Str("tenant_id", s.tenantID).Int("count", count).Msg("[Sync] Lazada products synced")
	return fmt.Sprintf("synced %d products from Lazada", count)
}

// syncTiktokProducts syncs products from TikTok API
func (s *CloneService) syncTiktokProducts(ctx context.Context) string {
	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "tiktok")
	if err != nil {
		return err.Error()
	}

	// Create client and sync
	client := tiktokPkg.NewClient(creds.AppKey, creds.AppSecret)
	client.SetCredentials(creds.AccessToken, creds.ShopCipher)

	syncSvc := tiktokService.NewSyncServiceWithTenant(client, s.db, s.tenantID)
	count, err := syncSvc.SyncProducts(ctx)
	if err != nil {
		return fmt.Sprintf("sync failed: %v", err)
	}

	log.Info().Str("tenant_id", s.tenantID).Int("count", count).Msg("[Sync] TikTok products synced")
	return fmt.Sprintf("synced %d products from TikTok", count)
}
