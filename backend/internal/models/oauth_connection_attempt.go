package models

import "time"

// OAuthConnectionAttempt tracks pending OAuth flows before store identity is known.
// The attempt_id is server-generated and embedded in the signed OAuth state.
type OAuthConnectionAttempt struct {
	ID              string     `gorm:"column:id;primaryKey;type:varchar(255)" json:"id"`
	TenantID        string     `gorm:"column:tenant_id;type:varchar(100);not null;index" json:"tenant_id"`
	Platform        string     `gorm:"column:platform;type:varchar(50);not null;index" json:"platform"`
	AttemptID       string     `gorm:"column:attempt_id;type:varchar(255);not null;uniqueIndex:idx_oauth_attempt" json:"attempt_id"`
	Status          string     `gorm:"column:status;type:varchar(50);not null;default:'pending'" json:"status"`
	IntendedStoreID string     `gorm:"column:intended_store_id;type:varchar(255)" json:"intended_store_id,omitempty"`
	SignedState     string     `gorm:"column:signed_state;type:text" json:"-"` // Redacted from API
	RedirectPath    string     `gorm:"column:redirect_path;type:varchar(500)" json:"redirect_path,omitempty"`
	ExpiresAt       time.Time  `gorm:"column:expires_at;not null" json:"expires_at"`
	CompletedAt     *time.Time `gorm:"column:completed_at" json:"completed_at,omitempty"`
	CreatedBy       string     `gorm:"column:created_by;type:varchar(100)" json:"created_by"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName returns the literal table name for GORM.
func (OAuthConnectionAttempt) TableName() string {
	return "oauth_connection_attempts"
}
