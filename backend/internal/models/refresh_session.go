package models

import (
	"time"
)

// RefreshSession represents a refresh token session for secure token management
// Table: refresh_sessions (in both system and tenant schemas)
type RefreshSession struct {
	ID             string     `gorm:"column:id;primaryKey" json:"id"`
	UserID         string     `gorm:"column:user_id;index;not null" json:"user_id"`
	TenantID       string     `gorm:"column:tenant_id;index" json:"tenant_id"`
	TokenHash      string     `gorm:"column:token_hash;uniqueIndex;not null" json:"-"` // Never expose hash
	ExpiresAt      time.Time  `gorm:"column:expires_at;index;not null" json:"expires_at"`
	RevokedAt      *time.Time `gorm:"column:revoked_at" json:"revoked_at,omitempty"`
	ReplacedByHash *string    `gorm:"column:replaced_by_hash" json:"-"` // For rotation tracking

	// Security metadata
	IPAddress string `gorm:"column:ip_address" json:"ip_address"`
	UserAgent string `gorm:"column:user_agent" json:"user_agent"`

	// Timestamps
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	LastUsedAt time.Time `gorm:"column:last_used_at" json:"last_used_at"`
}

// TableName specifies the table name for GORM
func (RefreshSession) TableName() string {
	return GetTableName("RefreshSession")
}

// IsValid checks if session is valid (not expired, not revoked)
func (s *RefreshSession) IsValid() bool {
	if s.RevokedAt != nil {
		return false
	}
	return time.Now().Before(s.ExpiresAt)
}

// IsReused checks if this token was already rotated (reuse attack detection)
func (s *RefreshSession) IsReused() bool {
	return s.ReplacedByHash != nil
}

// IsExpired checks if the session has expired
func (s *RefreshSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsRevoked checks if the session was explicitly revoked
func (s *RefreshSession) IsRevoked() bool {
	return s.RevokedAt != nil
}
