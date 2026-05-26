package models

import "time"

// CredentialAuditEvent records sanitized lifecycle events for credential operations.
// Never contains secret payloads — only metadata and reason codes.
type CredentialAuditEvent struct {
	ID              string    `gorm:"column:id;primaryKey;type:varchar(255)" json:"id"`
	TenantID        string    `gorm:"column:tenant_id;type:varchar(100);not null;index" json:"tenant_id"`
	Platform        string    `gorm:"column:platform;type:varchar(50);not null;index" json:"platform"`
	StoreIdentifier string    `gorm:"column:store_identifier;type:varchar(255)" json:"store_identifier,omitempty"`
	EventType       string    `gorm:"column:event_type;type:varchar(100);not null;index" json:"event_type"`
	Status          string    `gorm:"column:status;type:varchar(50);not null" json:"status"`
	Code            string    `gorm:"column:code;type:varchar(100)" json:"code,omitempty"`
	Actor           string    `gorm:"column:actor;type:varchar(100)" json:"actor"`
	ActorRole       string    `gorm:"column:actor_role;type:varchar(50)" json:"actor_role"`
	Metadata        JSONMap   `gorm:"column:metadata;type:jsonb;default:'{}'" json:"metadata,omitempty"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName returns the literal table name for GORM.
func (CredentialAuditEvent) TableName() string {
	return "credential_audit_events"
}
