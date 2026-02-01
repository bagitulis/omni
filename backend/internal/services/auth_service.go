package services

import (
	"context"
	"errors"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// AuthError represents authentication error
type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

// Common auth errors
var (
	ErrInvalidCredentials = &AuthError{Code: "INVALID_CREDENTIALS", Message: "Invalid username or password"}
	ErrAccountLocked      = &AuthError{Code: "ACCOUNT_LOCKED", Message: "Account is locked"}
	ErrUserNotFound       = &AuthError{Code: "USER_NOT_FOUND", Message: "User not found"}
	ErrInvalidToken       = &AuthError{Code: "INVALID_TOKEN", Message: "Invalid token"}
)

// Security settings - tuned for marketplace operations
const (
	DefaultMaxAttempts  = 10               // Increased from 5 - office users typo, password manager failures
	DefaultLockDuration = 15 * time.Minute // Reduced from 30min - less impact during flash sales
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo           *repositories.UserRepository
	auditRepo          *repositories.AuditRepository
	refreshSessionRepo *repositories.RefreshSessionRepository
	jwtService         *utils.JWTService
	maxAttempts        int
	lockDuration       time.Duration
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repositories.UserRepository, auditRepo *repositories.AuditRepository, jwtService *utils.JWTService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		auditRepo:    auditRepo,
		jwtService:   jwtService,
		maxAttempts:  DefaultMaxAttempts,
		lockDuration: DefaultLockDuration,
	}
}

// NewAuthServiceWithRefresh creates auth service with refresh session support
func NewAuthServiceWithRefresh(userRepo *repositories.UserRepository, auditRepo *repositories.AuditRepository, refreshSessionRepo *repositories.RefreshSessionRepository, jwtService *utils.JWTService) *AuthService {
	return &AuthService{
		userRepo:           userRepo,
		auditRepo:          auditRepo,
		refreshSessionRepo: refreshSessionRepo,
		jwtService:         jwtService,
		maxAttempts:        DefaultMaxAttempts,
		lockDuration:       DefaultLockDuration,
	}
}

// LoginRequest represents login request
type LoginRequest struct {
	Username  string
	Password  string
	TenantID  string // Passed from request context
	IPAddress string
	UserAgent string
}

// LoginResponse represents login response
type LoginResponse struct {
	User         *models.UserResponse
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.FindByUsernameOrEmail(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		s.logFailedLogin(ctx, "", req.TenantID, req.Username, req.IPAddress, req.UserAgent, "User not found")
		return nil, ErrInvalidCredentials
	}

	// Check if account is locked
	if user.IsLocked() {
		s.logFailedLogin(ctx, user.ID, req.TenantID, req.Username, req.IPAddress, req.UserAgent, "Account locked")
		return nil, ErrAccountLocked
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.userRepo.IncrementFailedAttempts(ctx, user.ID)

		// Lock account if max attempts exceeded
		if user.FailedLoginAttempts+1 >= s.maxAttempts {
			s.userRepo.LockAccount(ctx, user.ID, s.lockDuration)
		}

		s.logFailedLogin(ctx, user.ID, req.TenantID, req.Username, req.IPAddress, req.UserAgent, "Invalid password")
		return nil, ErrInvalidCredentials
	}

	// Reset failed attempts on successful login
	s.userRepo.ResetFailedAttempts(ctx, user.ID)

	// Generate access token (short-lived)
	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, req.TenantID, user.Role)
	if err != nil {
		return nil, err
	}

	// Generate refresh token and store session if repository available
	var refreshToken string
	expiresAt := time.Now().Add(utils.AccessTokenTTL)

	if s.refreshSessionRepo != nil {
		refreshToken, err = s.jwtService.GenerateRefreshToken()
		if err != nil {
			return nil, err
		}

		// Create refresh session in database
		session := &models.RefreshSession{
			UserID:    user.ID,
			TenantID:  req.TenantID,
			TokenHash: repositories.HashToken(refreshToken),
			ExpiresAt: time.Now().Add(utils.RefreshTokenTTL),
			IPAddress: req.IPAddress,
			UserAgent: req.UserAgent,
		}
		if err := s.refreshSessionRepo.Create(ctx, session); err != nil {
			// Log but don't fail login - fallback to access token only
			// In production, this might be treated as error
			refreshToken = ""
		}
	}

	// Log successful login
	s.logSuccessfulLogin(ctx, user.ID, req.TenantID, req.IPAddress, req.UserAgent)

	userResp := user.ToResponse()
	return &LoginResponse{
		User:         &userResp,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

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

// HashPassword hashes a password
func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// logFailedLogin logs failed login attempt
func (s *AuthService) logFailedLogin(ctx context.Context, userID, tenantID, username, ip, userAgent, reason string) {
	if s.auditRepo == nil {
		return
	}
	s.auditRepo.Create(ctx, &models.AuditLogEntry{
		TenantID:     tenantID,
		Action:       models.AuditActionLoginFailed,
		UserID:       userID,
		Status:       models.AuditStatusFailed,
		ErrorMessage: reason,
		IPAddress:    ip,
		UserAgent:    userAgent,
		Details:      map[string]interface{}{"username": username},
	})
}

// logSuccessfulLogin logs successful login
func (s *AuthService) logSuccessfulLogin(ctx context.Context, userID, tenantID, ip, userAgent string) {
	if s.auditRepo == nil {
		return
	}
	s.auditRepo.Create(ctx, &models.AuditLogEntry{
		TenantID:  tenantID,
		Action:    models.AuditActionUserLogin,
		UserID:    userID,
		Status:    models.AuditStatusSuccess,
		IPAddress: ip,
		UserAgent: userAgent,
	})
}

// ChangePassword changes user password and revokes all sessions for security
func (s *AuthService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Validate new password strength
	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	hash, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// SECURITY: Revoke all refresh sessions after password change
	// This ensures any stolen tokens are invalidated
	if s.refreshSessionRepo != nil {
		if err := s.refreshSessionRepo.RevokeAllForUser(ctx, userID); err != nil {
			// Log error but don't fail - password was already changed
			// In production, consider making this transactional
		}
	}

	return nil
}

// GenerateTokenForSwitch generates a new token for tenant switch
func (s *AuthService) GenerateTokenForSwitch(userID, tenantID, role string) (string, error) {
	return s.jwtService.GenerateAccessToken(userID, tenantID, role)
}
