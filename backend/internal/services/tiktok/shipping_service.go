package tiktok

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/repositories"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// TikTokClient defines the interface for TikTok API operations
// This allows for mocking in tests
type TikTokClient interface {
	ArrangeShipment(packageID string, req *tiktokPkg.ShipPackageRequest) (*tiktokPkg.ShipPackageResponse, error)
	GetShippingDocument(packageID, documentType string) (string, error)
	GetOrderPackages(orderID string) (*tiktokPkg.GetPackageDetailResponse, error)
	GetOrderDetail(orderIDs []string) (*tiktokPkg.OrderDetailResponse, error)
	GetHandoverTimeSlots(packageID string) (*tiktokPkg.HandoverTimeSlotsResponse, error)
	GetHandoverTimeSlotsForOrder(orderID string, lineItemIDs []string) (*tiktokPkg.HandoverTimeSlotsResponse, error)
	ResolveOrderToPackageID(orderID string) (string, string, error)
}

// Ensure tiktokPkg.Client implements TikTokClient
var _ TikTokClient = (*tiktokPkg.Client)(nil)

// ClientFactory is a function that returns a TikTokClient
type ClientFactory func(ctx context.Context, tenantID string) (TikTokClient, error)

// ShippingService handles TikTok shipping operations
type ShippingService struct {
	basePath      string
	clientFactory ClientFactory
}

// NewShippingService creates a new shipping service
func NewShippingService(basePath string) *ShippingService {
	s := &ShippingService{
		basePath: basePath,
	}
	s.clientFactory = s.defaultGetClient
	return s
}

// NewShippingServiceWithFactory creates a new shipping service with a custom client factory (for testing)
func NewShippingServiceWithFactory(basePath string, factory ClientFactory) *ShippingService {
	return &ShippingService{
		basePath:      basePath,
		clientFactory: factory,
	}
}

// defaultGetClient creates TikTok API client with tenant-specific credentials
func (s *ShippingService) defaultGetClient(ctx context.Context, tenantID string) (TikTokClient, error) {
	db, err := config.GetTenantDB(tenantID, s.basePath)
	if err != nil {
		return nil, err
	}

	// Use PlatformCredentialsRepository for key-value based config
	credRepo := repositories.NewPlatformCredentialsRepository(db)
	tenantCreds, err := credRepo.GetTiktokCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if tenantCreds.AccessToken == "" || tenantCreds.ShopCipher == "" {
		return nil, fmt.Errorf("missing TikTok credentials: accessToken or shopCipher not configured")
	}

	// Use tenant credentials for appKey/appSecret if available, otherwise fall back to global
	appKey := tenantCreds.AppKey
	appSecret := tenantCreds.AppSecret

	if appKey == "" || appSecret == "" {
		systemDB, err := config.GetSystemDB(s.basePath)
		if err != nil {
			return nil, err
		}
		globalRepo := repositories.NewGlobalConfigRepository(systemDB)
		globalCreds, err := globalRepo.GetTiktokCredentials(ctx)
		if err != nil {
			return nil, err
		}
		appKey = globalCreds.AppKey
		appSecret = globalCreds.AppSecret
	}

	client := tiktokPkg.NewClient(appKey, appSecret)
	client.SetCredentials(tenantCreds.AccessToken, tenantCreds.ShopCipher)
	return client, nil
}

