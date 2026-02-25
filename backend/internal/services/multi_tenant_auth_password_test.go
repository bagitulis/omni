package services

import (
	"testing"

	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthError_PasswordError verifies AuthError used for password validation.
func TestAuthError_PasswordError(t *testing.T) {
	err := &AuthError{Code: "INVALID_PASSWORD", Message: "Current password is incorrect"}

	assert.Equal(t, "INVALID_PASSWORD", err.Code)
	assert.Equal(t, "Current password is incorrect", err.Message)
	assert.Equal(t, "Current password is incorrect", err.Error())
}

// TestMultiTenantAuthService_ChangePasswordForTenant_PanicsWithNilTenantService
// confirms that calling ChangePasswordForTenant when tenantService is nil
// (as constructed via NewMultiTenantAuthService(nil, ...)) will panic because
// the method immediately dereferences s.tenantService.GetTenantDB.
// We use recover() so the test itself does not crash.
func TestMultiTenantAuthService_ChangePasswordForTenant_PanicWithNilTenantService(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/test/path")

	defer func() {
		r := recover()
		// We expect a non-nil panic (nil pointer dereference on tenantService)
		assert.NotNil(t, r, "expected panic when tenantService is nil")
	}()

	// This must panic — we are documenting the behaviour, not fixing it.
	// ChangePasswordForTenant calls s.tenantService.GetTenantDB which panics on nil.
	_ = svc.ChangePasswordForTenant(nil, "u1", "t1", "oldpw", "NewP@ssw0rd!")
}

// TestMultiTenantAuthService_Construction verifies the service is built correctly.
func TestMultiTenantAuthService_ChangePassword_ServiceConstruction(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/data/path")

	require.NotNil(t, svc)
	assert.Equal(t, "/data/path", svc.basePath)
	assert.Nil(t, svc.tenantService)
}

// TestAuthError_Wrapping verifies that AuthError satisfies the error interface
// and is usable in type assertions, as used in ChangePasswordForTenant.
func TestAuthError_Wrapping(t *testing.T) {
	var err error = &AuthError{Code: "INVALID_PASSWORD", Message: "wrong"}

	authErr, ok := err.(*AuthError)
	require.True(t, ok)
	assert.Equal(t, "INVALID_PASSWORD", authErr.Code)
}

// TestErrUserNotFound_UsedByChangePassword confirms ErrUserNotFound is accessible
// (it is returned when user lookup fails in ChangePasswordForTenant).
func TestErrUserNotFound_UsedByChangePassword(t *testing.T) {
	assert.NotNil(t, ErrUserNotFound)
	assert.Equal(t, "USER_NOT_FOUND", ErrUserNotFound.Code)
	assert.NotEmpty(t, ErrUserNotFound.Error())
}

// TestValidatePasswordStrength_ViaUtils documents the dependency used in
// ChangePasswordForTenant for new-password validation.
func TestValidatePasswordStrength_ViaUtils(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "strong password", password: "Str0ng!Pass", wantErr: false},
		{name: "short password", password: "Abc1!", wantErr: true},
		{name: "empty", password: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ValidatePasswordStrength(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
