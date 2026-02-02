package sync

import (
	"context"
	"time"
)

// OrderManager interface for platform-specific order managers
type OrderManager interface {
	GetOrderList(ctx context.Context, status string, days int) ([]Order, error)
	GetOrderDetails(ctx context.Context, orderIDs []string) ([]Order, error)
	GetOrderItems(ctx context.Context, orderIDs []string) (map[string][]OrderItem, error)
}

// Order represents a generic order
// Frontend expects order_no (from Node.js formatter)
type Order struct {
	ID            string  `json:"id"`
	OrderSN       string  `json:"-"` // Internal use only, not sent to frontend
	OrderNo       string  `json:"order_no"`
	Platform      string  `json:"platform"`
	Status        string  `json:"status"`
	Category      string  `json:"category"`
	TotalAmount   float64 `json:"total_amount"`
	Currency      string  `json:"currency"`
	BuyerUsername string  `json:"buyer_username"`
	// Payment and shipping info (for Order Manager display)
	PaymentMethod   string `json:"payment_method,omitempty"`
	ShippingType    string `json:"shipping_type,omitempty"`
	TrackingNumber  string `json:"tracking_number,omitempty"`
	ShippingCarrier string `json:"shipping_carrier,omitempty"`
	BuyerMessage    string `json:"buyer_message,omitempty"`
	Countdown       string `json:"countdown,omitempty"`
	// Flattened item fields (for frontend compatibility - one row per item)
	SKU           string  `json:"sku"`
	ProductName   string  `json:"product_name"`
	VariationName string  `json:"variation_name"`
	Quantity      int     `json:"qty"` // frontend expects "qty" not "quantity"
	Price         float64 `json:"price,omitempty"`
	ProductImage  string  `json:"product_image,omitempty"`
	// Legacy nested items (optional, for backward compatibility)
	Items     []OrderItem `json:"items,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID            string  `json:"id"`
	OrderID       string  `json:"order_id"`
	ItemID        int64   `json:"item_id,omitempty"`
	SKU           string  `json:"sku"`
	ProductName   string  `json:"product_name"`
	VariationName string  `json:"variation_name"`
	Quantity      int     `json:"quantity"`
	Price         float64 `json:"price"`
	TotalPrice    float64 `json:"total_price"`
	ProductImage  string  `json:"product_image,omitempty"`
	// Tracking info (for Lazada items which have per-item tracking)
	TrackingNumber  string `json:"tracking_number,omitempty"`
	ShippingCarrier string `json:"shipping_carrier,omitempty"`
}

// OrderRepository interface for order persistence
type OrderRepository interface {
	SaveOrders(ctx context.Context, platform PlatformType, orders []Order) error
	GetOrdersByStatus(ctx context.Context, platform PlatformType, status string, limit int) ([]Order, error)
	ClearOrdersByStatus(ctx context.Context, platform PlatformType, status string) error
	GetOrdersCount(ctx context.Context, platform PlatformType, status string) (int64, error)
}

// SyncResult represents result of a sync operation
type SyncResult struct {
	Platform PlatformType `json:"platform"`
	Success  bool         `json:"success"`
	Count    int          `json:"count"`
	Orders   []Order      `json:"orders,omitempty"`
	Error    string       `json:"error,omitempty"`
}
