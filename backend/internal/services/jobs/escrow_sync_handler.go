package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// EscrowSyncService defines the interface for platform-specific escrow sync operations.
// Implemented by Tasks 5/6 (Shopee/TikTok concrete services).
type EscrowSyncService interface {
	SyncMonthWithProgress(ctx context.Context, month, year int, forceResync bool, onProgress func(processed, total int, message string)) error
}

// EscrowSyncHandler handles escrow sync jobs for Shopee and TikTok.
// It uses EscrowSyncService interface to remain platform-agnostic.
type EscrowSyncHandler struct {
	systemDB          *gorm.DB
	shopeeSyncService EscrowSyncService
	tiktokSyncService EscrowSyncService
}

// NewEscrowSyncHandler creates a new EscrowSyncHandler.
func NewEscrowSyncHandler(systemDB *gorm.DB) *EscrowSyncHandler {
	return &EscrowSyncHandler{
		systemDB: systemDB,
	}
}

// SetShopeeSyncService injects the Shopee escrow sync service.
func (h *EscrowSyncHandler) SetShopeeSyncService(svc EscrowSyncService) {
	h.shopeeSyncService = svc
}

// SetTiktokSyncService injects the TikTok escrow sync service.
func (h *EscrowSyncHandler) SetTiktokSyncService(svc EscrowSyncService) {
	h.tiktokSyncService = svc
}

// HandleShopeeEscrowSync processes a Shopee escrow sync job.
// Compatible with jobs.JobHandler type.
func (h *EscrowSyncHandler) HandleShopeeEscrowSync(ctx context.Context, payload string) (string, error) {
	return h.handleEscrowSync(ctx, payload, "Shopee", h.shopeeSyncService)
}

// HandleTiktokEscrowSync processes a TikTok escrow sync job.
// Compatible with jobs.JobHandler type.
func (h *EscrowSyncHandler) HandleTiktokEscrowSync(ctx context.Context, payload string) (string, error) {
	return h.handleEscrowSync(ctx, payload, "TikTok", h.tiktokSyncService)
}

// handleEscrowSync is the shared logic for both Shopee and TikTok escrow sync handlers.
func (h *EscrowSyncHandler) handleEscrowSync(ctx context.Context, payload, platform string, svc EscrowSyncService) (string, error) {
	var data models.EscrowSyncJobData
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return "", fmt.Errorf("invalid escrow sync job data: %w", err)
	}

	if data.TenantID == "" {
		return "", fmt.Errorf("missing tenant_id in job data")
	}

	if svc == nil {
		return "", fmt.Errorf("%s escrow sync service not initialized", platform)
	}

	tenantDB, err := config.GetTenantDBByID(data.TenantID)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant database: %w", err)
	}

	qm := NewQueueManager(tenantDB, data.TenantID)

	// Build progress callback that updates job progress in the queue
	onProgress := func(processed, total int, message string) {
		percent := 0
		if total > 0 {
			percent = processed * 100 / total
			percent = min(percent, 100)
		}

		// Attempt to extract job ID from context for progress updates
		if jobID, ok := ctx.Value(models.ContextKeyJobID).(string); ok && jobID != "" {
			if err := qm.UpdateProgress(jobID, percent, processed, total, message); err != nil {
				// Non-fatal: log failure but continue processing
				_ = err
			}
		}
	}

	if err := svc.SyncMonthWithProgress(ctx, data.Month, data.Year, data.ForceResync, onProgress); err != nil {
		return "", fmt.Errorf("%s escrow sync failed: %w", platform, err)
	}

	return "ok", nil
}
