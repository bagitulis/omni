package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/middleware"
)

// RegisterImageRoutes registers image gallery routes
func RegisterImageRoutes(router *gin.RouterGroup, handler *handlers.ImageHandler) {
	images := router.Group("/images")
	images.Use(middleware.Auth())
	images.Use(middleware.Tenant())
	{
		images.POST("/upload", handler.Upload)
		images.GET("/gallery", handler.Gallery)
		images.GET("/:id", handler.GetByID)
		images.DELETE("/:id", handler.Delete)
	}
}
