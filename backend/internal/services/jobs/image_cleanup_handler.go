package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	imageService "github.com/omni/backend/internal/services/image"
	"gorm.io/gorm"
)

const (
	GracePeriod         = 24 * time.Hour
	JobTypeImageCleanup = "image_cleanup"
)

type ImageCleanupHandler struct {
	systemDB *gorm.DB
}

type imageCleanupJobData struct {
	TenantID string `json:"tenant_id"`
}

func NewImageCleanupHandler(systemDB *gorm.DB) *ImageCleanupHandler {
	return &ImageCleanupHandler{systemDB: systemDB}
}

// HandleImageCleanup - main job handler function
func (h *ImageCleanupHandler) HandleImageCleanup(ctx context.Context, payload string) (string, error) {
	var jobData imageCleanupJobData
	if err := json.Unmarshal([]byte(payload), &jobData); err != nil {
		return "", fmt.Errorf("invalid job data: %w", err)
	}

	tenantID := jobData.TenantID
	if tenantID == "" {
		return "", fmt.Errorf("missing tenant_id in job data")
	}

	log.Printf("[ImageCleanupHandler] Starting cleanup for tenant %s", tenantID)

	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant database: %w", err)
	}

	basePath := os.Getenv("UPLOAD_PATH")
	if basePath == "" {
		basePath = "uploads"
	}
	storageService := imageService.NewStorageService(basePath)

	cutoff := time.Now().Add(-GracePeriod)
	var images []models.Image
	if err := tenantDB.WithContext(ctx).
		Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoff).
		Find(&images).Error; err != nil {
		return "", fmt.Errorf("failed to fetch orphan images: %w", err)
	}

	if len(images) == 0 {
		return "no orphan images to cleanup", nil
	}

	total := len(images)
	deleted := 0
	failed := 0

	for _, img := range images {
		if err := storageService.DeleteImage(tenantID, img.LocalPath); err != nil {
			log.Printf("[ImageCleanupHandler] Failed to delete file for image %d: %v", img.ID, err)
			failed++
			continue
		}

		if err := tenantDB.WithContext(ctx).Unscoped().Delete(&img).Error; err != nil {
			log.Printf("[ImageCleanupHandler] Failed to delete image record %d: %v", img.ID, err)
			failed++
			continue
		}

		deleted++
	}

	summary := fmt.Sprintf("cleanup completed: total=%d deleted=%d failed=%d", total, deleted, failed)
	log.Printf("[ImageCleanupHandler] %s", summary)
	return summary, nil
}
