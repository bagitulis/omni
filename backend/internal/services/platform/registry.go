package platform

import (
	"context"
	"sync"
)

// TokenRefreshService interface for token refresh operations
type TokenRefreshService interface {
	RefreshShopeeToken(ctx context.Context, tenantID string) (accessToken, refreshToken string, err error)
	RefreshLazadaToken(ctx context.Context, tenantID string) (accessToken, refreshToken string, err error)
	RefreshTiktokToken(ctx context.Context, tenantID string) (accessToken, refreshToken string, err error)
}

// globalTokenRefreshService holds the registered token refresh service
var (
	globalTokenRefreshService TokenRefreshService
	tokenRefreshMu            sync.RWMutex
)

// RegisterTokenRefreshService registers the global token refresh service
// Should be called once during app initialization
func RegisterTokenRefreshService(svc TokenRefreshService) {
	tokenRefreshMu.Lock()
	defer tokenRefreshMu.Unlock()
	globalTokenRefreshService = svc
}

// GetTokenRefreshService returns the registered token refresh service
func GetTokenRefreshService() TokenRefreshService {
	tokenRefreshMu.RLock()
	defer tokenRefreshMu.RUnlock()
	return globalTokenRefreshService
}

// GetPlatformCoordinationService returns a NEW service for a tenant each time
// NO CACHING - always creates fresh instance to ensure latest tokens are used
func GetPlatformCoordinationService(tenantID string) *PlatformCoordinationService {
	if tenantID == "" {
		return nil
	}
	// Always create new instance - no caching
	return NewPlatformCoordinationService(tenantID)
}

// InitializePlatformService initializes platform service for a tenant
func InitializePlatformService(ctx context.Context, tenantID string) error {
	service := GetPlatformCoordinationService(tenantID)
	if service == nil {
		return nil
	}
	return service.InitializePlatforms(ctx)
}

// ClearPlatformServices is a no-op since we don't cache anymore
// Kept for API compatibility
func ClearPlatformServices() {
	// No-op - no cache to clear
}

// InvalidateTenantPlatformService is a no-op since we don't cache anymore
// Kept for API compatibility
func InvalidateTenantPlatformService(tenantID string) {
	// No-op - no cache to invalidate
}
