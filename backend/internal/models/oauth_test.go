package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthStateTableName(t *testing.T) {
	state := OAuthState{}
	assert.Equal(t, "oauth_states", state.TableName())
}

func TestOAuthStateIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "Expired",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "Not expired",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "Just expired",
			expiresAt: time.Now().Add(-1 * time.Second),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &OAuthState{
				ExpiresAt: tt.expiresAt,
			}
			assert.Equal(t, tt.want, state.IsExpired())
		})
	}
}

func TestOAuthStateJSONMarshalingSnakeCase(t *testing.T) {
	state := &OAuthState{
		ID:          "state123",
		TenantID:    "tenant1",
		Platform:    "shopee",
		State:       "abc123xyz",
		RedirectURL: "https://example.com/callback",
		Metadata:    `{"key":"value"}`,
		ExpiresAt:   time.Now(),
		CreatedAt:   time.Now(),
	}

	data, err := json.Marshal(state)
	require.NoError(t, err)

	// Verify snake_case JSON keys
	assert.Contains(t, string(data), `"tenant_id"`)
	assert.Contains(t, string(data), `"redirect_url"`)
	assert.Contains(t, string(data), `"expires_at"`)
	assert.Contains(t, string(data), `"created_at"`)
}

func TestOAuthLogTableName(t *testing.T) {
	log := OAuthLog{}
	assert.Equal(t, "oauth_logs", log.TableName())
}

func TestOAuthLogJSONMarshalingSnakeCase(t *testing.T) {
	processedAt := time.Now()
	log := &OAuthLog{
		ID:          "log123",
		TenantID:    "tenant1",
		Platform:    "shopee",
		EventType:   OAuthEventCallback,
		ShopID:      "shop123",
		Code:        "auth_code",
		State:       "state_value",
		Status:      OAuthStatusSuccess,
		ErrorMsg:    "",
		Metadata:    `{}`,
		ProcessedAt: &processedAt,
		CreatedAt:   time.Now(),
	}

	data, err := json.Marshal(log)
	require.NoError(t, err)

	// Verify snake_case JSON keys
	assert.Contains(t, string(data), `"tenant_id"`)
	assert.Contains(t, string(data), `"event_type"`)
	assert.Contains(t, string(data), `"shop_id"`)
	assert.Contains(t, string(data), `"processed_at"`)
	assert.Contains(t, string(data), `"created_at"`)
}

func TestOAuthEventTypeConstants(t *testing.T) {
	assert.Equal(t, "callback_received", OAuthEventCallback)
	assert.Equal(t, "token_exchanged", OAuthEventTokenExchange)
	assert.Equal(t, "token_refreshed", OAuthEventTokenRefresh)
	assert.Equal(t, "error", OAuthEventError)
}

func TestOAuthStatusConstants(t *testing.T) {
	assert.Equal(t, "received", OAuthStatusReceived)
	assert.Equal(t, "success", OAuthStatusSuccess)
	assert.Equal(t, "failed", OAuthStatusFailed)
}

func TestPlatformTypeConstants(t *testing.T) {
	assert.Equal(t, "shopee", PlatformShopee)
	assert.Equal(t, "lazada", PlatformLazada)
	assert.Equal(t, "tiktok", PlatformTiktok)
}
