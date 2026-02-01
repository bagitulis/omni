package services

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/utils"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// RefreshTokenForTenant refreshes tokens with rotation and reuse detection
// Returns: (newAccessToken, newRefreshToken, error)
func (s *MultiTenantAuthService) RefreshTokenForTenant(ctx context.Context, refreshToken, tenantID, ipAddress, userAgent string) (string, string, error) {
	tokenHash := repositories.HashToken(refreshToken)

	var db *gorm.DB
	var session *models.RefreshSession
	var refreshSessionRepo *repositories.RefreshSessionRepository
	var userRepo *repositories.UserRepository

	// Strategy: If tenantID provided, try that first. Otherwise search all DBs.
	if tenantID != "" {
		// Try specific tenant DB first
		tenantDB, err := s.tenantService.GetTenantDB(tenantID)
		if err == nil {
			refreshSessionRepo = repositories.NewRefreshSessionRepository(tenantDB)
			session, _ = refreshSessionRepo.FindByTokenHash(ctx, tokenHash)
			if session != nil {
				db = tenantDB
				userRepo = repositories.NewUserRepository(db)
			}
		}
	}

	// If not found yet, search all tenant DBs
	if session == nil {
		tenants, _ := s.tenantService.GetAvailableTenants(ctx)
		for _, tenant := range tenants {
			tenantDB, err := s.tenantService.GetTenantDB(tenant.ID)
			if err != nil {
				continue
			}
			refreshSessionRepo = repositories.NewRefreshSessionRepository(tenantDB)
			session, _ = refreshSessionRepo.FindByTokenHash(ctx, tokenHash)
			if session != nil {
				db = tenantDB
				userRepo = repositories.NewUserRepository(db)
				break
			}
		}
	}

	// Finally try system DB
	if session == nil {
		systemDB, err := s.tenantService.GetSystemDB()
		if err != nil {
			return "", "", ErrInvalidToken
		}
		refreshSessionRepo = repositories.NewRefreshSessionRepository(systemDB)
		session, _ = refreshSessionRepo.FindByTokenHash(ctx, tokenHash)
		if session != nil {
			db = systemDB
			userRepo = repositories.NewUserRepository(db)
		}
	}

	// Not found anywhere
	if session == nil {
		return "", "", ErrInvalidToken
	}

	// Validate session and check for token reuse
	if err := s.validateRefreshSession(ctx, session, refreshSessionRepo); err != nil {
		return "", "", err
	}

	// Get user for token generation
	user, err := userRepo.FindByID(ctx, session.UserID)
	if err != nil || user == nil {
		return "", "", ErrUserNotFound
	}

	// Generate new tokens
	newAccessToken, newRefreshToken, err := s.generateNewTokenPair(ctx, user, session)
	if err != nil {
		return "", "", err
	}

	// Rotate: mark old token as replaced and create new session
	if err := s.rotateRefreshSession(ctx, refreshSessionRepo, tokenHash, newRefreshToken, session, ipAddress, userAgent); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

// validateRefreshSession checks for token reuse and expiration
func (s *MultiTenantAuthService) validateRefreshSession(ctx context.Context, session *models.RefreshSession, repo *repositories.RefreshSessionRepository) error {
	// SECURITY: Check for token reuse attack
	if session.IsReused() {
		// Token was already rotated! This is a potential theft
		// Revoke ALL sessions for this user as precaution
		repo.RevokeSessionChain(ctx, session.UserID)
		return &AuthError{Code: "TOKEN_REUSE", Message: "Refresh token reuse detected, all sessions revoked"}
	}

	// Check if session is still valid
	if !session.IsValid() {
		return ErrInvalidToken
	}

	return nil
}

// generateNewTokenPair creates a new access and refresh token pair
func (s *MultiTenantAuthService) generateNewTokenPair(ctx context.Context, user *models.User, session *models.RefreshSession) (string, string, error) {
	newAccessToken, err := s.jwtService.GenerateAccessToken(user.ID, session.TenantID, user.Role)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

// rotateRefreshSession performs token rotation with session tracking
func (s *MultiTenantAuthService) rotateRefreshSession(ctx context.Context, repo *repositories.RefreshSessionRepository, oldTokenHash, newRefreshToken string, oldSession *models.RefreshSession, ipAddress, userAgent string) error {
	newSession := &models.RefreshSession{
		UserID:    oldSession.UserID,
		TenantID:  oldSession.TenantID,
		ExpiresAt: time.Now().Add(utils.RefreshTokenTTL),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	newTokenHash := repositories.HashToken(newRefreshToken)
	if err := repo.Rotate(ctx, oldTokenHash, newTokenHash, newSession); err != nil {
		log.Warn().Err(err).Msg("Failed to rotate refresh session")
		return err
	}
	return nil
}
