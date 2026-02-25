package services

import (
	"encoding/json"
	"testing"

	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoginStatusResponse_DefaultValues verifies zero-value behaviour.
func TestLoginStatusResponse_DefaultValues(t *testing.T) {
	var resp LoginStatusResponse
	assert.False(t, resp.IsLocked)
	assert.False(t, resp.RequiresCaptcha)
	assert.Equal(t, 0, resp.LockMinutesRemaining)
}

// TestLoginStatusResponse_JSONTags verifies snake_case JSON serialisation.
func TestLoginStatusResponse_JSONTags(t *testing.T) {
	resp := LoginStatusResponse{
		IsLocked:             true,
		RequiresCaptcha:      false,
		LockMinutesRemaining: 5,
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &m))

	_, hasIsLocked := m["is_locked"]
	_, hasRequiresCaptcha := m["requires_captcha"]
	assert.True(t, hasIsLocked, "expected JSON key 'is_locked'")
	assert.True(t, hasRequiresCaptcha, "expected JSON key 'requires_captcha'")
	assert.Equal(t, true, m["is_locked"])
	assert.Equal(t, false, m["requires_captcha"])
}

// TestLoginStatusResponse_Locked verifies a locked response.
func TestLoginStatusResponse_Locked(t *testing.T) {
	resp := LoginStatusResponse{
		IsLocked:             true,
		RequiresCaptcha:      false,
		LockMinutesRemaining: 7,
	}
	assert.True(t, resp.IsLocked)
	assert.False(t, resp.RequiresCaptcha)
	assert.Equal(t, 7, resp.LockMinutesRemaining)
}

// TestLoginStatusResponse_CaptchaRequired verifies a captcha-required response.
func TestLoginStatusResponse_CaptchaRequired(t *testing.T) {
	resp := LoginStatusResponse{
		IsLocked:        false,
		RequiresCaptcha: true,
	}
	assert.False(t, resp.IsLocked)
	assert.True(t, resp.RequiresCaptcha)
	assert.Equal(t, 0, resp.LockMinutesRemaining)
}

// TestSwitchTenant_AdminRole_Forbidden verifies FORBIDDEN guard for admin.
func TestSwitchTenant_AdminRole_Forbidden(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/test/path")

	_, err := svc.SwitchTenant(nil, "user-1", "admin", "tenant-x")
	require.Error(t, err)

	authErr, ok := err.(*AuthError)
	require.True(t, ok)
	assert.Equal(t, "FORBIDDEN", authErr.Code)
}

// TestSwitchTenant_OwnerRole_Forbidden verifies owner cannot switch tenants.
func TestSwitchTenant_OwnerRole_Forbidden(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/test/path")

	_, err := svc.SwitchTenant(nil, "user-1", "owner", "tenant-x")
	require.Error(t, err)

	authErr, ok := err.(*AuthError)
	require.True(t, ok)
	assert.Equal(t, "FORBIDDEN", authErr.Code)
}

// TestSwitchTenant_UserRole_Forbidden verifies user role cannot switch tenants.
func TestSwitchTenant_UserRole_Forbidden(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/test/path")

	_, err := svc.SwitchTenant(nil, "user-1", "user", "tenant-x")
	require.Error(t, err)

	authErr, ok := err.(*AuthError)
	require.True(t, ok)
	assert.Equal(t, "FORBIDDEN", authErr.Code)
}

// TestSwitchTenant_DeveloperRole_PanicsOnNilTenantService documents the expected
// panic when tenantService is nil and role="developer" (calls TenantExists).
func TestSwitchTenant_DeveloperRole_PanicsOnNilTenantService(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/test/path")

	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic when tenantService is nil and role=developer")
	}()

	// Developer role passes the role check; then calls s.tenantService.TenantExists → nil panic
	_, _ = svc.SwitchTenant(nil, "dev-user", "developer", "some-tenant")
}

// TestGetAvailableTenants_NilTenantService_Panics documents the nil-deref panic.
func TestGetAvailableTenants_NilTenantService_Panics(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/test/path")

	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic when tenantService is nil")
	}()

	_, _ = svc.GetAvailableTenants(nil)
}

// TestGetLoginStatus_NilTenantService_Panics documents the nil-deref panic.
func TestGetLoginStatus_NilTenantService_Panics(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/test/path")

	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic when tenantService is nil")
	}()

	_, _ = svc.GetLoginStatus(nil, "someuser")
}
