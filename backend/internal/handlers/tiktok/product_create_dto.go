package tiktok

import (
	"encoding/json"
)

// ProductDraft represents a draft product
type ProductDraft struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenantId"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	CategoryID  string          `json:"categoryId"`
	Images      []string        `json:"images"`
	SKUs        []ProductSKU    `json:"skus"`
	Attributes  json.RawMessage `json:"attributes"`
	Status      string          `json:"status"`
	CreatedAt   int64           `json:"createdAt"`
	UpdatedAt   int64           `json:"updatedAt"`
}

// ProductSKU represents a product SKU
type ProductSKU struct {
	SKU      string  `json:"sku"`
	Price    float64 `json:"price"`
	Stock    int     `json:"stock"`
	ImageURL string  `json:"imageUrl,omitempty"`
}

// Category represents a product category
type Category struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	ParentID string     `json:"parentId,omitempty"`
	Level    int        `json:"level"`
	IsLeaf   bool       `json:"isLeaf"`
	Children []Category `json:"children,omitempty"`
}

// Attribute represents a category attribute
type Attribute struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Options     []string `json:"options,omitempty"`
	InputType   string   `json:"inputType"`
	CustomValue bool     `json:"customValue"`
}

// CategoryRule represents category rules
type CategoryRule struct {
	MaxImages      int      `json:"maxImages"`
	MaxSKUs        int      `json:"maxSKUs"`
	MaxDescription int      `json:"maxDescription"`
	RequiredFields []string `json:"requiredFields"`
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
	MaxWeight   float64 `json:"maxWeight"`
	MaxSize     string  `json:"maxSize"`
	IsAvailable bool    `json:"isAvailable"`
}

// Warehouse represents a warehouse
type Warehouse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	IsDefault bool   `json:"isDefault"`
	Status    string `json:"status"`
}
