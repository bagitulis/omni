package tiktok

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// ImageHandler handles TikTok product image endpoints
type ImageHandler struct {
	basePath string
}

// NewImageHandler creates a new image handler
func NewImageHandler(basePath string) *ImageHandler {
	return &ImageHandler{basePath: basePath}
}

// UploadImageRequest represents image upload request
type UploadImageRequest struct {
	URL     string `json:"url" binding:"required"`
	UseCase string `json:"use_case"` // MAIN_IMAGE, ATTRIBUTE_IMAGE, etc.
}

// UploadImage handles POST /api/tiktok/products/images/upload
// Connects to real TikTok UploadImage SDK.
func (h *ImageHandler) UploadImage(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	var req UploadImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	if req.UseCase == "" {
		req.UseCase = "MAIN_IMAGE"
	}

	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get TikTok client: "+err.Error()))
		return
	}

	result, err := client.UploadImage(req.URL, req.UseCase)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorWithPlatform("tiktok", "", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"uri":    result.Data.URI,
		"url":    result.Data.URL,
		"width":  result.Data.Width,
		"height": result.Data.Height,
	}))
}

// BatchUploadRequest represents batch image upload request
type BatchUploadRequest struct {
	URLs    []string `json:"urls" binding:"required"`
	UseCase string   `json:"use_case"`
}

// BatchUploadImages handles POST /api/tiktok/products/images/batch-upload
// Uploads multiple images to TikTok CDN.
func (h *ImageHandler) BatchUploadImages(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	var req BatchUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	if req.UseCase == "" {
		req.UseCase = "MAIN_IMAGE"
	}

	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get TikTok client: "+err.Error()))
		return
	}

	type ImageResult struct {
		OriginalURL string `json:"original_url"`
		URI         string `json:"uri,omitempty"`
		URL         string `json:"url,omitempty"`
		Success     bool   `json:"success"`
		Error       string `json:"error,omitempty"`
	}

	results := make([]ImageResult, 0, len(req.URLs))
	for _, url := range req.URLs {
		result, err := client.UploadImage(url, req.UseCase)
		if err != nil {
			results = append(results, ImageResult{
				OriginalURL: url,
				Success:     false,
				Error:       err.Error(),
			})
			continue
		}
		results = append(results, ImageResult{
			OriginalURL: url,
			URI:         result.Data.URI,
			URL:         result.Data.URL,
			Success:     true,
		})
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"results": results,
		"total":   len(results),
	}))
}

// getTiktokClient creates TikTok API client for tenant

// GetUploadTasks handles GET /api/tiktok/products/image-upload-tasks
func (h *ImageHandler) GetUploadTasks(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.Error("TikTok image upload tasks not yet implemented"))
}

// getTiktokClient creates TikTok API client for tenant
func (h *ImageHandler) getTiktokClient(tenantID string) (*tiktokPkg.Client, error) {
	return NewTiktokClient(tenantID, h.basePath)
}
