package services

import (
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMultiTenantAuthService(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/test/path")

	assert.NotNil(t, svc)
	assert.Equal(t, "/test/path", svc.basePath)
}

func TestMultiTenantLoginRequest(t *testing.T) {
	req := MultiTenantLoginRequest{
		Username:     "testuser",
		Password:     "password123",
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
		CaptchaToken: "captcha-token",
	}

	assert.Equal(t, "testuser", req.Username)
	assert.Equal(t, "password123", req.Password)
	assert.Equal(t, "192.168.1.1", req.IPAddress)
	assert.Equal(t, "Mozilla/5.0", req.UserAgent)
	assert.Equal(t, "captcha-token", req.CaptchaToken)
}

func TestMultiTenantLoginResponse(t *testing.T) {
	expiresAt := time.Now().Add(24 * time.Hour)
	userResp := &models.UserResponse{
		ID:       "user-123",
		Username: "testuser",
		Role:     "admin",
	}

	resp := MultiTenantLoginResponse{
		User:         userResp,
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		TenantID:     "tenant-abc",
		ExpiresAt:    expiresAt,
	}

	assert.Equal(t, userResp, resp.User)
	assert.Equal(t, "access-token", resp.AccessToken)
	assert.Equal(t, "refresh-token", resp.RefreshToken)
	assert.Equal(t, "tenant-abc", resp.TenantID)
	assert.Equal(t, expiresAt, resp.ExpiresAt)
}

func TestLoginStatusResponse(t *testing.T) {
	tests := []struct {
		name        string
		response    LoginStatusResponse
		wantLock    bool
		wantCaptcha bool
	}{
		{
			name: "not locked, no captcha",
			response: LoginStatusResponse{
				IsLocked:        false,
				RequiresCaptcha: false,
			},
			wantLock:    false,
			wantCaptcha: false,
		},
		{
			name: "locked account",
			response: LoginStatusResponse{
				IsLocked:             true,
				RequiresCaptcha:      false,
				LockMinutesRemaining: 10,
			},
			wantLock:    true,
			wantCaptcha: false,
		},
		{
			name: "requires captcha",
			response: LoginStatusResponse{
				IsLocked:        false,
				RequiresCaptcha: true,
			},
			wantLock:    false,
			wantCaptcha: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantLock, tt.response.IsLocked)
			assert.Equal(t, tt.wantCaptcha, tt.response.RequiresCaptcha)
		})
	}
}

func TestMultiTenantAuthService_SwitchTenant_NonDeveloper(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/test/path")

	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{name: "admin role", role: "admin", wantErr: true},
		{name: "owner role", role: "owner", wantErr: true},
		{name: "user role", role: "user", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.SwitchTenant(nil, "user-123", tt.role, "new-tenant")
			if tt.wantErr {
				require.Error(t, err)
				authErr, ok := err.(*AuthError)
				require.True(t, ok)
				assert.Equal(t, "FORBIDDEN", authErr.Code)
			}
		})
	}
}

func TestDefaultSecurityConstants(t *testing.T) {
	// Verify default security settings exist
	assert.Equal(t, 10, DefaultMaxAttempts)
	assert.Equal(t, 15*time.Minute, DefaultLockDuration)
}

func TestLoginStatusResponse_LockMinutes(t *testing.T) {
	tests := []struct {
		name            string
		lockMinutes     int
		expectedMinutes int
	}{
		{name: "5 minutes", lockMinutes: 5, expectedMinutes: 5},
		{name: "15 minutes", lockMinutes: 15, expectedMinutes: 15},
		{name: "no lock", lockMinutes: 0, expectedMinutes: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := LoginStatusResponse{
				LockMinutesRemaining: tt.lockMinutes,
			}
			assert.Equal(t, tt.expectedMinutes, resp.LockMinutesRemaining)
		})
	}
}