// ArrangeShipment arranges shipment for a package
func (s *ShippingService) ArrangeShipment(ctx context.Context, tenantID, packageID string, req *tiktokPkg.ShipPackageRequest) (*tiktokPkg.ShipPackageResponse, error) {
	client, err := s.clientFactory(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return client.ArrangeShipment(packageID, req)
}

// GetShippingLabel retrieves shipping label URL
func (s *ShippingService) GetShippingLabel(ctx context.Context, tenantID, packageID, documentType string) (string, error) {
	client, err := s.clientFactory(ctx, tenantID)
	if err != nil {
		return "", err
	}

	// Try getting shipping document directly with packageID
	docURL, err := client.GetShippingDocument(packageID, documentType)
	if err == nil && docURL != "" {
		return docURL, nil
	}

	// If failed, maybe packageID is actually orderSN? Try to resolve packageID from orderSN
	// This handles the case where frontend sends order_sn because it doesn't have package_id
	pkgResp, pkgErr := client.GetOrderPackages(packageID) // Treat 'packageID' input as 'orderID'
	if pkgErr == nil && pkgResp != nil && len(pkgResp.Data.Packages) > 0 {
		realPackageID := pkgResp.Data.Packages[0].ID
		if realPackageID != "" && realPackageID != packageID {
			// Retry with resolved packageID
			return client.GetShippingDocument(realPackageID, documentType)
		}
	}

	// If resolution failed or didn't help, return original error
	return "", err
}

// GetHandoverTimeSlots retrieves available pickup/drop-off time slots
func (s *ShippingService) GetHandoverTimeSlots(ctx context.Context, tenantID, orderOrPackageID string) (*tiktokPkg.HandoverTimeSlotsResponse, error) {
	client, err := s.clientFactory(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// First try treating input as package ID
	resp, firstErr := client.GetHandoverTimeSlots(orderOrPackageID)
	if firstErr == nil && resp != nil && len(resp.Data.TimeSlots) > 0 {
		return resp, nil
	}

	// If failed, try treating input as order ID
	resp, secondErr := client.GetHandoverTimeSlotsForOrder(orderOrPackageID, nil)
	if secondErr == nil {
		return resp, nil
	}

	// Both attempts failed — include both errors for diagnosis
	if firstErr != nil {
		return nil, fmt.Errorf("handover time slots failed (as package: %v; as order: %w)", firstErr, secondErr)
	}
	return nil, fmt.Errorf("handover time slots failed: %w", secondErr)
}

// GetOrderDetail retrieves detailed order info including packages
func (s *ShippingService) GetOrderDetail(ctx context.Context, tenantID, orderID string) (*tiktokPkg.OrderDetailResponse, error) {
	client, err := s.clientFactory(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return client.GetOrderDetail([]string{orderID})
}

// ArrangeShipmentByOrder arranges shipment using order ID (resolves to package ID automatically)
func (s *ShippingService) ArrangeShipmentByOrder(ctx context.Context, tenantID, orderID string, req *tiktokPkg.ShipPackageRequest) (*tiktokPkg.ShipPackageResponse, error) {
	client, err := s.clientFactory(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Resolve order ID to package ID
	packageID, status, err := client.ResolveOrderToPackageID(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve order to package: %w", err)
	}

	// Check if package is already shipped
	if status == "IN_TRANSIT" || status == "DELIVERED" || status == "SHIPPED" {
		return nil, fmt.Errorf("package already shipped with status: %s", status)
	}

	return client.ArrangeShipment(packageID, req)
}

// GetShippingLabelByOrder retrieves shipping label using order ID
func (s *ShippingService) GetShippingLabelByOrder(ctx context.Context, tenantID, orderID, documentType string) (string, *tiktokPkg.OrderDetailData, error) {
	client, err := s.clientFactory(ctx, tenantID)
	if err != nil {
		return "", nil, err
	}

	// Get order detail to find package ID
	orderDetail, err := client.GetOrderDetail([]string{orderID})
	if err != nil {
		return "", nil, fmt.Errorf("failed to get order detail: %w", err)
	}

	if len(orderDetail.Data.Orders) == 0 {
		return "", nil, fmt.Errorf("order not found: %s", orderID)
	}

	order := &orderDetail.Data.Orders[0]
	if len(order.Packages) == 0 {
		return "", order, fmt.Errorf("no packages found for order: %s", orderID)
	}

	// Get shipping document for first package
	packageID := order.Packages[0].ID
	docURL, err := client.GetShippingDocument(packageID, documentType)
	if err != nil {
		return "", order, fmt.Errorf("failed to get shipping document: %w", err)
	}

	return docURL, order, nil
}
