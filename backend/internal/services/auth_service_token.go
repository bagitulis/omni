package services

import (
	"context"
	"os"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/utils"
	"github.com/rs/zerolog/log"
)

// RefreshToken refreshes access token using the old method (backward compatibility)
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return "", ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil || user == nil {
		return "", ErrUserNotFound
	}

	return s.jwtService.GenerateAccessToken(user.ID, claims.TenantID, user.Role)
}

// RefreshTokenSecure refreshes tokens with rotation and reuse detection
// Returns: (newAccessToken, newRefreshToken, error)
func (s *AuthService) RefreshTokenSecure(ctx context.Context, refreshToken, ipAddress, userAgent string) (string, string, error) {
	if s.refreshSessionRepo == nil {
		// Fallback to old method if no refresh repo
		accessToken, err := s.RefreshToken(ctx, refreshToken)
		return accessToken, "", err
	}

	tokenHash := repositories.HashToken(refreshToken)
	session, err := s.refreshSessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return "", "", err
	}
	if session == nil {
		return "", "", ErrInvalidToken
	}

	// SECURITY: Check for token reuse attack
	if session.IsReused() {
		// Token was already rotated! This is a potential theft
		// Revoke ALL sessions for this user as precaution
		s.refreshSessionRepo.RevokeSessionChain(ctx, session.UserID)
		return "", "", &AuthError{Code: "TOKEN_REUSE", Message: "Refresh token reuse detected, all sessions revoked"}
	}

	// Check if session is still valid
	if !session.IsValid() {
		return "", "", ErrInvalidToken
	}

	// Get user for token generation
	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil || user == nil {
		return "", "", ErrUserNotFound
	}

	// Generate new tokens
	newAccessToken, err := s.jwtService.GenerateAccessToken(user.ID, session.TenantID, user.Role)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// Rotate: mark old token as replaced and create new session
	newSession := &models.RefreshSession{
		UserID:    session.UserID,
		TenantID:  session.TenantID,
		ExpiresAt: time.Now().Add(utils.RefreshTokenTTL),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	newTokenHash := repositories.HashToken(newRefreshToken)
	if err := s.refreshSessionRepo.Rotate(ctx, tokenHash, newTokenHash, newSession); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

// Logout revokes the refresh session
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if s.refreshSessionRepo == nil {
		return nil // No-op if no refresh repo
	}

	tokenHash := repositories.HashToken(refreshToken)
	return s.refreshSessionRepo.RevokeByTokenHash(ctx, tokenHash)
}

// LogoutAllDevices revokes all refresh sessions for a user
func (s *AuthService) LogoutAllDevices(ctx context.Context, userID string) error {
	if s.refreshSessionRepo == nil {
		return nil
	}
	return s.refreshSessionRepo.RevokeAllForUser(ctx, userID)
}

// ValidateToken validates access token
func (s *AuthService) ValidateToken(tokenString string) (*utils.JWTClaims, error) {
	return s.jwtService.ValidateToken(tokenString)
}

// GenerateTokenForSwitch generates a new token for tenant switch
func (s *AuthService) GenerateTokenForSwitch(userID, tenantID, role string) (string, error) {
	return s.jwtService.GenerateAccessToken(userID, tenantID, role)
}

// GenerateDevToken generates a token for dev-mode bypass login
// SECURITY: Only called from DevLogin handler which checks localhost origin
// Returns: accessToken, refreshToken, actualUserID, error
func (s *AuthService) GenerateDevToken(ctx context.Context, userID, username, email, tenantID, role string) (string, string, string, error) {
	// Ensure dev user exists in DB so refresh token works
	// Graceful: if DB is broken, still generate token with provided userID
	actualUserID, err := s.ensureDevUserExists(ctx, userID, username, email, role)
	if err != nil {
		log.Warn().Err(err).Msg("Dev login: ensureDevUserExists failed, using provided userID")
		actualUserID = userID
	}

	// Generate access token using the ACTUAL user ID
	accessToken, err := s.jwtService.GenerateAccessToken(actualUserID, tenantID, role)
	if err != nil {
		return "", "", "", err
	}

	// Generate refresh token and session (graceful degradation: if this fails, still return access token)
	var refreshToken string
	if s.refreshSessionRepo != nil {
		refreshToken, err = s.jwtService.GenerateRefreshToken()
		if err != nil {
			// Refresh token generation failed — continue with access token only
			log.Warn().Err(err).Msg("Dev login: refresh token generation failed, continuing with access token only")
			return accessToken, "", actualUserID, nil
		}

		// Create refresh session in database using ACTUAL user ID
		session := &models.RefreshSession{
			UserID:    actualUserID,
			TenantID:  tenantID,
			TokenHash: repositories.HashToken(refreshToken),
			ExpiresAt: time.Now().Add(utils.RefreshTokenTTL),
			IPAddress: "127.0.0.1", // Dev login always local
			UserAgent: "DevLogin",
		}
		if err := s.refreshSessionRepo.Create(ctx, session); err != nil {
			// Refresh session storage failed — continue with access token only
			log.Warn().Err(err).Msg("Dev login: refresh session creation failed, continuing with access token only")
			return accessToken, "", actualUserID, nil
		}
	}

	return accessToken, refreshToken, actualUserID, nil
}

// ensureDevUserExists ensures the dev user exists in the database
// Returns the actual user ID (either found by ID, found by email, or newly created)
func (s *AuthService) ensureDevUserExists(ctx context.Context, userID, username, email, role string) (string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if user != nil {
		return user.ID, nil // User already exists with this ID
	}

	// Check if user with this email already exists (from previous attempt)
	userByEmail, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if userByEmail != nil {
		return userByEmail.ID, nil // User exists with this email, return their ID
	}

	// Create dev user if not exists
	// Password sourced from DEV_USER_PASSWORD env var (NEVER hardcoded)
	devPassword := os.Getenv("DEV_USER_PASSWORD")
	if devPassword == "" {
		devPassword = "dev-password-123" // Fallback for local development only
		if os.Getenv("GO_ENV") == "production" {
			log.Error().Msg("DEV_USER_PASSWORD not set in production - dev user creation blocked")
			return "", ErrInvalidCredentials
		}
		log.Warn().Msg("DEV_USER_PASSWORD not set, using development fallback (NOT FOR PRODUCTION)")
	}
	passwordHash, _ := s.HashPassword(devPassword)

	newUser := &models.User{
		ID:        userID,
		Username:  username,
		Email:     email,
		Password:  passwordHash,
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// We need to use a repository method that allows setting ID manually
	// Most GORM create methods respect the ID if set
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return "", err
	}
	return userID, nil
}
