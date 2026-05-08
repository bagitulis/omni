package shopee

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/image"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog"
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
	zlog := zerolog.Ctx(ctx)
	allItemIDs := make([]int64, 0)
	offset := 0
	pageSize := 100 // Max allowed by Shopee API.
	var totalCount int64
	firstPage := true

	const maxRetries = 3
	baseDelay := 2 * time.Second

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var listResp *shopeePkg.ProductListResponse
		var lastErr error

		for attempt := 0; attempt < maxRetries; attempt++ {
			listResp, lastErr = s.client.GetProductList(offset, pageSize)
			if lastErr == nil {
				break
			}

			zlog.Warn().Err(lastErr).Int("attempt", attempt+1).Int("offset", offset).Msg("Shopee GetProductList failed, retrying")

			if attempt >= maxRetries-1 {
				break
			}

			delay := baseDelay * time.Duration(1<<attempt)
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
			jitter := time.Duration(rand.Intn(500)) * time.Millisecond
			delay += jitter

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				continue
			}
		}

		if lastErr != nil {
			return nil, fmt.Errorf("shopee GetProductList failed after %d retries at offset %d: %w", maxRetries, offset, lastErr)
		}

		if firstPage {
			totalCount = listResp.Response.TotalCount
			firstPage = false
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

	if totalCount > 0 && int64(len(allItemIDs)) != totalCount {
		zlog.Warn().
			Int64("expected_total", totalCount).
			Int("actual_count", len(allItemIDs)).
			Msg("Shopee product count mismatch after pagination")
	}

	return allItemIDs, nil
}
