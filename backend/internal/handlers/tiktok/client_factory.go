package tiktok

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/repositories"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// NewTiktokClient creates a TikTok API client with tenant-specific credentials.
// This is the single source of truth for TikTok client initialization across all handlers.
// It resolves credentials from the tenant DB, falling back to global config for appKey/appSecret.
func NewTiktokClient(tenantID, basePath string) (*tiktokPkg.Client, error) {
	ctx := context.Background()
	db, err := config.GetTenantDB(tenantID, basePath)
	if err != nil {
		return nil, fmt.Errorf("tenant DB: %w", err)
	}

	// Use PlatformCredentialsRepository for key-value based config (current schema)
	credRepo := repositories.NewPlatformCredentialsRepository(db)
	tenantCreds, err := credRepo.GetTiktokCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("tenant credentials: %w", err)
	}
	if tenantCreds.AccessToken == "" || tenantCreds.ShopCipher == "" {
		return nil, fmt.Errorf("missing TikTok credentials: accessToken or shopCipher not configured")
	}

	// Use tenant credentials for appKey/appSecret if available, otherwise fall back to global
	appKey := tenantCreds.AppKey
	appSecret := tenantCreds.AppSecret

	if appKey == "" || appSecret == "" {
		systemDB, sysErr := config.GetSystemDB(basePath)
		if sysErr != nil {
			return nil, fmt.Errorf("system DB: %w", sysErr)
		}
		globalRepo := repositories.NewGlobalConfigRepository(systemDB)
		globalCreds, gErr := globalRepo.GetTiktokCredentials(ctx)
		if gErr != nil {
			return nil, fmt.Errorf("global credentials: %w", gErr)
		}
		appKey = globalCreds.AppKey
		appSecret = globalCreds.AppSecret
	}

	client := tiktokPkg.NewClient(appKey, appSecret)
	client.SetCredentials(tenantCreds.AccessToken, tenantCreds.ShopCipher)
	return client, nil
}
