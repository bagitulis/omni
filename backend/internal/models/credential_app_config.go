package models

import "time"

// CredentialAppConfig represents app credentials per tenant/platform.
// Encrypted fields are never exposed in JSON API responses.
type CredentialAppConfig struct {
	ID              string     `gorm:"column:id;primaryKey;type:varchar(255)" json:"id"`
	TenantID        string     `gorm:"column:tenant_id;type:varchar(100);not null;index" json:"tenant_id"`
	Platform        string     `gorm:"column:platform;type:varchar(50);not null;index" json:"platform"`
	StoreIdentifier string     `gorm:"column:store_identifier;type:varchar(255)" json:"store_identifier,omitempty"`
	Region          string     `gorm:"column:region;type:varchar(50);default:'id'" json:"region"`
	AppKey          string     `gorm:"column:app_key;type:text" json:"-"`       // NEVER expose — for Lazada/TikTok
	AppSecret       string     `gorm:"column:app_secret;type:text" json:"-"`    // NEVER expose
	PartnerID       int64      `gorm:"column:partner_id" json:"-"`              // NEVER expose — for Shopee
	PartnerKey      string     `gorm:"column:partner_key;type:text" json:"-"`   // NEVER expose — for Shopee
	Configured      bool       `gorm:"column:configured;default:false" json:"configured"`
	LastTestedAt    *time.Time `gorm:"column:last_tested_at" json:"last_tested_at,omitempty"`
	CreatedBy       string     `gorm:"column:created_by;type:varchar(100)" json:"created_by"`
	UpdatedBy       string     `gorm:"column:updated_by;type:varchar(100)" json:"updated_by"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName returns the literal table name for GORM.
func (CredentialAppConfig) TableName() string {
	return "credential_app_configs"
}

// CredentialAppConfigMaskedResponse is the safe API response without secrets.
type CredentialAppConfigMaskedResponse struct {
	ID              string     `json:"id"`
	Platform        string     `json:"platform"`
	StoreIdentifier string     `json:"store_identifier,omitempty"`
	Region          string     `json:"region"`
	Configured      bool       `json:"configured"`
	LastTestedAt    *time.Time `json:"last_tested_at,omitempty"`
	CreatedBy       string     `json:"created_by"`
	UpdatedBy       string     `json:"updated_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ToMaskedResponse returns a safe DTO with secrets redacted.
func (c *CredentialAppConfig) ToMaskedResponse() CredentialAppConfigMaskedResponse {
	return CredentialAppConfigMaskedResponse{
		ID:              c.ID,
		Platform:        c.Platform,
		StoreIdentifier: c.StoreIdentifier,
		Region:          c.Region,
		Configured:      c.Configured,
		LastTestedAt:    c.LastTestedAt,
		CreatedBy:       c.CreatedBy,
		UpdatedBy:       c.UpdatedBy,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}
