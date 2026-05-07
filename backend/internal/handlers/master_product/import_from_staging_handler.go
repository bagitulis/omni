// Package master_product provides staging import handlers for Master Product
package master_product

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	masterProductService "github.com/omni/backend/internal/services/master_product"
	"github.com/rs/zerolog/log"
)

// StagingImportHandler handles import-from-staging HTTP requests
type StagingImportHandler struct {
	basePath string
}

// NewStagingImportHandler creates a new staging import handler
func NewStagingImportHandler(basePath string) *StagingImportHandler {
	return &StagingImportHandler{
		basePath: basePath,
	}
}

// ImportFromShopee handles POST /api/master-products/import/from-staging/shopee
// Imports all Shopee staging rows into Master Products (no live API calls)
func (h *StagingImportHandler) ImportFromShopee(c *gin.Context) {
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

	svc := masterProductService.NewStagingImportService(db)
	result, err := svc.ImportFromShopeeStaging(c.Request.Context(), tenantID)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Msg("Failed to import from Shopee staging")
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Int("products_created", result.ProductsCreated).
		Int("skus_created", result.SkusCreated).
		Msg("Shopee staging import completed")

	c.JSON(http.StatusOK, response.Success(result))
}

// ImportFromTiktok handles POST /api/master-products/import/from-staging/tiktok
// Imports all TikTok staging rows into Master Products (no live API calls)
func (h *StagingImportHandler) ImportFromTiktok(c *gin.Context) {
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

	svc := masterProductService.NewStagingImportService(db)
	result, err := svc.ImportFromTiktokStaging(c.Request.Context(), tenantID)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Msg("Failed to import from TikTok staging")
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Int("products_created", result.ProductsCreated).
		Int("skus_created", result.SkusCreated).
		Msg("TikTok staging import completed")

	c.JSON(http.StatusOK, response.Success(result))
}

// ImportFromLazada handles POST /api/master-products/import/from-staging/lazada
// Imports all Lazada staging rows into Master Products (no live API calls)
func (h *StagingImportHandler) ImportFromLazada(c *gin.Context) {
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

	svc := masterProductService.NewStagingImportService(db)
	result, err := svc.ImportFromLazadaStaging(c.Request.Context(), tenantID)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Msg("Failed to import from Lazada staging")
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Int("products_created", result.ProductsCreated).
		Int("skus_created", result.SkusCreated).
		Msg("Lazada staging import completed")

	c.JSON(http.StatusOK, response.Success(result))
}

// CleanupInvalidProducts handles POST /api/master-products/import/cleanup
// Removes master products with invalid titles (platform-prefix-only, empty, garbage).
func (h *StagingImportHandler) CleanupInvalidProducts(c *gin.Context) {
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

	svc := masterProductService.NewStagingImportService(db)
	deleted, err := svc.CleanupInvalidProducts(c.Request.Context(), tenantID)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Msg("Failed to cleanup invalid products")
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Int("deleted", deleted).
		Msg("Invalid products cleanup completed")

	c.JSON(http.StatusOK, response.Success(map[string]int{"deleted": deleted}))
}
