package models

import "time"

// AnalyticsCacheMetadata tracks materialized view refresh metadata.
// Stored per-tenant schema — uses (tenant_id, view_name) as unique key
// for the ON CONFLICT upsert in cache_service.go.
type AnalyticsCacheMetadata struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ViewName      string    `gorm:"column:view_name;not null;uniqueIndex" json:"view_name"`
	LastRefresh   time.Time `gorm:"column:last_refresh" json:"last_refresh"`
	RefreshTimeMs int64     `gorm:"column:refresh_time_ms" json:"refresh_time_ms"`
	RowCount      int64     `gorm:"column:row_count" json:"row_count"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name
func (AnalyticsCacheMetadata) TableName() string {
	return "analytics_cache_metadata"
}
