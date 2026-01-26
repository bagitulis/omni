package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// respondUnauthorized sends unauthorized error response
func respondUnauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": msg})
}

// respondBadRequest sends bad request error response
func respondBadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": msg})
}

// respondNotFound sends not found error response
func respondNotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, gin.H{"success": false, "error": msg})
}

// respondInternalError sends internal server error response
func respondInternalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
}

// respondSuccess sends success response with data
// NOTE: Uses both "data" and "items" for frontend compatibility
func respondSuccess(c *gin.Context, data interface{}, count int) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"items":   data,
		"count":   count,
	})
}

// respondWithConfig sends success response with single config
func respondWithConfig(c *gin.Context, config interface{}) {
	c.JSON(http.StatusOK, gin.H{"success": true, "config": config})
}

// respondWithConfigs sends success response with multiple configs
func respondWithConfigs(c *gin.Context, configs interface{}) {
	c.JSON(http.StatusOK, gin.H{"success": true, "configs": configs})
}

// respondCreated sends created response with config
func respondCreated(c *gin.Context, config interface{}) {
	c.JSON(http.StatusCreated, gin.H{"success": true, "config": config})
}

// respondDeleted sends success response for deletion
func respondDeleted(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": msg})
}
