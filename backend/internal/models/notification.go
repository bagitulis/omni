package models

import "time"

// Notification represents an in-app notification (Facebook/Instagram style).
// Stored per-tenant schema — no tenant_id column needed.
type Notification struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Type      string    `gorm:"column:type;not null;size:20" json:"type"`           // success, error, warning, info
	Category  string    `gorm:"column:category;not null;size:30" json:"category"`   // sync, order, product, etc.
	Title     string    `gorm:"column:title;not null;size:255" json:"title"`
	Message   string    `gorm:"column:message;type:text" json:"message"`
	Read      bool      `gorm:"column:read;default:false" json:"read"`
	ActionURL string    `gorm:"column:action_url;size:500" json:"action_url,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at;index;not null" json:"created_at"`
}

// TableName returns the table name
func (Notification) TableName() string {
	return "notifications"
}

// NotificationSettings stores per-tenant notification preferences.
type NotificationSettings struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RetentionDays int       `gorm:"column:retention_days;default:30" json:"retention_days"` // 7, 30, 90, or 0 = manual
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name
func (NotificationSettings) TableName() string {
	return "notification_settings"
}
