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
	ItemID      string           `json:"item_id"`
	ModelID     string           `json:"model_id,omitempty"`
	SKU         string           `json:"sku"`
	ItemName    string           `json:"item_name"`
	SKUName     string           `json:"sku_name"`
	Image       string           `json:"image"`
	Price       float64          `json:"price"`
	Stock       int              `json:"stock"`
	Status      string           `json:"status"`
	LocalImages models.JSONArray `json:"local_images"`
	UpdatedAt   string           `json:"updated_at"`
}

// GetDBProducts handles GET /api/shopee/db/products
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

// GetProductList handles GET /api/shopee/db/products/list (alias)
func (h *DBProductHandler) GetProductList(c *gin.Context) {
	h.GetDBProducts(c)
}

// GetProductBase handles GET /api/shopee/db/products/base
func (h *DBProductHandler) GetProductBase(c *gin.Context) {
	h.GetDBProducts(c)
}

// GetProductModel handles GET /api/shopee/db/products/model
func (h *DBProductHandler) GetProductModel(c *gin.Context) {
	h.GetDBProducts(c)
}

// GetProductBaseByID handles GET /api/shopee/db/products/base/:itemId
func (h *DBProductHandler) GetProductBaseByID(c *gin.Context) {
	h.GetProductByID(c)
}

// GetProductModelsByID handles GET /api/shopee/db/products/models/:itemId
func (h *DBProductHandler) GetProductModelsByID(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}
	itemIDStr := c.Param("itemId")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid item_id"))
		return
	}
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	var skus []models.ShopeeSku
	db.Where("item_id = ?", itemID).Find(&skus)
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
		if sku.ModelID != nil {
			row.ModelID = strconv.FormatInt(*sku.ModelID, 10)
		}
		result = append(result, row)
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "products": result, "count": len(result)})
}

// GetProductVariationsByID handles GET /api/shopee/db/products/variations/:itemId
func (h *DBProductHandler) GetProductVariationsByID(c *gin.Context) {
	h.GetProductModelsByID(c)
}

// GetProductFull handles GET /api/shopee/db/products/full/:itemId
func (h *DBProductHandler) GetProductFull(c *gin.Context) {
	h.GetProductByID(c)
}

// GetProductByID handles GET /api/shopee/db/products/:itemId
func (h *DBProductHandler) GetProductByID(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}
	itemIDStr := c.Param("itemId")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid item_id"))
		return
	}
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	var product models.ShopeeProduct
	if tx := db.Where("item_id = ?", itemID).First(&product); tx.Error != nil {
		c.JSON(http.StatusNotFound, response.Error("Product not found"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": product})
}

// SearchProducts handles GET /api/shopee/db/products/search
func (h *DBProductHandler) SearchProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}
	sku := c.Query("sku")
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	var skus []models.ShopeeSku
	if sku != "" {
		db.Where("seller_sku LIKE ?", "%"+sku+"%").Find(&skus)
	} else {
		db.Limit(100).Find(&skus)
	}
	result := make([]FlattenedSkuRow, 0, len(skus))
	for _, s := range skus {
		result = append(result, FlattenedSkuRow{
			ItemID:    strconv.FormatInt(s.ItemID, 10),
			SKU:       s.SellerSku,
			SKUName:   s.VariantName,
			Price:     s.Price,
			Stock:     s.Quantity,
			UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "products": result, "count": len(result)})
}

// GetMasterProducts handles GET /api/shopee/db/products/master
func (h *DBProductHandler) GetMasterProducts(c *gin.Context) {
	h.GetDBProducts(c)
}

// getFlattenedSkuRows returns flattened SKU rows for frontend display
func (h *DBProductHandler) getFlattenedSkuRows(db *gorm.DB, offset, limit int) ([]FlattenedSkuRow, int64) {
	var total int64
	db.Model(&models.ShopeeSku{}).Count(&total)

	var products []models.ShopeeProduct
	db.Find(&products)

	productMap := make(map[int64]models.ShopeeProduct)
	for _, p := range products {
		productMap[p.ItemID] = p
	}

	var skus []models.ShopeeSku
	db.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&skus)

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
		if sku.ModelID != nil {
			row.ModelID = strconv.FormatInt(*sku.ModelID, 10)
		}
		if product, ok := productMap[sku.ItemID]; ok {
			row.ItemName = product.Name
			row.Status = product.Status
			row.Image = product.Image
			row.LocalImages = product.LocalImages
		}
		result = append(result, row)
	}

	return result, total
}
