package shopeesdk

import (
	"time"

	shopee "github.com/omni/backend/pkg/shopee"
)

// Config holds client credentials.
type Config struct {
	PartnerID     int64
	PartnerKey    string
	ShopID        int64
	AccessToken   string
	UseProduction bool
}

// OrderListParams controls order listing.
type OrderListParams struct {
	From           time.Time
	To             time.Time
	Status         string
	TimeRangeField string // create_time or update_time
}

// OrderDetailParams controls order detail lookups.
type OrderDetailParams struct {
	OrderSNs               []string
	ResponseOptionalFields string // comma separated fields; defaults to "item_list"
}

// RawCall represents an arbitrary Shopee API invocation.
type RawCall struct {
	Method string
	Path   string
	Query  map[string]string
	Body   interface{}
}

// Type aliases to reuse underlying transport models.
type (
	OrderListResponse               = shopee.GetOrderListResponse
	OrderDetailResponse             = shopee.GetOrderDetailResponse
	OrderDetailItem                 = shopee.OrderDetailItem
	GetEscrowDetailsRequest         = shopee.GetEscrowDetailsRequest
	GetEscrowDetailsResponse        = shopee.GetEscrowDetailsResponse
	ProductListResponse             = shopee.ProductListResponse
	ProductDetailResponse           = shopee.ProductDetailResponse
	ProductDetailWithImages         = shopee.ProductDetailWithImages
	ProductDetailWithImagesResponse = shopee.ProductDetailWithImagesResponse
	ModelListResponse               = shopee.ModelListResponse
	CreateProductRequest            = shopee.CreateProductRequest
	CreateProductResponse           = shopee.CreateProductResponse
	UpdateProductRequest            = shopee.UpdateProductRequest
	UpdateProductResponse           = shopee.UpdateProductResponse
	DeleteProductResponse           = shopee.DeleteProductResponse
	UpdateStockRequest              = shopee.UpdateStockRequest
	UpdateStockResponse             = shopee.UpdateStockResponse
	StockListItem                   = shopee.StockListItem
	SellerStock                     = shopee.SellerStock
	UpdatePriceRequest              = shopee.UpdatePriceRequest
	UpdatePriceResponse             = shopee.UpdatePriceResponse
	PriceInfo                       = shopee.PriceInfo
	UploadImageResponse             = shopee.UploadImageResponse
	GetWalletTransactionRequest     = shopee.GetWalletTransactionRequest
	WalletTransactionResponse       = shopee.WalletTransactionResponse
	ShipOrderRequest                = shopee.ShipOrderRequest
	ShipOrderResponse               = shopee.ShipOrderResponse
	CancelOrderRequest              = shopee.CancelOrderRequest
	CancelOrderResponse             = shopee.CancelOrderResponse
	GetShippingParameterResponse    = shopee.GetShippingParameterResponse
	GetTrackingNumberResponse       = shopee.GetTrackingNumberResponse
)
