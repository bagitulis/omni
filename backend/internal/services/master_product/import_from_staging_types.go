// Package master_product provides import service for Master Product
package master_product

import (
	"context"
	"errors"
	"fmt"
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
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, false, fmt.Errorf("create master product: %w", err)
	}

	return product, true, nil
}

// normalizeTitle returns a lowercase, trimmed version of s for deduplication lookups.
func normalizeTitle(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
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
		existing.VariantName = masterSku.VariantName
		existing.VariantData = masterSku.VariantData
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
