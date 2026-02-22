// Package master_product provides import service for Master Product
package master_product

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"gorm.io/gorm"
)

// StagingImportResult holds counters and errors from a staging-based import run.
type StagingImportResult struct {
	ProductsCreated int      `json:"products_created"`
	ProductsMatched int      `json:"products_matched"`
	ProductsSkipped int      `json:"products_skipped"`
	SkusCreated     int      `json:"skus_created"`
	SkusSkipped     int      `json:"skus_skipped"`
	LinksCreated    int      `json:"links_created"`
	Errors          []string `json:"errors"`
}

// NewStagingImportResult creates a result with initialized (non-nil) errors slice.
// This prevents JSON null serialization which crashes the frontend.
func NewStagingImportResult() StagingImportResult {
	return StagingImportResult{Errors: []string{}}
}

// StagingImportService processes staging rows and maps them to master products.
type StagingImportService struct {
	db   *gorm.DB
	repo *repositories.MasterProductRepository
}

// NewStagingImportService creates a new StagingImportService.
func NewStagingImportService(db *gorm.DB) *StagingImportService {
	return &StagingImportService{
		db:   db,
		repo: repositories.NewMasterProductRepository(db),
	}
}

// findOrCreateMasterProduct finds an existing master product by normalized title or creates one.
// Returns the product, a bool indicating whether it was created (true) or matched (false), and any error.
func (s *StagingImportService) findOrCreateMasterProduct(
	ctx context.Context,
	tenantID string,
	originalTitle string,
) (*models.MasterProduct, bool, error) {
	normalizedTitle := normalizeTitle(originalTitle)

	existing, err := s.repo.FindByExactTitle(ctx, tenantID, normalizedTitle)
	if err != nil {
		return nil, false, fmt.Errorf("find master product by title: %w", err)
	}

	if existing != nil {
		return existing, false, nil
	}

	// Not found — create a new master product using the trimmed original title.
	product := &models.MasterProduct{
		TenantID: tenantID,
		Title:    strings.TrimSpace(originalTitle),
		Status:   models.MasterProductStatusActive,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, false, fmt.Errorf("create master product: %w", err)
	}

	return product, true, nil
}

// whitespaceRe collapses runs of whitespace into a single space.
var whitespaceRe = regexp.MustCompile(`\s+`)

// normalizeTitle returns a lowercase, trimmed, whitespace-collapsed version of s
// for deduplication lookups. This prevents duplicates like "240 gr" vs "240gr"
// from being treated as different products.
func normalizeTitle(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	return whitespaceRe.ReplaceAllString(s, " ")
}

// findMasterProductByPlatformItemID looks up an existing MasterProduct via
// platform link (platform + platform_item_id). Returns nil if not found.
func (s *StagingImportService) findMasterProductByPlatformItemID(
	ctx context.Context,
	tenantID string,
	platform string,
	platformItemID string,
) (*models.MasterProduct, error) {
	trimmedID := strings.TrimSpace(platformItemID)
	if trimmedID == "" {
		return nil, nil
	}

	links, err := s.repo.FindPlatformLinksByItemID(ctx, tenantID, platform, trimmedID)
	if err != nil {
		return nil, fmt.Errorf("find %s platform links by item_id %s: %w", platform, trimmedID, err)
	}
	if len(links) == 0 {
		return nil, nil
	}

	product, err := s.repo.FindByTenantAndID(ctx, tenantID, links[0].MasterProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find master product %d for %s item_id %s: %w",
			links[0].MasterProductID, platform, trimmedID, err)
	}

	return product, nil
}

// findMasterProductBySellerSkus searches for an existing MasterProduct that
// already owns one of the given seller SKUs. Returns nil if none found.
func (s *StagingImportService) findMasterProductBySellerSkus(
	ctx context.Context,
	tenantID string,
	sellerSkus []string,
) (*models.MasterProduct, error) {
	for _, raw := range sellerSkus {
		sku := strings.TrimSpace(raw)
		if sku == "" {
			continue
		}

		existingSku, err := s.repo.FindBySku(ctx, tenantID, sku)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, fmt.Errorf("find master sku by seller_sku %s: %w", sku, err)
		}

		product, err := s.repo.FindByTenantAndID(ctx, tenantID, existingSku.MasterProductID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, fmt.Errorf("find master product %d for seller_sku %s: %w",
				existingSku.MasterProductID, sku, err)
		}

		return product, nil
	}

	return nil, nil
}

// upsertMasterSku creates a new SKU or updates an existing one by
// (tenant_id, master_product_id, seller_sku).
func (s *StagingImportService) upsertMasterSku(
	ctx context.Context,
	masterSku *models.MasterProductSku,
) (*models.MasterProductSku, bool, error) {
	if masterSku == nil {
		return nil, false, fmt.Errorf("master sku is nil")
	}

	masterSku.SellerSku = strings.TrimSpace(masterSku.SellerSku)
	if masterSku.SellerSku == "" {
		return nil, false, fmt.Errorf("seller_sku is required")
	}

	var existing models.MasterProductSku
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND master_product_id = ? AND seller_sku = ?",
			masterSku.TenantID, masterSku.MasterProductID, masterSku.SellerSku).
		First(&existing).Error
	if err == nil {
		incomingVariantName := strings.TrimSpace(masterSku.VariantName)
		if incomingVariantName != "" {
			existing.VariantName = incomingVariantName
		}
		if len(masterSku.VariantData) > 0 {
			existing.VariantData = masterSku.VariantData
		}
		existing.Price = masterSku.Price
		existing.Stock = masterSku.Stock

		if saveErr := s.repo.UpdateSku(ctx, &existing); saveErr != nil {
			return nil, false, fmt.Errorf("update existing master sku: %w", saveErr)
		}

		return &existing, false, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, fmt.Errorf("find existing master sku: %w", err)
	}

	if createErr := s.repo.CreateSku(ctx, masterSku); createErr != nil {
		return nil, false, fmt.Errorf("create master sku: %w", createErr)
	}

	return masterSku, true, nil
}
