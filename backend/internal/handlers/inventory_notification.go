package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services"
	"github.com/rs/zerolog/log"
)

// pushBulkOperationNotification sends a notification with structured metadata
// after a bulk stock or price sync operation completes.
// Called in a goroutine so it does not block the HTTP response.
func pushBulkOperationNotification(
	notifSvc *services.NotificationService,
	metadata *models.BulkOperationMetadata,
) {
	if notifSvc == nil || metadata == nil {
		return
	}

	// Determine notification type based on results
	notifType := models.NotifTypeSuccess
	if metadata.Failed > 0 && metadata.Succeeded == 0 {
		notifType = models.NotifTypeError
	} else if metadata.Failed > 0 {
		notifType = models.NotifTypeWarning
	}

	// Build human-readable title
	title := fmt.Sprintf("Bulk %s: %d succeeded, %d failed",
		operationTypeLabel(metadata.OperationType),
		metadata.Succeeded, metadata.Failed)

	// Category
	category := models.CatInventory

	// Marshal metadata to JSON for storage in Metadata field
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		log.Error().Err(err).Msg("[Notification] Failed to marshal bulk operation metadata")
		return
	}

	// Human-readable summary for the message field
	message := fmt.Sprintf("%d total, %d succeeded, %d failed across %d platforms",
		metadata.Total, metadata.Succeeded, metadata.Failed, len(metadata.Platforms))

	// Push notification
	notif, pushErr := notifSvc.Push(notifType, category, title, message, "", string(metadataJSON))
	if pushErr != nil {
		log.Error().Err(pushErr).Msg("[Notification] Failed to push bulk operation notification")
		return
	}


	log.Info().
		Int64("notification_id", notif.ID).
		Str("type", notifType).
		Int("succeeded", metadata.Succeeded).
		Int("failed", metadata.Failed).
		Msg("[Notification] Bulk operation notification pushed")
}

// operationTypeLabel returns a human-readable label for the operation type.
func operationTypeLabel(opType string) string {
	switch opType {
	case "stock_sync":
		return "Stock Sync"
	case "price_sync":
		return "Price Sync"
	default:
		return "Sync"
	}
}

// buildBulkOperationMetadata constructs BulkOperationMetadata from batch results.
func buildBulkOperationMetadata(operationType string, results []interface{}) *models.BulkOperationMetadata {
	meta := &models.BulkOperationMetadata{
		OperationType: operationType,
		Platforms:     make(map[string]models.PlatformSyncStats),
		FailedItems:   []models.FailedItemDetail{},
	}

	for _, raw := range results {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		meta.Total++

		sku, _ := item["sku"].(string)
		success, _ := item["success"].(bool)

		if success {
			meta.Succeeded++
		} else {
			meta.Failed++
			errMsg, _ := item["error"].(string)
			if errMsg != "" {
				meta.FailedItems = append(meta.FailedItems, models.FailedItemDetail{
					SKU:   sku,
					Error: errMsg,
				})
			}
		}

		// Extract per-platform stats if available
		if platforms, ok := item["platforms"].(map[string]interface{}); ok {
			for platform, pRaw := range platforms {
				pData, ok := pRaw.(map[string]interface{})
				if !ok {
					continue
				}
				stats := meta.Platforms[platform]
				pSuccess, _ := pData["success"].(bool)
				if pSuccess {
					stats.Succeeded++
				} else {
					stats.Failed++
					pErr, _ := pData["error"].(string)
					reqID, _ := pData["request_id"].(string)
					if pErr != "" {
						meta.FailedItems = append(meta.FailedItems, models.FailedItemDetail{
							SKU:       sku,
							Platform:  platform,
							Error:     pErr,
							RequestID: reqID,
						})
					}
					stats.RequestID = reqID
				}
				meta.Platforms[platform] = stats
			}
		}
	}

	return meta
}

// pushInventoryBatchNotification creates a NotificationService from the request context
// and pushes a bulk operation notification for the given operation type and results.
func (h *InventoryHandler) pushInventoryBatchNotification(c *gin.Context, operationType string, results []interface{}) {
	db, err := GetTenantDBFromContext(c, h.fallbackDB)
	if err != nil {
		log.Error().Err(err).Msg("[Notification] Failed to get tenant DB for batch notification")
		return
	}

	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		return
	}

	repo := repositories.NewNotificationRepository(db)
	notifSvc := services.NewNotificationService(repo).WithTenant(tenantID)
	meta := buildBulkOperationMetadata(operationType, results)
	pushBulkOperationNotification(notifSvc, meta)
}
