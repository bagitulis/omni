package sync

import (
	"context"
	"fmt"
	"sync"

	"github.com/omni/backend/internal/utils/logger"
)

// OrderSyncOperations handles order synchronization
type OrderSyncOperations struct {
	managers   map[PlatformType]OrderManager
	repository OrderRepository
	tenantID   string
	logger     *logger.Logger
}

// NewOrderSyncOperations creates new sync operations
func NewOrderSyncOperations(
	managers map[PlatformType]OrderManager,
	repository OrderRepository,
	tenantID string,
) *OrderSyncOperations {
	return &OrderSyncOperations{
		managers:   managers,
		repository: repository,
		tenantID:   tenantID,
		logger:     logger.Named("OrderSyncOperations"),
	}
}

// SyncByCategory syncs orders by category for all platforms
func (s *OrderSyncOperations) SyncByCategory(
	ctx context.Context,
	category OrderStatusCategory,
	days int,
	platforms []PlatformType,
) (map[PlatformType]SyncResult, error) {
	s.logger.WithFields(map[string]interface{}{
		"category":  category,
		"days":      days,
		"platforms": platforms,
		"tenant":    s.tenantID,
	}).Info("Starting order sync by category")

	if len(platforms) == 0 {
		platforms = AllPlatforms()
	}

	results := make(map[PlatformType]SyncResult)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, platform := range platforms {
		wg.Add(1)
		go func(p PlatformType) {
			defer wg.Done()
			result := s.syncPlatformOrders(ctx, p, category, days)
			mu.Lock()
			results[p] = result
			mu.Unlock()
		}(platform)
	}

	wg.Wait()
	return results, nil
}

// SyncPlatformOrders syncs orders for a specific platform
func (s *OrderSyncOperations) SyncPlatformOrders(
	ctx context.Context,
	platform PlatformType,
	category OrderStatusCategory,
	days int,
) ([]Order, error) {
	result := s.syncPlatformOrders(ctx, platform, category, days)
	if !result.Success {
		return nil, fmt.Errorf("%s", result.Error)
	}
	return result.Orders, nil
}

func (s *OrderSyncOperations) syncPlatformOrders(
	ctx context.Context,
	platform PlatformType,
	category OrderStatusCategory,
	days int,
) SyncResult {
	result := SyncResult{Platform: platform}

	manager, exists := s.managers[platform]
	if !exists {
		result.Error = fmt.Sprintf("order manager not found: %s", platform)
		return result
	}

	status := GetPlatformStatus(platform, category)
	if status == "" {
		result.Error = fmt.Sprintf("invalid status mapping for %s/%s", platform, category)
		return result
	}

	s.logger.WithFields(map[string]interface{}{
		"platform": platform,
		"status":   status,
		"days":     days,
	}).Info("Fetching orders from platform")

	// Fetch orders
	orders, err := manager.GetOrderList(ctx, status, days)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// Fetch order items if orders exist
	if len(orders) > 0 {
		if err := s.fetchOrderItems(ctx, platform, manager, orders); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"platform": platform,
				"error":    err.Error(),
			}).Warn("Failed to fetch order items, continuing without items")
		}
	}

	// Save to database
	if err := s.saveOrdersToDatabase(ctx, platform, orders, category); err != nil {
		result.Error = err.Error()
		return result
	}

	result.Success = true
	result.Count = len(orders)
	result.Orders = orders

	s.logger.WithFields(map[string]interface{}{
		"platform": platform,
		"count":    len(orders),
	}).Info("Successfully synced orders")

	return result
}

func (s *OrderSyncOperations) fetchOrderItems(
	ctx context.Context,
	platform PlatformType,
	manager OrderManager,
	orders []Order,
) error {
	// Check if orders already have items with images (e.g., Shopee enriched flow)
	// Skip fetch to avoid overwriting enriched data
	ordersWithItems := 0
	ordersWithImages := 0
	for _, o := range orders {
		if len(o.Items) > 0 {
			ordersWithItems++
			for _, item := range o.Items {
				if item.ProductImage != "" {
					ordersWithImages++
					break
				}
			}
		}
	}

	// If most orders already have items with images, skip fetch
	if ordersWithItems > 0 && ordersWithImages > len(orders)/2 {
		s.logger.WithFields(map[string]interface{}{
			"platform":           platform,
			"orders_with_items":  ordersWithItems,
			"orders_with_images": ordersWithImages,
		}).Info("Orders already enriched with items and images, skipping fetch")
		return nil
	}

	orderIDs := make([]string, len(orders))
	for i, o := range orders {
		if o.OrderSN != "" {
			orderIDs[i] = o.OrderSN
		} else {
			orderIDs[i] = o.ID
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"platform":    platform,
		"order_count": len(orderIDs),
	}).Info("Fetching order items from platform")

	// IMPORTANT: Platform APIs have batch limits
	// Shopee: max 50 order IDs per request
	// We'll batch by 50 to be safe for all platforms
	const batchSize = 50
	allItems := make(map[string][]OrderItem)

	for i := 0; i < len(orderIDs); i += batchSize {
		end := i + batchSize
		if end > len(orderIDs) {
			end = len(orderIDs)
		}
		batch := orderIDs[i:end]

		s.logger.WithFields(map[string]interface{}{
			"platform":   platform,
			"batch":      i/batchSize + 1,
			"batch_size": len(batch),
		}).Info("Fetching order items batch")

		items, err := manager.GetOrderItems(ctx, batch)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"platform": platform,
				"batch":    i/batchSize + 1,
				"error":    err.Error(),
			}).Warn("Failed to fetch order items batch, continuing with next batch")
			continue
		}

		// Merge items into allItems
		for orderID, orderItems := range items {
			allItems[orderID] = orderItems
		}
	}

	// Count how many items were found
	totalItems := 0
	matchedOrders := 0

	// Attach items to orders
	for i := range orders {
		orderID := orderIDs[i]
		if orderItems, ok := allItems[orderID]; ok {
			orders[i].Items = orderItems
			totalItems += len(orderItems)
			matchedOrders++
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"platform":       platform,
		"matched_orders": matchedOrders,
		"total_items":    totalItems,
	}).Info("Attached items to orders")

	return nil
}

func (s *OrderSyncOperations) saveOrdersToDatabase(
	ctx context.Context,
	platform PlatformType,
	orders []Order,
	category OrderStatusCategory,
) error {
	// IMPORTANT: Get status for this category to clear old orders first
	status := GetPlatformStatus(platform, category)

	// Clear old orders with this status BEFORE saving new ones
	// This ensures we don't have stale data (same as Node.js implementation)
	if err := s.repository.ClearOrdersByStatus(ctx, platform, status); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"platform": platform,
			"status":   status,
			"error":    err.Error(),
		}).Warn("Failed to clear old orders, continuing with save")
	} else {
		s.logger.WithFields(map[string]interface{}{
			"platform": platform,
			"status":   status,
		}).Info("Cleared old orders before saving new ones")
	}

	// If no new orders, just log and return (old orders were already cleared)
	if len(orders) == 0 {
		s.logger.WithFields(map[string]interface{}{
			"platform": platform,
			"category": category,
		}).Info("No new orders to save")
		return nil
	}

	// Set category and platform on orders
	for i := range orders {
		orders[i].Category = string(category)
		orders[i].Platform = string(platform)
		orders[i].Status = status // Ensure status is set correctly
	}

	if err := s.repository.SaveOrders(ctx, platform, orders); err != nil {
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"platform": platform,
		"category": category,
		"count":    len(orders),
	}).Info("Saved orders to database")

	return nil
}
