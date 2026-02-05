package services

import (
	"context"
	"testing"
	"time"

	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthError_Error(t *testing.T) {
	err := &AuthError{Code: "TEST_CODE", Message: "Test message"}
	assert.Equal(t, "Test message", err.Error())
}

func TestAuthService_HashPassword(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	authSvc := NewAuthService(nil, nil, jwtSvc)

	password := "MySecurePassword123!"
	hash, err := authSvc.HashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Verify hash is valid
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	assert.NoError(t, err)
}

func TestAuthService_ValidateToken(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret-key")
	authSvc := NewAuthService(nil, nil, jwtSvc)

	tests := []struct {
		name      string
		userID    string
		tenantID  string
		role      string
		wantError bool
	}{
		{
			name:      "valid token",
			userID:    "user-123",
			tenantID:  "tenant-abc",
			role:      "admin",
			wantError: false,
		},
		{
			name:      "owner role token",
			userID:    "user-456",
			tenantID:  "tenant-xyz",
			role:      "owner",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate token
			token, err := jwtSvc.GenerateAccessToken(tt.userID, tt.tenantID, tt.role)
			require.NoError(t, err)

			// Validate token
			claims, err := authSvc.ValidateToken(token)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.userID, claims.UserID)
				assert.Equal(t, tt.tenantID, claims.TenantID)
				assert.Equal(t, tt.role, claims.Role)
			}
		})
	}
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret-key")
	authSvc := NewAuthService(nil, nil, jwtSvc)

	tests := []struct {
		name  string
		token string
	}{
		{name: "empty token", token: ""},
		{name: "invalid format", token: "not-a-jwt-token"},
		{name: "wrong secret", token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzIn0.wrong"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := authSvc.ValidateToken(tt.token)
			assert.Error(t, err)
		})
	}
}

func TestAuthService_GenerateTokenForSwitch(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret-key")
	authSvc := NewAuthService(nil, nil, jwtSvc)

	tests := []struct {
		name     string
		userID   string
		tenantID string
		role     string
	}{
		{
			name:     "switch to new tenant",
			userID:   "user-123",
			tenantID: "new-tenant",
			role:     "admin",
		},
		{
			name:     "switch with different role",
			userID:   "user-456",
			tenantID: "another-tenant",
			role:     "owner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := authSvc.GenerateTokenForSwitch(tt.userID, tt.tenantID, tt.role)
			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// Verify token contains correct claims
			claims, err := jwtSvc.ValidateToken(token)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, claims.UserID)
			assert.Equal(t, tt.tenantID, claims.TenantID)
			assert.Equal(t, tt.role, claims.Role)
		})
	}
}

func TestAuthService_Logout_NoRefreshRepo(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	authSvc := NewAuthService(nil, nil, jwtSvc)

	// Logout should be no-op when no refresh repo
	err := authSvc.Logout(context.Background(), "some-token")
	assert.NoError(t, err)
}

func TestAuthService_LogoutAllDevices_NoRefreshRepo(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	authSvc := NewAuthService(nil, nil, jwtSvc)

	// LogoutAllDevices should be no-op when no refresh repo
	err := authSvc.LogoutAllDevices(context.Background(), "user-123")
	assert.NoError(t, err)
}

func TestAuthService_RefreshTokenSecure_NoRefreshRepo(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	authSvc := NewAuthService(nil, nil, jwtSvc)

	// RefreshTokenSecure without refresh repo should fallback to old method
	// which will fail since we pass invalid token
	_, _, err := authSvc.RefreshTokenSecure(context.Background(), "invalid-token", "127.0.0.1", "test-agent")
	assert.Error(t, err)
}

func TestDefaultSecuritySettings(t *testing.T) {
	// Verify default security settings are reasonable
	assert.Equal(t, 10, DefaultMaxAttempts)
	assert.Equal(t, 15*time.Minute, DefaultLockDuration)
}

func TestAuthErrors(t *testing.T) {
	// Verify error definitions
	tests := []struct {
		err      *AuthError
		code     string
		hasError bool
	}{
		{ErrInvalidCredentials, "INVALID_CREDENTIALS", true},
		{ErrAccountLocked, "ACCOUNT_LOCKED", true},
		{ErrUserNotFound, "USER_NOT_FOUND", true},
		{ErrInvalidToken, "INVALID_TOKEN", true},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			assert.Equal(t, tt.code, tt.err.Code)
			assert.NotEmpty(t, tt.err.Error())
		})
	}
}

func TestNewAuthService(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	authSvc := NewAuthService(nil, nil, jwtSvc)

	assert.NotNil(t, authSvc)
	assert.Equal(t, DefaultMaxAttempts, authSvc.maxAttempts)
	assert.Equal(t, DefaultLockDuration, authSvc.lockDuration)
}

func TestNewAuthServiceWithRefresh(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	authSvc := NewAuthServiceWithRefresh(nil, nil, nil, jwtSvc)

	assert.NotNil(t, authSvc)
	assert.Equal(t, DefaultMaxAttempts, authSvc.maxAttempts)
	assert.Equal(t, DefaultLockDuration, authSvc.lockDuration)
}
