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
	SourcePlatform string  `json:"sourcePlatform" binding:"required"`
	TargetPlatform string  `json:"targetPlatform" binding:"required"`
	SourceItemID   string  `json:"sourceItemId" binding:"required"`
	SKU            string  `json:"sku,omitempty"`            // SKU to lookup inventory data
	CategoryID     string  `json:"categoryId,omitempty"`
	UpdatePrice    bool    `json:"updatePrice,omitempty"`
	NewPrice       float64 `json:"newPrice,omitempty"`
	UseInventory   bool    `json:"useInventory,omitempty"`   // Use inventory data for price/stock (default: true)
	SaveAsDraft    bool    `json:"saveAsDraft,omitempty"`    // true = AS_DRAFT, false = LISTING (default: false = langsung aktif)
}

// BatchCloneRequest represents batch clone request
type BatchCloneRequest struct {
	SourcePlatform string   `json:"sourcePlatform" binding:"required"`
	TargetPlatform string   `json:"targetPlatform" binding:"required"`
	SourceItemIDs  []string `json:"sourceItemIds" binding:"required"`
	CategoryID     string   `json:"categoryId,omitempty"`
}

// CloneResult represents clone operation result
type CloneResult struct {
	ID             string     `json:"id"`
	SourcePlatform string     `json:"sourcePlatform"`
	TargetPlatform string     `json:"targetPlatform"`
	SourceItemID   string     `json:"sourceItemId"`
	TargetItemID   string     `json:"targetItemId,omitempty"`
	Status         string     `json:"status"`
	Message        string     `json:"message,omitempty"`
	Progress       int        `json:"progress"`
	StartedAt      time.Time  `json:"startedAt"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	SyncTriggered  bool       `json:"syncTriggered,omitempty"`  // True if product sync was triggered after clone
	SyncResult     string     `json:"syncResult,omitempty"`     // Result of the sync operation
}

// BatchCloneResult represents batch clone result
type BatchCloneResult struct {
	BatchID        string        `json:"batchId"`
	TotalRequested int           `json:"totalRequested"`
	TotalSuccess   int           `json:"totalSuccess"`
	TotalFailed    int           `json:"totalFailed"`
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
// Product Data DTOs
// =============================================================================

// ProductData represents product data for cloning
type ProductData struct {
	ItemID      string            `json:"itemId"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Price       float64           `json:"price"`
	Stock       int               `json:"stock"`
	CategoryID  string            `json:"categoryId"`
	Images      []string          `json:"images"`
	Variants    []ProductVariant  `json:"variants,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	SaveAsDraft bool              `json:"saveAsDraft,omitempty"` // TikTok: true = AS_DRAFT, false = LISTING
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
	TenantID       string          `json:"tenantId" gorm:"index"`
	SourcePlatform string          `json:"sourcePlatform"`
	TargetPlatform string          `json:"targetPlatform"`
	SourceItemID   string          `json:"sourceItemId"`
	TargetItemID   string          `json:"targetItemId,omitempty"`
	Status         string          `json:"status"`
	Progress       int             `json:"progress"`
	Message        string          `json:"message,omitempty"`
	Data           json.RawMessage `json:"data,omitempty"`
	StartedAt      time.Time       `json:"startedAt"`
	CompletedAt    *time.Time      `json:"completedAt,omitempty"`
}
