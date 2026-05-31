package services

import (
	"context"
	"strconv"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// syncPlatforms lists platforms to sync after auto-refresh.
var syncPlatforms = []string{models.PlatformShopee, models.PlatformLazada, models.PlatformTiktok}

// SyncCredentialTokens reads refreshed tokens from platform_configs and copies them
// to credential_connections. Called after auto_update_token refreshes tokens.
// The db parameter must be a tenant-scoped DB (from config.GetTenantDBWithContext)
// which has access to both the tenant schema (platform_configs) and public schema
// (credential_connections).
// Returns the number of connections synced.
func SyncCredentialTokens(ctx context.Context, db *gorm.DB, tenantID string) (synced int, err error) {
	platformConfigRepo := repositories.NewTenantPlatformConfigRepository(db)
	credentialRepo := repositories.NewCredentialRepository(db)

	for _, platformName := range syncPlatforms {
		n, syncErr := syncPlatformTokens(ctx, platformConfigRepo, credentialRepo, tenantID, platformName)
		if syncErr != nil {
			log.Warn().Err(syncErr).
				Str("platform", platformName).
				Str("tenant_id", tenantID).
				Msg("sync: platform sync error (non-fatal)")
			continue
		}
		synced += n
	}

	return synced, nil
}

// syncPlatformTokens syncs tokens for a single platform.
func syncPlatformTokens(
	ctx context.Context,
	platformConfigRepo *repositories.TenantPlatformConfigRepository,
	credentialRepo *repositories.CredentialRepository,
	tenantID, platformName string,
) (int, error) {
	tokenInfo, err := platformConfigRepo.GetTokenInfo(ctx, platformName)
	if err != nil {
		return 0, err
	}
	if tokenInfo == nil || tokenInfo.ShopID == 0 || tokenInfo.AccessToken == "" {
		log.Debug().
			Str("platform", platformName).
			Str("tenant_id", tenantID).
			Msg("sync: no token/shopID in platform config, skipping")
		return 0, nil
	}

	storeIdentifier := strconv.FormatInt(tokenInfo.ShopID, 10)

	conn, err := credentialRepo.GetConnection(ctx, tenantID, platformName, storeIdentifier)
	if err != nil {
		return 0, err
	}
	if conn == nil {
		log.Debug().
			Str("platform", platformName).
			Str("store", storeIdentifier).
			Str("tenant_id", tenantID).
			Msg("sync: no credential connection found, skipping")
		return 0, nil
	}

	// platform_configs stores tokenExpiry in milliseconds; credential_connections in seconds.
	platformExpirySec := tokenInfo.TokenExpiry / 1000
	if platformExpirySec <= conn.TokenExpiry {
		return 0, nil // platform_configs does not have a newer token
	}

	platformRefreshExpirySec := tokenInfo.RefreshTokenExpiry / 1000
	now := time.Now()

	updateConn := &models.CredentialConnection{
		TenantID:        tenantID,
		Platform:        platformName,
		StoreIdentifier: storeIdentifier,
		AccessToken:     tokenInfo.AccessToken,
		RefreshToken:    tokenInfo.RefreshToken,
		ShopCipher:      tokenInfo.ShopCipherOfSeller,
		TokenExpiry:     platformExpirySec,
		RefreshExpiry:   platformRefreshExpirySec,
		UpdatedBy:       "auto_sync",
	}

	if updateErr := credentialRepo.UpdateConnectionTokensWithVersion(
		ctx, updateConn, conn.Version, "connected", "", &now,
	); updateErr != nil {
		log.Warn().Err(updateErr).
			Str("platform", platformName).
			Str("store", storeIdentifier).
			Msg("sync: failed to update credential connection tokens")
		return 0, nil // best-effort: don't propagate individual failures
	}

	log.Debug().
		Str("platform", platformName).
		Str("store", storeIdentifier).
		Str("tenant_id", tenantID).
		Msg("sync: credential tokens synced")

	return 1, nil
}
