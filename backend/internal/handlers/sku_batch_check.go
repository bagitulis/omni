package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SkuBatchCheckHandler handles SKU batch check endpoints
type SkuBatchCheckHandler struct {
	fallbackDB *gorm.DB
}

// NewSkuBatchCheckHandler creates a new SKU batch check handler
func NewSkuBatchCheckHandler(db *gorm.DB) *SkuBatchCheckHandler {
	return &SkuBatchCheckHandler{fallbackDB: db}
}

// SkuCheckResult represents the result of checking a single SKU
type SkuCheckResult struct {
	Sku    string `json:"sku"`
	Lazada bool   `json:"lazada"`
	Shopee bool   `json:"shopee"`
	Tiktok bool   `json:"tiktok"`
}

// BatchCheckSkuRequest is the request body for batch SKU check
type BatchCheckSkuRequest struct {
	Skus []string `json:"skus" binding:"required"`
}

// BatchSavePlatformStatusRequest is the request body for saving platform status
type BatchSavePlatformStatusRequest struct {
	Results []SkuCheckResult `json:"results" binding:"required"`
}

// SkuPlatformCheckResult is the database model for cached SKU check results
type SkuPlatformCheckResult struct {
	ID        uint      `gorm:"primaryKey"`
	TenantID  string    `gorm:"column:tenant_id;index;not null"`
	Sku       string    `gorm:"column:sku;index;not null"`
	Lazada    bool      `gorm:"column:lazada;default:false"`
	Shopee    bool      `gorm:"column:shopee;default:false"`
	Tiktok    bool      `gorm:"column:tiktok;default:false"`
	CheckedAt time.Time `gorm:"column:checked_at"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName returns the table name for SkuPlatformCheckResult
func (SkuPlatformCheckResult) TableName() string {
	return "sku_platform_check_results"
}

// getDB returns the appropriate database for the current request
func (h *SkuBatchCheckHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// BatchCheckSku handles POST /api/inventory/batch-check-sku
// Check multiple SKUs across all platforms (Lazada, Shopee, TikTok)
func (h *SkuBatchCheckHandler) BatchCheckSku(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	var req BatchCheckSkuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "skus array is required",
		})
		return
	}

	if len(req.Skus) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "skus array cannot be empty",
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

	results := make([]SkuCheckResult, 0, len(req.Skus))

	for _, sku := range req.Skus {
		sku = strings.TrimSpace(sku)
		if sku == "" {
			continue
		}

		result := SkuCheckResult{Sku: sku}

		// Check Lazada SKU
		result.Lazada = h.checkLazadaSku(db, sku)

		// Check Shopee SKU
		result.Shopee = h.checkShopeeSku(db, sku)

		// Check TikTok SKU
		result.Tiktok = h.checkTiktokSku(db, sku)

		results = append(results, result)
	}

	log.Printf("[INFO] Batch SKU check completed - tenant: %s, total: %d, checked: %d",
		tenantID, len(req.Skus), len(results))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total":   len(req.Skus),
			"checked": len(results),
			"results": results,
		},
	})
}

// checkLazadaSku checks if SKU exists in Lazada
func (h *SkuBatchCheckHandler) checkLazadaSku(db *gorm.DB, sku string) bool {
	var count int64
	err := db.Table("lazada_skus").
		Where("sku_id = ? OR seller_sku = ?", sku, sku).
		Count(&count).Error
	if err != nil {
		log.Printf("[WARN] Error checking Lazada SKU %s: %v", sku, err)
		return false
	}
	return count > 0
}

// checkShopeeSku checks if SKU exists in Shopee
// Note: shopee_skus table only has seller_sku column, not model_sku
func (h *SkuBatchCheckHandler) checkShopeeSku(db *gorm.DB, sku string) bool {
	var count int64
	err := db.Table("shopee_skus").
		Where("seller_sku = ? OR CAST(model_id AS TEXT) = ?", sku, sku).
		Count(&count).Error
	if err != nil {
		log.Printf("[WARN] Error checking Shopee SKU %s: %v", sku, err)
		return false
	}
	return count > 0
}

// checkTiktokSku checks if SKU exists in TikTok
func (h *SkuBatchCheckHandler) checkTiktokSku(db *gorm.DB, sku string) bool {
	var count int64
	err := db.Table("tiktok_skus").
		Where("sku_id = ? OR seller_sku = ?", sku, sku).
		Count(&count).Error
	if err != nil {
		log.Printf("[WARN] Error checking TikTok SKU %s: %v", sku, err)
		return false
	}
	return count > 0
}
