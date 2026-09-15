package models

import "time"

// CredentialAppConfig represents app credentials per tenant/platform.
// Encrypted fields are never exposed in JSON API responses.
//
// Phase 8 additions (Shopee-focused, backward-compatible for other platforms):
//   - TestPartnerID / TestPartnerKey    — Shopee sandbox pair for dry-run
//     probing without impacting live traffic. Encrypted at rest, `json:"-"`.
//   - PartnerKeyExpiresAt               — Live partner-key expiry, drives the
//     UI's "expires in N days" warning banner.
//   - TestPartnerKeyExpiresAt           — Same for the sandbox pair.
//   - AppStatus (`online` | `offline`)  — All platforms; when offline, sync
//     workers should skip this platform.
//   - ActivePartnerEnv (`live` | `test`) — Shopee-only runtime switch; the
//     client reader picks partner_id/partner_key vs test_partner_id/key.
type CredentialAppConfig struct {
	ID              string     `gorm:"column:id;primaryKey;type:varchar(255)" json:"id"`
	TenantID        string     `gorm:"column:tenant_id;type:varchar(100);not null;index" json:"tenant_id"`
	Platform        string     `gorm:"column:platform;type:varchar(50);not null;index" json:"platform"`
	StoreIdentifier string     `gorm:"column:store_identifier;type:varchar(255)" json:"store_identifier,omitempty"`
	Region          string     `gorm:"column:region;type:varchar(50);default:'id'" json:"region"`
	AppKey          string     `gorm:"column:app_key;type:text" json:"-"`       // NEVER expose — for Lazada/TikTok
	AppSecret       string     `gorm:"column:app_secret;type:text" json:"-"`    // NEVER expose
	PartnerID       int64      `gorm:"column:partner_id" json:"-"`              // NEVER expose — for Shopee (live)
	PartnerKey      string     `gorm:"column:partner_key;type:text" json:"-"`   // NEVER expose — for Shopee (live)
	// Phase 8 — sandbox pair (Shopee only).
	TestPartnerID  int64  `gorm:"column:test_partner_id" json:"-"`
	TestPartnerKey string `gorm:"column:test_partner_key;type:text" json:"-"`
	// Phase 8 — key expiry & operational toggles.
	PartnerKeyExpiresAt     *time.Time `gorm:"column:partner_key_expires_at" json:"partner_key_expires_at,omitempty"`
	TestPartnerKeyExpiresAt *time.Time `gorm:"column:test_partner_key_expires_at" json:"test_partner_key_expires_at,omitempty"`
	AppStatus               string     `gorm:"column:app_status;type:varchar(16);not null;default:'online'" json:"app_status"`
	ActivePartnerEnv        string     `gorm:"column:active_partner_env;type:varchar(8);not null;default:'live'" json:"active_partner_env"`

	Configured   bool       `gorm:"column:configured;default:false" json:"configured"`
	LastTestedAt *time.Time `gorm:"column:last_tested_at" json:"last_tested_at,omitempty"`
	CreatedBy    string     `gorm:"column:created_by;type:varchar(100)" json:"created_by"`
	UpdatedBy    string     `gorm:"column:updated_by;type:varchar(100)" json:"updated_by"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName returns the literal table name for GORM.
func (CredentialAppConfig) TableName() string {
	return "credential_app_configs"
}

// CredentialAppConfigMaskedResponse is the safe API response without secrets.
//
// Phase 8: exposes derived booleans + expiry timestamps so the UI can render
// warning banners without ever seeing raw keys. `TestConfigured` is true
// when a Shopee sandbox pair is stored (both id + key present); consumers use
// it to render the sandbox tab / active-env toggle.
type CredentialAppConfigMaskedResponse struct {
	ID                      string     `json:"id"`
	Platform                string     `json:"platform"`
	StoreIdentifier         string     `json:"store_identifier,omitempty"`
	Region                  string     `json:"region"`
	Configured              bool       `json:"configured"`
	LastTestedAt            *time.Time `json:"last_tested_at,omitempty"`
	CreatedBy               string     `json:"created_by"`
	UpdatedBy               string     `json:"updated_by"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	// Phase 8 fields.
	PartnerKeyExpiresAt     *time.Time `json:"partner_key_expires_at,omitempty"`
	TestPartnerKeyExpiresAt *time.Time `json:"test_partner_key_expires_at,omitempty"`
	TestPartnerID           int64      `json:"test_partner_id,omitempty"` // ID is not a secret; UI shows for confirmation
	PartnerID               int64      `json:"partner_id,omitempty"`      // Same — ID visible, key is not
	TestConfigured          bool       `json:"test_configured"`
	AppStatus               string     `json:"app_status"`
	ActivePartnerEnv        string     `json:"active_partner_env"`
}

// ToMaskedResponse returns a safe DTO with secrets redacted.
func (c *CredentialAppConfig) ToMaskedResponse() CredentialAppConfigMaskedResponse {
	return CredentialAppConfigMaskedResponse{
		ID:                      c.ID,
		Platform:                c.Platform,
		StoreIdentifier:         c.StoreIdentifier,
		Region:                  c.Region,
		Configured:              c.Configured,
		LastTestedAt:            c.LastTestedAt,
		CreatedBy:               c.CreatedBy,
		UpdatedBy:               c.UpdatedBy,
		CreatedAt:               c.CreatedAt,
		UpdatedAt:               c.UpdatedAt,
		PartnerKeyExpiresAt:     c.PartnerKeyExpiresAt,
		TestPartnerKeyExpiresAt: c.TestPartnerKeyExpiresAt,
		TestPartnerID:           c.TestPartnerID,
		PartnerID:               c.PartnerID,
		TestConfigured:          c.TestPartnerID > 0 && c.TestPartnerKey != "",
		AppStatus:               nonEmptyOr(c.AppStatus, "online"),
		ActivePartnerEnv:        nonEmptyOr(c.ActivePartnerEnv, "live"),
	}
}

// nonEmptyOr returns fallback when v is empty. Guards masked-response defaults
// for legacy rows migrated before the new NOT NULL columns had values.
func nonEmptyOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
