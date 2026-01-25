package models

import "time"

// CopilotOAuthAccount represents a GitHub Copilot-linked account
// Table name: "copilot_oauth_accounts" (PostgreSQL)
type CopilotOAuthAccount struct {
	ID          string `gorm:"primaryKey" json:"id"`
	TenantID    string `gorm:"column:tenant_id;index;not null" json:"tenantId"`
	UserID      string `gorm:"column:user_id;index;not null" json:"userId"`
	GitHubID    string `gorm:"column:github_id;uniqueIndex;not null" json:"githubId"`
	Username    string `gorm:"column:username;not null" json:"username"`
	Email       string `gorm:"column:email;not null" json:"email"`
	DisplayName string `gorm:"column:display_name" json:"displayName"`
	AvatarURL   string `gorm:"column:avatar_url" json:"avatarUrl"`

	// OAuth tokens
	AccessToken  string    `gorm:"column:access_token;not null" json:"-"` // Never expose in JSON
	RefreshToken string    `gorm:"column:refresh_token" json:"-"`         // Never expose in JSON
	TokenType    string    `gorm:"column:token_type;default:bearer" json:"tokenType"`
	Scope        string    `gorm:"column:scope" json:"scope"`
	ExpiresAt    time.Time `gorm:"column:expires_at;index" json:"expiresAt"`

	// Copilot specific
	CopilotSeatActive bool       `gorm:"column:copilot_seat_active;default:false" json:"copilotSeatActive"`
	LastSyncAt        *time.Time `gorm:"column:last_sync_at" json:"lastSyncAt,omitempty"`

	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName specifies the table name for GORM
func (CopilotOAuthAccount) TableName() string {
	return GetTableName("CopilotOAuthAccount")
}

// IsTokenExpired checks if the OAuth token has expired
func (c *CopilotOAuthAccount) IsTokenExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// ShouldRefresh checks if token should be refreshed (expires within 1 hour)
func (c *CopilotOAuthAccount) ShouldRefresh() bool {
	return time.Now().Add(time.Hour).After(c.ExpiresAt)
}

// CopilotOAuthState stores OAuth state for GitHub Copilot flow
// Table name: "copilot_oauth_states" (PostgreSQL)
type CopilotOAuthState struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"column:tenant_id;index;not null" json:"tenantId"`
	State       string    `gorm:"column:state;uniqueIndex;not null" json:"state"`
	RedirectURL string    `gorm:"column:redirect_url" json:"redirectUrl,omitempty"`
	UserAgent   string    `gorm:"column:user_agent" json:"userAgent,omitempty"`
	IPAddress   string    `gorm:"column:ip_address" json:"ipAddress,omitempty"`
	ExpiresAt   time.Time `gorm:"column:expires_at;index" json:"expiresAt"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
}

// TableName specifies the table name for GORM
func (CopilotOAuthState) TableName() string {
	return GetTableName("CopilotOAuthState")
}

// IsExpired checks if the state has expired
func (s *CopilotOAuthState) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
