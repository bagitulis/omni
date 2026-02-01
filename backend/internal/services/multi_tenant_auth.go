package services

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/utils"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// MultiTenantAuthService handles authentication across all tenants
// Similar to Node.js findUserAcrossTenants pattern
type MultiTenantAuthService struct {
	tenantService *TenantService
	jwtService    *utils.JWTService
	basePath      string
}

// NewMultiTenantAuthService creates a new multi-tenant auth service
func NewMultiTenantAuthService(tenantService *TenantService, jwtService *utils.JWTService, basePath string) *MultiTenantAuthService {
	return &MultiTenantAuthService{
		tenantService: tenantService,
		jwtService:    jwtService,
		basePath:      basePath,
	}
}

// MultiTenantLoginRequest represents login request
type MultiTenantLoginRequest struct {
	Username     string
	Password     string
	IPAddress    string
	UserAgent    string
	CaptchaToken string
}

// MultiTenantLoginResponse represents login response
type MultiTenantLoginResponse struct {
	User         *models.UserResponse `json:"user"`
	AccessToken  string               `json:"token"`
	RefreshToken string               `json:"refresh_token"`
	TenantID     string               `json:"tenant_id"`
	ExpiresAt    time.Time            `json:"expires_at"`
}

// LoginAcrossTenants finds user across all tenants and authenticates
// This mimics Node.js findUserAcrossTenants behavior
// Priority: 1. System schema (developer accounts) 2. Tenant schemas
func (s *MultiTenantAuthService) LoginAcrossTenants(ctx context.Context, req *MultiTenantLoginRequest) (*MultiTenantLoginResponse, error) {
	// Step 1: Check system schema first for developer accounts
	log.Debug().
		Str("service", "multi_tenant_auth").
		Str("username", req.Username).
		Msg("Checking system schema for user")
	result, err := s.tryLoginInSystem(ctx, req)
	if err == nil && result != nil {
		log.Debug().
			Str("service", "multi_tenant_auth").
			Msg("Developer user found in system schema")
		return result, nil
	}
	if err != nil {
		// If account is locked, return immediately
		if authErr, ok := err.(*AuthError); ok {
			if authErr.Code == "ACCOUNT_LOCKED" {
				return nil, err
			}
		}
		log.Debug().
			Str("service", "multi_tenant_auth").
			Err(err).
			Msg("Not found in system")
	}

	// Step 2: Check all tenant schemas
	tenants, err := s.tenantService.GetAvailableTenants(ctx)
	if err != nil {
		return nil, err
	}

	log.Debug().
		Str("service", "multi_tenant_auth").
		Int("tenant_count", len(tenants)).
		Str("username", req.Username).
		Msg("Scanning tenants for user")

	for _, tenant := range tenants {
		log.Debug().
			Str("service", "multi_tenant_auth").
			Str("tenant_id", tenant.ID).
			Msg("Checking tenant")

		result, err := s.tryLoginInTenant(ctx, tenant.ID, req)
		if err != nil {
			log.Debug().
				Str("service", "multi_tenant_auth").
				Str("tenant_id", tenant.ID).
				Err(err).
				Msg("Not found in tenant")

			// If account is locked, return immediately
			if authErr, ok := err.(*AuthError); ok {
				if authErr.Code == "ACCOUNT_LOCKED" {
					return nil, err
				}
			}
			continue
		}

		if result != nil {
			log.Debug().
				Str("service", "multi_tenant_auth").
				Str("tenant_id", tenant.ID).
				Msg("User found in tenant")
			return result, nil
		}
	}

	return nil, ErrInvalidCredentials
}

