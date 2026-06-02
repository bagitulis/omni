package repositories

import (
	"time"

	"github.com/omni/backend/internal/models"
)

// PlatformConfig represents platform configuration.
// Used by PlatformConfigAdapter which assembles these structs from key-value data.
type PlatformConfig struct {
	ID           string `gorm:"primaryKey"`
	TenantID     string `gorm:"index"`
	Platform     string `gorm:"index"`
	ShopID       string
	ShopIDInt    int64
	ShopName     string
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
	Region       string
	IsActive     bool `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName specifies table name - uses dynamic naming for PostgreSQL compatibility
func (PlatformConfig) TableName() string {
	return models.GetTableName("PlatformConfig")
}
