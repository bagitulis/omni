// Package master_product provides import handlers for Master Product
package master_product

import (
	"encoding/json"
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

	platform := c.Query("platform")
	itemID := c.Query("item_id")
	if platform != "" || itemID != "" {
		h.previewFromPlatform(c, tenantID, platform, itemID)
		return
	}

	h.previewFromFile(c, tenantID)
}

// Import handles POST /api/master-products/import
// Imports a product from platform to Master Product OR from validated file rows.
func (h *ImportHandler) Import(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
		return
	}

	payload, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request payload"))
		return
	}

	var fileReq FileImportRequest
	if err := json.Unmarshal(payload, &fileReq); err == nil && len(fileReq.Rows) > 0 {
		h.importFromRows(c, tenantID, fileReq)
		return
	}

	var req ImportRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	h.importFromPlatform(c, tenantID, req)
}

func (h *ImportHandler) previewFromPlatform(c *gin.Context, tenantID, platform, itemIDStr string) {

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	importService := masterProductService.NewImportService(db, h.basePath)

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

func (h *ImportHandler) importFromPlatform(c *gin.Context, tenantID string, req ImportRequest) {
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}
	importService := masterProductService.NewImportService(db, h.basePath)

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
