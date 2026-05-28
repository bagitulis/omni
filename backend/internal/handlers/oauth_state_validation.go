package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/oauth"
	"github.com/rs/zerolog/log"
)

func isSupportedOAuthPlatform(platform string) bool {
	return platform == models.PlatformShopee || platform == models.PlatformLazada || platform == models.PlatformTiktok
}

func sanitizeRedirectPath(path string) string {
	if path == "" {
		return "/settings"
	}
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return "/settings"
	}
	if !isAllowedOAuthRedirectPath(path) {
		return "/settings"
	}
	return path
}

func isAllowedOAuthRedirectPath(path string) bool {
	switch path {
	case "/settings", "/settings/platforms":
		return true
	default:
		return false
	}
}

func validateStateClaimsForCallback(claims oauth.StateClaims) error {
	if claims.TenantID == "" || claims.UserID == "" || claims.Platform == "" || claims.Marketplace == "" || claims.AttemptID == "" || claims.Nonce == "" || claims.CSRFNonce == "" || claims.ExpiresAt == 0 {
		return fmt.Errorf("invalid_state")
	}
	if claims.Nonce != claims.CSRFNonce {
		return fmt.Errorf("invalid_state")
	}
	if !isAllowedOAuthRedirectPath(claims.RedirectURI) || claims.RedirectPath != claims.RedirectURI {
		return fmt.Errorf("invalid_redirect")
	}
	return nil
}

func redirectURLAllowed(frontendURL, redirectURL, redirectPath string) bool {
	if !isAllowedOAuthRedirectPath(redirectPath) || redirectURL == "" || frontendURL == "" {
		return false
	}
	parsedFrontend, err := url.Parse(frontendURL)
	if err != nil {
		return false
	}
	parsedRedirect, err := url.Parse(redirectURL)
	if err != nil {
		return false
	}
	return parsedRedirect.Scheme == parsedFrontend.Scheme && parsedRedirect.Host == parsedFrontend.Host && parsedRedirect.Path == redirectPath
}

func oauthCallbackStatusCode(status string) int {
	switch status {
	case "no_session":
		return http.StatusUnauthorized
	case "tenant_mismatch", "user_mismatch":
		return http.StatusForbidden
	case "expired_state", "expired", "replayed_state", "invalid_state", "invalid_attempt", "wrong_platform", "wrong_scope", "invalid_redirect":
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}

func stateMatchesClaims(state *models.OAuthState, claims oauth.StateClaims) bool {
	return state.TenantID == claims.TenantID && state.Platform == claims.Platform && state.AttemptID == claims.AttemptID && state.Intent == claims.Intent && state.StoreID == claims.StoreID && state.CSRFNonce == claims.CSRFNonce && state.ExpiresAt.Unix() == claims.ExpiresAt
}

func sanitizeOAuthError(err error) string {
	if err == nil {
		return "failed"
	}
	return sanitizeOAuthErrorText(err.Error())
}

func sanitizeOAuthErrorText(value string) string {
	safe := strings.ToLower(strings.TrimSpace(value))
	safe = strings.ReplaceAll(safe, " ", "_")
	allowed := map[string]bool{"success": true, "no_session": true, "tenant_mismatch": true, "user_mismatch": true, "invalid_state": true, "invalid_redirect": true, "expired_state": true, "expired": true, "replayed": true, "replayed_state": true, "wrong_platform": true, "wrong_scope": true, "invalid_attempt": true, "duplicate_store": true, "store_mismatch": true, "invalid_platform": true, "failed": true}
	if allowed[safe] {
		return safe
	}
	log.Warn().Str("oauth_error_code", "redacted").Msg("OAuth callback failed")
	return "failed"
}
