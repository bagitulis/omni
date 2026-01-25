package request

// TiktokShipOrderRequest for shipping a TikTok order
type TiktokShipOrderRequest struct {
	OrderID          string `json:"orderId" binding:"required"`
	TrackingNumber   string `json:"trackingNumber" binding:"required"`
	ShippingProvider string `json:"shippingProvider"`
}

// TiktokCancelOrderRequest for cancelling a TikTok order
type TiktokCancelOrderRequest struct {
	OrderID      string `json:"orderId" binding:"required"`
	CancelReason string `json:"cancelReason" binding:"required"`
}

// TiktokCreateProductRequest for creating a TikTok product
type TiktokCreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	CategoryID  string  `json:"categoryId" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	Stock       int     `json:"stock" binding:"required"`
	SKU         string  `json:"sku"`
}

// TiktokUpdateProductRequest for updating a TikTok product
type TiktokUpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Status      string  `json:"status"`
}
