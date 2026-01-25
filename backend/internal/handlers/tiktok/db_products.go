package tiktok

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// DBProductHandler handles TikTok DB product requests
type DBProductHandler struct {
	basePath string
}

// NewDBProductHandler creates a new DB product handler
func NewDBProductHandler(basePath string) *DBProductHandler {
	return &DBProductHandler{basePath: basePath}
}

// MasterProductItem represents a flattened SKU row
type MasterProductItem struct {
	ProductID   string  `json:"product_id"`
	SKU         string  `json:"sku"`
	SellerSKU   string  `json:"seller_sku"`
	SkuID       string  `json:"sku_id,omitempty"`
	VariantName string  `json:"variant_name"`
	ItemName    string  `json:"item_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Status      string  `json:"status"`
	UpdatedAt   string  `json:"updated_at"`
}

// GetDBProducts handles GET /api/tiktok/db/products
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit > 1000 {
		limit = 1000
	}

	products, total := h.getMasterProducts(db, offset, limit, tenantID)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"products": products,
		"total":    total,
		"offset":   offset,
		"limit":    limit,
		"count":    len(products),
	})
}

// GetMasterProducts handles GET /api/tiktok/db/products/master
func (h *DBProductHandler) GetMasterProducts(c *gin.Context) {
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100000"))

	products, total := h.getMasterProducts(db, offset, limit, tenantID)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"products": products,
		"total":    total,
		"offset":   offset,
		"limit":    limit,
		"count":    len(products),
	})
}

// getMasterProducts returns flattened product-SKU rows
// NOTE: tenant_id filter removed - we use schema isolation (per-tenant schema)
// so db connection is already scoped to tenant
func (h *DBProductHandler) getMasterProducts(db *gorm.DB, offset, limit int, tenantID string) ([]MasterProductItem, int64) {
	var total int64
	db.Model(&models.TiktokSku{}).Count(&total)

	// Query products (no tenant_id filter - schema isolation handles this)
	var products []models.TiktokProduct
	db.Find(&products)

	// Query SKUs with pagination
	var skus []models.TiktokSku
	db.Offset(offset).Limit(limit).Find(&skus)

	// Build product map by ID
	productMap := make(map[uint]models.TiktokProduct)
	for _, p := range products {
		productMap[p.ID] = p
	}

	result := make([]MasterProductItem, 0, len(skus))
	for _, sku := range skus {
		item := MasterProductItem{
			SKU:         sku.SellerSku,
			SellerSKU:   sku.SellerSku,
			SkuID:       sku.SkuID,
			VariantName: sku.VariantName,
			Price:       sku.Price,
			Quantity:    sku.Quantity,
			UpdatedAt:   sku.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if product, ok := productMap[sku.ProductID]; ok {
			item.ProductID = product.ProductID
			item.ItemName = product.Name
			item.Status = product.Status
		}
		result = append(result, item)
	}

	return result, total
}
