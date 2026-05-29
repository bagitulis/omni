package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm/clause"
)

// ================== Platform Link Operations ==================

// CreatePlatformLink creates a new platform link
func (r *MasterProductRepository) CreatePlatformLink(ctx context.Context, link *models.MasterProductPlatformLink) error {
	return r.db.WithContext(ctx).Create(link).Error
}

// FindPlatformLinks returns all platform links for a master product
func (r *MasterProductRepository) FindPlatformLinks(ctx context.Context, masterProductID uint) ([]models.MasterProductPlatformLink, error) {
	var links []models.MasterProductPlatformLink
	err := r.db.WithContext(ctx).
		Where("master_product_id = ?", masterProductID).
		Order("platform ASC").
		Find(&links).Error
	return links, err
}

// FindPlatformLinkByPlatform finds a link for a specific platform
func (r *MasterProductRepository) FindPlatformLinkByPlatform(ctx context.Context, masterProductID uint, platform string) (*models.MasterProductPlatformLink, error) {
	var link models.MasterProductPlatformLink
	err := r.db.WithContext(ctx).
		Where("master_product_id = ? AND platform = ?", masterProductID, platform).
		First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// UpdateLinkStatus updates the sync status of a platform link
func (r *MasterProductRepository) UpdateLinkStatus(ctx context.Context, linkID uint, status string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.MasterProductPlatformLink{}).
		Where("id = ?", linkID).
		Updates(map[string]interface{}{
			"sync_status":    status,
			"last_synced_at": &now,
			"updated_at":     now,
		}).Error
}

// DeletePlatformLink removes a platform link
func (r *MasterProductRepository) DeletePlatformLink(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.MasterProductPlatformLink{}, id).Error
}

// UpsertPlatformLink creates or updates a platform link.
// Uses ON CONFLICT on the partial unique index idx_platform_links_unique
// (platform, platform_product_id) WHERE platform_product_id IS NOT NULL.
func (r *MasterProductRepository) UpsertPlatformLink(ctx context.Context, link *models.MasterProductPlatformLink) error {
	if link == nil {
		return errors.New("platform link is nil")
	}

	now := time.Now()
	if link.CreatedAt.IsZero() {
		link.CreatedAt = now
	}
	link.UpdatedAt = now

	// Use ON CONFLICT for atomic upsert — prevents race conditions
	// The unique index covers (platform, platform_product_id) WHERE platform_product_id IS NOT NULL
	if link.PlatformProductID != "" {
		return r.db.WithContext(ctx).
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "platform"}, {Name: "platform_product_id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"master_product_id", "master_sku_id", "platform_item_id",
					"platform_sku_id", "sync_status", "last_synced_at",
					"error_message", "updated_at",
				}),
			}).Create(link).Error
	}

	// Fallback for links without platform_product_id — use find-then-create
	var existing models.MasterProductPlatformLink
	err := r.db.WithContext(ctx).Where(
		"platform = ? AND COALESCE(platform_sku_id, '') = COALESCE(?, '')",
		link.Platform, link.PlatformSkuID,
	).First(&existing).Error

	if err == nil {
		updates := map[string]interface{}{
			"master_product_id": link.MasterProductID,
			"master_sku_id":     link.MasterSkuID,
			"platform_item_id":  link.PlatformItemID,
			"sync_status":       link.SyncStatus,
			"last_synced_at":    link.LastSyncedAt,
			"updated_at":        now,
		}
		return r.db.WithContext(ctx).Model(&existing).Updates(updates).Error
	}

	return r.db.WithContext(ctx).Create(link).Error
}


// FindPlatformLinksByItemID finds platform links by platform item ID
// Used to check if a platform product is already imported
func (r *MasterProductRepository) FindPlatformLinksByItemID(ctx context.Context, tenantID, platform, platformItemID string) ([]models.MasterProductPlatformLink, error) {
	var links []models.MasterProductPlatformLink

	err := r.db.WithContext(ctx).
		Joins("JOIN master_products ON master_products.id = master_product_platform_links.master_product_id").
		Where("master_products.tenant_id = ? AND master_product_platform_links.platform = ? AND master_product_platform_links.platform_item_id = ?",
			tenantID, platform, platformItemID).
		Find(&links).Error

	return links, err
}
