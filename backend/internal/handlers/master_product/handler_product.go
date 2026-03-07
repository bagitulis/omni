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

// Update handles PUT /api/master-products/:id
// Updates an existing master product
func (h *Handler) Update(c *gin.Context) {
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
	service := masterProductService.NewService(db)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid product ID"))
		return
	}

	var input masterProductService.UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	product, err := service.Update(c.Request.Context(), tenantID, uint(id), input)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Uint64("product_id", id).
			Msg("Failed to update master product")

		// Handle validation errors
		if errors.Is(err, masterProductService.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, response.Error("Product not found"))
			return
		}
		if errors.Is(err, masterProductService.ErrTitleRequired) {
			c.JSON(http.StatusBadRequest, response.Error("Title is required"))
			return
		}
		if errors.Is(err, masterProductService.ErrTitleTooLong) {
			c.JSON(http.StatusBadRequest, response.Error("Title exceeds 120 characters"))
			return
		}
		if errors.Is(err, masterProductService.ErrDescriptionTooLong) {
			c.JSON(http.StatusBadRequest, response.Error("Description exceeds 5000 characters"))
			return
		}
		if errors.Is(err, masterProductService.ErrTooManyImages) {
			c.JSON(http.StatusBadRequest, response.Error("Maximum 8 images allowed"))
			return
		}
		if errors.Is(err, masterProductService.ErrTooManySKUs) {
			c.JSON(http.StatusBadRequest, response.Error("Maximum 50 SKUs allowed"))
			return
		}
		if errors.Is(err, masterProductService.ErrSellerSkuRequired) {
			c.JSON(http.StatusBadRequest, response.Error("Seller SKU is required for each SKU"))
			return
		}
		if errors.Is(err, masterProductService.ErrSkuNotFound) {
			c.JSON(http.StatusNotFound, response.Error("SKU not found"))
			return
		}
		if errors.Is(err, masterProductService.ErrTenantIDRequired) {
			c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
			return
		}

		c.JSON(http.StatusInternalServerError, response.Error("Failed to update product: "+err.Error()))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Uint("product_id", product.ID).
		Msg("Master product updated via API")

	c.JSON(http.StatusOK, response.Success(product))
}

// Delete handles DELETE /api/master-products/:id
// Deletes a master product and its associated SKUs
func (h *Handler) Delete(c *gin.Context) {
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
	service := masterProductService.NewService(db)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid product ID"))
		return
	}

	err = service.Delete(c.Request.Context(), tenantID, uint(id))
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Uint64("product_id", id).
			Msg("Failed to delete master product")

		if errors.Is(err, masterProductService.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, response.Error("Product not found"))
			return
		}
		if errors.Is(err, masterProductService.ErrTenantIDRequired) {
			c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
			return
		}

		c.JSON(http.StatusInternalServerError, response.Error("Failed to delete product: "+err.Error()))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Uint64("product_id", id).
		Msg("Master product deleted via API")

	c.JSON(http.StatusOK, response.Success(gin.H{"deleted": true}))
}
