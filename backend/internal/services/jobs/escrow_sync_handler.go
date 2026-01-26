// Package jobs provides escrow sync handler for background job execution
package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/analytics"
	"gorm.io/gorm"
)

// EscrowSyncHandler handles escrow sync jobs in background
type EscrowSyncHandler struct {
	systemDB *gorm.DB
}

// NewEscrowSyncHandler creates a new escrow sync handler
func NewEscrowSyncHandler(systemDB *gorm.DB) *EscrowSyncHandler {
	return &EscrowSyncHandler{systemDB: systemDB}
}

// HandleShopeeEscrowSync handles Shopee escrow sync job
func (h *EscrowSyncHandler) HandleShopeeEscrowSync(ctx context.Context, payload string) (string, error) {
	return h.handleEscrowSync(ctx, payload, "shopee")
}

// HandleTiktokEscrowSync handles TikTok escrow sync job
func (h *EscrowSyncHandler) HandleTiktokEscrowSync(ctx context.Context, payload string) (string, error) {
	return h.handleEscrowSync(ctx, payload, "tiktok")
}

// handleEscrowSync is the common handler for escrow sync jobs
func (h *EscrowSyncHandler) handleEscrowSync(ctx context.Context, payload, platform string) (string, error) {
	// Parse job data
	var jobData models.EscrowSyncJobData
	if err := json.Unmarshal([]byte(payload), &jobData); err != nil {
		return "", fmt.Errorf("invalid job data: %w", err)
	}

	// Validate tenant ID
	tenantID := jobData.TenantID
	if tenantID == "" {
		return "", fmt.Errorf("missing tenant_id in job data")
	}

	// Extract job ID from context for progress updates
	jobID, _ := ctx.Value(models.ContextKeyJobID).(string)

	log.Printf("[EscrowSyncHandler] Starting %s escrow sync for tenant %s, month=%d, year=%d, job_id=%s",
		platform, tenantID, jobData.Month, jobData.Year, jobID)

	// Get tenant database connection
	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant database: %w", err)
	}

	// Create progress callback using job ID
	var qm *QueueManager
	if jobID != "" {
		qm = NewQueueManager(tenantDB, tenantID)
	}

	progressCallback := func(percent, processed, total int, message string) {
		if qm != nil && jobID != "" {
			if err := qm.UpdateProgress(jobID, percent, processed, total, message); err != nil {
				log.Printf("[EscrowSyncHandler] Failed to update progress: %v", err)
			}
		}
	}

	// Get base path for credential service
	basePath := config.GetDataDir()

	// Execute sync based on platform
	var result *analytics.SyncResultWithProgress
	switch platform {
	case "shopee":
		result, err = h.syncShopeeEscrow(ctx, tenantDB, tenantID, basePath, jobData, progressCallback)
	case "tiktok":
		result, err = h.syncTiktokEscrow(ctx, tenantDB, tenantID, basePath, jobData, progressCallback)
	default:
		return "", fmt.Errorf("unknown platform: %s", platform)
	}

	if err != nil {
		log.Printf("[EscrowSyncHandler] %s escrow sync failed: %v", platform, err)
		return "", err
	}

	// Store result
	resultJSON, _ := json.Marshal(result)
	log.Printf("[EscrowSyncHandler] %s escrow sync completed: %s", platform, string(resultJSON))

	return string(resultJSON), nil
}

// syncShopeeEscrow executes Shopee escrow sync
func (h *EscrowSyncHandler) syncShopeeEscrow(
	ctx context.Context,
	tenantDB *gorm.DB,
	tenantID, basePath string,
	jobData models.EscrowSyncJobData,
	onProgress analytics.ProgressCallback,
) (*analytics.SyncResultWithProgress, error) {
	syncService := analytics.NewShopeeEscrowSyncService(tenantDB, tenantID, basePath)
	return syncService.SyncMonthWithProgress(ctx, jobData.Month, jobData.Year, jobData.ForceResync, onProgress)
}

// syncTiktokEscrow executes TikTok escrow sync
func (h *EscrowSyncHandler) syncTiktokEscrow(
	ctx context.Context,
	tenantDB *gorm.DB,
	tenantID, basePath string,
	jobData models.EscrowSyncJobData,
	onProgress analytics.ProgressCallback,
) (*analytics.SyncResultWithProgress, error) {
	syncService := analytics.NewTiktokEscrowSyncService(tenantDB, tenantID, basePath)
	return syncService.SyncMonthWithProgress(ctx, jobData.Month, jobData.Year, jobData.ForceResync, onProgress)
}
