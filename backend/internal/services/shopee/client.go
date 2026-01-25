package shopee

import "context"

// APIClient defines the interface for Shopee API operations
// This allows for dependency injection and testing
// NOTE: Methods must match signatures in pkg/shopee/client.go
type APIClient interface {
	// Wallet operations
	GetWalletBalance(ctx context.Context) (map[string]interface{}, error)

	// Shipping operations
	GetShippingOptions(ctx context.Context, orderSn string) (map[string]interface{}, error)
	GetTrackingInfo(ctx context.Context, orderSn string) (map[string]interface{}, error)
	GetShipmentInfo(ctx context.Context, orderSn string) (map[string]interface{}, error)
}
