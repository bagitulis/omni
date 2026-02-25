package services

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	authSvc := NewAuthService(nil, nil, jwtSvc)

	tests := []struct {
		name  string
		token string
	}{
		{name: "empty token", token: ""},
		{name: "invalid format", token: "not-a-valid-jwt"},
		{name: "random string", token: "abc.def.ghi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// JWT validation fails before userRepo is called, so no panic with nil userRepo
			_, err := authSvc.RefreshToken(context.Background(), tt.token)
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidToken, err)
		})
	}
}

func TestAuthService_RefreshToken_WrongSecret(t *testing.T) {
	// Token signed with different secret
	otherJwtSvc := utils.NewJWTService("other-secret")
	token, err := otherJwtSvc.GenerateAccessToken("user-1", "tenant-1", "admin")
	require.NoError(t, err)

	// Service with different secret cannot validate it
	jwtSvc := utils.NewJWTService("test-secret")
	authSvc := NewAuthService(nil, nil, jwtSvc)

	_, err = authSvc.RefreshToken(context.Background(), token)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestAuthService_ValidateToken_AfterGenerate(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret-for-validate")
	authSvc := NewAuthService(nil, nil, jwtSvc)

	token, err := jwtSvc.GenerateAccessToken("user-99", "tenant-99", "owner")
	require.NoError(t, err)

	claims, err := authSvc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-99", claims.UserID)
	assert.Equal(t, "tenant-99", claims.TenantID)
	assert.Equal(t, "owner", claims.Role)
}

func TestAuthService_GenerateTokenForSwitch_RoundTrip(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret-switch")
	authSvc := NewAuthService(nil, nil, jwtSvc)

	token, err := authSvc.GenerateTokenForSwitch("u1", "t1", "developer")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := authSvc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "u1", claims.UserID)
	assert.Equal(t, "t1", claims.TenantID)
	assert.Equal(t, "developer", claims.Role)
}

func TestAuthService_Logout_WithNilRefreshRepo(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	authSvc := NewAuthService(nil, nil, jwtSvc)

	// Should be a no-op (nil refresh repo)
	err := authSvc.Logout(context.Background(), "any-token")
	assert.NoError(t, err)
}

func TestAuthService_LogoutAllDevices_WithNilRefreshRepo(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	authSvc := NewAuthService(nil, nil, jwtSvc)

	// Should be a no-op (nil refresh repo)
	err := authSvc.LogoutAllDevices(context.Background(), "user-123")
	assert.NoError(t, err)
}

func TestAuthService_RefreshTokenSecure_FallbackInvalidToken(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	// nil refreshSessionRepo → falls back to old RefreshToken path
	authSvc := NewAuthService(nil, nil, jwtSvc)

	_, _, err := authSvc.RefreshTokenSecure(context.Background(), "bad-token", "127.0.0.1", "go-test")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
}
