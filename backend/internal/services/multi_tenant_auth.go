package services

import (
	"context"
	"log"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
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
	log.Printf("[MultiTenantAuth] 🔍 Checking system schema for user: %s", req.Username)
	result, err := s.tryLoginInSystem(ctx, req)
	if err == nil && result != nil {
		log.Printf("[MultiTenantAuth] ✅ Developer user found in system schema")
		return result, nil
	}
	if err != nil {
		// If account is locked, return immediately
		if authErr, ok := err.(*AuthError); ok {
			if authErr.Code == "ACCOUNT_LOCKED" {
				return nil, err
			}
		}
		log.Printf("[MultiTenantAuth] ❌ Not found in system: %v", err)
	}

	// Step 2: Check all tenant schemas
	tenants, err := s.tenantService.GetAvailableTenants(ctx)
	if err != nil {
		return nil, err
	}

	log.Printf("[MultiTenantAuth] 🔍 Scanning %d tenants for user: %s", len(tenants), req.Username)

	for _, tenant := range tenants {
		log.Printf("[MultiTenantAuth] 🔍 Checking tenant: %s", tenant.ID)

		result, err := s.tryLoginInTenant(ctx, tenant.ID, req)
		if err != nil {
			log.Printf("[MultiTenantAuth] ❌ Not found in %s: %v", tenant.ID, err)

			// If account is locked, return immediately
			if authErr, ok := err.(*AuthError); ok {
				if authErr.Code == "ACCOUNT_LOCKED" {
					return nil, err
				}
			}
			continue
		}

		if result != nil {
			log.Printf("[MultiTenantAuth] ✅ User found in tenant: %s", tenant.ID)
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

		// Lock account if max attempts exceeded (5 attempts)
		if user.FailedLoginAttempts+1 >= 5 {
			userRepo.LockAccount(ctx, user.ID, 30*time.Minute)
		}

		return nil, ErrInvalidCredentials
	}

	// Reset failed attempts on successful login
	userRepo.ResetFailedAttempts(ctx, user.ID)

	// Generate JWT token with tenantID
	accessToken, err := s.jwtService.GenerateToken(user.ID, tenantID, user.Role)
	if err != nil {
		return nil, err
	}

	userResp := user.ToResponse()
	return &MultiTenantLoginResponse{
		User:         &userResp,
		AccessToken:  accessToken,
		RefreshToken: accessToken, // Simplified - same as access token
		TenantID:     tenantID,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
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

		// Lock account if max attempts exceeded (5 attempts)
		if user.FailedLoginAttempts+1 >= 5 {
			userRepo.LockAccount(ctx, user.ID, 30*time.Minute)
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

	// Generate JWT token - developer gets first tenant as default
	accessToken, err := s.jwtService.GenerateToken(user.ID, defaultTenantID, user.Role)
	if err != nil {
		return nil, err
	}

	userResp := user.ToResponse()
	return &MultiTenantLoginResponse{
		User:         &userResp,
		AccessToken:  accessToken,
		RefreshToken: accessToken,
		TenantID:     defaultTenantID,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
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
