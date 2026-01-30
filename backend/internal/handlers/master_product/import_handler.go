// Package master_product provides import handlers for Master Product
package master_product

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	masterProductService "github.com/omni/backend/internal/services/master_product"
	"github.com/rs/zerolog/log"
)

// ImportHandler handles import-related HTTP requests
type ImportHandler struct {
	basePath string
}

// NewImportHandler creates a new import handler
func NewImportHandler(basePath string) *ImportHandler {
	return &ImportHandler{
		basePath: basePath,
	}
}

// Preview handles GET /api/master-products/import/preview
// Shows what will be imported without creating
func (h *ImportHandler) Preview(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	importService := masterProductService.NewImportService(db, h.basePath)

	platform := c.Query("platform")
	itemIDStr := c.Query("item_id")

	if platform == "" || itemIDStr == "" {
		c.JSON(http.StatusBadRequest, response.Error("platform and item_id are required"))
		return
	}

	// Currently only Shopee is supported
	if platform != "shopee" {
		c.JSON(http.StatusBadRequest, response.Error("Only shopee platform is currently supported"))
		return
	}

	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid item_id"))
		return
	}

	result, err := importService.PreviewFromShopee(c.Request.Context(), tenantID, itemID)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Int64("item_id", itemID).
			Msg("Failed to preview import")

		status := http.StatusInternalServerError
		if err == masterProductService.ErrShopeeProductNotFound {
			status = http.StatusNotFound
		} else if err == masterProductService.ErrShopeeNotConfigured {
			status = http.StatusBadRequest
		}

		c.JSON(status, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// Import handles POST /api/master-products/import
// Imports a product from platform to Master Product
func (h *ImportHandler) Import(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	importService := masterProductService.NewImportService(db, h.basePath)

	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	// Currently only Shopee is supported
	if req.Platform != "shopee" {
		c.JSON(http.StatusBadRequest, response.Error("Only shopee platform is currently supported"))
		return
	}

	itemID, err := strconv.ParseInt(req.ItemID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid item_id"))
		return
	}

	result, err := importService.ImportFromShopee(c.Request.Context(), tenantID, itemID)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Int64("item_id", itemID).
			Msg("Failed to import from Shopee")

		status := http.StatusInternalServerError
		switch err {
		case masterProductService.ErrShopeeProductNotFound:
			status = http.StatusNotFound
		case masterProductService.ErrProductAlreadyExists:
			status = http.StatusConflict
		case masterProductService.ErrShopeeNotConfigured:
			status = http.StatusBadRequest
		case masterProductService.ErrNoSellerSku:
			status = http.StatusBadRequest
		}

		c.JSON(status, response.Error(err.Error()))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Uint("master_product_id", result.MasterProduct.ID).
		Int("skus_imported", result.SkusImported).
		Msg("Import completed via API")

	c.JSON(http.StatusCreated, response.Success(result))
}

// GetMappingStatus handles GET /api/master-products/:id/mapping
// Returns SKU mapping status for a product
func (h *ImportHandler) GetMappingStatus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	mapper := masterProductService.NewSkuMapper(db, tenantID)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid product ID"))
		return
	}

	status, err := mapper.GetMappingStatus(c.Request.Context(), uint(id))
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Uint64("product_id", id).
			Msg("Failed to get mapping status")

		if err == masterProductService.ErrProductNotFound {
			c.JSON(http.StatusNotFound, response.Error("Product not found"))
			return
		}

		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(status))
}

// AutoMap handles POST /api/master-products/mapping/auto
// Automatically finds platform SKUs by seller_sku
func (h *ImportHandler) AutoMap(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	mapper := masterProductService.NewSkuMapper(db, tenantID)

	var req AutoMapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	results, err := mapper.AutoMapBySku(c.Request.Context(), req.SellerSku)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Str("seller_sku", req.SellerSku).
			Msg("Failed to auto-map SKU")

		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(results))
}

// ManualLink handles POST /api/master-products/mapping/link
// Manually links a master SKU to a platform SKU
func (h *ImportHandler) ManualLink(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	mapper := masterProductService.NewSkuMapper(db, tenantID)

	var req ManualLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	err = mapper.ManualLink(c.Request.Context(), req.MasterSkuID, req.Platform, req.PlatformItemID, req.PlatformSkuID)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Uint("master_sku_id", req.MasterSkuID).
			Str("platform", req.Platform).
			Msg("Failed to create manual link")

		status := http.StatusInternalServerError
		switch err {
		case masterProductService.ErrSkuNotFound:
			status = http.StatusNotFound
		case masterProductService.ErrSkuAlreadyLinked:
			status = http.StatusConflict
		}

		c.JSON(status, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success(gin.H{"linked": true}))
}

// Unlink handles DELETE /api/master-products/mapping/link
// Removes a platform link from a master SKU
func (h *ImportHandler) Unlink(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	mapper := masterProductService.NewSkuMapper(db, tenantID)

	var req UnlinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	err = mapper.UnlinkSku(c.Request.Context(), req.MasterSkuID, req.Platform)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Uint("master_sku_id", req.MasterSkuID).
			Str("platform", req.Platform).
			Msg("Failed to unlink SKU")

		if err == masterProductService.ErrSkuNotFound {
			c.JSON(http.StatusNotFound, response.Error("SKU not found"))
			return
		}

		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"unlinked": true}))
}
