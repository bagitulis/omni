package models

import "time"

// Notification represents an in-app notification (Facebook/Instagram style).
// Stored per-tenant schema — no tenant_id column needed.
//
// V2 (2026-09-16) adds: severity, dedup_key/dedup_count, source, actor_user_id,
// recipient_user_id (NULL = tenant-wide), expires_at, snoozed_until, updated_at.
// Legacy Read column is retained but no longer written by new code paths; a
// separate NotificationRead join table tracks per-user read state.
type Notification struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Type            string     `gorm:"column:type;not null;size:20" json:"type"`
	Category        string     `gorm:"column:category;not null;size:30" json:"category"`
	Severity        int16      `gorm:"column:severity;not null;default:20;index" json:"severity"`
	Title           string     `gorm:"column:title;not null;size:255" json:"title"`
	Message         string     `gorm:"column:message;type:text" json:"message"`
	Metadata        string     `gorm:"column:metadata;type:jsonb;default:'{}'" json:"metadata,omitempty"`
	ActionURL       string     `gorm:"column:action_url;size:500" json:"action_url,omitempty"`
	DedupKey        *string    `gorm:"column:dedup_key;size:200" json:"dedup_key,omitempty"`
	DedupCount      int        `gorm:"column:dedup_count;not null;default:1" json:"dedup_count"`
	Source          string     `gorm:"column:source;size:60" json:"source,omitempty"`
	ActorUserID     *int64     `gorm:"column:actor_user_id" json:"actor_user_id,omitempty"`
	RecipientUserID *int64     `gorm:"column:recipient_user_id;index" json:"recipient_user_id,omitempty"`
	ExpiresAt       *time.Time `gorm:"column:expires_at" json:"expires_at,omitempty"`
	SnoozedUntil    *time.Time `gorm:"column:snoozed_until" json:"snoozed_until,omitempty"`
	Read            bool       `gorm:"column:read;default:false" json:"-"` // legacy, kept for backward compat during migration window; not serialized
	CreatedAt       time.Time  `gorm:"column:created_at;index;not null" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null" json:"updated_at"`
}

// TableName returns the table name
func (Notification) TableName() string {
	return "notifications"
}

// NotificationRead tracks per-user read state for a notification.
// Composite PK (notification_id, user_id).
type NotificationRead struct {
	NotificationID int64     `gorm:"column:notification_id;primaryKey;not null" json:"notification_id"`
	UserID         int64     `gorm:"column:user_id;primaryKey;not null;index" json:"user_id"`
	ReadAt         time.Time `gorm:"column:read_at;not null;default:CURRENT_TIMESTAMP" json:"read_at"`
}

// TableName returns the table name
func (NotificationRead) TableName() string {
	return "notification_reads"
}

// NotificationSettings stores per-tenant notification preferences.
type NotificationSettings struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RetentionDays    int       `gorm:"column:retention_days;default:30" json:"retention_days"`             // 7, 30, 90, or 0 = manual
	MinSeverityToast int16     `gorm:"column:min_severity_toast;not null;default:20" json:"min_severity_toast"` // FE suppresses toast below this
	UIPrefs          string    `gorm:"column:ui_prefs;type:jsonb;default:'{}'" json:"ui_prefs,omitempty"`
	CreatedAt        time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name
func (NotificationSettings) TableName() string {
	return "notification_settings"
}
