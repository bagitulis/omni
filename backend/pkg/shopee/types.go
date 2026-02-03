package shopee

// GetEscrowDetailsRequest represents request for escrow details
type GetEscrowDetailsRequest struct {
	OrderSNList []string `json:"order_sn_list"`
}

// GetEscrowDetailsResponse represents response for escrow details batch
type GetEscrowDetailsResponse struct {
	Response []EscrowDetailWrapper `json:"response"`
	Error    string                `json:"error"`
	Message  string                `json:"message"`
}

// EscrowDetailWrapper wraps escrow_detail in batch response
type EscrowDetailWrapper struct {
	EscrowDetail *EscrowOrder `json:"escrow_detail"`
}

// EscrowOrder represents an order in escrow response
type EscrowOrder struct {
	OrderSN          string          `json:"order_sn"`
	BuyerUsername    string          `json:"buyer_user_name"`
	PayTime          int64           `json:"pay_time"`
	OrderStatus      string          `json:"order_status"`
	TotalAmount      float64         `json:"escrow_amount"`
	OrderIncome      EscrowOrderData `json:"order_income"`
	BuyerPaymentInfo map[string]any  `json:"buyer_payment_info"`
}

// EscrowOrderData represents the order_income object from Shopee API
type EscrowOrderData struct {
	EscrowAmount               float64          `json:"escrow_amount"`
	BuyerTotalAmount           float64          `json:"buyer_total_amount"`
	OriginalPrice              float64          `json:"original_price"`
	SellerDiscount             float64          `json:"seller_discount"`
	ShopeeDiscount             float64          `json:"shopee_discount"`
	VoucherFromSeller          float64          `json:"voucher_from_seller"`
	VoucherFromShopee          float64          `json:"voucher_from_shopee"`
	Coins                      float64          `json:"coins"`
	BuyerPaidShippingFee       float64          `json:"buyer_paid_shipping_fee"`
	BuyerTransactionFee        float64          `json:"buyer_transaction_fee"`
	CrossBorderTax             float64          `json:"cross_border_tax"`
	PaymentPromotion           float64          `json:"payment_promotion"`
	CommissionFee              float64          `json:"commission_fee"`
	ServiceFee                 float64          `json:"service_fee"`
	SellerTransactionFee       float64          `json:"seller_transaction_fee"`
	SellerLostCompensation     float64          `json:"seller_lost_compensation"`
	SellerCoinCashBack         float64          `json:"seller_coin_cash_back"`
	EscrowTax                  float64          `json:"escrow_tax"`
	FinalShippingFee           float64          `json:"final_shipping_fee"`
	ActualShippingFee          float64          `json:"actual_shipping_fee"`
	ShopeeShippingRebate       float64          `json:"shopee_shipping_rebate"`
	ShippingFeeDiscountFrom3pl float64          `json:"shipping_fee_discount_from_3pl"`
	SellerShippingDiscount     float64          `json:"seller_shipping_discount"`
	EstimatedShippingFee       float64          `json:"estimated_shipping_fee"`
	SellerOrderProcessingFee   float64          `json:"seller_order_processing_fee"`
	DrcAdjustableRefund        float64          `json:"drc_adjustable_refund"`
	BuyerPaymentMethod         string           `json:"buyer_payment_method"`
	Items                      []EscrowItemData `json:"items"`
}

// EscrowItemData represents an item in escrow order_income
type EscrowItemData struct {
	ItemID                    int64   `json:"item_id"`
	ModelID                   int64   `json:"model_id"`
	ItemName                  string  `json:"item_name"`
	ModelName                 string  `json:"model_name"`
	ItemSKU                   string  `json:"item_sku"`
	ModelSKU                  string  `json:"model_sku"`
	QuantityPurchased         int     `json:"quantity_purchased"`
	OriginalPrice             float64 `json:"original_price"`
	SellingPrice              float64 `json:"selling_price"`
	DiscountedPrice           float64 `json:"discounted_price"`
	SellerDiscount            float64 `json:"seller_discount"`
	ShopeeDiscount            float64 `json:"shopee_discount"`
	DiscountFromCoin          float64 `json:"discount_from_coin"`
	DiscountFromVoucherSeller float64 `json:"discount_from_voucher_seller"`
	DiscountFromVoucherShopee float64 `json:"discount_from_voucher_shopee"`
	AmsCommissionFee          float64 `json:"ams_commission_fee"`
	SellerOrderProcessingFee  float64 `json:"seller_order_processing_fee"`
}

// GetOrderListRequest represents request for order list
type GetOrderListRequest struct {
	TimeFrom       int64  `json:"time_from"`
	TimeTo         int64  `json:"time_to"`
	TimeRangeField string `json:"time_range_field"`
	PageSize       int    `json:"page_size"`
	Cursor         string `json:"cursor,omitempty"`
}

// GetOrderListResponse represents response for order list
type GetOrderListResponse struct {
	Response struct {
		OrderList  []OrderBasic `json:"order_list"`
		NextCursor string       `json:"next_cursor"`
		More       bool         `json:"more"`
	} `json:"response"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// OrderBasic represents basic order info
type OrderBasic struct {
	OrderSN string `json:"order_sn"`
}

// GetOrderDetailResponse represents response for order details
type GetOrderDetailResponse struct {
	Response struct {
		OrderList []OrderDetailItem `json:"order_list"`
	} `json:"response"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// OrderDetailItem represents an order detail item
type OrderDetailItem struct {
	OrderSN                 string            `json:"order_sn"`
	OrderStatus             string            `json:"order_status"`
	TotalAmount             float64           `json:"total_amount"`
	Currency                string            `json:"currency"`
	CreateTime              int64             `json:"create_time"`
	UpdateTime              int64             `json:"update_time"`
	PaymentMethod           string            `json:"payment_method"`
	BuyerUserID             int64             `json:"buyer_user_id"`
	BuyerUsername           string            `json:"buyer_username"`
	ShippingCarrier         string            `json:"shipping_carrier"`
	CheckoutShippingCarrier string            `json:"checkout_shipping_carrier"`
	ShipByDate              int64             `json:"ship_by_date"` // Deadline timestamp for shipping
	DaysToShip              int               `json:"days_to_ship"` // Number of days to ship
	TrackingNo              string            `json:"tracking_no"`  // Tracking number
	Note                    string            `json:"note"`
	MessageToSeller         string            `json:"message_to_seller"`
	ItemList                []OrderItemDetail `json:"item_list"` // Order items
}

// OrderItemDetail represents an item in an order
type OrderItemDetail struct {
	ItemID                 int64          `json:"item_id"`
	ModelID                int64          `json:"model_id"`
	ItemName               string         `json:"item_name"`
	ModelName              string         `json:"model_name"`
	ItemSKU                string         `json:"item_sku"`
	ModelSKU               string         `json:"model_sku"`
	ModelQuantityPurchased int            `json:"model_quantity_purchased"`
	ModelOriginalPrice     float64        `json:"model_original_price"`
	ModelDiscountedPrice   float64        `json:"model_discounted_price"`
	ImageInfo              *ItemImageInfo `json:"image_info,omitempty"` // Product image info from Shopee API
}

// ItemImageInfo represents image info for an order item
type ItemImageInfo struct {
	ImageURL string `json:"image_url"` // Product image URL
}
