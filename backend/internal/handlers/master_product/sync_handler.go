// Package master_product provides sync handlers
package master_product

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	masterProductService "github.com/omni/backend/internal/services/master_product"
	"github.com/rs/zerolog/log"
)

// SyncHandler handles sync-related HTTP requests
type SyncHandler struct {
	basePath string
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(basePath string) *SyncHandler {
	return &SyncHandler{
		basePath: basePath,
	}
}

// SyncRequest represents the sync request body
type SyncRequest struct {
	TargetPlatform string `json:"target_platform" binding:"required"`
	DryRun         bool   `json:"dry_run"`
}

// BackfillImagesRequest represents backfill request body
type BackfillImagesRequest struct {
	Limit int  `json:"limit"`
	Force bool `json:"force"`
}

// Sync handles POST /api/master-products/:id/sync
func (h *SyncHandler) Sync(c *gin.Context) {
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
	syncService := masterProductService.NewSyncService(db, h.basePath)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid product ID"))
		return
	}

	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	// Validate platform
	if req.TargetPlatform != "shopee" && req.TargetPlatform != "tiktok" && req.TargetPlatform != "lazada" {
		c.JSON(http.StatusBadRequest, response.Error("Invalid platform. Must be: shopee, tiktok, or lazada"))
		return
	}

	if req.DryRun {
		// Dry run - just validate without syncing
		c.JSON(http.StatusOK, response.Success(gin.H{
			"dry_run":           true,
			"would_sync_to":     req.TargetPlatform,
			"master_product_id": id,
		}))
		return
	}

	result, err := syncService.SyncToPlatform(c.Request.Context(), tenantID, uint(id), req.TargetPlatform)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Uint64("product_id", id).
			Str("platform", req.TargetPlatform).
			Msg("Sync failed")

		status := http.StatusInternalServerError
		switch err {
		case masterProductService.ErrMasterProductNotFound:
			status = http.StatusNotFound
		case masterProductService.ErrUnsupportedPlatform:
			status = http.StatusBadRequest
		case masterProductService.ErrNoSkusToSync:
			status = http.StatusBadRequest
		}

		c.JSON(status, response.Error(err.Error()))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Uint64("product_id", id).
		Str("platform", req.TargetPlatform).
		Int("skus_synced", result.SkusSynced).
		Msg("Sync completed")

	c.JSON(http.StatusOK, response.Success(result))
}

// GetSyncStatus handles GET /api/master-products/:id/sync-status
func (h *SyncHandler) GetSyncStatus(c *gin.Context) {
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
	syncService := masterProductService.NewSyncService(db, h.basePath)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid product ID"))
		return
	}

	status, err := syncService.GetSyncStatus(c.Request.Context(), tenantID, uint(id))
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Uint64("product_id", id).
			Msg("Failed to get sync status")

		if err == masterProductService.ErrMasterProductNotFound {
			c.JSON(http.StatusNotFound, response.Error("Product not found"))
			return
		}

		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(status))
}

// BackfillImages handles POST /api/master-products/images/backfill
func (h *SyncHandler) BackfillImages(c *gin.Context) {
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

	var req BackfillImagesRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	result, err := service.BackfillImages(c.Request.Context(), tenantID, limit, req.Force)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Int("limit", limit).
			Bool("force", req.Force).
			Msg("Failed to backfill master product images")

		if errors.Is(err, masterProductService.ErrTenantIDRequired) {
			c.JSON(http.StatusUnauthorized, response.Error("Missing tenant ID"))
			return
		}

		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Int("limit", limit).
		Bool("force", req.Force).
		Int("processed", result.Processed).
		Int("updated", result.Updated).
		Int("skipped", result.Skipped).
		Int("errors", result.Errors).
		Msg("Master product image backfill completed")

	c.JSON(http.StatusOK, response.Success(result))
}
