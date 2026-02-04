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
	return client.GetShippingDocument(packageID, documentType)
}
