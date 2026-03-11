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

func (h *ImportHandler) newSkuMapper(tenantID string) (*masterProductService.SkuMapper, error) {
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return nil, err
	}

	return masterProductService.NewSkuMapper(db, tenantID), nil
}

// GetMappingStatus handles GET /api/master-products/:id/mapping
// Returns SKU mapping status for a product
func (h *ImportHandler) GetMappingStatus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	mapper, err := h.newSkuMapper(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	mapper, err := h.newSkuMapper(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	mapper, err := h.newSkuMapper(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	var req ManualLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	err = mapper.ManualLink(
		c.Request.Context(),
		req.MasterSkuID,
		req.Platform,
		req.PlatformItemID,
		req.PlatformSkuID,
	)
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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	mapper, err := h.newSkuMapper(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

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
