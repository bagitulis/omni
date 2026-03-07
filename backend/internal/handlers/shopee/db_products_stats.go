package shopee

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
)

// GetDBStats handles GET /api/shopee/db/stats
func (h *DBProductHandler) GetDBStats(c *gin.Context) {
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
	var totalProducts int64
	var totalModels int64
	db.Model(&models.ShopeeProduct{}).Count(&totalProducts)
	db.Model(&models.ShopeeSku{}).Count(&totalModels)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_products":   totalProducts,
			"total_models":     totalModels,
			"total_variations": totalModels,
		},
	})
}

// GetUnprocessedItems handles GET /api/shopee/db/sync/unprocessed
func (h *DBProductHandler) GetUnprocessedItems(c *gin.Context) {
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
	var products []models.ShopeeProduct
	db.Where("status = ?", "UNLIST").Limit(100).Find(&products)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": products, "count": len(products)})
}

// GetItemsWithoutModels handles GET /api/shopee/db/sync/no-models
func (h *DBProductHandler) GetItemsWithoutModels(c *gin.Context) {
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
	var products []models.ShopeeProduct
	db.Where("item_id NOT IN (SELECT DISTINCT item_id FROM shopee_skus)").Limit(100).Find(&products)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": products, "count": len(products)})
}

// GetSyncLogs handles GET /api/shopee/db/sync/logs
func (h *DBProductHandler) GetSyncLogs(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": []gin.H{}, "count": 0})
}

// GetProductsByStatus handles GET /api/shopee/db/products/status/:status
func (h *DBProductHandler) GetProductsByStatus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}
	status := c.Param("status")
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
	var products []models.ShopeeProduct
	var total int64
	db.Model(&models.ShopeeProduct{}).Where("status = ?", status).Count(&total)
	db.Where("status = ?", status).Offset(offset).Limit(limit).Find(&products)
	c.JSON(http.StatusOK, gin.H{"success": true, "products": products, "total": total, "count": len(products)})
}
