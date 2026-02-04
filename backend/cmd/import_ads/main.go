//go:build tools
// +build tools

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/ads"
	"gorm.io/gorm"
)

func main() {
	// Parse flags
	tenantID := flag.String("tenant", "yumna_bertigamart", "Tenant ID")
	shopeeDir := flag.String("shopee", "", "Directory containing Shopee CSV files")
	tiktokDir := flag.String("tiktok", "", "Directory containing TikTok Excel files")
	flag.Parse()

	if *shopeeDir == "" && *tiktokDir == "" {
		log.Fatal("At least one of -shopee or -tiktok directory must be specified")
	}

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

	// Get tenant DB
	db, err := config.GetTenantDBWithContext(*tenantID, cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to connect to tenant database: %v", err)
	}

	ctx := context.Background()

	// Import Shopee ads
	if *shopeeDir != "" {
		if err := importShopeeAds(ctx, db, *tenantID, *shopeeDir); err != nil {
			log.Printf("Error importing Shopee ads: %v", err)
		}
	}

	// Import TikTok ads
	if *tiktokDir != "" {
		if err := importTiktokAds(ctx, db, *tenantID, *tiktokDir); err != nil {
			log.Printf("Error importing TikTok ads: %v", err)
		}
	}

	log.Println("Import completed!")
}

// getExistingBatches returns a set of existing filenames for a platform
func getExistingShopeeFiles(db *gorm.DB) map[string]bool {
	existing := make(map[string]bool)
	var batches []models.ShopeeAdsUploadBatch
	db.Select("file_name").Find(&batches)
	for _, b := range batches {
		existing[b.FileName] = true
	}
	return existing
}

func getExistingTiktokFiles(db *gorm.DB) map[string]bool {
	existing := make(map[string]bool)
	var batches []models.TiktokAdsUploadBatch
	db.Select("file_name").Find(&batches)
	for _, b := range batches {
		existing[b.FileName] = true
	}
	return existing
}

// printProgress displays a progress bar
func printProgress(current, total int, filename string, status string) {
	pct := float64(current) / float64(total) * 100
	barWidth := 30
	filled := int(float64(barWidth) * float64(current) / float64(total))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	// Truncate filename if too long
	displayName := filename
	if len(displayName) > 40 {
		displayName = displayName[:37] + "..."
	}

	fmt.Printf("\r[%s] %5.1f%% (%d/%d) %s - %s", bar, pct, current, total, status, displayName)
}

func importShopeeAds(ctx context.Context, db *gorm.DB, tenantID, dir string) error {
	fmt.Println("\n========== SHOPEE ADS IMPORT ==========")
	fmt.Printf("Directory: %s\n", dir)

	// Find all CSV files
	files, err := filepath.Glob(filepath.Join(dir, "*.csv"))
	if err != nil {
		return fmt.Errorf("failed to list CSV files: %w", err)
	}

	// Filter only ads files
	var adsFiles []string
	for _, f := range files {
		if strings.Contains(filepath.Base(f), "Iklan-Produk") {
			adsFiles = append(adsFiles, f)
		}
	}

	// Sort files by name (chronological order)
	sort.Strings(adsFiles)
	total := len(adsFiles)
	fmt.Printf("Found %d Shopee ads CSV files\n\n", total)

	if total == 0 {
		return nil
	}

	// Get existing batches to skip
	existing := getExistingShopeeFiles(db)
	fmt.Printf("Already imported: %d files\n", len(existing))

	svc := ads.NewShopeeAdsService(db, tenantID)

	var imported, skipped, failed int

	for i, file := range adsFiles {
		filename := filepath.Base(file)

		// Check if already imported
		if existing[filename] {
			skipped++
			printProgress(i+1, total, filename, "SKIP")
			continue
		}

		printProgress(i+1, total, filename, "LOAD")

		data, err := os.ReadFile(file)
		if err != nil {
			failed++
			fmt.Printf("\n  ❌ Error reading: %v\n", err)
			continue
		}

		result, err := svc.ParseCSV(ctx, data, filename)
		if err != nil {
			failed++
			fmt.Printf("\n  ❌ Error parsing: %v\n", err)
			continue
		}

		printProgress(i+1, total, filename, "SAVE")

		if err := svc.SaveBatch(ctx, filename, result); err != nil {
			failed++
			fmt.Printf("\n  ❌ Error saving: %v\n", err)
			continue
		}

		imported++
		printProgress(i+1, total, filename, fmt.Sprintf("OK(%d)", len(result.Data)))
	}

	fmt.Printf("\n\n✅ Shopee Import Summary:\n")
	fmt.Printf("   - Imported: %d files\n", imported)
	fmt.Printf("   - Skipped:  %d files (already exist)\n", skipped)
	fmt.Printf("   - Failed:   %d files\n", failed)

	return nil
}

func importTiktokAds(ctx context.Context, db *gorm.DB, tenantID, dir string) error {
	fmt.Println("\n========== TIKTOK ADS IMPORT ==========")
	fmt.Printf("Directory: %s\n", dir)

	// Find all Excel files
	files, err := filepath.Glob(filepath.Join(dir, "*.xlsx"))
	if err != nil {
		return fmt.Errorf("failed to list Excel files: %w", err)
	}

	// Filter only creative data files
	var creativeFiles []string
	for _, f := range files {
		if strings.Contains(filepath.Base(f), "creative data") {
			creativeFiles = append(creativeFiles, f)
		}
	}

	// Sort files by name (chronological order)
	sort.Strings(creativeFiles)
	total := len(creativeFiles)
	fmt.Printf("Found %d TikTok creative Excel files\n\n", total)

	if total == 0 {
		return nil
	}

	// Get existing batches to skip
	existing := getExistingTiktokFiles(db)
	fmt.Printf("Already imported: %d files\n", len(existing))

	svc := ads.NewTiktokAdsService(db, tenantID)

	var imported, skipped, failed int

	for i, file := range creativeFiles {
		filename := filepath.Base(file)

		// Check if already imported
		if existing[filename] {
			skipped++
			printProgress(i+1, total, filename, "SKIP")
			continue
		}

		printProgress(i+1, total, filename, "LOAD")

		data, err := os.ReadFile(file)
		if err != nil {
			failed++
			fmt.Printf("\n  ❌ Error reading: %v\n", err)
			continue
		}

		result, err := svc.ParseExcel(ctx, data, filename)
		if err != nil {
			failed++
			fmt.Printf("\n  ❌ Error parsing: %v\n", err)
			continue
		}

		printProgress(i+1, total, filename, "SAVE")

		if err := svc.SaveBatch(ctx, filename, result); err != nil {
			failed++
			fmt.Printf("\n  ❌ Error saving: %v\n", err)
			continue
		}

		imported++
		printProgress(i+1, total, filename, fmt.Sprintf("OK(%d)", len(result.Data)))
	}

	fmt.Printf("\n\n✅ TikTok Import Summary:\n")
	fmt.Printf("   - Imported: %d files\n", imported)
	fmt.Printf("   - Skipped:  %d files (already exist)\n", skipped)
	fmt.Printf("   - Failed:   %d files\n", failed)

	return nil
}
