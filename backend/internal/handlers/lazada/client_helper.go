// Package lazada provides Lazada handler helpers
package lazada

import (
	"context"
	"errors"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/repositories"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
)

// ErrMissingAccessToken indicates that the tenant has no access token configured
var ErrMissingAccessToken = errors.New("lazada access token not configured")

// GetLazadaClient creates a Lazada API client for a tenant.
// This is the shared implementation used by all Lazada handlers.
// It uses PlatformCredentialsRepository for tenant-specific credentials,
// with fallback to GlobalConfigRepository for app credentials.
func GetLazadaClient(tenantID, basePath string) (*lazadaPkg.Client, error) {
	ctx := context.Background()

	db, err := config.GetTenantDB(tenantID, basePath)
	if err != nil {
		return nil, err
	}

	// Get platform credentials from tenant's key-value storage
	// NOTE: PostgreSQL uses schema isolation, NOT tenant_id column
	credRepo := repositories.NewPlatformCredentialsRepository(db)
	tenantCreds, err := credRepo.GetLazadaCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if tenantCreds.AccessToken == "" {
		return nil, ErrMissingAccessToken
	}

	// Use tenant credentials for appKey/appSecret if available
	appKey := tenantCreds.AppKey
	appSecret := tenantCreds.AppSecret
	region := tenantCreds.Region
	if region == "" {
		region = "ID" // Default to Indonesia
	}

	// Fall back to global credentials if tenant doesn't have app credentials
	if appKey == "" || appSecret == "" {
		systemDB, err := config.GetSystemDB(basePath)
		if err != nil {
			return nil, err
		}
		globalRepo := repositories.NewGlobalConfigRepository(systemDB)
		globalCreds, err := globalRepo.GetLazadaCredentials(ctx)
		if err != nil {
			return nil, err
		}
		appKey = globalCreds.AppKey
		appSecret = globalCreds.AppSecret
		if region == "" {
			region = globalCreds.Region
		}
	}

	client := lazadaPkg.NewClient(appKey, appSecret, region)
	client.SetAccessToken(tenantCreds.AccessToken)
	return client, nil
}