// tryLoginInTenant attempts to login in a specific tenant
func (s *MultiTenantAuthService) tryLoginInTenant(ctx context.Context, tenantID string, req *MultiTenantLoginRequest) (*MultiTenantLoginResponse, error) {
	db, err := s.tenantService.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}

	userRepo := repositories.NewUserRepository(db)

	// Find user by username or email
	user, err := userRepo.FindByUsernameOrEmail(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Check if account is locked
	if user.IsLocked() {
		return nil, ErrAccountLocked
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Increment failed attempts
		userRepo.IncrementFailedAttempts(ctx, user.ID)

		// Lock account if max attempts exceeded (10 attempts - tuned for office users)
		if user.FailedLoginAttempts+1 >= DefaultMaxAttempts {
			userRepo.LockAccount(ctx, user.ID, DefaultLockDuration)
		}

		return nil, ErrInvalidCredentials
	}

	// Reset failed attempts on successful login
	userRepo.ResetFailedAttempts(ctx, user.ID)

	// Generate access token
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, tenantID, user.Role)
	if err != nil {
		return nil, err
	}

	// Generate secure refresh token
	refreshToken, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store refresh session in database
	refreshSessionRepo := repositories.NewRefreshSessionRepository(db)
	refreshSession := &models.RefreshSession{
		UserID:    user.ID,
		TenantID:  tenantID,
		TokenHash: repositories.HashToken(refreshToken),
		ExpiresAt: time.Now().Add(utils.RefreshTokenTTL),
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
	}
	if err := refreshSessionRepo.Create(ctx, refreshSession); err != nil {
		log.Warn().Err(err).Msg("Failed to store refresh session, continuing without it")
		// Continue without refresh session - fallback to JWT-only refresh
		refreshToken = accessToken
	}

	userResp := user.ToResponse()
	return &MultiTenantLoginResponse{
		User:         &userResp,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TenantID:     tenantID,
		ExpiresAt:    time.Now().Add(utils.AccessTokenTTL),
	}, nil
}

// tryLoginInSystem attempts to login for developer accounts in system schema
// Developer accounts are stored in system.users and can switch to any tenant
func (s *MultiTenantAuthService) tryLoginInSystem(ctx context.Context, req *MultiTenantLoginRequest) (*MultiTenantLoginResponse, error) {
	db, err := s.tenantService.GetSystemDB()
	if err != nil {
		return nil, err
	}

	userRepo := repositories.NewUserRepository(db)

	// Find user by username or email in system schema
	user, err := userRepo.FindByUsernameOrEmail(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Check if account is locked
	if user.IsLocked() {
		return nil, ErrAccountLocked
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Increment failed attempts
		userRepo.IncrementFailedAttempts(ctx, user.ID)

		// Lock account if max attempts exceeded (10 attempts - tuned for office users)
		if user.FailedLoginAttempts+1 >= DefaultMaxAttempts {
			userRepo.LockAccount(ctx, user.ID, DefaultLockDuration)
		}

		return nil, ErrInvalidCredentials
	}

	// Reset failed attempts on successful login
	userRepo.ResetFailedAttempts(ctx, user.ID)

	// For developer accounts, use first available tenant as default
	// They can switch tenant later via /api/auth/switch-tenant
	defaultTenantID := ""
	tenants, _ := s.tenantService.GetAvailableTenants(ctx)
	if len(tenants) > 0 {
		defaultTenantID = tenants[0].ID
	}

	// Generate access token
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, defaultTenantID, user.Role)
	if err != nil {
		return nil, err
	}

	// Generate secure refresh token
	refreshToken, err := s.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Store refresh session in system database
	refreshSessionRepo := repositories.NewRefreshSessionRepository(db)
	refreshSession := &models.RefreshSession{
		UserID:    user.ID,
		TenantID:  defaultTenantID,
		TokenHash: repositories.HashToken(refreshToken),
		ExpiresAt: time.Now().Add(utils.RefreshTokenTTL),
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
	}
	if err := refreshSessionRepo.Create(ctx, refreshSession); err != nil {
		log.Warn().Err(err).Msg("Failed to store refresh session for developer, continuing without it")
		refreshToken = accessToken
	}

	userResp := user.ToResponse()
	return &MultiTenantLoginResponse{
		User:         &userResp,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TenantID:     defaultTenantID,
		ExpiresAt:    time.Now().Add(utils.AccessTokenTTL),
	}, nil
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

// LoginStatusResponse represents login status
type LoginStatusResponse struct {
	IsLocked             bool `json:"is_locked"`
	RequiresCaptcha      bool `json:"requires_captcha"`
	LockMinutesRemaining int  `json:"lock_minutes_remaining,omitempty"`
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

	// SECURITY: Check for token reuse attack
	if session.IsReused() {
		// Token was already rotated! This is a potential theft
		// Revoke ALL sessions for this user as precaution
		refreshSessionRepo.RevokeSessionChain(ctx, session.UserID)
		return "", "", &AuthError{Code: "TOKEN_REUSE", Message: "Refresh token reuse detected, all sessions revoked"}
	}

	// Check if session is still valid
	if !session.IsValid() {
		return "", "", ErrInvalidToken
	}

	// Get user for token generation
	user, err := userRepo.FindByID(ctx, session.UserID)
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
	if err := refreshSessionRepo.Rotate(ctx, tokenHash, newTokenHash, newSession); err != nil {
		log.Warn().Err(err).Msg("Failed to rotate refresh session")
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}
