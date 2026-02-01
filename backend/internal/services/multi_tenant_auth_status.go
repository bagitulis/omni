package services

import (
	"context"

	"github.com/omni/backend/internal/repositories"
)

// LoginStatusResponse represents login status
type LoginStatusResponse struct {
	IsLocked             bool `json:"is_locked"`
	RequiresCaptcha      bool `json:"requires_captcha"`
	LockMinutesRemaining int  `json:"lock_minutes_remaining,omitempty"`
}

// GetLoginStatus checks login status across all tenants (for CAPTCHA)
func (s *MultiTenantAuthService) GetLoginStatus(ctx context.Context, username string) (*LoginStatusResponse, error) {
	tenants, err := s.tenantService.GetAvailableTenants(ctx)
	if err != nil {
		return nil, err
	}

	for _, tenant := range tenants {
		db, err := s.tenantService.GetTenantDB(tenant.ID)
		if err != nil {
			continue
		}

		userRepo := repositories.NewUserRepository(db)
		user, err := userRepo.FindByUsernameOrEmail(ctx, username)
		if err != nil || user == nil {
			continue
		}

		// Check if locked or needs CAPTCHA
		if user.IsLocked() {
			return &LoginStatusResponse{
				IsLocked:             true,
				RequiresCaptcha:      false,
				LockMinutesRemaining: user.LockMinutesRemaining(),
			}, nil
		}

		// Require CAPTCHA after 3 failed attempts
		if user.FailedLoginAttempts >= 3 {
			return &LoginStatusResponse{
				IsLocked:        false,
				RequiresCaptcha: true,
			}, nil
		}
	}

	return &LoginStatusResponse{
		IsLocked:        false,
		RequiresCaptcha: false,
	}, nil
}

// SwitchTenant generates new token for different tenant (developer only)
func (s *MultiTenantAuthService) SwitchTenant(ctx context.Context, userID, currentRole, newTenantID string) (string, error) {
	// Only developer can switch tenants
	if currentRole != "developer" {
		return "", &AuthError{Code: "FORBIDDEN", Message: "Only developers can switch tenants"}
	}

	// Verify tenant exists
	if !s.tenantService.TenantExists(ctx, newTenantID) {
		return "", &AuthError{Code: "TENANT_NOT_FOUND", Message: "Tenant not found"}
	}

	// Generate new token for the target tenant
	return s.jwtService.GenerateToken(userID, newTenantID, currentRole)
}

// GetAvailableTenants returns tenants for UI display
func (s *MultiTenantAuthService) GetAvailableTenants(ctx context.Context) ([]TenantInfo, error) {
	return s.tenantService.GetAvailableTenants(ctx)
}
