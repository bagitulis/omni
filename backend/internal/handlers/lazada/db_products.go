package lazada

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// DBProductHandler handles Lazada DB product requests
type DBProductHandler struct {
	basePath string
}

// NewDBProductHandler creates a new DB product handler
func NewDBProductHandler(basePath string) *DBProductHandler {
	return &DBProductHandler{basePath: basePath}
}

// FlattenedSkuRow represents a flattened SKU row for frontend display
// Format matches frontend productManagerConfig.ts lazadaConfig (snake_case)
type FlattenedSkuRow struct {
	ItemID      string         `json:"item_id"`      // Product item_id
	SkuID       string         `json:"sku_id"`       // SKU ID from Lazada
	SkuName     string         `json:"sku_name"`     // seller_sku
	ItemName    string         `json:"item_name"`    // Product name
	VariantName string         `json:"variant_name"` // Variant name
	Price       float64        `json:"price"`        // Current price
	Quantity    int            `json:"quantity"`     // Stock quantity
	Status      string         `json:"status"`       // Product status
	LocalImages models.JSONMap `json:"local_images"` // Local image paths
	UpdatedAt   string         `json:"updated_at"`   // Last update time
}

// GetDBProducts handles GET /api/lazada/db/products
// Returns flattened SKU rows for frontend ProductTable display
// Matches frontend productManagerConfig.ts lazadaConfig columnFields
func (h *DBProductHandler) GetDBProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))
	if limit > 10000 {
		limit = 10000
	}

	products, total := h.getFlattenedSkuRows(db, offset, limit)

	log.Printf("[Lazada DB] GetDBProducts: found %d rows (total: %d)", len(products), total)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"products": products,
		"total":    total,
		"offset":   offset,
		"limit":    limit,
		"count":    len(products),
	})
}

// GetMasterProducts handles GET /api/lazada/db/products/master
// Returns same flattened SKU rows (alias for compatibility)
func (h *DBProductHandler) GetMasterProducts(c *gin.Context) {
	h.GetDBProducts(c) // Same logic, reuse
}

// getFlattenedSkuRows returns flattened SKU rows for frontend display
// Each row = 1 SKU with product info, matches frontend columnFields
// If no SKUs exist, falls back to showing products directly
func (h *DBProductHandler) getFlattenedSkuRows(db *gorm.DB, offset, limit int) ([]FlattenedSkuRow, int64) {
	// First check if there are any SKUs
	var skuCount int64
	db.Model(&models.LazadaSku{}).Count(&skuCount)
	log.Printf("[Lazada DB] SKU count: %d", skuCount)

	// If no SKUs, fall back to products directly
	if skuCount == 0 {
		log.Printf("[Lazada DB] No SKUs found, falling back to products")
		return h.getProductsAsFlattenedRows(db, offset, limit)
	}

	// Query all products for lookup by ItemID
	var products []models.LazadaProduct
	db.Find(&products)

	// Build product lookup map by ItemID (string)
	productMap := make(map[string]models.LazadaProduct)
	for _, p := range products {
		productMap[p.ItemID] = p
	}

	// Query SKUs with pagination
	var skus []models.LazadaSku
	db.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&skus)

	// Build flattened result
	result := make([]FlattenedSkuRow, 0, len(skus))
	for _, sku := range skus {
		// Use Name from SKU if available, otherwise SellerSku
		skuName := sku.Name
		if skuName == "" {
			skuName = sku.SellerSku
		}

		row := FlattenedSkuRow{
			ItemID:      sku.ItemID,
			SkuID:       sku.SkuID,
			SkuName:     skuName,
			VariantName: sku.VariantName,
			Price:       sku.Price,
			Quantity:    sku.Quantity,
			UpdatedAt:   sku.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		// Get product info by ItemID
		if product, ok := productMap[sku.ItemID]; ok {
			row.ItemName = product.Name
			row.Status = product.Status
			row.LocalImages = product.LocalImages
		}
		result = append(result, row)
	}

	return result, skuCount
}

// getProductsAsFlattenedRows returns products as flattened rows (fallback when no SKUs)
func (h *DBProductHandler) getProductsAsFlattenedRows(db *gorm.DB, offset, limit int) ([]FlattenedSkuRow, int64) {
	var total int64
	db.Model(&models.LazadaProduct{}).Count(&total)

	var products []models.LazadaProduct
	db.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&products)

	result := make([]FlattenedSkuRow, 0, len(products))
	for _, p := range products {
		row := FlattenedSkuRow{
			ItemID:      p.ItemID,
			ItemName:    p.Name,
			Price:       p.Price,
			Quantity:    p.Quantity,
			Status:      p.Status,
			LocalImages: p.LocalImages,
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		result = append(result, row)
	}

	return result, total
}
