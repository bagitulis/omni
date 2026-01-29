package models

import "time"

// GlobalConfig represents config stored in system schema (PostgreSQL)
// Matches Node.js Prisma schema for GlobalConfig
type GlobalConfig struct {
	ID          string    `gorm:"primaryKey;type:varchar(100)" json:"id"`
	Platform    string    `gorm:"column:platform;not null" json:"platform"`
	ConfigKey   string    `gorm:"column:config_key;not null" json:"config_key"`
	ConfigValue string    `gorm:"column:config_value" json:"config_value"`
	IsEncrypted bool      `gorm:"column:is_encrypted;default:false" json:"is_encrypted"`
	Description string    `gorm:"column:description" json:"description"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (GlobalConfig) TableName() string {
	return GetTableName("GlobalConfig")
}

// PlatformCredentials holds decrypted platform API credentials
type PlatformCredentials struct {
	PartnerID   int64  // Shopee
	PartnerKey  string // Shopee
	AppKey      string // Lazada, TikTok
	AppSecret   string // Lazada, TikTok
	ShopID      int64
	AccessToken string
	Region      string // Lazada region
}
