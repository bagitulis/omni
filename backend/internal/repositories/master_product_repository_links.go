package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
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

// UpsertPlatformLink creates or updates a platform link
func (r *MasterProductRepository) UpsertPlatformLink(ctx context.Context, link *models.MasterProductPlatformLink) error {
	if link == nil {
		return errors.New("platform link is nil")
	}
	now := time.Now()

	var existingBySku models.MasterProductPlatformLink
	err := r.db.WithContext(ctx).
		Where("master_product_id = ? AND platform = ? AND COALESCE(master_sku_id, 0) = COALESCE(?, 0)",
			link.MasterProductID, link.Platform, link.MasterSkuID).
		First(&existingBySku).Error
	if err == nil {
		existingBySku.MasterProductID = link.MasterProductID
		existingBySku.MasterSkuID = link.MasterSkuID
		existingBySku.Platform = link.Platform
		existingBySku.PlatformProductID = link.PlatformProductID
		existingBySku.PlatformItemID = link.PlatformItemID
		existingBySku.PlatformSkuID = link.PlatformSkuID
		existingBySku.SyncStatus = link.SyncStatus
		existingBySku.LastSyncedAt = link.LastSyncedAt
		existingBySku.UpdatedAt = now

		return r.db.WithContext(ctx).Save(&existingBySku).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	platformProductID := strings.TrimSpace(link.PlatformProductID)
	if platformProductID == "" {
		platformProductID = strings.TrimSpace(link.PlatformItemID)
		if platformProductID != "" {
			link.PlatformProductID = platformProductID
		}
	}

	if platformProductID != "" {
		link.PlatformProductID = platformProductID
		if link.CreatedAt.IsZero() {
			link.CreatedAt = now
		}
		link.UpdatedAt = now

		var existing models.MasterProductPlatformLink
		err := r.db.WithContext(ctx).
			Where("platform = ? AND platform_product_id = ?", link.Platform, platformProductID).
			First(&existing).Error
		if err == nil {
			existing.MasterProductID = link.MasterProductID
			existing.MasterSkuID = link.MasterSkuID
			existing.PlatformItemID = link.PlatformItemID
			existing.PlatformSkuID = link.PlatformSkuID
			existing.SyncStatus = link.SyncStatus
			existing.LastSyncedAt = link.LastSyncedAt
			existing.UpdatedAt = now

			return r.db.WithContext(ctx).Save(&existing).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		return r.db.WithContext(ctx).Create(link).Error
	}

	if link.CreatedAt.IsZero() {
		link.CreatedAt = now
	}
	link.UpdatedAt = now

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
