package handlers

import (
	"fmt"
	"net/url"

	"github.com/omni/backend/internal/services/oauth"
)

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
