package tiktok

import (
	"encoding/json"
)

// ProductDraft represents a draft product
type ProductDraft struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	CategoryID  string          `json:"category_id"`
	Images      []string        `json:"images"`
	SKUs        []ProductSKU    `json:"skus"`
	Attributes  json.RawMessage `json:"attributes"`
	Status      string          `json:"status"`
	CreatedAt   int64           `json:"created_at"`
	UpdatedAt   int64           `json:"updated_at"`
}

// ProductSKU represents a product SKU
type ProductSKU struct {
	SKU      string  `json:"sku"`
	Price    float64 `json:"price"`
	Stock    int     `json:"stock"`
	ImageURL string  `json:"image_url,omitempty"`
}

// Category represents a product category
type Category struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	ParentID string     `json:"parent_id,omitempty"`
	Level    int        `json:"level"`
	IsLeaf   bool       `json:"is_leaf"`
	Children []Category `json:"children,omitempty"`
}

// Attribute represents a category attribute
type Attribute struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Options     []string `json:"options,omitempty"`
	InputType   string   `json:"input_type"`
	CustomValue bool     `json:"custom_value"`
}

// CategoryRule represents category rules
type CategoryRule struct {
	MaxImages      int      `json:"max_images"`
	MaxSKUs        int      `json:"max_skus"`
	MaxDescription int      `json:"max_description"`
	RequiredFields []string `json:"required_fields"`
}

// Brand represents a brand
type Brand struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DeliveryOption represents delivery options
type DeliveryOption struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	MaxWeight   float64 `json:"max_weight"`
	MaxSize     string  `json:"max_size"`
	IsAvailable bool    `json:"is_available"`
}

// Warehouse represents a warehouse
type Warehouse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	IsDefault bool   `json:"is_default"`
	Status    string `json:"status"`
}
