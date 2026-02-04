package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

type productImagesRow struct {
	ID          uint             `gorm:"column:id"`
	LocalImages models.JSONArray `gorm:"column:local_images"`
}

type masterImagesRow struct {
	ID     uint             `gorm:"column:id"`
	Images models.JSONArray `gorm:"column:images"`
}

func main() {
	dryRun := flag.Bool("dry-run", false, "preview changes without modifying files or database")
	flag.Parse()

	tenantID := os.Getenv("TENANT_ID")
	if tenantID == "" {
		fmt.Println("FAILED: TENANT_ID is required")
		os.Exit(1)
	}

	basePath := os.Getenv("UPLOAD_PATH")
	if basePath == "" {
		basePath = "uploads"
	}

	category := "products"
	root := filepath.Join(basePath, tenantID, category)
	if _, statErr := os.Stat(root); statErr != nil {
		if os.IsNotExist(statErr) {
			fmt.Printf("No uploads found at %s\n", root)
			return
		}
		fmt.Printf("FAILED: unable to access %s: %v\n", root, statErr)
		os.Exit(1)
	}
	mapping := make(map[string]string)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext != ".webp" && ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		hash := hashBytes(data)
		canonicalName := fmt.Sprintf("asset_%s%s", hash, ext)

		relPath, relErr := filepath.Rel(basePath, path)
		if relErr != nil {
			return relErr
		}
		oldRel := filepath.ToSlash(relPath)
		newRel := filepath.ToSlash(filepath.Join(tenantID, category, canonicalName))
		if oldRel == newRel {
			return nil
		}

		newFull := filepath.Join(basePath, filepath.FromSlash(newRel))
		if !*dryRun {
			if _, statErr := os.Stat(newFull); statErr == nil {
				if removeErr := os.Remove(path); removeErr != nil {
					return removeErr
				}
			} else {
				if renameErr := os.Rename(path, newFull); renameErr != nil {
					return renameErr
				}
			}
		}

		mapping[oldRel] = newRel
		return nil
	})
	if err != nil {
		fmt.Printf("FAILED: scan error: %v\n", err)
		os.Exit(1)
	}

	if len(mapping) == 0 {
		fmt.Println("No duplicates detected")
		return
	}

	expanded := expandMapping(mapping)
	if *dryRun {
		fmt.Printf("Dry run complete. Dedupe candidates: %d\n", len(mapping))
		return
	}

	cfg := config.Load()
	if cfg.DBDriver != "postgres" {
		fmt.Println("FAILED: DB_DRIVER must be postgres")
		os.Exit(1)
	}

	config.SetDatabaseDriver(config.DriverPostgres, &config.PostgresConfig{
		Host:     cfg.PGHost,
		Port:     cfg.PGPort,
		User:     cfg.PGUser,
		Password: cfg.PGPassword,
		DBName:   cfg.PGDatabase,
		SSLMode:  cfg.PGSSLMode,
	})

	db, err := config.GetTenantDB(tenantID, cfg.DatabasePath)
	if err != nil {
		fmt.Printf("FAILED: unable to connect tenant DB: %v\n", err)
		os.Exit(1)
	}

	if err := updateProductImages(db, models.GetTableName("ShopeeProduct"), expanded); err != nil {
		fail(err)
	}
	if err := updateProductImages(db, models.GetTableName("TiktokProduct"), expanded); err != nil {
		fail(err)
	}
	if err := updateProductImages(db, models.GetTableName("LazadaProduct"), expanded); err != nil {
		fail(err)
	}
	if err := updateMasterImages(db, expanded); err != nil {
		fail(err)
	}
	if err := updateOrderItemImages(db, expanded); err != nil {
		fail(err)
	}
	if err := updateGalleryImages(db, mapping); err != nil {
		fail(err)
	}

	fmt.Printf("Deduped %d files and updated references\n", len(mapping))
}

func fail(err error) {
	fmt.Printf("FAILED: %v\n", err)
	os.Exit(1)
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func expandMapping(mapping map[string]string) map[string]string {
	expanded := make(map[string]string, len(mapping)*2)
	for oldRel, newRel := range mapping {
		oldRel = filepath.ToSlash(oldRel)
		newRel = filepath.ToSlash(newRel)
		expanded[oldRel] = newRel
		expanded["/uploads/"+oldRel] = "/uploads/" + newRel
	}
	return expanded
}

func updateProductImages(db *gorm.DB, tableName string, mapping map[string]string) error {
	offset := 0
	limit := 200
	for {
		var rows []productImagesRow
		err := db.Table(tableName).
			Select("id", "local_images").
			Order("id ASC").
			Offset(offset).
			Limit(limit).
			Find(&rows).Error
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			updated, changed := replaceJSONArray(row.LocalImages, mapping)
			if !changed {
				continue
			}
			if err := db.Table(tableName).
				Where("id = ?", row.ID).
				Update("local_images", updated).Error; err != nil {
				return err
			}
		}
		offset += limit
	}
	return nil
}

func updateMasterImages(db *gorm.DB, mapping map[string]string) error {
	offset := 0
	limit := 200
	for {
		var rows []masterImagesRow
		err := db.Table(models.GetTableName("MasterProduct")).
			Select("id", "images").
			Order("id ASC").
			Offset(offset).
			Limit(limit).
			Find(&rows).Error
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			updated, changed := replaceJSONArray(row.Images, mapping)
			if !changed {
				continue
			}
			if err := db.Table(models.GetTableName("MasterProduct")).
				Where("id = ?", row.ID).
				Update("images", updated).Error; err != nil {
				return err
			}
		}
		offset += limit
	}
	return nil
}

func updateOrderItemImages(db *gorm.DB, mapping map[string]string) error {
	for oldPath, newPath := range mapping {
		if err := db.Model(&models.TiktokOrderItem{}).
			Where("product_image = ?", oldPath).
			Update("product_image", newPath).Error; err != nil {
			return err
		}
		if err := db.Model(&models.LazadaOrderItem{}).
			Where("product_image = ?", oldPath).
			Update("product_image", newPath).Error; err != nil {
			return err
		}
	}
	return nil
}

func updateGalleryImages(db *gorm.DB, mapping map[string]string) error {
	for oldPath, newPath := range mapping {
		if err := db.Model(&models.Image{}).
			Where("local_path = ?", oldPath).
			Update("local_path", newPath).Error; err != nil {
			return err
		}
	}
	return nil
}

func replaceJSONArray(arr models.JSONArray, mapping map[string]string) (models.JSONArray, bool) {
	if len(arr) == 0 {
		return arr, false
	}

	changed := false
	updated := make(models.JSONArray, 0, len(arr))
	for _, entry := range arr {
		value, ok := entry.(string)
		if !ok || value == "" {
			updated = append(updated, entry)
			continue
		}
		if replacement, exists := mapping[value]; exists {
			updated = append(updated, replacement)
			changed = true
			continue
		}
		updated = append(updated, value)
	}

	return updated, changed
}
