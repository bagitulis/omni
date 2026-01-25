package sku

import (
	"context"
	"sync"

	"gorm.io/gorm"
)

// PlatformSKUChecker defines interface for checking SKU on platforms
type PlatformSKUChecker interface {
	CheckSKU(ctx context.Context, sku string) (*SKUStatus, error)
}

// SKUStatus represents SKU status on a platform
type SKUStatus struct {
	Exists    bool   `json:"exists"`
	ItemID    int64  `json:"item_id,omitempty"`
	ProductID string `json:"product_id,omitempty"`
	Status    string `json:"status,omitempty"` // active, deleted, etc
	Error     string `json:"error,omitempty"`
}

// PlatformSKUStatus represents SKU status across platforms
type PlatformSKUStatus struct {
	SKU     string               `json:"sku"`
	Shopee  *SKUStatus           `json:"shopee,omitempty"`
	Lazada  *SKUStatus           `json:"lazada,omitempty"`
	Tiktok  *SKUStatus           `json:"tiktok,omitempty"`
}

// CheckService handles SKU checking operations
// NO CACHING - always queries fresh data
type CheckService struct {
	db       *gorm.DB
	tenantID string
}

// NewCheckService creates a new SKU check service
func NewCheckService(db *gorm.DB, tenantID string) *CheckService {
	return &CheckService{
		db:       db,
		tenantID: tenantID,
	}
}

// CheckSingle checks a single SKU across platforms
// NO CACHING - always queries fresh data from APIs
func (s *CheckService) CheckSingle(ctx context.Context, sku string, apis map[string]PlatformSKUChecker) *PlatformSKUStatus {

	result := &PlatformSKUStatus{SKU: sku}
	var wg sync.WaitGroup
	var mu sync.Mutex

	for platform, api := range apis {
		wg.Add(1)
		go func(p string, a PlatformSKUChecker) {
			defer wg.Done()
			status, _ := a.CheckSKU(ctx, sku)
			mu.Lock()
			switch p {
			case "shopee":
				result.Shopee = status
			case "lazada":
				result.Lazada = status
			case "tiktok":
				result.Tiktok = status
			}
			mu.Unlock()
		}(platform, api)
	}

	wg.Wait()

	return result
}

// CheckBatch checks multiple SKUs
func (s *CheckService) CheckBatch(ctx context.Context, skus []string, apis map[string]PlatformSKUChecker) []PlatformSKUStatus {
	results := make([]PlatformSKUStatus, len(skus))
	var wg sync.WaitGroup

	// Use semaphore to limit concurrency
	sem := make(chan struct{}, 5)

	for i, sku := range skus {
		wg.Add(1)
		go func(idx int, s string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			
			// Placeholder for parallel execution
			_ = idx
			_ = s
		}(i, sku)
	}

	// Simple sequential implementation
	for i, sku := range skus {
		results[i] = *s.CheckSingle(ctx, sku, apis)
	}

	return results
}

// GetCachedStatus is a no-op - no caching
// Kept for API compatibility, returns nil
func (s *CheckService) GetCachedStatus(sku string) *PlatformSKUStatus {
	return nil // No caching
}

// InvalidateCache is a no-op - no caching
// Kept for API compatibility
func (s *CheckService) InvalidateCache(sku string) {
	// No-op - no cache to invalidate
}

// ClearCache is a no-op - no caching
// Kept for API compatibility
func (s *CheckService) ClearCache() {
	// No-op - no cache to clear
}
