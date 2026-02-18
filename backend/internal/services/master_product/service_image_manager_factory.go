package master_product

import (
	"os"

	imageService "github.com/omni/backend/internal/services/image"
	"gorm.io/gorm"
)

func newDefaultImageManager(db *gorm.DB) *ImageManager {
	uploadPath := os.Getenv("UPLOAD_PATH")
	if uploadPath == "" {
		uploadPath = "uploads"
	}

	webpService := imageService.NewWebPService()
	storageService := imageService.NewStorageService(uploadPath)
	thumbnailService := imageService.NewThumbnailService(webpService, storageService)
	referenceService := imageService.NewReferenceService(db)
	dedupService := imageService.NewDedupService(db, webpService, storageService, thumbnailService, referenceService)

	return NewImageManager(db, dedupService, referenceService)
}
