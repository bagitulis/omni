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
	ProductID   string           `json:"product_id"`
	SKU         string           `json:"sku"`
	SellerSKU   string           `json:"seller_sku"`
	SkuID       string           `json:"sku_id,omitempty"`
	VariantName string           `json:"variant_name"`
	ItemName    string           `json:"item_name"`
	Image       string           `json:"image"` // Remote image URL (fallback when local_images empty)
	Price       float64          `json:"price"`
	Quantity    int              `json:"quantity"`
	Status      string           `json:"status"`
	LocalImages models.JSONArray `json:"local_images"`
	UpdatedAt   string           `json:"updated_at"`
}

// GetDBProducts handles GET /api/tiktok/db/products
func (h *DBProductHandler) GetDBProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
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

	products, total, err := h.getMasterProducts(db, offset, limit, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database query failed: "+err.Error()))
		return
	}

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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100000"))

	products, total, err := h.getMasterProducts(db, offset, limit, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database query failed: "+err.Error()))
		return
	}

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
func (h *DBProductHandler) getMasterProducts(db *gorm.DB, offset, limit int, tenantID string) ([]MasterProductItem, int64, error) {
	var total int64
	if result := db.Model(&models.TiktokSku{}).Count(&total); result.Error != nil {
		return nil, 0, result.Error
	}

	// Query SKUs with pagination first to know which products we need
	var skus []models.TiktokSku
	if result := db.Offset(offset).Limit(limit).Find(&skus); result.Error != nil {
		return nil, 0, result.Error
	}

	// Query products — limit to those referenced by SKUs for efficiency
	var products []models.TiktokProduct
	if result := db.Find(&products); result.Error != nil {
		return nil, 0, result.Error
	}

	// Build product map by ID
	productMap := make(map[string]models.TiktokProduct)
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
			item.Image = product.Image
			item.LocalImages = product.LocalImages
		}
		result = append(result, item)
	}

	return result, total, nil
}

// GetProductList handles GET /api/tiktok/db/products/list (alias)
func (h *DBProductHandler) GetProductList(c *gin.Context) {
	h.GetDBProducts(c)
}

// GetProductByID handles GET /api/tiktok/db/products/:productId
func (h *DBProductHandler) GetProductByID(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}
	productID := c.Param("productId")
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	var product models.TiktokProduct
	if tx := db.Where("product_id = ?", productID).First(&product); tx.Error != nil {
		c.JSON(http.StatusNotFound, response.Error("Product not found"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": product})
}

// DeleteProduct handles DELETE /api/tiktok/db/products/:productId
func (h *DBProductHandler) DeleteProduct(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}
	productID := c.Param("productId")
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	if tx := db.Where("product_id = ?", productID).Delete(&models.TiktokProduct{}); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to delete product"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetProductsByStatus handles GET /api/tiktok/db/products/status/:status
func (h *DBProductHandler) GetProductsByStatus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}
	status := c.Param("status")
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
	var products []models.TiktokProduct
	var total int64
	if result := db.Model(&models.TiktokProduct{}).Where("status = ?", status).Count(&total); result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to count products"))
		return
	}
	if result := db.Where("status = ?", status).Offset(offset).Limit(limit).Find(&products); result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to query products"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "products": products, "total": total, "count": len(products)})
}

// SearchProducts handles GET /api/tiktok/db/search
func (h *DBProductHandler) SearchProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}
	query := c.Query("q")
	field := c.DefaultQuery("field", "name")
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	var products []models.TiktokProduct
	var dbErr error
	if query != "" && field == "name" {
		dbErr = db.Where("name LIKE ?", "%"+query+"%").Limit(100).Find(&products).Error
	} else if query != "" {
		dbErr = db.Where("product_id = ?", query).Limit(1).Find(&products).Error
	} else {
		dbErr = db.Limit(100).Find(&products).Error
	}
	if dbErr != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to search products"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": products, "count": len(products)})
}

// GetStatistics handles GET /api/tiktok/db/statistics
func (h *DBProductHandler) GetStatistics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	var totalProducts int64
	var activeProducts int64
	var inactiveProducts int64
	if result := db.Model(&models.TiktokProduct{}).Count(&totalProducts); result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to query statistics"))
		return
	}
	if result := db.Model(&models.TiktokProduct{}).Where("status = ?", "ACTIVATE").Count(&activeProducts); result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to query statistics"))
		return
	}
	if result := db.Model(&models.TiktokProduct{}).Where("status != ?", "ACTIVATE").Count(&inactiveProducts); result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to query statistics"))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_products":    totalProducts,
			"active_products":   activeProducts,
			"inactive_products": inactiveProducts,
		},
	})
}
