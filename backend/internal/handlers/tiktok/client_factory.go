package tiktok

import (
	"fmt"
	"github.com/omni/backend/internal/services"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// NewTiktokClient creates a TikTok API client with tenant-specific credentials.
// This is the single source of truth for TikTok client initialization across all handlers.
// It resolves credentials through CredentialService with canonical-first lookup and legacy fallback.
func NewTiktokClient(tenantID, basePath string) (*tiktokPkg.Client, error) {
	credService := services.NewCredentialService(basePath)
	creds, err := credService.GetPlatformCredentials(tenantID, "tiktok")
	if err != nil {
		return nil, fmt.Errorf("tiktok credentials: %w", err)
	}
	if creds.AccessToken == "" || creds.ShopCipher == "" {
		return nil, fmt.Errorf("missing TikTok credentials: accessToken or shopCipher not configured")
	}
	client := tiktokPkg.NewClient(creds.AppKey, creds.AppSecret)
	client.SetCredentials(creds.AccessToken, creds.ShopCipher)
	return client, nil
}
