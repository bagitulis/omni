package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/omni/backend/internal/models"
)

// Bug A — real remote OAuth refresh wire.
//
// Before this file, `POST /credentials/platforms/:p/connections/:s/refresh`
// only flipped the DB status column and wrote a success audit event; it
// never contacted the platform OAuth endpoint. This lied to callers: a
// hard-expired refresh token still produced HTTP 200 "success".
//
// The fix injects a CredentialTokenRefresher into CredentialApiService and
// delegates the actual token exchange to it. TokenManager already
// implements the interface (it does the real HTTP call to Shopee / Lazada /
// TikTok auth endpoints); the service layer just needs to call it.

// CredentialTokenRefresher is the minimal interface the credential API
// service needs from a token manager. Kept small so tests can substitute a
// fake without pulling in the whole TokenManager surface.
type CredentialTokenRefresher interface {
	RefreshShopeeToken(ctx context.Context, tenantID string) (accessToken, refreshToken string, err error)
	RefreshLazadaToken(ctx context.Context, tenantID string) (accessToken, refreshToken string, err error)
	RefreshTiktokToken(ctx context.Context, tenantID string) (accessToken, refreshToken string, err error)
}

// errRemoteRefreshFailed wraps an underlying platform-refresh error with a
// stable identity so higher-level code can distinguish "remote refresh
// failed" from "wire missing" or "unknown platform".
func errRemoteRefreshFailed(msg string) error {
	return fmt.Errorf("remote refresh failed: %s", msg)
}

// errRefresherNotWired is returned when the service was constructed without
// a refresher. Callers should return an HTTP 500 with this message so the
// operator knows to fix the wiring at app start.
var errRefresherNotWired = errors.New("token refresher not wired: credential API service cannot refresh remote tokens")

// invokeRemoteRefresh dispatches to the right platform-specific method on
// the injected refresher. Pure translation — no DB, no audit, no state
// change — so it's trivial to unit-test with a fake refresher.
func invokeRemoteRefresh(ctx context.Context, r CredentialTokenRefresher, platform, tenantID string) (accessToken, refreshToken string, err error) {
	if r == nil {
		return "", "", errRefresherNotWired
	}
	switch platform {
	case models.PlatformShopee:
		return r.RefreshShopeeToken(ctx, tenantID)
	case models.PlatformLazada:
		return r.RefreshLazadaToken(ctx, tenantID)
	case models.PlatformTiktok:
		return r.RefreshTiktokToken(ctx, tenantID)
	default:
		return "", "", fmt.Errorf("unsupported platform for refresh: %q", platform)
	}
}
