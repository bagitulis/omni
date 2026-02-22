package shopee

import (
	"context"

	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/image"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"gorm.io/gorm"
)

// ProductSyncService handles syncing products from Shopee API to database.
type ProductSyncService struct {
	client   *shopeePkg.Client
	db       *gorm.DB
	prodRepo *repositories.ShopeeProductRepository
	skuRepo  *repositories.ShopeeSkuRepository
	imgMgr   image.Manager
	tenantID string
}

// NewProductSyncService creates a new product sync service with tenant ID.
func NewProductSyncService(client *shopeePkg.Client, db *gorm.DB, tenantID string) *ProductSyncService {
	return &ProductSyncService{
		client:   client,
		db:       db,
		prodRepo: repositories.NewShopeeProductRepository(db),
		skuRepo:  repositories.NewShopeeSkuRepository(db),
		imgMgr:   image.NewManager(db, ""),
		tenantID: tenantID,
	}
}

// SyncProducts fetches products from Shopee API and saves to database.
// Implements pagination using has_next_page and next_offset from API.
func (s *ProductSyncService) SyncProducts(ctx context.Context) (int, error) {
	allItemIDs, err := s.fetchAllItemIDs(ctx)
	if err != nil {
		return 0, err
	}

	return s.syncProductsByItemIDs(ctx, allItemIDs)
}

// SyncProductsByIDs syncs specific Shopee products by item IDs.
// Skips cache clear — only fetches and upserts the given items.
func (s *ProductSyncService) SyncProductsByIDs(ctx context.Context, itemIDs []int64) (int, error) {
	if len(itemIDs) == 0 {
		return 0, nil
	}

	count := 0
	txProdRepo := repositories.NewShopeeProductRepository(s.db)
	txSkuRepo := repositories.NewShopeeSkuRepository(s.db)
	err := s.syncItemDetailBatches(ctx, s.db, txProdRepo, txSkuRepo, itemIDs, &count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *ProductSyncService) fetchAllItemIDs(ctx context.Context) ([]int64, error) {
	allItemIDs := make([]int64, 0)
	offset := 0
	pageSize := 100 // Max allowed by Shopee API.

	for {
		listResp, err := s.client.GetProductList(offset, pageSize)
		if err != nil {
			return nil, err
		}

		if len(listResp.Response.Item) == 0 {
			break
		}

		for _, item := range listResp.Response.Item {
			allItemIDs = append(allItemIDs, item.ItemID)
		}

		if !listResp.Response.HasNextPage {
			break
		}

		offset = listResp.Response.NextOffset
	}

	return allItemIDs, nil
}
