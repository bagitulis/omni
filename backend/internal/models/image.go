// Package models contains database models for the application
package models

import "time"

// Image represents an image stored in the system with content-based deduplication
type Image struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"index;not null" json:"tenant_id"`
	ContentHash string    `gorm:"size:64;index;not null" json:"content_hash"`
	OriginalURL string    `gorm:"index" json:"original_url,omitempty"`
	LocalPath   string    `gorm:"not null" json:"local_path"` // base path without size suffix (e.g., "uploads/tenant/images/hash")
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	FileSize    int64     `json:"file_size"`
	RefCount    int       `gorm:"default:1" json:"ref_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName returns the PostgreSQL table name
func (Image) TableName() string {
	return GetTableName("Image")
}
