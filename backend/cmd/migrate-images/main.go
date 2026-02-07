package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/image"
	"gorm.io/gorm"
)

func main() {
	// Load config
	cfg := config.Load()

	// Set up database driver
	if cfg.DBDriver == "postgres" {
		pgConfig := &config.PostgresConfig{
			Host:     cfg.PGHost,
			Port:     cfg.PGPort,
			User:     cfg.PGUser,
			Password: cfg.PGPassword,
			DBName:   cfg.PGDatabase,
			SSLMode:  cfg.PGSSLMode,
		}
		config.SetDatabaseDriver(config.DriverPostgres, pgConfig)
	}

	// Initialize system DB (needed for initial connection setup usually)
	_, err := config.GetSystemDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to connect to system database: %v", err)
	}

	// Tenants to migrate - using the same list as seed command
	tenants := []string{"yumna_bertigamart", "tika_nusseyba"}

	for _, tenantID := range tenants {
		log.Printf("Starting migration for tenant: %s", tenantID)
		if err := migrateTenantImages(tenantID, cfg.DatabasePath); err != nil {
			log.Printf("❌ Failed to migrate tenant %s: %v", tenantID, err)
		} else {
			log.Printf("✅ Successfully migrated tenant %s", tenantID)
		}
	}
}

func migrateTenantImages(tenantID, basePath string) error {
	db, err := config.GetTenantDB(tenantID, basePath)
	if err != nil {
		return fmt.Errorf("failed to get tenant db: %w", err)
	}

	// Initialize Services
	webpService := image.NewWebPService()
	storageService := image.NewStorageService("uploads") // Default upload path
	if envPath := os.Getenv("UPLOAD_PATH"); envPath != "" {
		storageService = image.NewStorageService(envPath)
	}
	thumbnailService := image.NewThumbnailService(webpService, storageService)
	refService := image.NewReferenceService(db)
	dedupService := image.NewDedupService(db, webpService, storageService, thumbnailService, refService)

	// Fetch MasterProducts with images
	// Use batching to avoid loading all into memory
	var products []models.MasterProduct

	// Filtering where images is not null and not empty JSON array.
	query := db.Where("images IS NOT NULL AND jsonb_array_length(images) > 0")

	ctx := context.Background()

	err = query.FindInBatches(&products, 100, func(tx *gorm.DB, batch int) error {

		log.Printf("Processing batch %d...", batch)

		for _, product := range products {
			log.Printf("Processing Product ID %d...", product.ID)

			// product.Images is models.JSONArray ([]interface{})
			for i, imgRaw := range product.Images {
				imgURL, ok := imgRaw.(string)
				if !ok {
					log.Printf("  ⚠️ Skipping non-string image at index %d for product %d", i, product.ID)
					continue
				}

				if imgURL == "" {
					continue
				}

				// Download Image
				log.Printf("  Downloading: %s", imgURL)
				imgData, err := downloadImage(imgURL)
				if err != nil {
					log.Printf("  ❌ Failed to download %s: %v", imgURL, err)
					continue
				}

				// Dedup/Create Image
				imgRecord, isNew, err := dedupService.FindOrCreate(ctx, imgData, imgURL, tenantID, "products")
				if err != nil {
					log.Printf("  ❌ DedupService error for %s: %v", imgURL, err)
					continue
				}
				if isNew {
					log.Printf("  ✨ Created new image record: %d", imgRecord.ID)
				} else {
					log.Printf("  🔗 Found existing image record: %d", imgRecord.ID)
				}

				// Link to MasterProduct (Create Join Table Entry)
				// Check if link already exists to avoid duplicates/errors
				var existingLink models.MasterProductImage
				checkLink := db.Where("product_id = ? AND image_id = ?", product.ID, imgRecord.ID).First(&existingLink)
				if checkLink.Error == nil {
					log.Printf("  ⚠️ Link already exists for Product %d - Image %d", product.ID, imgRecord.ID)
					continue
				}

				link := models.MasterProductImage{
					ProductID: product.ID,
					ImageID:   uint(imgRecord.ID),
					SortOrder: i,
					Role:      "gallery", // Default role
				}
				if i == 0 {
					link.Role = "main" // First image is main
				}

				if err := db.Create(&link).Error; err != nil {
					log.Printf("  ❌ Failed to link image %d to product %d: %v", imgRecord.ID, product.ID, err)
				} else {
					log.Printf("  ✅ Linked image %d to product %d", imgRecord.ID, product.ID)
				}
			}
		}
		return nil
	}).Error

	if err != nil {
		return fmt.Errorf("failed to process products: %w", err)
	}

	return nil
}

func downloadImage(url string) ([]byte, error) {
	// Set a timeout for downloads
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}
