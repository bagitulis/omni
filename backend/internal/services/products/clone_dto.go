// Package products provides product management DTOs
package products

import (
	"encoding/json"
	"time"
)

// =============================================================================
// Clone Request/Response DTOs
// =============================================================================

// CloneRequest represents a product clone request
type CloneRequest struct {
	SourcePlatform string  `json:"source_platform" binding:"required"`
	TargetPlatform string  `json:"target_platform" binding:"required"`
	SourceItemID   string  `json:"source_item_id" binding:"required"`
	SKU            string  `json:"sku,omitempty"` // SKU to lookup inventory data
	CategoryID     string  `json:"category_id,omitempty"`
	UpdatePrice    bool    `json:"update_price,omitempty"`
	NewPrice       float64 `json:"new_price,omitempty"`
	UseInventory   bool    `json:"use_inventory,omitempty"` // Use inventory data for price/stock (default: true)
	SaveAsDraft    bool    `json:"save_as_draft,omitempty"` // true = AS_DRAFT, false = LISTING (default: false = langsung aktif)
}

// BatchCloneRequest represents batch clone request
type BatchCloneRequest struct {
	SourcePlatform string   `json:"source_platform" binding:"required"`
	TargetPlatform string   `json:"target_platform" binding:"required"`
	SourceItemIDs  []string `json:"source_item_ids" binding:"required"`
	CategoryID     string   `json:"category_id,omitempty"`
}

// CloneResult represents clone operation result
type CloneResult struct {
	ID             string     `json:"id"`
	SourcePlatform string     `json:"source_platform"`
	TargetPlatform string     `json:"target_platform"`
	SourceItemID   string     `json:"source_item_id"`
	TargetItemID   string     `json:"target_item_id,omitempty"`
	Status         string     `json:"status"`
	Message        string     `json:"message,omitempty"`
	Progress       int        `json:"progress"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	SyncTriggered  bool       `json:"sync_triggered,omitempty"` // True if product sync was triggered after clone
	SyncResult     string     `json:"sync_result,omitempty"`    // Result of the sync operation
}

// BatchCloneResult represents batch clone result
type BatchCloneResult struct {
	BatchID        string        `json:"batch_id"`
	TotalRequested int           `json:"total_requested"`
	TotalSuccess   int           `json:"total_success"`
	TotalFailed    int           `json:"total_failed"`
	Status         string        `json:"status"`
	Results        []CloneResult `json:"results"`
}

// PlatformStatus represents SKU existence status on each platform
type PlatformStatus struct {
	Shopee bool `json:"shopee"`
	Lazada bool `json:"lazada"`
	TikTok bool `json:"tiktok"`
}

// CloneTargetsResult represents available clone targets for a SKU
type CloneTargetsResult struct {
	SKU     string         `json:"sku"`
	Status  PlatformStatus `json:"status"`
	Sources []string       `json:"sources"` // Platforms where SKU exists
	Targets []string       `json:"targets"` // Platforms where SKU can be cloned to
}

// =============================================================================
// Conflict Detection DTOs
// =============================================================================

// ConflictResult represents a potential clone conflict
type ConflictResult struct {
	HasConflict   bool            `json:"has_conflict"`
	SourceProduct *ProductSummary `json:"source_product,omitempty"`
	TargetProduct *ProductSummary `json:"target_product,omitempty"`
	Differences   []Difference    `json:"differences,omitempty"`
	Adjustments   *AdjustmentInfo `json:"adjustments,omitempty"`
}

// ProductSummary contains a summary of product data for comparison
type ProductSummary struct {
	Platform string   `json:"platform"`
	ItemID   string   `json:"item_id"`
	SKU      string   `json:"sku"`
	Name     string   `json:"name"`
	Price    float64  `json:"price"`
	Stock    int      `json:"stock"`
	Images   []string `json:"images"`
}

// Difference represents a single field difference between products
type Difference struct {
	Field       string `json:"field"`
	SourceValue string `json:"source_value"`
	TargetValue string `json:"target_value"`
}

// AdjustmentInfo contains info about adjustments that will be made
type AdjustmentInfo struct {
	TitleWillTruncate bool   `json:"title_will_truncate"`
	DescWillTruncate  bool   `json:"desc_will_truncate"`
	OriginalTitle     string `json:"original_title,omitempty"`
	AdjustedTitle     string `json:"adjusted_title,omitempty"`
	TitleLimit        int    `json:"title_limit"`
	DescLimit         int    `json:"desc_limit"`
}

// =============================================================================
// Product Data DTOs
// =============================================================================

// ProductData represents product data for cloning
type ProductData struct {
	ItemID      string            `json:"item_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Price       float64           `json:"price"`
	Stock       int               `json:"stock"`
	CategoryID  string            `json:"category_id"`
	Images      []string          `json:"images"`
	Variants    []ProductVariant  `json:"variants,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	SaveAsDraft bool              `json:"save_as_draft,omitempty"` // TikTok: true = AS_DRAFT, false = LISTING
}

// ProductVariant represents a product variant/SKU
type ProductVariant struct {
	SKU   string  `json:"sku"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// =============================================================================
// Clone Job (for async tracking)
// =============================================================================

// CloneJob represents a persisted clone job for async tracking
type CloneJob struct {
	ID             string          `json:"id" gorm:"primaryKey"`
	TenantID       string          `json:"tenant_id" gorm:"index"`
	SourcePlatform string          `json:"source_platform"`
	TargetPlatform string          `json:"target_platform"`
	SourceItemID   string          `json:"source_item_id"`
	TargetItemID   string          `json:"target_item_id,omitempty"`
	Status         string          `json:"status"`
	Progress       int             `json:"progress"`
	Message        string          `json:"message,omitempty"`
	Data           json.RawMessage `json:"data,omitempty"`
	StartedAt      time.Time       `json:"started_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
}
