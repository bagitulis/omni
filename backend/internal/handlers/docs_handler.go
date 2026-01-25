package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DocsHandler handles API documentation endpoints
type DocsHandler struct{}

// NewDocsHandler creates a new docs handler
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// GetAPIInfo returns API information and available endpoints
// @Summary Get API information
// @Description Returns general API info and available endpoints
// @Tags Documentation
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/docs [get]
func (h *DocsHandler) GetAPIInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"name":        "Omni Backend API",
			"version":     "1.0.0",
			"description": "Multi-platform e-commerce management API",
			"docs":        "/api/docs/swagger",
			"endpoints": gin.H{
				"health":       "/api/health",
				"auth":         "/api/auth/*",
				"users":        "/api/users/*",
				"shopee":       "/api/shopee/*",
				"lazada":       "/api/lazada/*",
				"tiktok":       "/api/tiktok/*",
				"inventory":    "/api/inventory/*",
				"analytics":    "/api/analytics/*",
				"webhooks":     "/api/webhooks/*",
				"oauth":        "/api/platform-auth/*",
				"tokens":       "/api/tokens/*",
				"jobs":         "/api/jobs/*",
				"settings":     "/api/settings/*",
			},
			"authentication": gin.H{
				"type":   "Bearer Token (JWT)",
				"header": "Authorization: Bearer <token>",
			},
			"headers": gin.H{
				"x-tenant-id":  "Required for multi-tenant requests",
				"x-csrf-token": "Required for mutating requests (POST, PUT, DELETE)",
				"x-request-id": "Optional, auto-generated if not provided",
			},
		},
	})
}

// GetSwaggerJSON serves the Swagger JSON specification
func (h *DocsHandler) GetSwaggerJSON(c *gin.Context) {
	c.JSON(http.StatusOK, generateSwaggerSpec())
}

// generateSwaggerSpec generates OpenAPI 3.0 specification
func generateSwaggerSpec() map[string]interface{} {
	return map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       "Omni Backend API",
			"description": "Multi-platform e-commerce management API (Shopee, Lazada, TikTok)",
			"version":     "1.0.0",
			"contact": map[string]interface{}{
				"name": "Support",
			},
		},
		"servers": []map[string]interface{}{
			{"url": "/api", "description": "API Server"},
		},
		"components": map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"bearerAuth": map[string]interface{}{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
			"schemas": map[string]interface{}{
				"Error": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": false},
						"error":   map[string]interface{}{"type": "string"},
						"type":    map[string]interface{}{"type": "string"},
					},
				},
				"Success": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data":    map[string]interface{}{"type": "object"},
					},
				},
			},
		},
		"security": []map[string]interface{}{
			{"bearerAuth": []string{}},
		},
		"paths": generatePaths(),
	}
}

func generatePaths() map[string]interface{} {
	return map[string]interface{}{
		"/health": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "Health check",
				"description": "Returns service health status",
				"tags":        []string{"Health"},
				"security":    []map[string]interface{}{},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Service is healthy",
					},
				},
			},
		},
		"/auth/login": map[string]interface{}{
			"post": map[string]interface{}{
				"summary":     "User login",
				"description": "Authenticate user and return JWT token",
				"tags":        []string{"Authentication"},
				"security":    []map[string]interface{}{},
				"requestBody": map[string]interface{}{
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"username": map[string]interface{}{"type": "string"},
									"password": map[string]interface{}{"type": "string"},
								},
								"required": []string{"username", "password"},
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Login successful"},
					"401": map[string]interface{}{"description": "Invalid credentials"},
				},
			},
		},
		"/shopee/orders": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "List Shopee orders",
				"description": "Get paginated list of Shopee orders",
				"tags":        []string{"Shopee"},
				"parameters": []map[string]interface{}{
					{"name": "page", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 1}},
					{"name": "pageSize", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 20}},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Orders list"},
					"401": map[string]interface{}{"description": "Unauthorized"},
				},
			},
		},
		"/lazada/orders": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "List Lazada orders",
				"description": "Get paginated list of Lazada orders",
				"tags":        []string{"Lazada"},
				"parameters": []map[string]interface{}{
					{"name": "page", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 1}},
					{"name": "pageSize", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 20}},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Orders list"},
				},
			},
		},
		"/tiktok/orders": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "List TikTok orders",
				"description": "Get paginated list of TikTok orders",
				"tags":        []string{"TikTok"},
				"parameters": []map[string]interface{}{
					{"name": "page", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 1}},
					{"name": "pageSize", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 20}},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Orders list"},
				},
			},
		},
	}
}
