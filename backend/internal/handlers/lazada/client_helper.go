// Package lazada provides Lazada handler helpers
package lazada

import (
	"errors"

	"github.com/omni/backend/internal/services"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
)

// ErrMissingAccessToken indicates that the tenant has no access token configured
var ErrMissingAccessToken = errors.New("lazada access token not configured")

// GetLazadaClient creates a Lazada API client for a tenant.
// This is the shared implementation used by all Lazada handlers.
// It uses CredentialService for canonical-first tenant credentials with legacy fallback.
func GetLazadaClient(tenantID, basePath string) (*lazadaPkg.Client, error) {
	credService := services.NewCredentialService(basePath)
	creds, err := credService.GetPlatformCredentials(tenantID, "lazada")
	if err != nil {
		return nil, err
	}
	if creds.AccessToken == "" {
		return nil, ErrMissingAccessToken
	}
	region := creds.Region
	if region == "" {
		region = "ID" // Default to Indonesia
	}

	client := lazadaPkg.NewClient(creds.AppKey, creds.AppSecret, region)
	client.SetAccessToken(creds.AccessToken)
	return client, nil
}
