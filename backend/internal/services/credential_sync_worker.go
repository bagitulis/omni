package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// retryDualWrite retries a dual-write operation with linear backoff.
// Respects context cancellation between attempts.
// Returns the last error if all attempts fail.
func retryDualWrite(ctx context.Context, maxAttempts int, fn func() error) error {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if attempt < maxAttempts {
				backoff := time.Duration(attempt*100) * time.Millisecond
				log.Debug().Err(err).
					Int("attempt", attempt).
					Int("max_attempts", maxAttempts).
					Dur("backoff", backoff).
					Msg("dual-write retry")
				select {
				case <-time.After(backoff):
				case <-ctx.Done():
					return fmt.Errorf("dual-write cancelled: %w", ctx.Err())
				}
			}
		} else {
			return nil
		}
	}
	return fmt.Errorf("dual-write failed after %d attempts: %w", maxAttempts, lastErr)
}

// isStaleVersionError checks if an error is due to optimistic lock conflict.
// Uses sentinel error matching for robustness against message refactors.
func isStaleVersionError(err error) bool {
	return errors.Is(err, repositories.ErrStaleVersion)
}

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

	// Both platform_configs and credential_connections store tokenExpiry in milliseconds.
	platformExpiryMs := tokenInfo.TokenExpiry
	if platformExpiryMs <= conn.TokenExpiry {
		return 0, nil // platform_configs does not have a newer token
	}

	platformRefreshExpiryMs := tokenInfo.RefreshTokenExpiry
	now := time.Now()

	updateConn := &models.CredentialConnection{
		TenantID:        tenantID,
		Platform:        platformName,
		StoreIdentifier: storeIdentifier,
		AccessToken:     tokenInfo.AccessToken,
		RefreshToken:    tokenInfo.RefreshToken,
		ShopCipher:      tokenInfo.ShopCipherOfSeller,
		TokenExpiry:     platformExpiryMs,
		RefreshExpiry:   platformRefreshExpiryMs,
		UpdatedBy:       "auto_sync",
	}

	if updateErr := credentialRepo.UpdateConnectionTokensWithVersion(
		ctx, updateConn, conn.Version, "connected", "", &now,
	); updateErr != nil {
		// Retry once on optimistic lock conflict (stale version)
		if isStaleVersionError(updateErr) {
			log.Debug().
				Str("platform", platformName).
				Str("store", storeIdentifier).
				Msg("sync: version conflict, re-reading connection for retry")

			refreshedConn, reReadErr := credentialRepo.GetConnection(ctx, tenantID, platformName, storeIdentifier)
			if reReadErr != nil || refreshedConn == nil {
				log.Warn().Err(reReadErr).
					Str("platform", platformName).
					Str("store", storeIdentifier).
					Msg("sync: re-read failed after version conflict")
				return 0, nil
			}

			now = time.Now()
			if retryErr := credentialRepo.UpdateConnectionTokensWithVersion(
				ctx, updateConn, refreshedConn.Version, "connected", "", &now,
			); retryErr != nil {
				log.Warn().Err(retryErr).
					Str("platform", platformName).
					Str("store", storeIdentifier).
					Msg("sync: retry also failed after version conflict")
				return 0, nil
			}

			log.Debug().
				Str("platform", platformName).
				Str("store", storeIdentifier).
				Msg("sync: credential tokens synced (after retry)")
			return 1, nil
		}

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
