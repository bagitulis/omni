package services

import "fmt"

// PlatformFactory creates platform-specific services
// Implements Factory Pattern for OOP
type PlatformFactory struct {
	shopeeOrder   OrderService
	lazadaOrder   OrderService
	tiktokOrder   OrderService
	shopeeProduct ProductService
	lazadaProduct ProductService
	tiktokProduct ProductService
}

// NewPlatformFactory creates a new factory with initialized services
func NewPlatformFactory() *PlatformFactory {
	return &PlatformFactory{
		// Services will be injected during initialization
	}
}

// SetShopeeOrderService sets the Shopee order service
func (f *PlatformFactory) SetShopeeOrderService(s OrderService) {
	f.shopeeOrder = s
}

// SetLazadaOrderService sets the Lazada order service
func (f *PlatformFactory) SetLazadaOrderService(s OrderService) {
	f.lazadaOrder = s
}

// SetTiktokOrderService sets the TikTok order service
func (f *PlatformFactory) SetTiktokOrderService(s OrderService) {
	f.tiktokOrder = s
}

// GetOrderService returns the order service for a platform
func (f *PlatformFactory) GetOrderService(platform string) (OrderService, error) {
	switch platform {
	case "shopee":
		if f.shopeeOrder == nil {
			return nil, fmt.Errorf("shopee order service not initialized")
		}
		return f.shopeeOrder, nil
	case "lazada":
		if f.lazadaOrder == nil {
			return nil, fmt.Errorf("lazada order service not initialized")
		}
		return f.lazadaOrder, nil
	case "tiktok":
		if f.tiktokOrder == nil {
			return nil, fmt.Errorf("tiktok order service not initialized")
		}
		return f.tiktokOrder, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

// GetProductService returns the product service for a platform
func (f *PlatformFactory) GetProductService(platform string) (ProductService, error) {
	switch platform {
	case "shopee":
		return f.shopeeProduct, nil
	case "lazada":
		return f.lazadaProduct, nil
	case "tiktok":
		return f.tiktokProduct, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

// SupportedPlatforms returns list of supported platforms
func (f *PlatformFactory) SupportedPlatforms() []string {
	return []string{"shopee", "lazada", "tiktok"}
}
