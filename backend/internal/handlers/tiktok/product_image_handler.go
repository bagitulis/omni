package tiktok

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
)

// ProductImageHandler handles TikTok product image endpoints
type ProductImageHandler struct {
	basePath string
}

// NewProductImageHandler creates a new product image handler
func NewProductImageHandler(basePath string) *ProductImageHandler {
	return &ProductImageHandler{basePath: basePath}
}

// UploadImageRequest represents image upload request
type UploadImageRequest struct {
	ImageURL  string `json:"imageUrl,omitempty"`
	ImageData string `json:"imageData,omitempty"`
	UseCase   string `json:"useCase"`
}

// ImageUploadResult represents upload result
type ImageUploadResult struct {
	URI     string `json:"uri"`
	URL     string `json:"url"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Success bool   `json:"success"`
}

// UploadImage handles POST /api/tiktok/products/upload-image
func (h *ProductImageHandler) UploadImage(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req UploadImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	// In production, upload to TikTok API
	result := ImageUploadResult{
		URI:     "tiktok://image/" + generateID(),
		URL:     "https://p16-oec-sg.ibyteimg.com/tos-alisg-i-omjb3jn5vq/sample.png",
		Width:   800,
		Height:  800,
		Success: true,
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// ImageUploadTask represents an image upload task
type ImageUploadTask struct {
	TaskID    string `json:"taskId"`
	Status    string `json:"status"`
	ImageURI  string `json:"imageUri,omitempty"`
	Error     string `json:"error,omitempty"`
	CreatedAt int64  `json:"createdAt"`
}

// GetUploadTasks handles GET /api/tiktok/products/image-upload-tasks
func (h *ProductImageHandler) GetUploadTasks(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	// In production, fetch from database or TikTok API
	tasks := []ImageUploadTask{}

	c.JSON(http.StatusOK, response.Success(gin.H{"tasks": tasks}))
}
