package lazada

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/repositories"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// SyncProductsWithDetails fetches all products from Lazada API, saves to DB, and returns flattened product rows.
func (s *SyncService) SyncProductsWithDetails(ctx context.Context, offset, limit int) ([]map[string]interface{}, int, error) {
	zlog := zerolog.Ctx(ctx)
	zlog.Info().Int("offset", offset).Int("limit", limit).Msg("Starting Lazada product sync with pagination")

	pageOffset := 0
	pageLimit := 50
	totalProducts := 0
	allRows := make([]map[string]interface{}, 0)
	allProducts := make([]lazadaPkg.Product, 0)
	consecutiveEmpty := 0
	const maxConsecutiveEmpty = 3

	for {
		zlog.Info().Int("page_offset", pageOffset).Int("page_limit", pageLimit).Msg("Fetching Lazada product page")

		resp, err := s.client.GetProductsWithContext(ctx, pageOffset, pageLimit)
		if err != nil {
			zlog.Error().Err(err).Int("page_offset", pageOffset).Msg("Failed to fetch Lazada products page")
			return nil, 0, err
		}

		totalProducts = resp.Data.TotalProducts
		products := resp.Data.Products
		zlog.Info().Int("fetched_products", len(products)).Int("total_products", totalProducts).Msg("Fetched Lazada products page")

		if len(products) == 0 {
			if pageOffset >= totalProducts || totalProducts == 0 {
				break
			}
			// Empty page but haven't fetched all — transient issue
			consecutiveEmpty++
			if consecutiveEmpty >= maxConsecutiveEmpty {
				zlog.Warn().Int("page_offset", pageOffset).Int("total_products", totalProducts).Msg("Breaking after max consecutive empty pages")
				break
			}
			zlog.Warn().Int("page_offset", pageOffset).Int("total_products", totalProducts).Int("consecutive_empty", consecutiveEmpty).Msg("Empty page received before all products fetched, continuing")
			pageOffset += pageLimit
			continue
		}

		consecutiveEmpty = 0
		for _, product := range products {
			allProducts = append(allProducts, product)
			allRows = append(allRows, s.buildProductRows(product)...)
		}

		pageOffset += pageLimit
		if pageOffset >= totalProducts {
			break
		}
	}

	if totalProducts > 0 && len(allProducts) < totalProducts {
		zlog.Warn().Int("expected", totalProducts).Int("fetched", len(allProducts)).Msg("Lazada sync fetched fewer products than API reported")
	}

	if len(allProducts) == 0 {
		zlog.Warn().Msg("Lazada API returned 0 products - skipping sync to avoid data loss")
		return []map[string]interface{}{}, 0, nil
	}

	savedProductCount := 0
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.clearLazadaProductCacheWithDB(ctx, tx); err != nil {
			return err
		}

		txProdRepo := repositories.NewLazadaProductRepository(tx)
		for _, product := range allProducts {
			if err := s.persistProductWithSkus(ctx, tx, txProdRepo, product); err != nil {
				return err
			}
			savedProductCount++
		}

		return nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("sync lazada products with details: %w", err)
	}

	zlog.Info().Int("saved_products", savedProductCount).Int("total_rows", len(allRows)).Msg("Completed Lazada product sync with details")
	return allRows, savedProductCount, nil
}
