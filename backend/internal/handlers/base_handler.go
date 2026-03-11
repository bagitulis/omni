package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"gorm.io/gorm"
)

// BaseHandler provides common functionality for all handlers
// Embed this struct in your handlers to reduce code duplication
type BaseHandler struct {
	FallbackDB *gorm.DB
}

// NewBaseHandler creates a new BaseHandler
func NewBaseHandler(fallbackDB *gorm.DB) BaseHandler {
	return BaseHandler{FallbackDB: fallbackDB}
}

// GetDB returns the tenant database connection from context
// This is the standard method for getting DB in handlers
func (h *BaseHandler) GetDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.FallbackDB)
}

// GetTenantID extracts and validates the tenant ID from context
// Returns the tenant ID and true if valid, or sends error response and returns false
func (h *BaseHandler) GetTenantID(c *gin.Context) (string, bool) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		h.ErrorResponse(c, http.StatusUnauthorized, "Missing tenant_id")
		return "", false
	}
	return tenantID, true
}

// ErrorResponse sends a standardized error response
func (h *BaseHandler) ErrorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"error":   message,
	})
}

// ErrorResponseWithErr sends an error response with error details
func (h *BaseHandler) ErrorResponseWithErr(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{
		"success": false,
		"error":   err.Error(),
	})
}

// SuccessResponse sends a standardized success response with data
func (h *BaseHandler) SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// SuccessResponseWithMessage sends a success response with message
func (h *BaseHandler) SuccessResponseWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}

// SuccessResponseRaw sends raw data as JSON response
func (h *BaseHandler) SuccessResponseRaw(c *gin.Context, response gin.H) {
	c.JSON(http.StatusOK, response)
}

// PaginatedResponse sends a paginated response
func (h *BaseHandler) PaginatedResponse(c *gin.Context, data interface{}, total int64, page, limit int) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}
