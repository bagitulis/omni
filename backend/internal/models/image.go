// Package models contains database models for the application
package models

import "time"

// Image represents an image stored in the gallery
type Image struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	TenantID    string     `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	Filename    string     `gorm:"column:filename;not null" json:"filename"`
	OriginalURL string     `gorm:"column:original_url;index" json:"original_url,omitempty"` // Source URL for dedup
	LocalPath   string     `gorm:"column:local_path;not null" json:"local_path"`
	ContentHash string     `gorm:"column:content_hash;size:64;index" json:"content_hash"` // [NEW] SHA256 hash
	Width       int        `gorm:"column:width" json:"width,omitempty"`
	Height      int        `gorm:"column:height" json:"height,omitempty"`
	Size        int64      `gorm:"column:size" json:"size"`
	MimeType    string     `gorm:"column:mime_type" json:"mime_type"`
	Category    string     `gorm:"column:category;index" json:"category"` // "products" or "gallery"
	ProductID   *uint      `gorm:"column:product_id;index" json:"product_id,omitempty"`
	RefCount    int        `gorm:"column:ref_count;default:0" json:"ref_count"`              // [NEW] Reference count
	Thumbnails  JSONMap    `gorm:"column:thumbnails;type:jsonb" json:"thumbnails,omitempty"` // [NEW] {small, medium, large}
	DeletedAt   *time.Time `gorm:"column:deleted_at;index" json:"deleted_at,omitempty"`      // [NEW] Soft delete
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// CREATE UNIQUE INDEX idx_images_tenant_hash ON images(tenant_id, content_hash);

// TableName returns the PostgreSQL table name
func (Image) TableName() string {
	return GetTableName("Image")
}
