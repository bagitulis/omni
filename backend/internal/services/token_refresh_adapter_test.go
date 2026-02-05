package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTokenRefreshAdapter(t *testing.T) {
	tm := NewTokenManager(nil, nil, "/test/path")
	adapter := NewTokenRefreshAdapter(tm)

	assert.NotNil(t, adapter)
	assert.Equal(t, tm, adapter.tokenManager)
}

func TestTokenRefreshAdapter_NilCheck(t *testing.T) {
	// Test that adapter can be created with a nil token manager
	adapter := NewTokenRefreshAdapter(nil)
	assert.NotNil(t, adapter)
	assert.Nil(t, adapter.tokenManager)
}

func TestTokenRefreshAdapter_Structure(t *testing.T) {
	tm := NewTokenManager(nil, nil, "/custom/path")
	adapter := NewTokenRefreshAdapter(tm)

	// Verify the adapter properly wraps the token manager
	assert.NotNil(t, adapter.tokenManager)
	assert.Equal(t, "/custom/path", adapter.tokenManager.basePath)
}

func TestTokenRefreshAdapter_AdapterPattern(t *testing.T) {
	// Test the adapter pattern is correctly implemented
	tm := NewTokenManager(nil, nil, "/test")
	adapter := NewTokenRefreshAdapter(tm)

	// The adapter should wrap the token manager
	assert.Same(t, tm, adapter.tokenManager)
}
