package tiktok

import (
	"context"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/image"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// SyncService handles syncing data from TikTok API
type SyncService struct {
	client    *tiktokPkg.Client
	db        *gorm.DB
	orderRepo *repositories.TiktokOrderRepository
	prodRepo  *repositories.TiktokProductRepository
	skuRepo   *repositories.TiktokSkuRepository
	imgMgr    image.Manager
	tenantID  string
}

// NewSyncService creates a new sync service
func NewSyncService(client *tiktokPkg.Client, db *gorm.DB) *SyncService {
	return &SyncService{
		client:    client,
		db:        db,
		orderRepo: repositories.NewTiktokOrderRepository(db),
		prodRepo:  repositories.NewTiktokProductRepository(db),
		skuRepo:   repositories.NewTiktokSkuRepository(db),
		imgMgr:    image.NewManager(db, ""),
	}
}

// NewSyncServiceWithTenant creates a new sync service with tenant ID
func NewSyncServiceWithTenant(client *tiktokPkg.Client, db *gorm.DB, tenantID string) *SyncService {
	return &SyncService{
		client:    client,
		db:        db,
		orderRepo: repositories.NewTiktokOrderRepository(db),
		prodRepo:  repositories.NewTiktokProductRepository(db),
		skuRepo:   repositories.NewTiktokSkuRepository(db),
		imgMgr:    image.NewManager(db, ""),
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
		// Get ship by date - prefer rts_sla_time, fallback to shipping_due_time
		var shipByDate *int64
		if order.RtsSlaTime > 0 {
			shipByDate = &order.RtsSlaTime
		} else if order.ShippingDueTime > 0 {
			shipByDate = &order.ShippingDueTime
		}

		dbOrder := &models.TiktokOrder{
			TenantID:        s.tenantID,
			OrderSN:         order.OrderID, // API returns OrderID, we store as OrderSN
			OrderStatus:     order.OrderStatus,
			ShippingCarrier: order.ShippingProvider,
			TrackingNumber:  order.TrackingNumber,
			BuyerUsername:   order.BuyerEmail,
			BuyerMessage:    order.BuyerMessage,
			ShipByDate:      shipByDate,
		}

		if err := s.orderRepo.Upsert(ctx, dbOrder); err == nil {
			count++
		}
	}

	return count, nil
}

// SyncProducts fetches products from TikTok API and saves to database
func (s *SyncService) SyncProducts(ctx context.Context) (int, error) {
	zlog := zerolog.Ctx(ctx)

	products, totalCount, err := s.fetchAllProducts(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to search products from TikTok API")
		return 0, err
	}

	zlog.Info().Int("total_products", len(products)).Int("total_count", totalCount).Msg("Fetched products from TikTok API")

	if len(products) == 0 {
		zlog.Warn().Msg("TikTok API returned 0 products - skipping sync to prevent data loss")
		return 0, nil
	}

	count := 0
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.clearTiktokProductCacheWithDB(ctx, tx); err != nil {
			return err
		}

		txProdRepo := repositories.NewTiktokProductRepository(tx)
		return s.syncProductsWithDB(ctx, tx, txProdRepo, products, &count)
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}
