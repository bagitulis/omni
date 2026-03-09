package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/utils/logger"
)

var shippingFilesLogger = logger.Named("ShippingFilesHandler")

// ShippingFilesHandler handles shipping files endpoints
type ShippingFilesHandler struct {
	basePath string
}

// NewShippingFilesHandler creates a new shipping files handler
func NewShippingFilesHandler(basePath string) *ShippingFilesHandler {
	return &ShippingFilesHandler{basePath: basePath}
}

// ProcessShippingFileRequest represents the request for processing a shipping file
type ProcessShippingFileRequest struct {
	Filename string `json:"filename" binding:"required"`
}

// ShippingFilesData represents the files data in the response
type ShippingFilesData struct {
	Files   []string `json:"files"`
	Message string   `json:"message,omitempty"`
}

// GetShippingFiles handles GET /api/shipping/files
func (h *ShippingFilesHandler) GetShippingFiles(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	// Construct shipping directory path: {basePath}/data/{tenantID}/shipping/
	shippingDir := filepath.Join(h.basePath, "data", tenantID, "shipping")

	logger.WithFields(map[string]interface{}{
		"tenant_id":    tenantID,
		"shipping_dir": shippingDir,
	}).Info("Getting shipping files")

	// Check if directory exists
	if _, err := os.Stat(shippingDir); os.IsNotExist(err) {
		// Directory doesn't exist - return empty list with message
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": ShippingFilesData{
				Files:   []string{},
				Message: "No shipping files found",
			},
		})
		return
	}

	// Read directory contents
	entries, err := os.ReadDir(shippingDir)
	if err != nil {
		logger.WithFields(map[string]interface{}{
			"tenant_id": tenantID,
			"error":     err.Error(),
		}).Error("Failed to read shipping directory")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to read shipping directory"})
		return
	}

	// Collect file names (skip directories)
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	// If no files found, return empty list
	if len(files) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": ShippingFilesData{
				Files:   []string{},
				Message: "No shipping files found",
			},
		})
		return
	}

	// Return files list
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": ShippingFilesData{
			Files: files,
		},
	})
}

// ProcessShippingFile handles POST /api/shipping/process-file
func (h *ShippingFilesHandler) ProcessShippingFile(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	var req ProcessShippingFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	shippingFilesLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
		"filename": req.Filename,
	}).Info("Processing shipping file")

	// Validate filename is not empty
	if req.Filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "filename is required"})
		return
	}

	// NOT_IMPLEMENTED: Actual file processing logic not yet built.
	// When implementing: parse shipping file → extract tracking numbers → update order statuses
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error":   "Shipping file processing is not yet implemented",
	})
}
