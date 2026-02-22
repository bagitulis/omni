package master_product

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	lazadaService "github.com/omni/backend/internal/services/lazada"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	tiktokService "github.com/omni/backend/internal/services/tiktok"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// SyncSelectedResult contains the result of syncing selected products.
type SyncSelectedResult struct {
	Synced  int      `json:"synced"`
	Failed  int      `json:"failed"`
	Details []string `json:"details"`
}

// SyncSelectedProducts syncs specific products by fetching latest data
// from each linked platform's API and upserting into staging tables.
// Unlike full sync, this does NOT clear cache — only updates selected items.
func (s *Service) SyncSelectedProducts(ctx context.Context, tenantID string, productIDs []uint, systemDB *gorm.DB) (*SyncSelectedResult, error) {
	if len(productIDs) == 0 {
		return &SyncSelectedResult{}, nil
	}

	// Load platform links for selected products
	var links []models.MasterProductPlatformLink
	if err := s.db.WithContext(ctx).
		Where("master_product_id IN ? AND (platform_product_id != '' OR platform_item_id != '')", productIDs).
		Find(&links).Error; err != nil {
		return nil, fmt.Errorf("failed to load platform links: %w", err)
	}

	log.Info().Int("product_count", len(productIDs)).Int("link_count", len(links)).Msg("SyncSelectedProducts: loaded platform links")

	if len(links) == 0 {
		return &SyncSelectedResult{Details: []string{"No platform links found for selected products"}}, nil
	}

	// Group links by platform
	platformLinks := make(map[string][]models.MasterProductPlatformLink)
	for _, link := range links {
		log.Debug().
			Str("platform", link.Platform).
			Str("product_id", link.PlatformProductID).
			Str("item_id", link.PlatformItemID).
			Uint("master_product_id", link.MasterProductID).
			Msg("SyncSelectedProducts: link")
		platformLinks[link.Platform] = append(platformLinks[link.Platform], link)
	}

	result := &SyncSelectedResult{}
	credRepo := repositories.NewPlatformCredentialsRepository(s.db)
	globalConfigRepo := repositories.NewGlobalConfigRepository(systemDB)

	for platform, pLinks := range platformLinks {
		count, err := s.syncPlatformItems(ctx, tenantID, platform, pLinks, credRepo, globalConfigRepo)
		if err != nil {
			result.Failed += len(pLinks)
			result.Details = append(result.Details, fmt.Sprintf("%s: %v", platform, err))
			log.Warn().Err(err).Str("platform", platform).Msg("Per-product sync failed")
			continue
		}
		result.Synced += count
		result.Details = append(result.Details, fmt.Sprintf("%s: %d synced", platform, count))
	}

	return result, nil
}

// syncPlatformItems syncs items for a single platform.
func (s *Service) syncPlatformItems(
	ctx context.Context,
	tenantID, platform string,
	links []models.MasterProductPlatformLink,
	credRepo *repositories.PlatformCredentialsRepository,
	globalConfigRepo *repositories.GlobalConfigRepository,
) (int, error) {
	switch platform {
	case "shopee":
		return s.syncShopeeItems(ctx, tenantID, links, credRepo, globalConfigRepo)
	case "tiktok":
		return s.syncTiktokItems(ctx, tenantID, links, credRepo, globalConfigRepo)
	case "lazada":
		return s.syncLazadaItems(ctx, tenantID, links, credRepo, globalConfigRepo)
	default:
		return 0, fmt.Errorf("unknown platform: %s", platform)
	}
}

func (s *Service) syncShopeeItems(
	ctx context.Context, tenantID string,
	links []models.MasterProductPlatformLink,
	credRepo *repositories.PlatformCredentialsRepository,
	globalConfigRepo *repositories.GlobalConfigRepository,
) (int, error) {
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

	// Collect unique item IDs
	itemIDs := collectShopeeItemIDs(links)
	if len(itemIDs) == 0 {
		return 0, nil
	}

	// Use existing sync service for upsert
	svc := shopeeService.NewProductSyncService(client, s.db, tenantID)
	return svc.SyncProductsByIDs(ctx, itemIDs)
}

func (s *Service) syncTiktokItems(
	ctx context.Context, tenantID string,
	links []models.MasterProductPlatformLink,
	credRepo *repositories.PlatformCredentialsRepository,
	globalConfigRepo *repositories.GlobalConfigRepository,
) (int, error) {
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

	svc := tiktokService.NewSyncServiceWithTenant(client, s.db, tenantID)
	productIDs := collectPlatformProductIDs(links)
	return svc.SyncProductsByIDs(ctx, productIDs)
}

func (s *Service) syncLazadaItems(
	ctx context.Context, tenantID string,
	links []models.MasterProductPlatformLink,
	credRepo *repositories.PlatformCredentialsRepository,
	globalConfigRepo *repositories.GlobalConfigRepository,
) (int, error) {
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

	svc := lazadaService.NewSyncServiceWithTenant(client, s.db, tenantID)
	itemIDs := collectLazadaItemIDs(links)
	return svc.SyncProductsByIDs(ctx, itemIDs)
}

// collectShopeeItemIDs extracts unique int64 item IDs from platform links.
func collectShopeeItemIDs(links []models.MasterProductPlatformLink) []int64 {
	seen := make(map[int64]struct{})
	var ids []int64
	for _, link := range links {
		id, err := strconv.ParseInt(strings.TrimSpace(link.PlatformItemID), 10, 64)
		if err != nil || id == 0 {
			continue
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	return ids
}

// collectPlatformProductIDs extracts unique string product IDs from platform links.
func collectPlatformProductIDs(links []models.MasterProductPlatformLink) []string {
	seen := make(map[string]struct{})
	var ids []string
	for _, link := range links {
		pid := strings.TrimSpace(link.PlatformProductID)
		if pid == "" {
			continue
		}
		if _, ok := seen[pid]; !ok {
			seen[pid] = struct{}{}
			ids = append(ids, pid)
		}
	}
	return ids
}

// collectLazadaItemIDs extracts unique int64 item IDs from Lazada platform links.
func collectLazadaItemIDs(links []models.MasterProductPlatformLink) []int64 {
	seen := make(map[int64]struct{})
	var ids []int64
	for _, link := range links {
		// Lazada uses item_id (numeric)
		idStr := strings.TrimSpace(link.PlatformItemID)
		if idStr == "" {
			idStr = strings.TrimSpace(link.PlatformProductID)
		}
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id == 0 {
			continue
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	return ids
}
