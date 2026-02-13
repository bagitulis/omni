package sync

import (
	"context"
	"fmt"
	"sync"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/services/platform"
	"github.com/omni/backend/internal/utils/logger"
)

// OrderSyncService coordinates order sync operations
// SRP: Facade for order sync, delegates to specialized services
type OrderSyncService struct {
	managers       map[PlatformType]OrderManager
	repository     OrderRepository
	syncOperations *OrderSyncOperations
	queryService   *OrderQueryService
	tenantID       string
	logger         *logger.Logger
	initialized    bool
	mu             sync.RWMutex
}

// NewOrderSyncService creates a new order sync service
func NewOrderSyncService(tenantID string) *OrderSyncService {
	return &OrderSyncService{
		managers: make(map[PlatformType]OrderManager),
		tenantID: tenantID,
		logger:   logger.Named("OrderSyncService"),
	}
}

// Initialize initializes the service with platform managers
func (s *OrderSyncService) Initialize(
	managers map[PlatformType]OrderManager,
	repository OrderRepository,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return nil
	}

	s.logger.WithTenantID(s.tenantID).Info("Initializing OrderSyncService")

	s.managers = managers
	s.repository = repository
	s.syncOperations = NewOrderSyncOperations(managers, repository, s.tenantID)
	s.queryService = NewOrderQueryService(repository)
	s.initialized = true

	s.logger.WithTenantID(s.tenantID).Info("OrderSyncService initialized successfully")
	return nil
}

// IsInitialized returns whether the service is initialized
func (s *OrderSyncService) IsInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.initialized
}

// GetTenantID returns the tenant ID
func (s *OrderSyncService) GetTenantID() string {
	return s.tenantID
}

// SyncByCategory syncs orders by category for all platforms
func (s *OrderSyncService) SyncByCategory(
	ctx context.Context,
	category OrderStatusCategory,
	days int,
	platforms []PlatformType,
) (map[PlatformType]SyncResult, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("service not initialized")
	}
	return s.syncOperations.SyncByCategory(ctx, category, days, platforms)
}

// SyncPlatformOrders syncs orders for a specific platform
func (s *OrderSyncService) SyncPlatformOrders(
	ctx context.Context,
	platform PlatformType,
	category OrderStatusCategory,
	days int,
) ([]Order, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("service not initialized")
	}
	return s.syncOperations.SyncPlatformOrders(ctx, platform, category, days)
}

// GetOrdersByCategory gets orders by category
func (s *OrderSyncService) GetOrdersByCategory(
	ctx context.Context,
	category OrderStatusCategory,
	platform *PlatformType,
) ([]Order, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("service not initialized")
	}
	return s.queryService.GetOrdersByCategory(ctx, category, platform)
}

// GetPlatformOrders gets orders for a specific platform
func (s *OrderSyncService) GetPlatformOrders(
	ctx context.Context,
	platform PlatformType,
	category *OrderStatusCategory,
) ([]Order, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("service not initialized")
	}
	return s.queryService.GetPlatformOrders(ctx, platform, category)
}

// GetOrderDetails gets detailed order information
func (s *OrderSyncService) GetOrderDetails(
	ctx context.Context,
	platform PlatformType,
	orderIDs []string,
) ([]Order, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("service not initialized")
	}

	manager, exists := s.managers[platform]
	if !exists {
		return nil, fmt.Errorf("order manager not found: %s", platform)
	}

	s.logger.WithFields(map[string]interface{}{
		"platform": platform,
		"count":    len(orderIDs),
	}).Info("Fetching order details")

	return manager.GetOrderDetails(ctx, orderIDs)
}

// OrderQueryService handles order queries
type OrderQueryService struct {
	repository OrderRepository
}

// NewOrderQueryService creates a new query service
func NewOrderQueryService(repository OrderRepository) *OrderQueryService {
	return &OrderQueryService{repository: repository}
}

// GetOrdersByCategory gets orders by category
func (q *OrderQueryService) GetOrdersByCategory(
	ctx context.Context,
	category OrderStatusCategory,
	platform *PlatformType,
) ([]Order, error) {
	var allOrders []Order

	platforms := AllPlatforms()
	if platform != nil {
		platforms = []PlatformType{*platform}
	}

	for _, p := range platforms {
		status := GetPlatformStatus(p, category)
		if status == "" {
			continue
		}
		orders, err := q.repository.GetOrdersByStatus(ctx, p, status, 0)
		if err != nil {
			continue
		}
		allOrders = append(allOrders, orders...)
	}

	return allOrders, nil
}

// GetPlatformOrders gets orders for a specific platform
func (q *OrderQueryService) GetPlatformOrders(
	ctx context.Context,
	platform PlatformType,
	category *OrderStatusCategory,
) ([]Order, error) {
	status := ""
	if category != nil {
		status = GetPlatformStatus(platform, *category)
	}
	return q.repository.GetOrdersByStatus(ctx, platform, status, 0)
}

// GetOrderSyncService creates a NEW order sync service for a tenant each time
// NO CACHING - always creates fresh instance to ensure latest tokens are used
func GetOrderSyncService(tenantID string) (*OrderSyncService, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("missing tenant ID")
	}

	// Always create new instance - no caching
	instance := NewOrderSyncService(tenantID)

	// Initialize with platforms
	if err := initializeServiceWithPlatforms(instance, tenantID); err != nil {
		// Log warning but don't fail - service can work without platforms
		instance.logger.WithTenantID(tenantID).Warn("Failed to initialize platforms: " + err.Error())
	}

	return instance, nil
}

// initializeServiceWithPlatforms initializes the service with platform managers
func initializeServiceWithPlatforms(service *OrderSyncService, tenantID string) error {
	ctx := context.Background()

	// Get platform coordination service
	coordService := platform.GetPlatformCoordinationService(tenantID)
	if err := coordService.InitializePlatforms(ctx); err != nil {
		return fmt.Errorf("initialize platforms: %w", err)
	}

	// Get tenant DB for ImageService
	db, dbErr := config.GetTenantDBByID(tenantID)

	// Create order managers from platform API clients
	managers := make(map[PlatformType]OrderManager)

	if shopeeClient := coordService.GetShopeeClient(); shopeeClient != nil {
		// Create ImageService if DB is available
		var imageService *ImageService
		if dbErr == nil && db != nil && shopeeClient.GetClient() != nil {
			imageService = NewImageService(shopeeClient.GetClient(), db, tenantID)
		}
		managers[PlatformShopee] = NewShopeeOrderManager(shopeeClient, tenantID, imageService)
	}

	if lazadaClient := coordService.GetLazadaClient(); lazadaClient != nil {
		managers[PlatformLazada] = NewLazadaOrderManager(lazadaClient, tenantID)
	}

	if tiktokClient := coordService.GetTiktokClient(); tiktokClient != nil {
		managers[PlatformTiktok] = NewTiktokOrderManager(tiktokClient, tenantID)
	}

	// Create repository
	repository := NewGormOrderRepository(tenantID)

	// Initialize service
	return service.Initialize(managers, repository)
}

// ClearOrderSyncInstances is a no-op since we don't cache anymore
// Kept for API compatibility
func ClearOrderSyncInstances() {
	// No-op - no cache to clear
}

// InvalidateTenantInstance is a no-op since we don't cache anymore
// Kept for API compatibility
func InvalidateTenantInstance(tenantID string) {
	// No-op - no cache to invalidate
}
