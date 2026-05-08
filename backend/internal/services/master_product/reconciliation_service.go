package master_product

import (
	"context"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// SyncStatusStale marks links pointing to products no longer found in staging.
const SyncStatusStale = "stale"

// ReconciliationResult holds the outcome of a reconciliation run.
type ReconciliationResult struct {
	TotalLinks    int `json:"total_links"`
	StaleDetected int `json:"stale_detected"`
	StaleCleared  int `json:"stale_cleared"`
	Errors        int `json:"errors"`
}

// ReconciliationService detects orphan platform links after product sync.
// An orphan link points to a platform product that no longer exists in staging tables.
type ReconciliationService struct {
	db       *gorm.DB
	tenantID string
}

// NewReconciliationService creates a new reconciliation service.
func NewReconciliationService(db *gorm.DB, tenantID string) *ReconciliationService {
	return &ReconciliationService{db: db, tenantID: tenantID}
}

// Reconcile checks all platform links and marks stale ones.
// A link is stale if its platform product no longer exists in the corresponding staging table.
// Links that were previously stale but now have a matching product are cleared back to "synced".
func (r *ReconciliationService) Reconcile(ctx context.Context) (*ReconciliationResult, error) {
	zlog := zerolog.Ctx(ctx)
	result := &ReconciliationResult{}

	var links []models.MasterProductPlatformLink
	if err := r.db.WithContext(ctx).
		Joins("JOIN master_products ON master_products.id = master_product_platform_links.master_product_id").
		Where("master_products.tenant_id = ?", r.tenantID).
		Find(&links).Error; err != nil {
		return nil, fmt.Errorf("reconciliation: load platform links: %w", err)
	}

	result.TotalLinks = len(links)
	if len(links) == 0 {
		return result, nil
	}

	for _, link := range links {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		exists, err := r.platformProductExists(ctx, link)
		if err != nil {
			zlog.Warn().Err(err).Uint("link_id", link.ID).Str("platform", link.Platform).Msg("Reconciliation: error checking link")
			result.Errors++
			continue
		}

		if !exists && link.SyncStatus != SyncStatusStale {
			// Mark as stale
			if err := r.markLinkStale(ctx, link.ID); err != nil {
				zlog.Warn().Err(err).Uint("link_id", link.ID).Msg("Reconciliation: failed to mark stale")
				result.Errors++
				continue
			}
			result.StaleDetected++
		} else if exists && link.SyncStatus == SyncStatusStale {
			// Product reappeared — clear stale status
			if err := r.clearStaleStatus(ctx, link.ID); err != nil {
				zlog.Warn().Err(err).Uint("link_id", link.ID).Msg("Reconciliation: failed to clear stale")
				result.Errors++
				continue
			}
			result.StaleCleared++
		}
	}

	zlog.Info().
		Int("total_links", result.TotalLinks).
		Int("stale_detected", result.StaleDetected).
		Int("stale_cleared", result.StaleCleared).
		Int("errors", result.Errors).
		Msg("Reconciliation completed")

	return result, nil
}

// platformProductExists checks if the linked platform product still exists in staging.
func (r *ReconciliationService) platformProductExists(ctx context.Context, link models.MasterProductPlatformLink) (bool, error) {
	var count int64

	switch link.Platform {
	case "shopee":
		err := r.db.WithContext(ctx).
			Table("shopee_products").
			Where("tenant_id = ? AND item_id = ?", r.tenantID, link.PlatformItemID).
			Count(&count).Error
		return count > 0, err

	case "tiktok":
		err := r.db.WithContext(ctx).
			Table("tiktok_products").
			Where("tenant_id = ? AND product_id = ?", r.tenantID, link.PlatformProductID).
			Count(&count).Error
		return count > 0, err

	case "lazada":
		// Lazada uses item_id stored in PlatformItemID
		err := r.db.WithContext(ctx).
			Table("lazada_products").
			Where("tenant_id = ? AND item_id = ?", r.tenantID, link.PlatformItemID).
			Count(&count).Error
		return count > 0, err

	default:
		return false, fmt.Errorf("unknown platform: %s", link.Platform)
	}
}

// markLinkStale sets a link's sync_status to "stale" with timestamp.
func (r *ReconciliationService) markLinkStale(ctx context.Context, linkID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.MasterProductPlatformLink{}).
		Where("id = ?", linkID).
		Updates(map[string]interface{}{
			"sync_status":    SyncStatusStale,
			"last_synced_at": now,
			"error_message":  "Product not found in staging after sync",
		}).Error
}

// clearStaleStatus resets a previously stale link back to synced.
func (r *ReconciliationService) clearStaleStatus(ctx context.Context, linkID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.MasterProductPlatformLink{}).
		Where("id = ?", linkID).
		Updates(map[string]interface{}{
			"sync_status":    models.SyncStatusSynced,
			"last_synced_at": now,
			"error_message":  "",
		}).Error
}
