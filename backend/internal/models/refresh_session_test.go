package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshSessionTableName(t *testing.T) {
	session := RefreshSession{}
	assert.Equal(t, "refresh_sessions", session.TableName())
}

func TestRefreshSessionIsValid(t *testing.T) {
	tests := []struct {
		name      string
		revokedAt *time.Time
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "Valid session",
			revokedAt: nil,
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      true,
		},
		{
			name:      "Expired session",
			revokedAt: nil,
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      false,
		},
		{
			name:      "Revoked session",
			revokedAt: timePtr(time.Now()),
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "Revoked and expired",
			revokedAt: timePtr(time.Now()),
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &RefreshSession{
				RevokedAt: tt.revokedAt,
				ExpiresAt: tt.expiresAt,
			}
			assert.Equal(t, tt.want, session.IsValid())
		})
	}
}

func TestRefreshSessionIsReused(t *testing.T) {
	tests := []struct {
		name           string
		replacedByHash *string
		want           bool
	}{
		{
			name:           "Not reused",
			replacedByHash: nil,
			want:           false,
		},
		{
			name:           "Reused",
			replacedByHash: stringPtr("new_hash"),
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &RefreshSession{
				ReplacedByHash: tt.replacedByHash,
			}
			assert.Equal(t, tt.want, session.IsReused())
		})
	}
}

func TestRefreshSessionIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "Not expired",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "Expired",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &RefreshSession{
				ExpiresAt: tt.expiresAt,
			}
			assert.Equal(t, tt.want, session.IsExpired())
		})
	}
}

func TestRefreshSessionIsRevoked(t *testing.T) {
	tests := []struct {
		name      string
		revokedAt *time.Time
		want      bool
	}{
		{
			name:      "Not revoked",
			revokedAt: nil,
			want:      false,
		},
		{
			name:      "Revoked",
			revokedAt: timePtr(time.Now()),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &RefreshSession{
				RevokedAt: tt.revokedAt,
			}
			assert.Equal(t, tt.want, session.IsRevoked())
		})
	}
}

func TestRefreshSessionJSONMarshalingSnakeCase(t *testing.T) {
	revokedAt := time.Now().Add(1 * time.Hour)
	session := &RefreshSession{
		ID:             "session123",
		UserID:         "user123",
		TenantID:       "tenant1",
		TokenHash:      "hash_value",
		ExpiresAt:      time.Now().Add(1 * time.Hour),
		RevokedAt:      &revokedAt,
		ReplacedByHash: nil,
		IPAddress:      "127.0.0.1",
		UserAgent:      "Mozilla/5.0",
		CreatedAt:      time.Now(),
		LastUsedAt:     time.Now(),
	}

	data, err := json.Marshal(session)
	require.NoError(t, err)

	// Verify snake_case JSON keys
	assert.Contains(t, string(data), `"user_id"`)
	assert.Contains(t, string(data), `"tenant_id"`)
	assert.Contains(t, string(data), `"expires_at"`)
	assert.Contains(t, string(data), `"revoked_at"`)
	assert.Contains(t, string(data), `"ip_address"`)
	assert.Contains(t, string(data), `"user_agent"`)
	assert.Contains(t, string(data), `"created_at"`)
	assert.Contains(t, string(data), `"last_used_at"`)

	// Verify TokenHash and ReplacedByHash are NOT exposed (json:"-")
	assert.NotContains(t, string(data), `"token_hash"`)
	assert.NotContains(t, string(data), `"replaced_by_hash"`)
}

func TestRefreshSessionCompoundValidation(t *testing.T) {
	// Test a realistic scenario: session created, then revoked, then rotated
	now := time.Now()
	session := &RefreshSession{
		ID:             "session123",
		UserID:         "user123",
		TenantID:       "tenant1",
		TokenHash:      "hash_value",
		ExpiresAt:      now.Add(7 * 24 * time.Hour), // 7 days
		RevokedAt:      timePtr(now.Add(1 * time.Hour)),
		ReplacedByHash: stringPtr("new_hash"),
		CreatedAt:      now,
		LastUsedAt:     now,
	}

	// Should not be valid because revoked
	assert.False(t, session.IsValid())
	// Should be revoked
	assert.True(t, session.IsRevoked())
	// Should be reused (rotated)
	assert.True(t, session.IsReused())
	// Should not be expired yet
	assert.False(t, session.IsExpired())
}
