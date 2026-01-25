package tiktok

import (
	"context"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"gorm.io/gorm"
)

// SyncService handles syncing data from TikTok API
type SyncService struct {
	client    *tiktokPkg.Client
	orderRepo *repositories.TiktokOrderRepository
	prodRepo  *repositories.TiktokProductRepository
	skuRepo   *repositories.TiktokSkuRepository
	tenantID  string
}

// NewSyncService creates a new sync service
func NewSyncService(client *tiktokPkg.Client, db *gorm.DB) *SyncService {
	return &SyncService{
		client:    client,
		orderRepo: repositories.NewTiktokOrderRepository(db),
		prodRepo:  repositories.NewTiktokProductRepository(db),
		skuRepo:   repositories.NewTiktokSkuRepository(db),
	}
}

// NewSyncServiceWithTenant creates a new sync service with tenant ID
func NewSyncServiceWithTenant(client *tiktokPkg.Client, db *gorm.DB, tenantID string) *SyncService {
	return &SyncService{
		client:    client,
		orderRepo: repositories.NewTiktokOrderRepository(db),
		prodRepo:  repositories.NewTiktokProductRepository(db),
		skuRepo:   repositories.NewTiktokSkuRepository(db),
		tenantID:  tenantID,
	}
}

// SyncOrders fetches orders from TikTok API and saves to database
func (s *SyncService) SyncOrders(ctx context.Context, status string) (int, error) {
	resp, err := s.client.GetOrders(status, 100)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, order := range resp.Data.OrderList {
		dbOrder := &models.TiktokOrder{
			OrderSN:     order.OrderID, // API returns OrderID, we store as OrderSN
			OrderStatus: order.OrderStatus,
		}

		if err := s.orderRepo.Upsert(ctx, dbOrder); err == nil {
			count++
		}
	}

	return count, nil
}

// SyncProducts fetches products from TikTok API and saves to database
func (s *SyncService) SyncProducts(ctx context.Context) (int, error) {
	resp, err := s.client.GetProducts(100)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, prod := range resp.Data.Products {
		dbProd := &models.TiktokProduct{
			ProductID:   prod.ProductID,
			Name:        prod.Name,
			Description: prod.Description,
			Price:       prod.Price,
			Status:      prod.Status,
		}

		if err := s.prodRepo.Upsert(ctx, dbProd); err == nil {
			count++
		}
	}

	return count, nil
}
