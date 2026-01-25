package services

import "context"

// OrderFilter represents order query filters
type OrderFilter struct {
	Platform  string
	Status    string
	StartDate string
	EndDate   string
	Page      int
	PageSize  int
}

// OrderService defines the interface for order operations
// All platform services (Shopee, Lazada, TikTok) must implement this
type OrderService interface {
	GetOrders(ctx context.Context, tenantID string, filter OrderFilter) ([]Order, int, error)
	GetOrderByID(ctx context.Context, tenantID string, orderID string) (*Order, error)
	SyncOrders(ctx context.Context, tenantID string) error
}

// ProductService defines the interface for product operations
type ProductService interface {
	GetProducts(ctx context.Context, tenantID string, page, pageSize int) ([]Product, int, error)
	GetProductByID(ctx context.Context, tenantID string, productID string) (*Product, error)
	SyncProducts(ctx context.Context, tenantID string) error
}

// TokenService defines the interface for OAuth token operations
type TokenService interface {
	GetAccessToken(ctx context.Context, tenantID, platform string) (string, error)
	RefreshToken(ctx context.Context, tenantID, platform string) error
	IsTokenValid(ctx context.Context, tenantID, platform string) bool
}

// Order represents a generic order (platform-agnostic)
type Order struct {
	ID           string  `json:"id"`
	Platform     string  `json:"platform"`
	OrderSN      string  `json:"orderSn"`
	Status       string  `json:"status"`
	TotalAmount  float64 `json:"totalAmount"`
	Currency     string  `json:"currency"`
	BuyerName    string  `json:"buyerName"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

// Product represents a generic product (platform-agnostic)
type Product struct {
	ID          string   `json:"id"`
	Platform    string   `json:"platform"`
	Name        string   `json:"name"`
	SKU         string   `json:"sku"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	Status      string   `json:"status"`
	Images      []string `json:"images"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}
