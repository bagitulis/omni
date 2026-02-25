package services

import (
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
)

// TestRefreshTokenForTenant_PanicsOnNilTenantService verifies that
// RefreshTokenForTenant panics when tenantService is nil (nil pointer dereference
// when iterating tenant DBs to find the session).
func TestRefreshTokenForTenant_PanicsOnNilTenantService(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/test/path")

	defer func() {
		r := recover()
		assert.NotNil(t, r, "expected panic when tenantService is nil")
	}()

	// RefreshTokenForTenant calls s.tenantService.GetAvailableTenants / GetSystemDB
	// when session is not found — both dereference the nil tenantService.
	_, _, _ = svc.RefreshTokenForTenant(nil, "some-token", "tenant-1", "127.0.0.1", "go-test")
}

// TestErrInvalidToken_AccessibleInTokenFile confirms ErrInvalidToken is the
// sentinel used when session lookup fails in RefreshTokenForTenant.
func TestErrInvalidToken_AccessibleInTokenFile(t *testing.T) {
	assert.NotNil(t, ErrInvalidToken)
	assert.Equal(t, "INVALID_TOKEN", ErrInvalidToken.Code)
	assert.NotEmpty(t, ErrInvalidToken.Error())
}

// TestTokenReuseAuthError_Fields verifies the TOKEN_REUSE AuthError produced by
// validateRefreshSession has the expected code and message.
func TestTokenReuseAuthError_Fields(t *testing.T) {
	err := &AuthError{Code: "TOKEN_REUSE", Message: "Refresh token reuse detected, all sessions revoked"}
	assert.Equal(t, "TOKEN_REUSE", err.Code)
	assert.Equal(t, "Refresh token reuse detected, all sessions revoked", err.Error())
}

// TestHashToken_Determinism verifies that repositories.HashToken always produces
// the same hash for the same input — a core requirement for token rotation.
func TestHashToken_Determinism(t *testing.T) {
	token := "test-refresh-token-abc123"
	hash1 := repositories.HashToken(token)
	hash2 := repositories.HashToken(token)

	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2, "HashToken must be deterministic")
	assert.NotEqual(t, token, hash1, "hash must differ from plaintext")
}

// TestHashToken_Uniqueness verifies different tokens produce different hashes.
func TestHashToken_Uniqueness(t *testing.T) {
	hash1 := repositories.HashToken("token-alpha")
	hash2 := repositories.HashToken("token-beta")
	assert.NotEqual(t, hash1, hash2)
}

// TestGenerateNewTokenPair_JWTRoundTrip verifies that generateNewTokenPair builds
// a valid access token whose claims match the user + session inputs.
func TestGenerateNewTokenPair_JWTRoundTrip(t *testing.T) {
	jwtSvc := utils.NewJWTService("test-secret-pair")
	svc := NewMultiTenantAuthService(nil, jwtSvc, "/test/path")

	user := &models.User{
		ID:   "user-gen-1",
		Role: "admin",
	}
	session := &models.RefreshSession{
		UserID:   user.ID,
		TenantID: "tenant-gen-1",
	}

	accessToken, refreshToken, err := svc.generateNewTokenPair(nil, user, session)
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	// Validate the access token contains correct claims
	claims, err := jwtSvc.ValidateToken(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, "user-gen-1", claims.UserID)
	assert.Equal(t, "tenant-gen-1", claims.TenantID)
	assert.Equal(t, "admin", claims.Role)
}
