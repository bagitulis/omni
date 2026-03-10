package shopee

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/request"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// CreateProduct handles POST /api/shopee/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req request.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	// Get Shopee credentials
	client, err := h.getShopeeClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get Shopee client: "+err.Error()))
		return
	}

	// Build Shopee API request
	shopeeReq := shopeePkg.CreateProductRequest{
		ItemName:      req.Name,
		Description:   req.Description,
		OriginalPrice: req.OriginalPrice,
		Weight:        req.Weight,
		CategoryID:    req.CategoryID,
		ItemStatus:    "NORMAL",
		Image: shopeePkg.ImageInfo{
			ImageURLList: req.Images,
		},
	}

	// Call Shopee API
	result, err := client.CreateProduct(shopeeReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorWithPlatform("shopee", "", err.Error()))
		return
	}

	// Save to local database
	product := models.ShopeeProduct{
		ItemID:   result.Response.ItemID,
		Name:     req.Name,
		Price:    req.OriginalPrice, // Map OriginalPrice from req to Price in DB
		Quantity: req.Stock,         // Map Stock from req to Quantity in DB
		Status:   "NORMAL",
	}
	if err := h.saveProduct(tenantID, &product); err != nil {
		log.Printf("[WARN] [Shopee/CreateProduct] Local DB save failed for item %d: %v", result.Response.ItemID, err)
	}

	c.JSON(http.StatusCreated, response.Success(map[string]interface{}{
		"message": "Product created successfully",
		"item_id": result.Response.ItemID,
		"name":    req.Name,
	}))
}

// UpdateProduct handles PUT /api/shopee/products/:itemId
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	itemIDStr := c.Param("itemId")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid item ID"))
		return
	}

	var req request.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	// Get Shopee credentials
	client, err := h.getShopeeClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get Shopee client: "+err.Error()))
		return
	}

	// Build update request
	shopeeReq := shopeePkg.UpdateProductRequest{
		ItemID:        itemID,
		ItemName:      req.Name,
		Description:   req.Description,
		OriginalPrice: req.OriginalPrice,
	}

	// Call Shopee API
	result, err := client.UpdateProduct(shopeeReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorWithPlatform("shopee", "", err.Error()))
		return
	}

	// Update local database
	if err := h.updateLocalProduct(tenantID, itemID, req.Name); err != nil {
		log.Printf("[WARN] [Shopee/UpdateProduct] Local DB update failed for item %d: %v", itemID, err)
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"message": "Product updated successfully",
		"item_id": result.Response.ItemID,
	}))
}

// DeleteProduct handles DELETE /api/shopee/products/:itemId
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	itemIDStr := c.Param("itemId")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid item ID"))
		return
	}

	// Get Shopee credentials
	client, err := h.getShopeeClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get Shopee client: "+err.Error()))
		return
	}

	// Call Shopee API
	result, err := client.DeleteProduct(itemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorWithPlatform("shopee", "", err.Error()))
		return
	}

	// Delete from local database
	if err := h.deleteLocalProduct(tenantID, itemID); err != nil {
		log.Printf("[WARN] [Shopee/DeleteProduct] Local DB delete failed for item %d: %v", itemID, err)
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"message": "Product deleted successfully",
		"item_id": result.Response.ItemID,
	}))
}

// Helper functions

func (h *ProductHandler) getShopeeClient(tenantID string) (*shopeePkg.Client, error) {
	credService := services.NewCredentialService(h.basePath)
	creds, err := credService.GetPlatformCredentials(tenantID, "shopee")
	if err != nil {
		return nil, err
	}

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)
	return client, nil
}

func (h *ProductHandler) saveProduct(tenantID string, product *models.ShopeeProduct) error {
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return err
	}
	return db.Create(product).Error
}

func (h *ProductHandler) updateLocalProduct(tenantID string, itemID int64, name string) error {
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return err
	}
	return db.Model(&models.ShopeeProduct{}).
		Where("item_id = ?", itemID).
		Update("name", name).Error
}

func (h *ProductHandler) deleteLocalProduct(tenantID string, itemID int64) error {
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return err
	}
	return db.Where("item_id = ?", itemID).
		Delete(&models.ShopeeProduct{}).Error
}
