// Package master_product provides SKU mapping service
package master_product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// SKU Mapper errors
var (
	ErrSkuAlreadyLinked    = errors.New("SKU is already linked to this platform")
	ErrAmbiguousSku        = errors.New("SKU found in multiple products - manual linking required")
	ErrPlatformSkuNotFound = errors.New("platform SKU not found")
)

// SkuMapper handles SKU mapping between Master Product and platforms
type SkuMapper struct {
	db       *gorm.DB
	repo     *repositories.MasterProductRepository
	tenantID string
}

// NewSkuMapper creates a new SKU mapper
func NewSkuMapper(db *gorm.DB, tenantID string) *SkuMapper {
	return &SkuMapper{
		db:       db,
		repo:     repositories.NewMasterProductRepository(db),
		tenantID: tenantID,
	}
}

// MappingStatus represents the mapping status for a master product
type MappingStatus struct {
	MasterProductID uint               `json:"master_product_id"`
	Platforms       []PlatformStatus   `json:"platforms"`
	SKUs            []SkuMappingStatus `json:"skus"`
}

// PlatformStatus represents mapping status for a platform
type PlatformStatus struct {
	Platform   string `json:"platform"`
	IsLinked   bool   `json:"is_linked"`
	ItemID     string `json:"item_id,omitempty"`
	SyncStatus string `json:"sync_status,omitempty"`
	LastSynced string `json:"last_synced,omitempty"`
}

// SkuMappingStatus represents mapping status for a SKU
type SkuMappingStatus struct {
	MasterSkuID uint             `json:"master_sku_id"`
	SellerSku   string           `json:"seller_sku"`
	VariantName string           `json:"variant_name"`
	Platforms   []PlatformStatus `json:"platforms"`
}

// AutoMapResult represents the result of auto-mapping
type AutoMapResult struct {
	Platform       string `json:"platform"`
	Found          bool   `json:"found"`
	PlatformSkuID  string `json:"platform_sku_id,omitempty"`
	PlatformItemID string `json:"platform_item_id,omitempty"`
	Ambiguous      bool   `json:"ambiguous"`
	Message        string `json:"message,omitempty"`
}

