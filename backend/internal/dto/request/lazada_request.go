package request

// LazadaShipOrderRequest for shipping a Lazada order
type LazadaShipOrderRequest struct {
	OrderID        string `json:"orderId" binding:"required"`
	TrackingNumber string `json:"trackingNumber" binding:"required"`
	ShippingType   string `json:"shippingType"`
}

// LazadaCancelOrderRequest for cancelling a Lazada order
type LazadaCancelOrderRequest struct {
	OrderID      string `json:"orderId" binding:"required"`
	CancelReason string `json:"cancelReason" binding:"required"`
}

// LazadaCreateProductRequest for creating a Lazada product
type LazadaCreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Brand       string  `json:"brand"`
	Price       float64 `json:"price" binding:"required"`
	Stock       int     `json:"stock" binding:"required"`
	SKU         string  `json:"sku"`
}

// LazadaUpdateProductRequest for updating a Lazada product
type LazadaUpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Status      string  `json:"status"`
}
