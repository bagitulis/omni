// Package lazada provides Lazada handler DTOs
package lazada

// CreateProductRequest represents create product request
type CreateProductRequest struct {
	Name            string  `json:"name" binding:"required"`
	Description     string  `json:"description" binding:"required"`
	Brand           string  `json:"brand,omitempty"`
	PrimaryCategory int64   `json:"primary_category" binding:"required"`
	SellerSku       string  `json:"seller_sku" binding:"required"`
	Price           float64 `json:"price" binding:"required"`
	Quantity        int     `json:"quantity" binding:"required"`
}

// UpdateProductRequest represents update product request
type UpdateProductRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}
