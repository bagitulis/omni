package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers/ml"
	"github.com/omni/backend/internal/middleware"
)

// RegisterMLReportRoutes registers ML report routes
func RegisterMLReportRoutes(router *gin.RouterGroup, basePath string) {
	handler := ml.NewReportHandler(basePath)

	mlGroup := router.Group("/ml/reports")
	mlGroup.Use(middleware.Auth())
	mlGroup.Use(middleware.Tenant())
	{
		mlGroup.POST("/generate", handler.Generate)
		mlGroup.GET("/:platform/list", handler.List)
		mlGroup.GET("/:platform/latest", handler.GetLatest)
		mlGroup.GET("/:platform/:filename", handler.GetByFilename)
	}
}
