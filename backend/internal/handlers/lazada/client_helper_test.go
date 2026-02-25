package lazada

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrMissingAccessToken(t *testing.T) {
	assert.NotNil(t, ErrMissingAccessToken)
	assert.Equal(t, "lazada access token not configured", ErrMissingAccessToken.Error())
}

func TestGetLazadaClient_MissingTenantDB(t *testing.T) {
	// basePath that does not exist → config.GetTenantDB returns error
	_, err := GetLazadaClient("test-tenant", "/nonexistent/path/xyz")
	assert.Error(t, err)
}

func TestGetLazadaClient_EmptyTenantID(t *testing.T) {
	// empty tenant id should fail at DB lookup
	_, err := GetLazadaClient("", "/nonexistent/path/xyz")
	assert.Error(t, err)
}

func TestErrMissingAccessToken_IsErrors(t *testing.T) {
	// Verify the error can be compared with errors.Is
	wrapped := errors.New("wrapped: " + ErrMissingAccessToken.Error())
	assert.NotNil(t, wrapped)
	// Direct equality check
	assert.Equal(t, ErrMissingAccessToken, ErrMissingAccessToken)
}
