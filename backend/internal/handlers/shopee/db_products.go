package shopee

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

// DBProductHandler handles Shopee DB product requests
type DBProductHandler struct {
	basePath string
}

// NewDBProductHandler creates a new DB product handler
func NewDBProductHandler(basePath string) *DBProductHandler {
	return &DBProductHandler{basePath: basePath}
}

// FlattenedSkuRow represents a flattened SKU row for frontend display
// Format matches frontend productManagerConfig.ts (snake_case)
type FlattenedSkuRow struct {
	ItemID      string           `json:"item_id"`            // Matches frontend columnFields
	ModelID     string           `json:"model_id,omitempty"` // Matches frontend columnFields
	SKU         string           `json:"sku"`                // seller_sku from DB
	ItemName    string           `json:"item_name"`          // Product name from ShopeeProduct
	SKUName     string           `json:"sku_name"`           // Variant name from ShopeeSku
	Price       float64          `json:"price"`              // Current price
	Stock       int              `json:"stock"`              // Quantity/stock
	Status      string           `json:"status"`             // Product status
	LocalImages models.JSONArray `json:"local_images"`       // Local image paths
	UpdatedAt   string           `json:"updated_at"`         // Last update time
}

// GetDBProducts handles GET /api/shopee/db/products
// Returns flattened SKU rows for frontend ProductTable display
// Matches frontend productManagerConfig.ts shopeeConfig columnFields
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

	log.Printf("[Shopee DB] GetDBProducts: found %d rows (total: %d)", len(products), total)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"products": products,
		"total":    total,
		"offset":   offset,
		"limit":    limit,
		"count":    len(products),
	})
}

// GetMasterProducts handles GET /api/shopee/db/products/master
// Returns same flattened SKU rows (alias for compatibility)
func (h *DBProductHandler) GetMasterProducts(c *gin.Context) {
	h.GetDBProducts(c) // Same logic, reuse
}

// getFlattenedSkuRows returns flattened SKU rows for frontend display
// Each row = 1 SKU with product info, matches frontend columnFields
func (h *DBProductHandler) getFlattenedSkuRows(db *gorm.DB, offset, limit int) ([]FlattenedSkuRow, int64) {
	var total int64
	db.Model(&models.ShopeeSku{}).Count(&total)

	// Query all products for lookup
	var products []models.ShopeeProduct
	db.Find(&products)

	// Build product lookup map by ItemID
	productMap := make(map[int64]models.ShopeeProduct)
	for _, p := range products {
		productMap[p.ItemID] = p
	}

	// Query SKUs with pagination
	var skus []models.ShopeeSku
	db.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&skus)

	// Build flattened result
	result := make([]FlattenedSkuRow, 0, len(skus))
	for _, sku := range skus {
		row := FlattenedSkuRow{
			ItemID:    strconv.FormatInt(sku.ItemID, 10),
			SKU:       sku.SellerSku,
			SKUName:   sku.VariantName,
			Price:     sku.Price,
			Stock:     sku.Quantity,
			UpdatedAt: sku.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		// Set ModelID if present
		if sku.ModelID != nil {
			row.ModelID = strconv.FormatInt(*sku.ModelID, 10)
		}
		// Get product info
		if product, ok := productMap[sku.ItemID]; ok {
			row.ItemName = product.Name
			row.Status = product.Status
			row.LocalImages = product.LocalImages
		}
		result = append(result, row)
	}

	return result, total
}
