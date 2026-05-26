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

// CredentialConnectionMaskedResponse is the safe API response DTO without secrets.
type CredentialConnectionMaskedResponse struct {
	ID                  string     `json:"id"`
	Platform            string     `json:"platform"`
	StoreIdentifierMask string     `json:"store_identifier_mask"`
	Status              string     `json:"status"`
	Region              string     `json:"region"`
	TokenExpiry         int64      `json:"token_expiry"`
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
		Region:              c.Region,
		TokenExpiry:         c.TokenExpiry,
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
