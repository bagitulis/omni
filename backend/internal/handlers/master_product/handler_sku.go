// Package master_product provides HTTP handlers for Master Product management
package master_product

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	masterProductService "github.com/omni/backend/internal/services/master_product"
	"github.com/rs/zerolog/log"
)

// UpdateSkuInput represents input for updating a SKU
type UpdateSkuInput struct {
	Price *float64 `json:"price,omitempty"`
	Stock *int     `json:"stock,omitempty"`
}

// UpdateSku handles PUT /api/master-products/skus/:id
// Updates price and stock for a single SKU
func (h *Handler) UpdateSku(c *gin.Context) {
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
	service := masterProductService.NewService(db)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid SKU ID"))
		return
	}

	var input UpdateSkuInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	// Build CreateSkuInput with existing values where not provided
	skuInput := masterProductService.CreateSkuInput{}
	if input.Price != nil {
		skuInput.Price = *input.Price
	}
	if input.Stock != nil {
		skuInput.Stock = *input.Stock
	}

	sku, err := service.UpdateSku(c.Request.Context(), tenantID, uint(id), skuInput)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Uint64("sku_id", id).
			Msg("Failed to update SKU")

		if errors.Is(err, masterProductService.ErrSkuNotFound) {
			c.JSON(http.StatusNotFound, response.Error("SKU not found"))
			return
		}
		if errors.Is(err, masterProductService.ErrTenantIDRequired) {
			c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
			return
		}

		c.JSON(http.StatusInternalServerError, response.Error("Failed to update SKU: "+err.Error()))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Uint64("sku_id", id).
		Msg("SKU updated via API")

	c.JSON(http.StatusOK, response.Success(sku))
}

// BatchUpdateSkusInput represents input for batch updating SKUs
type BatchUpdateSkusInput struct {
	SkuIDs []uint   `json:"sku_ids" binding:"required,min=1"`
	Price  *float64 `json:"price,omitempty"`
	Stock  *int     `json:"stock,omitempty"`
}

// BatchUpdateSkus handles PUT /api/master-products/skus/batch
// Updates price and/or stock for multiple SKUs at once
func (h *Handler) BatchUpdateSkus(c *gin.Context) {
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
	service := masterProductService.NewService(db)

	var input BatchUpdateSkusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	if len(input.SkuIDs) == 0 {
		c.JSON(http.StatusBadRequest, response.Error("At least one SKU ID is required"))
		return
	}

	if input.Price == nil && input.Stock == nil {
		c.JSON(http.StatusBadRequest, response.Error("At least one of price or stock must be provided"))
		return
	}

	batchInput := masterProductService.BatchUpdateSkuInput{
		SkuIDs: input.SkuIDs,
		Price:  input.Price,
		Stock:  input.Stock,
	}

	result, err := service.BatchUpdateSkus(c.Request.Context(), tenantID, batchInput)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Int("sku_count", len(input.SkuIDs)).
			Msg("Failed to batch update SKUs")

		c.JSON(http.StatusInternalServerError, response.Error("Failed to batch update SKUs: "+err.Error()))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Int("updated", result.Updated).
		Int("failed", result.Failed).
		Msg("Batch SKU update completed via API")

	c.JSON(http.StatusOK, response.Success(result))
}
