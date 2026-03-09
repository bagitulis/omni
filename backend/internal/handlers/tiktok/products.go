package tiktok

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/repositories"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// ProductHandler handles TikTok product HTTP requests
type ProductHandler struct {
	basePath string
}

// NewProductHandler creates a new product handler
func NewProductHandler(basePath string) *ProductHandler {
	return &ProductHandler{basePath: basePath}
}

// GetProducts handles GET /api/tiktok/products
func (h *ProductHandler) GetProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	page, pageSize := parsePagination(c)

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewTiktokProductRepository(db)
	products, total, err := repo.FindAll(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch products"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMeta(products, buildPaginationMeta(int(total), page, pageSize)))
}

// GetProductByID handles GET /api/tiktok/products/:productId
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	productID := c.Param("productId")

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewTiktokProductRepository(db)
	product, err := repo.FindByProductID(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Product not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(product))
}

// CreateProduct handles POST /api/tiktok/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req tiktokPkg.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get TikTok client: "+err.Error()))
		return
	}

	resp, err := client.CreateProduct(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to create product: "+err.Error()))
		return
	}

	if resp.Code != 0 {
		c.JSON(http.StatusBadRequest, response.ErrorWithPlatform("tiktok", strconv.Itoa(resp.Code), resp.Message))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"productId": resp.Data.ProductID,
	}))
}

// UpdateProduct handles PUT /api/tiktok/products/:productId
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	productID := c.Param("productId")
	var req tiktokPkg.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	req.ProductID = productID

	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get TikTok client: "+err.Error()))
		return
	}

	resp, err := client.UpdateProduct(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to update product: "+err.Error()))
		return
	}

	if resp.Code != 0 {
		c.JSON(http.StatusBadRequest, response.ErrorWithPlatform("tiktok", strconv.Itoa(resp.Code), resp.Message))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"productId": productID,
	}))
}

// DeleteProduct handles DELETE /api/tiktok/products/:productId
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	productID := c.Param("productId")

	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get TikTok client: "+err.Error()))
		return
	}

	resp, err := client.DeleteProduct(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to delete product: "+err.Error()))
		return
	}

	if resp.Code != 0 {
		c.JSON(http.StatusBadRequest, response.ErrorWithPlatform("tiktok", strconv.Itoa(resp.Code), resp.Message))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"deleted": true,
	}))
}

// getTiktokClient creates TikTok API client for tenant
func (h *ProductHandler) getTiktokClient(tenantID string) (*tiktokPkg.Client, error) {
	ctx := context.Background()
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return nil, err
	}

	// Use PlatformCredentialsRepository for key-value based config (current schema)
	credRepo := repositories.NewPlatformCredentialsRepository(db)
	tenantCreds, err := credRepo.GetTiktokCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if tenantCreds.AccessToken == "" || tenantCreds.ShopCipher == "" {
		return nil, fmt.Errorf("missing TikTok credentials: accessToken or shopCipher not configured")
	}

	// Use tenant credentials for appKey/appSecret if available, otherwise fall back to global
	appKey := tenantCreds.AppKey
	appSecret := tenantCreds.AppSecret

	if appKey == "" || appSecret == "" {
		systemDB, err := config.GetSystemDB(h.basePath)
		if err != nil {
			return nil, err
		}
		globalRepo := repositories.NewGlobalConfigRepository(systemDB)
		globalCreds, err := globalRepo.GetTiktokCredentials(ctx)
		if err != nil {
			return nil, err
		}
		appKey = globalCreds.AppKey
		appSecret = globalCreds.AppSecret
	}

	client := tiktokPkg.NewClient(appKey, appSecret)
	client.SetCredentials(tenantCreds.AccessToken, tenantCreds.ShopCipher)
	return client, nil
}
