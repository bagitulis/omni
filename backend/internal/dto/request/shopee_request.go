package request

// ShipOrderRequest for shipping an order
type ShipOrderRequest struct {
	OrderSN         string `json:"order_sn" binding:"required"`
	TrackingNumber  string `json:"tracking_number" binding:"required"`
	ShippingCarrier string `json:"shipping_carrier"`
}

// CancelOrderRequest for cancelling an order
type CancelOrderRequest struct {
	OrderSN      string `json:"order_sn" binding:"required"`
	CancelReason string `json:"cancel_reason" binding:"required"`
}

// CreateProductRequest for creating a product
type CreateProductRequest struct {
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description"`
	CategoryID    int64    `json:"category_id" binding:"required"`
	OriginalPrice float64  `json:"original_price" binding:"required"`
	Stock         int      `json:"stock" binding:"required"`
	Images        []string `json:"images"`
	SKU           string   `json:"sku"`
	Weight        float64  `json:"weight"`
}

// UpdateProductRequest for updating a product
type UpdateProductRequest struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	OriginalPrice float64  `json:"original_price"`
	Stock         int      `json:"stock"`
	Images        []string `json:"images"`
	Status        string   `json:"status"`
}

// UpdateStockRequest for updating stock only
type UpdateStockRequest struct {
	Stock int `json:"stock" binding:"required"`
}

// UpdatePriceRequest for updating price only
type UpdatePriceRequest struct {
	OriginalPrice float64 `json:"original_price" binding:"required"`
}
