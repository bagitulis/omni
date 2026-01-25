package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/middleware"
)

// RegisterWholesaleRoutes registers wholesale routes
func RegisterWholesaleRoutes(router *gin.RouterGroup, handler *handlers.WholesaleHandler) {
	wholesale := router.Group("/wholesale")
	wholesale.Use(middleware.Auth())
	wholesale.Use(middleware.Tenant())
	{
		wholesale.GET("/settings", handler.GetSettings)
		wholesale.PUT("/settings", handler.UpdateSettings)
		wholesale.POST("/calculate", handler.Calculate)
		wholesale.POST("/apply", handler.Apply)
	}
}

// RegisterSKUCheckRoutes registers SKU check routes
func RegisterSKUCheckRoutes(router *gin.RouterGroup, handler *handlers.SKUCheckHandler) {
	sku := router.Group("/sku")
	sku.Use(middleware.Auth())
	sku.Use(middleware.Tenant())
	{
		sku.GET("/check/:sku", handler.CheckSingle)
		sku.POST("/batch-check", handler.CheckBatch)
		sku.GET("/cached/:sku", handler.GetCachedStatus)
	}
}
