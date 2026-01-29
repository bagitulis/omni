package handlers

// CategoryItem represents a category
type CategoryItem struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	ParentID string         `json:"parent_id,omitempty"`
	Level    int            `json:"level"`
	IsLeaf   bool           `json:"is_leaf"`
	Children []CategoryItem `json:"children,omitempty"`
}

// AttributeItem represents an attribute
type AttributeItem struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Required  bool     `json:"required"`
	Options   []string `json:"options,omitempty"`
	InputType string   `json:"input_type"`
}

// BrandItem represents a brand
type BrandItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LogisticsItem represents a logistics option
type LogisticsItem struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Enabled   bool    `json:"enabled"`
	Fee       float64 `json:"fee"`
	MaxWeight float64 `json:"max_weight"`
}

// WarehouseItem represents a warehouse
type WarehouseItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	IsDefault bool   `json:"is_default"`
}

// UploadImageRequest represents image upload request
type UploadImageRequest struct {
	Platform  string `json:"platform" binding:"required"`
	ImageURL  string `json:"image_url,omitempty"`
	ImageData string `json:"image_data,omitempty"`
}

// ValidateProductRequest represents validate product request
type ValidateProductRequest struct {
	Platform string                 `json:"platform" binding:"required"`
	Product  map[string]interface{} `json:"product" binding:"required"`
}

// ValidationResult represents validation result
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ProductTemplate represents a product template
type ProductTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Platform    string                 `json:"platform"`
	CategoryID  string                 `json:"category_id"`
	Fields      map[string]interface{} `json:"fields"`
	Description string                 `json:"description,omitempty"`
}
