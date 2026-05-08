package lazada

import (
	"context"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/image"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// SyncService handles syncing data from Lazada API.
type SyncService struct {
	client    *lazadaPkg.Client
	db        *gorm.DB
	orderRepo *repositories.LazadaOrderRepository
	prodRepo  *repositories.LazadaProductRepository
	imgMgr    image.Manager
	tenantID  string
}

// NewSyncService creates a new sync service.
func NewSyncService(client *lazadaPkg.Client, db *gorm.DB) *SyncService {
	return &SyncService{
		client:    client,
		db:        db,
		orderRepo: repositories.NewLazadaOrderRepository(db),
		prodRepo:  repositories.NewLazadaProductRepository(db),
		imgMgr:    image.NewManager(db, ""),
	}
}

// NewSyncServiceWithTenant creates a new sync service with tenant ID.
func NewSyncServiceWithTenant(client *lazadaPkg.Client, db *gorm.DB, tenantID string) *SyncService {
	return &SyncService{
		client:    client,
		db:        db,
		orderRepo: repositories.NewLazadaOrderRepository(db),
		prodRepo:  repositories.NewLazadaProductRepository(db),
		imgMgr:    image.NewManager(db, ""),
		tenantID:  tenantID,
	}
}

// SyncOrders fetches orders from Lazada API and saves to database.
func (s *SyncService) SyncOrders(ctx context.Context, status string) (int, error) {
	resp, err := s.client.GetOrders(status, 0, 100)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, order := range resp.Data.Orders {
		var shipByDate *int64
		if order.PromisedShipDate != "" {
			for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
				if parsedTime, parseErr := time.Parse(layout, order.PromisedShipDate); parseErr == nil {
					ts := parsedTime.Unix()
					shipByDate = &ts
					break
				}
			}
		}

		dbOrder := &models.LazadaOrder{
			TenantID:        s.tenantID,
			OrderSN:         order.OrderID,
			OrderStatus:     order.Status,
			BuyerUsername:   order.CustomerName,
			ShippingCarrier: order.ShippingType,
			ShipByDate:      shipByDate,
		}

		if err := s.orderRepo.Upsert(ctx, dbOrder); err == nil {
			count++
		} else {
			zlog := zerolog.Ctx(ctx)
			zlog.Error().Err(err).Str("order_sn", order.OrderID).Msg("Failed to upsert order")
		}
	}

	return count, nil
}

// SyncProducts fetches products from Lazada API and saves to database.
func (s *SyncService) SyncProducts(ctx context.Context) (int, error) {
	zlog := zerolog.Ctx(ctx)

	// Fetch ALL products with pagination (API returns max 50 per page)
	const pageSize = 50
	var products []lazadaPkg.Product
	var totalProducts int
	for offset := 0; ; offset += pageSize {
		resp, err := s.client.GetProductsWithContext(ctx, offset, pageSize)
		if err != nil {
			return 0, fmt.Errorf("lazada fetch products offset=%d: %w", offset, err)
		}
		totalProducts = resp.Data.TotalProducts
		products = append(products, resp.Data.Products...)
		zlog.Info().
			Int("fetched", len(resp.Data.Products)).
			Int("total_so_far", len(products)).
			Int("total_api", totalProducts).
			Msg("Lazada sync fetched page")

		if len(resp.Data.Products) < pageSize || len(products) >= totalProducts {
			break
		}
	}

	if totalProducts > 0 && len(products) < totalProducts {
		zlog.Warn().Int("expected", totalProducts).Int("fetched", len(products)).Msg("Lazada sync fetched fewer products than API reported")
	}

	zlog.Info().Int("count", len(products)).Msg("Lazada sync fetched all products")

	if len(products) == 0 {
		zlog.Warn().Msg("Lazada API returned 0 products - skipping sync to avoid data loss")
		return 0, nil
	}

	count := 0
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.clearLazadaProductCacheWithDB(ctx, tx); err != nil {
			return err
		}

		txProdRepo := repositories.NewLazadaProductRepository(tx)
		for _, product := range products {
			if err := s.persistProductWithSkus(ctx, tx, txProdRepo, product); err != nil {
				return err
			}
			count++
		}

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("sync lazada products: %w", err)
	}

	return count, nil
}

// SyncProductsByIDs syncs specific Lazada products by item IDs.
// Skips cache clear — only fetches detail and upserts the given items.
func (s *SyncService) SyncProductsByIDs(ctx context.Context, itemIDs []int64) (int, error) {
	zlog := zerolog.Ctx(ctx)
	if len(itemIDs) == 0 {
		zlog.Info().Msg("[Lazada SyncProductsByIDs] No item IDs provided")
		return 0, nil
	}
	zlog.Info().Int("count", len(itemIDs)).Msg("[Lazada SyncProductsByIDs] Starting sync")

	count := 0
	txProdRepo := repositories.NewLazadaProductRepository(s.db)
	for _, itemID := range itemIDs {
		zlog.Info().Int64("item_id", itemID).Msg("[Lazada SyncProductsByIDs] Fetching product")
		product, err := s.client.GetProductItem(ctx, itemID)
		if err != nil {
			zlog.Error().Err(err).Int64("item_id", itemID).Msg("[Lazada SyncProductsByIDs] API call FAILED")
			// Return API/auth errors so they appear in sync response
			return count, fmt.Errorf("item %d: %w", itemID, err)
		}
		zlog.Info().Str("name", product.Name).Int("skus", len(product.Skus)).Msg("[Lazada SyncProductsByIDs] Fetched")
		for i, sku := range product.Skus {
			zlog.Debug().Int("index", i).Str("seller_sku", sku.SellerSku).Int("qty", sku.Quantity).Int("avail", sku.Available).Float64("price", sku.Price).Msg("[Lazada SyncProductsByIDs] SKU")
		}
		if err := s.persistProductWithSkus(ctx, s.db, txProdRepo, *product); err != nil {
			zlog.Error().Err(err).Int64("item_id", itemID).Msg("[Lazada SyncProductsByIDs] Upsert FAILED")
			continue
		}
		zlog.Info().Int64("item_id", itemID).Msg("[Lazada SyncProductsByIDs] ✅ Persisted")
		count++
	}
	zlog.Info().Int("synced", count).Int("total", len(itemIDs)).Msg("[Lazada SyncProductsByIDs] Done")
	return count, nil
}
