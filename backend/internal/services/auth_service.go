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

// AuthService handles authentication operations
type AuthService struct {
	userRepo     *repositories.UserRepository
	auditRepo    *repositories.AuditRepository
	jwtService   *utils.JWTService
	maxAttempts  int
	lockDuration time.Duration
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repositories.UserRepository, auditRepo *repositories.AuditRepository, jwtService *utils.JWTService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		auditRepo:    auditRepo,
		jwtService:   jwtService,
		maxAttempts:  5,
		lockDuration: 30 * time.Minute,
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

	// Generate tokens - tenantID comes from request, not user
	accessToken, err := s.jwtService.GenerateToken(user.ID, req.TenantID, user.Role)
	if err != nil {
		return nil, err
	}

	// Use same token for refresh (simplified - in production use separate refresh token)
	refreshToken := accessToken

	// Log successful login
	s.logSuccessfulLogin(ctx, user.ID, req.TenantID, req.IPAddress, req.UserAgent)

	userResp := user.ToResponse()
	return &LoginResponse{
		User:         &userResp,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}, nil
}

// RefreshToken refreshes access token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return "", ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil || user == nil {
		return "", ErrUserNotFound
	}

	return s.jwtService.GenerateToken(user.ID, claims.TenantID, user.Role)
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

// ChangePassword changes user password
func (s *AuthService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	hash, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hash
	return s.userRepo.Update(ctx, user)
}

// GenerateTokenForSwitch generates a new token for tenant switch
func (s *AuthService) GenerateTokenForSwitch(userID, tenantID, role string) (string, error) {
	return s.jwtService.GenerateToken(userID, tenantID, role)
}
