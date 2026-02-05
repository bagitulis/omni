package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserTableName(t *testing.T) {
	user := User{}
	assert.Equal(t, "users", user.TableName())
}

func TestUserToResponse(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:        "user123",
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "secret123",
		Role:      "admin",
		CreatedAt: now,
		UpdatedAt: now,
	}

	response := user.ToResponse()

	assert.Equal(t, "user123", response.ID)
	assert.Equal(t, "testuser", response.Username)
	assert.Equal(t, "test@example.com", response.Email)
	assert.Equal(t, "admin", response.Role)
	assert.Equal(t, now, response.CreatedAt)
	assert.Equal(t, now, response.UpdatedAt)
}

func TestUserIsLocked(t *testing.T) {
	tests := []struct {
		name     string
		lockTime *time.Time
		want     bool
	}{
		{
			name:     "No lock time",
			lockTime: nil,
			want:     false,
		},
		{
			name:     "Lock time in past",
			lockTime: timePtr(time.Now().Add(-10 * time.Minute)),
			want:     false,
		},
		{
			name:     "Lock time in future",
			lockTime: timePtr(time.Now().Add(10 * time.Minute)),
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				AccountLockedUntil: tt.lockTime,
			}
			assert.Equal(t, tt.want, user.IsLocked())
		})
	}
}

func TestUserLockMinutesRemaining(t *testing.T) {
	tests := []struct {
		name     string
		lockTime *time.Time
		check    func(int) bool
	}{
		{
			name:     "No lock time",
			lockTime: nil,
			check:    func(minutes int) bool { return minutes == 0 },
		},
		{
			name:     "Lock time in past",
			lockTime: timePtr(time.Now().Add(-10 * time.Minute)),
			check:    func(minutes int) bool { return minutes == 0 },
		},
		{
			name:     "Lock time 5 minutes in future",
			lockTime: timePtr(time.Now().Add(5 * time.Minute)),
			check:    func(minutes int) bool { return minutes >= 4 && minutes <= 6 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				AccountLockedUntil: tt.lockTime,
			}
			minutes := user.LockMinutesRemaining()
			assert.True(t, tt.check(minutes), "LockMinutesRemaining returned %d, but check failed", minutes)
		})
	}
}

func TestUserIsOAuthUser(t *testing.T) {
	tests := []struct {
		name          string
		oauthProvider *string
		want          bool
	}{
		{
			name:          "No OAuth provider",
			oauthProvider: nil,
			want:          false,
		},
		{
			name:          "Empty OAuth provider",
			oauthProvider: stringPtr(""),
			want:          false,
		},
		{
			name:          "GitHub OAuth provider",
			oauthProvider: stringPtr("github"),
			want:          true,
		},
		{
			name:          "Copilot OAuth provider",
			oauthProvider: stringPtr("copilot"),
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				OAuthProvider: tt.oauthProvider,
			}
			assert.Equal(t, tt.want, user.IsOAuthUser())
		})
	}
}

func TestUserGetOAuthProvider(t *testing.T) {
	tests := []struct {
		name          string
		oauthProvider *string
		want          string
	}{
		{
			name:          "No OAuth provider",
			oauthProvider: nil,
			want:          "",
		},
		{
			name:          "GitHub OAuth provider",
			oauthProvider: stringPtr("github"),
			want:          "github",
		},
		{
			name:          "Copilot OAuth provider",
			oauthProvider: stringPtr("copilot"),
			want:          "copilot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				OAuthProvider: tt.oauthProvider,
			}
			assert.Equal(t, tt.want, user.GetOAuthProvider())
		})
	}
}

func TestUserIsOAuthTokenExpired(t *testing.T) {
	tests := []struct {
		name           string
		tokenExpiresAt *time.Time
		want           bool
	}{
		{
			name:           "No expiry time",
			tokenExpiresAt: nil,
			want:           false,
		},
		{
			name:           "Token expired",
			tokenExpiresAt: timePtr(time.Now().Add(-1 * time.Hour)),
			want:           true,
		},
		{
			name:           "Token not expired",
			tokenExpiresAt: timePtr(time.Now().Add(1 * time.Hour)),
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				OAuthTokenExpiresAt: tt.tokenExpiresAt,
			}
			assert.Equal(t, tt.want, user.IsOAuthTokenExpired())
		})
	}
}

func TestUserJSONMarshalingSnakeCase(t *testing.T) {
	user := &User{
		ID:                  "user123",
		Username:            "testuser",
		Email:               "test@example.com",
		Password:            "secret",
		Role:                "admin",
		FailedLoginAttempts: 2,
		AccountLockedUntil:  timePtr(time.Now()),
		LastFailedLogin:     timePtr(time.Now()),
		OAuthProvider:       stringPtr("github"),
		OAuthID:             stringPtr("gh123"),
		OAuthAccessToken:    stringPtr("token"),
		OAuthRefreshToken:   stringPtr("refresh"),
		OAuthTokenExpiresAt: timePtr(time.Now()),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Verify snake_case JSON keys
	assert.Contains(t, string(data), `"failed_login_attempts"`)
	assert.Contains(t, string(data), `"account_locked_until"`)
	assert.Contains(t, string(data), `"last_failed_login"`)
	assert.Contains(t, string(data), `"oauth_provider"`)
	assert.Contains(t, string(data), `"oauth_id"`)
	assert.Contains(t, string(data), `"oauth_token_expires_at"`)
	assert.Contains(t, string(data), `"created_at"`)
	assert.Contains(t, string(data), `"updated_at"`)

	// Verify password is NOT exposed (json:"-")
	assert.NotContains(t, string(data), `"password"`)
	assert.NotContains(t, string(data), `"oauth_access_token"`)
	assert.NotContains(t, string(data), `"oauth_refresh_token"`)
}

func TestUserResponseJSONMarshalingSnakeCase(t *testing.T) {
	response := UserResponse{
		ID:        "user123",
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	// Verify snake_case JSON keys
	assert.Contains(t, string(data), `"created_at"`)
	assert.Contains(t, string(data), `"updated_at"`)
}

func TestValidRoles(t *testing.T) {
	roles := ValidRoles()
	assert.Equal(t, []string{RoleOwner, RoleAdmin, RoleUser, RoleDeveloper, RoleService}, roles)
	assert.Len(t, roles, 5)
}

func TestIsValidRole(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{"owner", true},
		{"admin", true},
		{"user", true},
		{"developer", true},
		{"service", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidRole(tt.role))
		})
	}
}

func TestValidOAuthProviders(t *testing.T) {
	providers := ValidOAuthProviders()
	assert.Equal(t, []string{OAuthProviderGitHub, OAuthProviderCopilot}, providers)
	assert.Len(t, providers, 2)
}

func TestIsValidOAuthProvider(t *testing.T) {
	tests := []struct {
		provider string
		want     bool
	}{
		{OAuthProviderGitHub, true},
		{OAuthProviderCopilot, true},
		{"google", false},
		{"", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidOAuthProvider(tt.provider))
		})
	}
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func stringPtr(s string) *string {
	return &s
}
