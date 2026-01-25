package shopee

import (
	"github.com/omni/backend/internal/repositories"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"gorm.io/gorm"
)

// SyncService handles syncing data from Shopee API to database
// This is a facade that combines OrderSyncService and ProductSyncService
type SyncService struct {
	*OrderSyncService
	*ProductSyncService
}

// NewSyncServiceWithTenant creates a new sync service with tenant ID
func NewSyncServiceWithTenant(client *shopeePkg.Client, db *gorm.DB, tenantID string) *SyncService {
	return &SyncService{
		OrderSyncService:   NewOrderSyncService(client, db, tenantID),
		ProductSyncService: NewProductSyncService(client, db, tenantID),
	}
}

// NewSyncServiceForProducts creates a sync service for product operations only
func NewSyncServiceForProducts(client *shopeePkg.Client, db *gorm.DB, tenantID string) *ProductSyncService {
	return &ProductSyncService{
		client:   client,
		db:       db,
		prodRepo: repositories.NewShopeeProductRepository(db),
		skuRepo:  repositories.NewShopeeSkuRepository(db),
		tenantID: tenantID,
	}
}
