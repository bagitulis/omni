package master_product

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	masterProductService "github.com/omni/backend/internal/services/master_product"
	"github.com/rs/zerolog/log"
)

func (h *ImportHandler) newImportService(tenantID string) (*masterProductService.ImportService, error) {
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return nil, err
	}

	return masterProductService.NewImportService(db, h.basePath), nil
}

func (h *ImportHandler) previewFromFile(c *gin.Context, tenantID string) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("file is required"))
		return
	}

	fileReader, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("failed to open upload file"))
		return
	}
	defer func() {
		_ = fileReader.Close()
	}()

	importService, err := h.newImportService(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	result, err := importService.PreviewFromFile(c.Request.Context(), fileHeader.Filename, fileReader)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Str("filename", fileHeader.Filename).
			Msg("Failed to preview import file")

		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func (h *ImportHandler) importFromRows(c *gin.Context, tenantID string, req FileImportRequest) {
	importService, err := h.newImportService(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	rows := make([]masterProductService.FileImportRow, 0, len(req.Rows))
	for _, row := range req.Rows {
		rows = append(rows, masterProductService.FileImportRow{
			RowNumber:   row.RowNumber,
			ItemName:    row.ItemName,
			ItemSku:     row.ItemSku,
			VariantName: row.VariantName,
			Price:       row.Price,
			Stock:       row.Stock,
			BatchKey:    row.BatchKey,
			Description: row.Description,
			ImageUrls:   row.ImageUrls,
			Valid:       row.Valid,
			Errors:      row.Errors,
		})
	}

	result, err := importService.ImportFromRows(c.Request.Context(), tenantID, rows)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, masterProductService.ErrTenantIDRequired) || strings.Contains(err.Error(), "rows are required") {
			status = http.StatusBadRequest
		}

		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Int("row_count", len(rows)).
			Msg("Failed to import rows")

		c.JSON(status, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// DownloadTemplate handles GET /api/master-products/import/template.
// Supported formats: xlsx (default) and csv.
func (h *ImportHandler) DownloadTemplate(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	importService, err := h.newImportService(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "xlsx")))
	fileBytes, contentType, filename, err := importService.BuildImportTemplate(format)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unsupported template format") {
			c.JSON(http.StatusBadRequest, response.Error(err.Error()))
			return
		}

		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Str("format", format).
			Msg("Failed to build import template")

		c.JSON(http.StatusInternalServerError, response.Error("Failed to build import template"))
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, contentType, fileBytes)
}
