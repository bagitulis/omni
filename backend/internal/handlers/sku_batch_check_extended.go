package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BatchSavePlatformStatus handles POST /api/inventory/batch-save-platform-status
// Save SKU platform check results to database for caching
func (h *SkuBatchCheckHandler) BatchSavePlatformStatus(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "tenant ID required",
		})
		return
	}

	var req BatchSavePlatformStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "results array is required",
		})
		return
	}

	if len(req.Results) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "results array cannot be empty",
		})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		log.Printf("[ERROR] Failed to get tenant database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection error",
		})
		return
	}

	// Ensure table exists
	if err := h.ensureTableExists(db); err != nil {
		log.Printf("[ERROR] Failed to ensure table exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create table",
		})
		return
	}

	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to start transaction",
		})
		return
	}

	// Delete old records for this tenant
	deleteResult := tx.Where("tenant_id = ?", tenantID).Delete(&SkuPlatformCheckResult{})
	if deleteResult.Error != nil {
		tx.Rollback()
		log.Printf("[ERROR] Failed to delete old records: %v", deleteResult.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete old records",
		})
		return
	}
	deletedCount := deleteResult.RowsAffected

	// Insert new records
	now := time.Now()
	records := make([]SkuPlatformCheckResult, 0, len(req.Results))
	for _, r := range req.Results {
		records = append(records, SkuPlatformCheckResult{
			TenantID:  tenantID,
			Sku:       r.Sku,
			Lazada:    r.Lazada,
			Shopee:    r.Shopee,
			Tiktok:    r.Tiktok,
			CheckedAt: now,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	if err := tx.CreateInBatches(records, 100).Error; err != nil {
		tx.Rollback()
		log.Printf("[ERROR] Failed to insert new records: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save platform status",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("[ERROR] Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to commit changes",
		})
		return
	}

	log.Printf("[INFO] Batch save platform status - tenant: %s, deleted: %d, saved: %d",
		tenantID, deletedCount, len(records))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total":   len(req.Results),
			"saved":   len(records),
			"deleted": deletedCount,
		},
	})
}

// GetPlatformStatus handles GET /api/inventory/platform-status
// Load cached SKU platform check results from database
func (h *SkuBatchCheckHandler) GetPlatformStatus(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "tenant ID required",
		})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		log.Printf("[ERROR] Failed to get tenant database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection error",
		})
		return
	}

	// Check if table exists first
	if !db.Migrator().HasTable(&SkuPlatformCheckResult{}) {
		// Table doesn't exist yet, return empty results
		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"count":       0,
			"lastChecked": nil,
			"results":     []interface{}{},
		})
		return
	}

	var records []SkuPlatformCheckResult
	if err := db.Where("tenant_id = ?", tenantID).
		Order("checked_at DESC").
		Find(&records).Error; err != nil {
		log.Printf("[ERROR] Failed to load platform status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to load platform status",
		})
		return
	}

	// Convert to response format
	results := make([]SkuCheckResult, 0, len(records))
	var lastChecked *time.Time
	for _, r := range records {
		if lastChecked == nil {
			lastChecked = &r.CheckedAt
		}
		results = append(results, SkuCheckResult{
			Sku:    r.Sku,
			Lazada: r.Lazada,
			Shopee: r.Shopee,
			Tiktok: r.Tiktok,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"count":       len(results),
		"lastChecked": lastChecked,
		"results":     results,
	})
}

// ensureTableExists creates the table if it doesn't exist
func (h *SkuBatchCheckHandler) ensureTableExists(db *gorm.DB) error {
	if !db.Migrator().HasTable(&SkuPlatformCheckResult{}) {
		return db.AutoMigrate(&SkuPlatformCheckResult{})
	}
	return nil
}
