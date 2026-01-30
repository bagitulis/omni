// Package master_product provides HTTP handlers for Master Product management
package master_product

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	masterProductService "github.com/omni/backend/internal/services/master_product"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Handler handles Master Product HTTP requests
type Handler struct {
	service *masterProductService.Service
}

// NewHandler creates a new Master Product handler
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		service: masterProductService.NewService(db),
	}
}

// List handles GET /api/master-products
// Returns paginated list of master products for the tenant
func (h *Handler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	search := c.Query("search")

	filter := masterProductService.ListFilter{
		Page:   page,
		Limit:  limit,
		Status: status,
		Search: search,
	}

	result, err := h.service.List(c.Request.Context(), tenantID, filter)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Msg("Failed to list master products")

		if errors.Is(err, masterProductService.ErrTenantIDRequired) {
			c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
			return
		}

		c.JSON(http.StatusInternalServerError, response.Error("Failed to list products: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMeta(result.Data, &response.Meta{
		Total:    int(result.Total),
		Page:     result.Page,
		PageSize: result.Limit,
	}))
}

// GetByID handles GET /api/master-products/:id
// Returns a single master product with its SKUs
func (h *Handler) GetByID(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid product ID"))
		return
	}

	product, err := h.service.GetByID(c.Request.Context(), tenantID, uint(id))
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Uint64("product_id", id).
			Msg("Failed to get master product")

		if errors.Is(err, masterProductService.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, response.Error("Product not found"))
			return
		}
		if errors.Is(err, masterProductService.ErrTenantIDRequired) {
			c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
			return
		}

		c.JSON(http.StatusInternalServerError, response.Error("Failed to get product: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(product))
}

// Create handles POST /api/master-products
// Creates a new master product with optional SKUs
func (h *Handler) Create(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
		return
	}

	var input masterProductService.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	product, err := h.service.Create(c.Request.Context(), tenantID, input)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Str("title", input.Title).
			Msg("Failed to create master product")

		// Handle validation errors
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
		if errors.Is(err, masterProductService.ErrTenantIDRequired) {
			c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
			return
		}

		c.JSON(http.StatusInternalServerError, response.Error("Failed to create product: "+err.Error()))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Uint("product_id", product.ID).
		Str("title", product.Title).
		Msg("Master product created via API")

	c.JSON(http.StatusCreated, response.Success(product))
}

// Update handles PUT /api/master-products/:id
// Updates an existing master product
func (h *Handler) Update(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
		return
	}

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

	product, err := h.service.Update(c.Request.Context(), tenantID, uint(id), input)
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
		if errors.Is(err, masterProductService.ErrTenantIDRequired) {
			c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid product ID"))
		return
	}

	err = h.service.Delete(c.Request.Context(), tenantID, uint(id))
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
			c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
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
