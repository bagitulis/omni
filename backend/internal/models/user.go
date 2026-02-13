package models

import (
	"sync"
	"time"
)

// DatabaseDriver type for determining table naming convention
type DatabaseDriver string

const (
	DriverPostgres DatabaseDriver = "postgres"
)

var (
	currentDriver DatabaseDriver = DriverPostgres
	driverMu      sync.RWMutex
)

// SetDatabaseDriver sets the global database driver for table naming
func SetDatabaseDriver(driver DatabaseDriver) {
	driverMu.Lock()
	defer driverMu.Unlock()
	currentDriver = driver
}

// GetDatabaseDriver returns the current database driver
func GetDatabaseDriver() DatabaseDriver {
	driverMu.RLock()
	defer driverMu.RUnlock()
	return currentDriver
}

// User represents a user in the system
// Table name: "users" (PostgreSQL)
// AGENTS.MD: JSON tags MUST be snake_case
type User struct {
	ID                  string     `gorm:"column:id;primaryKey" json:"id"`
	Username            string     `gorm:"column:username;unique;not null" json:"username"`
	Email               string     `gorm:"column:email;unique;not null" json:"email"`
	Password            string     `gorm:"column:password" json:"-"` // Never expose in JSON
	Role                string     `gorm:"column:role;default:owner" json:"role"`
	FailedLoginAttempts int        `gorm:"column:failed_login_attempts;default:0" json:"failed_login_attempts"`
	AccountLockedUntil  *time.Time `gorm:"column:account_locked_until" json:"account_locked_until,omitempty"`
	LastFailedLogin     *time.Time `gorm:"column:last_failed_login" json:"last_failed_login,omitempty"`

	// OAuth fields
	OAuthProvider       *string    `gorm:"column:oauth_provider" json:"oauth_provider,omitempty"` // github, copilot
	OAuthID             *string    `gorm:"column:oauth_id" json:"oauth_id,omitempty"`             // Provider's user ID
	OAuthAccessToken    *string    `gorm:"column:oauth_access_token" json:"-"`                    // Never expose in JSON
	OAuthRefreshToken   *string    `gorm:"column:oauth_refresh_token" json:"-"`                   // Never expose in JSON
	OAuthTokenExpiresAt *time.Time `gorm:"column:oauth_token_expires_at" json:"oauth_token_expires_at,omitempty"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName specifies the table name for GORM (PostgreSQL: "users")
func (User) TableName() string {
	return GetTableName("User")
}

// UserResponse is the safe user response without password
// AGENTS.MD: JSON tags MUST be snake_case
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts User to safe response
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// IsLocked checks if user account is currently locked
func (u *User) IsLocked() bool {
	if u.AccountLockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.AccountLockedUntil)
}

// LockMinutesRemaining returns minutes until unlock
func (u *User) LockMinutesRemaining() int {
	if u.AccountLockedUntil == nil {
		return 0
	}
	remaining := time.Until(*u.AccountLockedUntil)
	if remaining <= 0 {
		return 0
	}
	return int(remaining.Minutes()) + 1
}

// IsOAuthUser checks if user is authenticated via OAuth
func (u *User) IsOAuthUser() bool {
	return u.OAuthProvider != nil && *u.OAuthProvider != ""
}

// GetOAuthProvider returns OAuth provider if exists
func (u *User) GetOAuthProvider() string {
	if u.OAuthProvider == nil {
		return ""
	}
	return *u.OAuthProvider
}

// IsOAuthTokenExpired checks if OAuth token has expired
func (u *User) IsOAuthTokenExpired() bool {
	if u.OAuthTokenExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.OAuthTokenExpiresAt)
}

// UserRole constants
const (
	RoleOwner     = "owner"
	RoleAdmin     = "admin"
	RoleUser      = "user"
	RoleDeveloper = "developer"
	RoleService   = "service"
)

// OAuthProvider constants
const (
	OAuthProviderGitHub  = "github"
	OAuthProviderCopilot = "copilot"
)

// ValidOAuthProviders returns list of valid OAuth providers
func ValidOAuthProviders() []string {
	return []string{OAuthProviderGitHub, OAuthProviderCopilot}
}

// IsValidOAuthProvider checks if OAuth provider is valid
func IsValidOAuthProvider(provider string) bool {
	for _, p := range ValidOAuthProviders() {
		if p == provider {
			return true
		}
	}
	return false
}

// ValidRoles returns list of valid roles
func ValidRoles() []string {
	return []string{RoleOwner, RoleAdmin, RoleUser, RoleDeveloper, RoleService}
}

// IsValidRole checks if role is valid
func IsValidRole(role string) bool {
	for _, r := range ValidRoles() {
		if r == role {
			return true
		}
	}
	return false
}
