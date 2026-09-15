package models

import "time"

// CredentialConnection represents one platform store connection row.
// Active uniqueness: tenant_id + platform + store_identifier where disabled_at IS NULL.
type CredentialConnection struct {
	ID              string     `gorm:"column:id;primaryKey;type:varchar(255)" json:"id"`
	TenantID        string     `gorm:"column:tenant_id;type:varchar(100);not null;index" json:"tenant_id"`
	Platform        string     `gorm:"column:platform;type:varchar(50);not null;index" json:"platform"`
	StoreIdentifier string     `gorm:"column:store_identifier;type:varchar(255);not null" json:"store_identifier"`
	StoreName       string     `gorm:"column:store_name;type:varchar(255)" json:"store_name"`
	Status          string     `gorm:"column:status;type:varchar(50);not null;default:'disconnected'" json:"status"`
	Region          string     `gorm:"column:region;type:varchar(50)" json:"region"`
	AccessToken     string     `gorm:"column:access_token;type:text" json:"-"`        // NEVER expose in JSON
	RefreshToken    string     `gorm:"column:refresh_token;type:text" json:"-"`       // NEVER expose in JSON
	ShopCipher      string     `gorm:"column:shop_cipher;type:text" json:"-"`         // NEVER expose in JSON
	TokenExpiry     int64      `gorm:"column:token_expiry" json:"token_expiry"`
	RefreshExpiry   int64      `gorm:"column:refresh_expiry" json:"refresh_expiry"`
	LastRefreshAt   *time.Time `gorm:"column:last_refresh_at" json:"last_refresh_at"`
	Version         int        `gorm:"column:version;default:1" json:"version"`
	DisabledAt      *time.Time `gorm:"column:disabled_at" json:"disabled_at,omitempty"`
	DisabledReason  string     `gorm:"column:disabled_reason;type:varchar(255)" json:"disabled_reason,omitempty"`
	DisabledBy      string     `gorm:"column:disabled_by;type:varchar(100)" json:"disabled_by,omitempty"`
	CreatedBy       string     `gorm:"column:created_by;type:varchar(100)" json:"created_by"`
	UpdatedBy       string     `gorm:"column:updated_by;type:varchar(100)" json:"updated_by"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName returns the literal table name for GORM.
func (CredentialConnection) TableName() string {
	return "credential_connections"
}

// IsAccessTokenExpired reports whether the access token is past its recorded
// expiry (or has no expiry recorded — treated as expired for safety).
func (c *CredentialConnection) IsAccessTokenExpired() bool {
	if c.TokenExpiry == 0 {
		return true
	}
	return time.Now().UnixMilli() >= c.TokenExpiry
}

// IsRefreshTokenExpired reports whether the refresh token is past its
// recorded expiry. Zero = treated as expired (safer default).
func (c *CredentialConnection) IsRefreshTokenExpired() bool {
	if c.RefreshExpiry == 0 {
		return true
	}
	return time.Now().UnixMilli() >= c.RefreshExpiry
}

// RefreshExpiryDaysLeft returns whole days remaining before the refresh
// token expires. Returns 0 for unset (0) or already-expired timestamps —
// callers combine this with IsRefreshTokenExpired to distinguish "0 days
// left, expiring today" from "already dead N days ago".
func (c *CredentialConnection) RefreshExpiryDaysLeft() int {
	if c.RefreshExpiry == 0 {
		return 0
	}
	deltaMs := c.RefreshExpiry - time.Now().UnixMilli()
	if deltaMs <= 0 {
		return 0
	}
	return int(deltaMs / 86_400_000)
}

// IsRefreshExpiringWithin reports whether the refresh token is expiring
// (or already expired) inside a `windowDays` warning window. Phase 11.5
// notifier fires when this returns true so sellers can re-authorize
// BEFORE their refresh_token dies + they lose the auto-recovery path.
// A window of 0 or negative disables the warning (returns false always).
func (c *CredentialConnection) IsRefreshExpiringWithin(windowDays int) bool {
	if windowDays <= 0 {
		return false
	}
	if c.RefreshExpiry == 0 {
		return true // treat unset as expired — matches IsRefreshTokenExpired
	}
	deltaMs := c.RefreshExpiry - time.Now().UnixMilli()
	windowMs := int64(windowDays) * 86_400_000
	return deltaMs <= windowMs
}

// EffectiveStatus is the status the API should report to the frontend, honoring
// token expiry. Rules (in order):
//  1. row disabled → "disconnected"
//  2. refresh token expired → "expired" (needs full re-OAuth, no automatic recovery)
//  3. access token expired but refresh still alive → "refresh_required"
//  4. explicit non-"connected" status (action_required, incomplete, ...) → pass through
//  5. otherwise → "connected"
func (c *CredentialConnection) EffectiveStatus() string {
	if c.DisabledAt != nil {
		return "disconnected"
	}
	if c.IsRefreshTokenExpired() {
		return "expired"
	}
	if c.IsAccessTokenExpired() {
		return "refresh_required"
	}
	if c.Status != "" && c.Status != "connected" {
		return c.Status
	}
	return "connected"
}

// CredentialConnectionMaskedResponse is the safe API response DTO without secrets.
//
// Phase 9 / Bug D: exposes RefreshExpiry + EffectiveStatus so the UI can
// render an accurate connection badge (e.g. "expired" for a row whose row
// column says "connected" but whose tokens are dead) without a second
// round-trip.
type CredentialConnectionMaskedResponse struct {
	ID                  string     `json:"id"`
	Platform            string     `json:"platform"`
	StoreIdentifierMask string     `json:"store_identifier_mask"`
	Status              string     `json:"status"`
	EffectiveStatus     string     `json:"effective_status"`
	Region              string     `json:"region"`
	TokenExpiry         int64      `json:"token_expiry"`
	RefreshExpiry       int64      `json:"refresh_expiry"`
	LastRefreshAt       *time.Time `json:"last_refresh_at"`
	Version             int        `json:"version"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// ToMaskedResponse returns a safe DTO with secrets redacted.
func (c *CredentialConnection) ToMaskedResponse() CredentialConnectionMaskedResponse {
	mask := "***" + lastN(c.StoreIdentifier, 3)
	return CredentialConnectionMaskedResponse{
		ID:                  c.ID,
		Platform:            c.Platform,
		StoreIdentifierMask: mask,
		Status:              c.Status,
		EffectiveStatus:     c.EffectiveStatus(),
		Region:              c.Region,
		TokenExpiry:         c.TokenExpiry,
		RefreshExpiry:       c.RefreshExpiry,
		LastRefreshAt:       c.LastRefreshAt,
		Version:             c.Version,
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
	}
}

func lastN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

// EnsureLastN is a compile-time check utility for the mask helper.
func (c *CredentialConnection) StoreIdentifierMask() string {
	return "***" + lastN(c.StoreIdentifier, 3)
}
