// Package models contains database models for the application
package models

import "time"

// Image represents an image stored in the gallery
type Image struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	Filename    string    `gorm:"column:filename;not null" json:"filename"`
	OriginalURL string    `gorm:"column:original_url" json:"original_url,omitempty"`
	LocalPath   string    `gorm:"column:local_path;not null" json:"local_path"`
	Width       int       `gorm:"column:width" json:"width,omitempty"`
	Height      int       `gorm:"column:height" json:"height,omitempty"`
	Size        int64     `gorm:"column:size" json:"size"`
	MimeType    string    `gorm:"column:mime_type" json:"mime_type"`
	Category    string    `gorm:"column:category;index" json:"category"` // "products" or "gallery"
	ProductID   *uint     `gorm:"column:product_id;index" json:"product_id,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the PostgreSQL table name
func (Image) TableName() string {
	return GetTableName("Image")
}
