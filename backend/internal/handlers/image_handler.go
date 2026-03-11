package handlers

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/image"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ImageHandler handles image gallery endpoints
type ImageHandler struct {
	fallbackDB     *gorm.DB
	storageService *image.StorageService
	webpService    *image.WebPService
}

// NewImageHandler creates a new image handler
func NewImageHandler(db *gorm.DB) *ImageHandler {
	basePath := os.Getenv("UPLOAD_PATH")
	if basePath == "" {
		basePath = "uploads"
	}
	return &ImageHandler{
		fallbackDB:     db,
		storageService: image.NewStorageService(basePath),
		webpService:    image.NewWebPService(),
	}
}

// getDB returns the tenant database with proper schema context
func (h *ImageHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// Upload handles POST /api/images/upload
// Accepts multipart form with "image" field
func (h *ImageHandler) Upload(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get tenant database"})
		return
	}

	// Get file from form
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No image file provided"})
		return
	}

	// Get optional category (default: gallery)
	category := c.DefaultPostForm("category", "gallery")
	if category != "products" && category != "gallery" {
		category = "gallery"
	}

	// Read file data
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to read uploaded file"})
		return
	}

	// Get image dimensions before conversion
	width, height, _ := image.GetImageDimensions(data)

	// Convert to WebP (graceful - returns original if fails)
	finalData, _ := h.webpService.ConvertToWebP(data)
	ext := ".webp"

	// If WebP conversion failed (still not WebP), try JPEG
	if !image.IsWebP(finalData) {
		jpegData, jpegErr := h.webpService.ConvertToJPEG(data)
		if jpegErr != nil {
			log.Info().Msgf("Warning: JPEG conversion failed for %s: %v", file.Filename, jpegErr)
			// Use original data as-is
			finalData = data
		} else {
			finalData = jpegData
			ext = ".jpg"
		}
	}

	// Generate filename
	originalName := file.Filename
	// If both conversions failed, use original extension
	if !image.IsWebP(finalData) && !image.IsJPEG(finalData) {
		ext = getExtension(originalName)
	}

	baseName := strings.TrimSuffix(originalName, getExtension(originalName))
	filename := baseName + ext

	// Save to local storage
	localPath, err := h.storageService.SaveImage(tenantID, category, filename, finalData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to save image: " + err.Error()})
		return
	}

	// Create database record
	img := &models.Image{
		TenantID:  tenantID,
		LocalPath: localPath,
		Width:     width,
		Height:    height,
		FileSize:  int64(len(finalData)),
	}

	if err := db.Create(img).Error; err != nil {
		// Try to clean up the saved file
		if delErr := h.storageService.DeleteImage(tenantID, localPath); delErr != nil {
			log.Info().Msgf("Warning: failed to clean up image file after DB error: %v", delErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to save image metadata"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    img,
	})
}

// Gallery handles GET /api/images/gallery
// Query params: category, search, page, limit
func (h *ImageHandler) Gallery(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get tenant database"})
		return
	}

	// Parse query params
	category := c.Query("category")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Build query
	query := db.Model(&models.Image{}).Where("tenant_id = ?", tenantID)

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if search != "" {
		query = query.Where("filename ILIKE ?", "%"+search+"%")
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to count images"})
		return
	}

	// Get paginated results
	var images []models.Image
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&images).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch images"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    images,
		"meta": gin.H{
			"total":     total,
			"page":      page,
			"page_size": limit,
			"pages":     (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetByID handles GET /api/images/:id
func (h *ImageHandler) GetByID(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get tenant database"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Image ID required"})
		return
	}

	var img models.Image
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&img).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Image not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    img,
	})
}

// Delete handles DELETE /api/images/:id
func (h *ImageHandler) Delete(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get tenant database"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Image ID required"})
		return
	}

	// Find the image first
	var img models.Image
	if err := db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&img).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Image not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch image"})
		return
	}

	// Delete file from storage
	if err := h.storageService.DeleteImage(tenantID, img.LocalPath); err != nil {
		log.Info().Msgf("Warning: failed to delete image file %s: %v", img.LocalPath, err)
	}

	// Delete database record
	if err := db.Delete(&img).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete image record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Image deleted successfully",
	})
}

// getExtension returns the file extension including the dot
func getExtension(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ""
}
