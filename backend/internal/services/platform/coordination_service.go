package platform

import (
	"context"
	"fmt"
	"sync"

	"github.com/omni/backend/internal/utils/logger"
)

// PlatformType represents supported e-commerce platforms
type PlatformType string

const (
	PlatformShopee PlatformType = "shopee"
	PlatformLazada PlatformType = "lazada"
	PlatformTiktok PlatformType = "tiktok"
)

// ConfigManager interface for platform config management
type ConfigManager interface {
	LoadConfig(ctx context.Context) error
	GetConfig(key string) (string, bool)
	SetConfig(key, value string) error
	GetAccessToken() string
	GetRefreshToken() string
	SetTokens(accessToken, refreshToken string) error
	GetPlatform() PlatformType
	GetTenantID() string
}

// APIClient interface for platform API operations
type APIClient interface {
	// Order operations
	GetOrderList(ctx context.Context, status string, days int) ([]map[string]interface{}, error)
	GetOrderDetails(ctx context.Context, orderIDs []string) ([]map[string]interface{}, error)

	// Product operations (optional)
	GetProductList(ctx context.Context, offset, limit int) ([]map[string]interface{}, error)

	// Check if client is initialized
	IsInitialized() bool
}

// PlatformCoordinationService manages all platform clients and configs
// Multi-tenant: Each tenant has its own instance
type PlatformCoordinationService struct {
	platformClients map[PlatformType]APIClient
	configManagers  map[PlatformType]ConfigManager
	tenantID        string
	logger          *logger.Logger
	initialized     bool
	mu              sync.RWMutex
}

// NewPlatformCoordinationService creates a new service for a tenant
func NewPlatformCoordinationService(tenantID string) *PlatformCoordinationService {
	return &PlatformCoordinationService{
		platformClients: make(map[PlatformType]APIClient),
		configManagers:  make(map[PlatformType]ConfigManager),
		tenantID:        tenantID,
		logger:          logger.Named("PlatformCoord"),
	}
}

// InitializePlatforms initializes all platform clients for this tenant
func (s *PlatformCoordinationService) InitializePlatforms(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return nil
	}

	s.logger.WithTenantID(s.tenantID).Info("Initializing all platform clients...")

	// Initialize Shopee
	if err := s.initShopee(ctx); err != nil {
		s.logger.WithTenantID(s.tenantID).Warn("Failed to init Shopee: " + err.Error())
	}

	// Initialize Lazada
	if err := s.initLazada(ctx); err != nil {
		s.logger.WithTenantID(s.tenantID).Warn("Failed to init Lazada: " + err.Error())
	}

	// Initialize TikTok
	if err := s.initTiktok(ctx); err != nil {
		s.logger.WithTenantID(s.tenantID).Warn("Failed to init TikTok: " + err.Error())
	}

	s.initialized = true
	s.logger.WithTenantID(s.tenantID).Info("Platform clients initialized")
	return nil
}

// initShopee initializes Shopee client
func (s *PlatformCoordinationService) initShopee(ctx context.Context) error {
	config := NewShopeeConfigManager(s.tenantID)
	if err := config.LoadConfig(ctx); err != nil {
		return fmt.Errorf("load shopee config: %w", err)
	}

	// Setup token refresher if available
	if refreshSvc := GetTokenRefreshService(); refreshSvc != nil {
		config.SetTokenRefresher(func(ctx context.Context, tenantID string) (string, string, error) {
			return refreshSvc.RefreshShopeeToken(ctx, tenantID)
		})
	}

	s.configManagers[PlatformShopee] = config

	client := NewShopeeAPIClient(config)
	s.platformClients[PlatformShopee] = client
	return nil
}

// initLazada initializes Lazada client
func (s *PlatformCoordinationService) initLazada(ctx context.Context) error {
	config := NewLazadaConfigManager(s.tenantID)
	if err := config.LoadConfig(ctx); err != nil {
		return fmt.Errorf("load lazada config: %w", err)
	}

	// Setup token refresher if available
	if refreshSvc := GetTokenRefreshService(); refreshSvc != nil {
		config.SetTokenRefresher(func(ctx context.Context, tenantID string) (string, string, error) {
			return refreshSvc.RefreshLazadaToken(ctx, tenantID)
		})
	}

	s.configManagers[PlatformLazada] = config

	client := NewLazadaAPIClient(config)
	s.platformClients[PlatformLazada] = client
	return nil
}

// initTiktok initializes TikTok client
func (s *PlatformCoordinationService) initTiktok(ctx context.Context) error {
	config := NewTiktokConfigManager(s.tenantID)
	if err := config.LoadConfig(ctx); err != nil {
		return fmt.Errorf("load tiktok config: %w", err)
	}

	// Setup token refresher if available
	if refreshSvc := GetTokenRefreshService(); refreshSvc != nil {
		config.SetTokenRefresher(func(ctx context.Context, tenantID string) (string, string, error) {
			return refreshSvc.RefreshTiktokToken(ctx, tenantID)
		})
	}

	s.configManagers[PlatformTiktok] = config

	client := NewTiktokAPIClient(config)
	s.platformClients[PlatformTiktok] = client
	return nil
}

// GetClient returns platform API client
func (s *PlatformCoordinationService) GetClient(platform PlatformType) (APIClient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.platformClients[platform]
	if !ok {
		return nil, fmt.Errorf("platform client not initialized: %s", platform)
	}
	return client, nil
}

// GetConfigManager returns platform config manager
func (s *PlatformCoordinationService) GetConfigManager(platform PlatformType) (ConfigManager, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, ok := s.configManagers[platform]
	if !ok {
		return nil, fmt.Errorf("config manager not found: %s", platform)
	}
	return config, nil
}

// GetAllClients returns all initialized clients
func (s *PlatformCoordinationService) GetAllClients() map[PlatformType]APIClient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[PlatformType]APIClient)
	for k, v := range s.platformClients {
		result[k] = v
	}
	return result
}

// IsInitialized returns whether service is initialized
func (s *PlatformCoordinationService) IsInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.initialized
}

// GetTenantID returns the tenant ID
func (s *PlatformCoordinationService) GetTenantID() string {
	return s.tenantID
}

// GetShopeeClient returns the Shopee API client
func (s *PlatformCoordinationService) GetShopeeClient() *ShopeeAPIClient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if client, ok := s.platformClients[PlatformShopee]; ok {
		if typed, ok := client.(*ShopeeAPIClient); ok {
			return typed
		}
	}
	return nil
}

// GetLazadaClient returns the Lazada API client
func (s *PlatformCoordinationService) GetLazadaClient() *LazadaAPIClient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if client, ok := s.platformClients[PlatformLazada]; ok {
		if typed, ok := client.(*LazadaAPIClient); ok {
			return typed
		}
	}
	return nil
}

// GetTiktokClient returns the TikTok API client
func (s *PlatformCoordinationService) GetTiktokClient() *TiktokAPIClient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if client, ok := s.platformClients[PlatformTiktok]; ok {
		if typed, ok := client.(*TiktokAPIClient); ok {
			return typed
		}
	}
	return nil
}
