package tiktok

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// CreateHandler handles TikTok product creation endpoints
type CreateHandler struct {
	basePath string
}

// NewCreateHandler creates a new create handler
func NewCreateHandler(basePath string) *CreateHandler {
	return &CreateHandler{basePath: basePath}
}

// SaveDraftRequest represents draft save request
type SaveDraftRequest struct {
	Title       string                       `json:"title" binding:"required"`
	Description string                       `json:"description"`
	CategoryID  string                       `json:"category_id"`
	MainImages  []tiktokPkg.ImageInfo        `json:"main_images"`
	Skus        []tiktokPkg.CreateProductSku `json:"skus"`
	Weight      string                       `json:"weight"`
	WeightUnit  string                       `json:"weight_unit"`
}

// SaveDraft handles POST /api/tiktok/products/draft
// Creates product as draft via real TikTok API.
func (h *CreateHandler) SaveDraft(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	var req SaveDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get TikTok client: "+err.Error()))
		return
	}

	weightUnit := req.WeightUnit
	if weightUnit == "" {
		weightUnit = "KILOGRAM"
	}

	createReq := tiktokPkg.CreateProductRequest{
		Title:       req.Title,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		MainImages:  req.MainImages,
		Skus:        req.Skus,
		PackageWeight: tiktokPkg.PackageWeight{
			Value: req.Weight,
			Unit:  weightUnit,
		},
		SaveMode: "AS_DRAFT",
	}

	resp, err := client.CreateProduct(c.Request.Context(), createReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to create draft: "+err.Error()))
		return
	}

	if resp.Code != 0 {
		c.JSON(http.StatusBadRequest, response.ErrorWithPlatform("tiktok",
			fmt.Sprintf("%d", resp.Code), resp.Message))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"product_id": resp.Data.ProductID,
		"skus":       resp.Data.Skus,
		"message":    "Draft saved successfully",
	}))
}

// PublishDraftRequest represents draft publish request
type PublishDraftRequest struct {
	Title       string                       `json:"title" binding:"required"`
	Description string                       `json:"description"`
	CategoryID  string                       `json:"category_id"`
	MainImages  []tiktokPkg.ImageInfo        `json:"main_images"`
	Skus        []tiktokPkg.CreateProductSku `json:"skus"`
	Weight      string                       `json:"weight"`
	WeightUnit  string                       `json:"weight_unit"`
}

// PublishDraft handles POST /api/tiktok/products/draft/publish
// Creates product as LISTING (published) via real TikTok API.
func (h *CreateHandler) PublishDraft(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	var req PublishDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get TikTok client: "+err.Error()))
		return
	}

	weightUnit := req.WeightUnit
	if weightUnit == "" {
		weightUnit = "KILOGRAM"
	}

	createReq := tiktokPkg.CreateProductRequest{
		Title:       req.Title,
		Description: req.Description,
		CategoryID:  req.CategoryID,
		MainImages:  req.MainImages,
		Skus:        req.Skus,
		PackageWeight: tiktokPkg.PackageWeight{
			Value: req.Weight,
			Unit:  weightUnit,
		},
		SaveMode: "LISTING",
	}

	resp, err := client.CreateProduct(c.Request.Context(), createReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to publish product: "+err.Error()))
		return
	}

	if resp.Code != 0 {
		c.JSON(http.StatusBadRequest, response.ErrorWithPlatform("tiktok",
			fmt.Sprintf("%d", resp.Code), resp.Message))
		return
	}

	c.JSON(http.StatusCreated, response.Success(gin.H{
		"product_id": resp.Data.ProductID,
		"skus":       resp.Data.Skus,
		"message":    "Product published successfully",
	}))
}

// GetCategories handles GET /api/tiktok/products/categories
// Returns 501 — TikTok Category API SDK not yet implemented.
func (h *CreateHandler) GetCategories(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.Error("TikTok Categories API not yet implemented"))
}

// GetAttributes handles GET /api/tiktok/products/attributes
// Returns 501 — TikTok Attributes API SDK not yet implemented.
func (h *CreateHandler) GetAttributes(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.Error("TikTok Attributes API not yet implemented"))
}

// GetBrands handles GET /api/tiktok/products/brands
// Returns 501 — TikTok Brand API SDK not yet implemented.
func (h *CreateHandler) GetBrands(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.Error("TikTok Brands API not yet implemented"))
}

// GetWarehouses handles GET /api/tiktok/products/warehouses
// Connects to real TikTok GetWarehouses SDK (F17).
func (h *CreateHandler) GetWarehouses(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get TikTok client: "+err.Error()))
		return
	}

	warehouseResp, err := client.GetWarehouses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorWithPlatform("tiktok", "", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"warehouses": warehouseResp.Data.Warehouses,
	}))
}

// GetProductsFromDB handles GET /api/tiktok/products/db
func (h *CreateHandler) GetProductsFromDB(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.Error("TikTok product DB listing not yet implemented"))
}

// GetProductFromDB handles GET /api/tiktok/products/db/:productId
func (h *CreateHandler) GetProductFromDB(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.Error("TikTok product DB detail not yet implemented"))
}

// SearchProducts handles GET /api/tiktok/products/search (DB search)
func (h *CreateHandler) SearchProducts(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.Error("TikTok product search not yet implemented"))
}

// GetRules handles GET /api/tiktok/products/categories/:categoryId/rules
func (h *CreateHandler) GetRules(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.Error("TikTok Category Rules API not yet implemented"))
}

// GetDeliveryOptions handles GET /api/tiktok/products/delivery-options
// Returns 501 — TikTok Delivery Options SDK not yet implemented.
func (h *CreateHandler) GetDeliveryOptions(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.Error("TikTok Delivery Options API not yet implemented"))
}

// getTiktokClient creates TikTok API client for tenant
func (h *CreateHandler) getTiktokClient(tenantID string) (*tiktokPkg.Client, error) {
	return NewTiktokClient(tenantID, h.basePath)
}
