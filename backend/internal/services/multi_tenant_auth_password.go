package services

import (
	"context"

	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// ChangePasswordForTenant changes password for a user in their tenant
func (s *MultiTenantAuthService) ChangePasswordForTenant(ctx context.Context, userID, tenantID, oldPassword, newPassword string) error {
	// Get tenant DB
	db, err := s.tenantService.GetTenantDB(tenantID)
	if err != nil {
		// Try system DB for developer accounts
		db, err = s.tenantService.GetSystemDB()
		if err != nil {
			return err
		}
	}

	userRepo := repositories.NewUserRepository(db)

	// Find user
	user, err := userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return &AuthError{Code: "INVALID_PASSWORD", Message: "Current password is incorrect"}
	}

	// Validate new password strength
	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	user.Password = string(hashedPassword)
	if err := userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Revoke all refresh sessions for security
	refreshSessionRepo := repositories.NewRefreshSessionRepository(db)
	_ = refreshSessionRepo.RevokeAllForUser(ctx, userID)

	return nil
}