// GetMappingStatus returns the current mapping status for a master product
func (m *SkuMapper) GetMappingStatus(ctx context.Context, masterProductID uint) (*MappingStatus, error) {
	// Verify product exists and belongs to tenant
	product, err := m.repo.FindByTenantAndID(ctx, m.tenantID, masterProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	status := &MappingStatus{
		MasterProductID: masterProductID,
		Platforms:       []PlatformStatus{},
		SKUs:            []SkuMappingStatus{},
	}

	// Get product-level platform links
	links, err := m.repo.FindPlatformLinks(ctx, masterProductID)
	if err != nil {
		return nil, err
	}

	// Build platform status (product level - where master_sku_id is null)
	platformMap := make(map[string]PlatformStatus)
	for _, platform := range []string{models.PlatformShopee, models.PlatformTiktok, models.PlatformLazada} {
		platformMap[platform] = PlatformStatus{
			Platform: platform,
			IsLinked: false,
		}
	}

	for _, link := range links {
		if link.MasterSkuID == nil {
			ps := platformMap[link.Platform]
			ps.IsLinked = true
			ps.ItemID = link.PlatformItemID
			ps.SyncStatus = link.SyncStatus
			if link.LastSyncedAt != nil {
				ps.LastSynced = link.LastSyncedAt.Format(time.RFC3339)
			}
			platformMap[link.Platform] = ps
		}
	}

	for _, ps := range platformMap {
		status.Platforms = append(status.Platforms, ps)
	}

	// Build SKU status
	for _, sku := range product.SKUs {
		skuStatus := SkuMappingStatus{
			MasterSkuID: sku.ID,
			SellerSku:   sku.SellerSku,
			VariantName: sku.VariantName,
			Platforms:   []PlatformStatus{},
		}

		// Initialize all platforms as unlinked
		skuPlatformMap := make(map[string]PlatformStatus)
		for _, platform := range []string{models.PlatformShopee, models.PlatformTiktok, models.PlatformLazada} {
			skuPlatformMap[platform] = PlatformStatus{
				Platform: platform,
				IsLinked: false,
			}
		}

		// Update with actual links
		for _, link := range sku.PlatformLinks {
			ps := skuPlatformMap[link.Platform]
			ps.IsLinked = true
			ps.ItemID = link.PlatformSkuID
			ps.SyncStatus = link.SyncStatus
			if link.LastSyncedAt != nil {
				ps.LastSynced = link.LastSyncedAt.Format(time.RFC3339)
			}
			skuPlatformMap[link.Platform] = ps
		}

		for _, ps := range skuPlatformMap {
			skuStatus.Platforms = append(skuStatus.Platforms, ps)
		}

		status.SKUs = append(status.SKUs, skuStatus)
	}

	return status, nil
}

// AutoMapBySku attempts to automatically find and link platform SKUs by seller_sku
func (m *SkuMapper) AutoMapBySku(ctx context.Context, sellerSku string) ([]AutoMapResult, error) {
	results := []AutoMapResult{}

	// Find Shopee SKU
	shopeeResult := m.findShopeeSku(ctx, sellerSku)
	results = append(results, shopeeResult)

	// Find TikTok SKU
	tiktokResult := m.findTiktokSku(ctx, sellerSku)
	results = append(results, tiktokResult)

	// Find Lazada SKU
	lazadaResult := m.findLazadaSku(ctx, sellerSku)
	results = append(results, lazadaResult)

	return results, nil
}

// ManualLink creates a manual link between a master SKU and a platform SKU
func (m *SkuMapper) ManualLink(ctx context.Context, masterSkuID uint, platform, platformItemID, platformSkuID string) error {
	// Find the master SKU
	var sku models.MasterProductSku
	err := m.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", masterSkuID, m.tenantID).
		First(&sku).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSkuNotFound
		}
		return err
	}

	// Check if already linked
	existingLinks, err := m.repo.FindPlatformLinks(ctx, sku.MasterProductID)
	if err != nil {
		return err
	}

	for _, link := range existingLinks {
		if link.MasterSkuID != nil && *link.MasterSkuID == masterSkuID && link.Platform == platform {
			return ErrSkuAlreadyLinked
		}
	}

	// Create new link
	now := time.Now()
	link := &models.MasterProductPlatformLink{
		MasterProductID: sku.MasterProductID,
		MasterSkuID:     &masterSkuID,
		Platform:        platform,
		PlatformItemID:  platformItemID,
		PlatformSkuID:   platformSkuID,
		SyncStatus:      models.SyncStatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := m.repo.CreatePlatformLink(ctx, link); err != nil {
		return fmt.Errorf("failed to create platform link: %w", err)
	}

	log.Info().
		Str("tenant_id", m.tenantID).
		Uint("master_sku_id", masterSkuID).
		Str("platform", platform).
		Str("platform_sku_id", platformSkuID).
		Msg("Manual SKU link created")

	return nil
}

// UnlinkSku removes a platform link from a master SKU
func (m *SkuMapper) UnlinkSku(ctx context.Context, masterSkuID uint, platform string) error {
	// Find the master SKU
	var sku models.MasterProductSku
	err := m.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", masterSkuID, m.tenantID).
		First(&sku).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSkuNotFound
		}
		return err
	}

	// Find and delete the link
	links, err := m.repo.FindPlatformLinks(ctx, sku.MasterProductID)
	if err != nil {
		return err
	}

	for _, link := range links {
		if link.MasterSkuID != nil && *link.MasterSkuID == masterSkuID && link.Platform == platform {
			if err := m.repo.DeletePlatformLink(ctx, link.ID); err != nil {
				return err
			}

			log.Info().
				Str("tenant_id", m.tenantID).
				Uint("master_sku_id", masterSkuID).
				Str("platform", platform).
				Msg("SKU link removed")

			return nil
		}
	}

	return nil // No link found, nothing to delete
}
