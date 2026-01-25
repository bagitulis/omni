package models

import "time"

// OAuthState stores OAuth state for security validation
// Matches Prisma schema: OAuthState
type OAuthState struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"index;not null" json:"tenantId"`
	Platform    string    `gorm:"index;not null" json:"platform"` // shopee, lazada, tiktok
	State       string    `gorm:"uniqueIndex;not null" json:"state"`
	RedirectURL string    `json:"redirectUrl,omitempty"`
	Metadata    string    `json:"metadata,omitempty"` // JSON string
	ExpiresAt   time.Time `gorm:"index" json:"expiresAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

// TableName specifies the table name for GORM
func (OAuthState) TableName() string {
	return GetTableName("OAuthState")
}

// IsExpired checks if the state has expired
func (o *OAuthState) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

// OAuthLog logs all OAuth callbacks and token exchanges
// Matches Prisma schema: OAuthLog
type OAuthLog struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	TenantID    string     `gorm:"index;not null" json:"tenantId"`
	Platform    string     `gorm:"index;not null" json:"platform"`
	EventType   string     `gorm:"index;not null" json:"eventType"` // callback_received, token_exchanged, token_refreshed, error
	ShopID      string     `json:"shopId,omitempty"`
	Code        string     `json:"code,omitempty"`
	State       string     `json:"state,omitempty"`
	Status      string     `gorm:"default:received;index" json:"status"` // received, success, failed
	ErrorMsg    string     `json:"errorMsg,omitempty"`
	Metadata    string     `json:"metadata,omitempty"` // JSON string
	ProcessedAt *time.Time `json:"processedAt,omitempty"`
	CreatedAt   time.Time  `gorm:"index" json:"createdAt"`
}

// TableName specifies the table name for GORM
func (OAuthLog) TableName() string {
	return GetTableName("OAuthLog")
}

// OAuthEventType constants
const (
	OAuthEventCallback      = "callback_received"
	OAuthEventTokenExchange = "token_exchanged"
	OAuthEventTokenRefresh  = "token_refreshed"
	OAuthEventError         = "error"
)

// OAuthStatus constants
const (
	OAuthStatusReceived = "received"
	OAuthStatusSuccess  = "success"
	OAuthStatusFailed   = "failed"
)

// PlatformType constants
const (
	PlatformShopee = "shopee"
	PlatformLazada = "lazada"
	PlatformTiktok = "tiktok"
)

// ValidPlatforms returns list of valid platforms
func ValidPlatforms() []string {
	return []string{PlatformShopee, PlatformLazada, PlatformTiktok}
}

// IsValidPlatform checks if platform is valid
func IsValidPlatform(platform string) bool {
	for _, p := range ValidPlatforms() {
		if p == platform {
			return true
		}
	}
	return false
}
