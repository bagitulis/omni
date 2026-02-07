package models

import "time"

// PlatformConfigKV represents a key-value entry in platform_configs table
// NOTE: This table uses schema-level multi-tenancy, so NO tenant_id column needed
// Each tenant has their own schema (e.g., tenant_yumna_bertigamart)
// Matches PostgreSQL structure: id, platform, config_key, config_value, data_type, is_encrypted, metadata
type PlatformConfigKV struct {
	ID          string    `gorm:"column:id;primaryKey;type:varchar(255)" json:"id"`
	Platform    string    `gorm:"column:platform;type:varchar(50);not null;index" json:"platform"`
	ConfigKey   string    `gorm:"column:config_key;type:varchar(255);not null" json:"config_key"`
	ConfigValue string    `gorm:"column:config_value;type:text" json:"config_value"`
	DataType    string    `gorm:"column:data_type;type:varchar(50);default:'string'" json:"data_type"`
	IsEncrypted bool      `gorm:"column:is_encrypted;default:true" json:"is_encrypted"`
	Metadata    JSONMap   `gorm:"column:metadata;type:jsonb;default:'{}'" json:"metadata,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (PlatformConfigKV) TableName() string { return GetTableName("PlatformConfig") }

// TenantPlatformCredentials is a structured view of credentials from key-value storage
// This is NOT a database model, just a helper struct to hold parsed credentials
type TenantPlatformCredentials struct {
	Platform     string
	ShopID       string
	ShopIDInt    int64
	ShopName     string
	ShopCipher   string // TikTok specific - encrypted shop identifier for API calls
	AccessToken  string
	RefreshToken string
	TokenExpiry  string
	AppKey       string
	AppSecret    string
	Code         string
	ExpiresAt    int64
	Region       string
}
