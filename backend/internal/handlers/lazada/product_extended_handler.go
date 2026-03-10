package lazada

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/repositories"
)

// ProductExtendedHandler handles extended Lazada product endpoints
type ProductExtendedHandler struct {
	basePath string
}

// NewProductExtendedHandler creates a new product extended handler
func NewProductExtendedHandler(basePath string) *ProductExtendedHandler {
	return &ProductExtendedHandler{basePath: basePath}
}

// GetProductsFromDB handles GET /api/lazada/products/db
func (h *ProductExtendedHandler) GetProductsFromDB(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewLazadaProductRepository(db)
	products, total, err := repo.FindAll(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch products"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMeta(products, &response.Meta{
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (int(total) + pageSize - 1) / pageSize,
	}))
}

// GetProductFromDB handles GET /api/lazada/products/db/:itemId
func (h *ProductExtendedHandler) GetProductFromDB(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	itemIDStr := c.Param("itemId")
	if itemIDStr == "" {
		c.JSON(http.StatusBadRequest, response.Error("Invalid itemId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewLazadaProductRepository(db)
	product, err := repo.FindByItemID(c.Request.Context(), itemIDStr)
	if err != nil || product == nil {
		c.JSON(http.StatusNotFound, response.Error("Product not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(product))
}

// GetCategories handles GET /api/lazada/products/categories
// Fetches real category tree from Lazada API.
func (h *ProductExtendedHandler) GetCategories(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	client, err := GetLazadaClient(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get Lazada client: "+err.Error()))
		return
	}

	categories, err := client.GetCategoryTree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorWithPlatform("lazada", "", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"categories": categories}))
}

// GetAttributes handles GET /api/lazada/products/attributes/:categoryId
// Fetches real category attributes from Lazada API.
func (h *ProductExtendedHandler) GetAttributes(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	categoryID, err := strconv.ParseInt(c.Param("categoryId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid categoryId"))
		return
	}

	client, err := GetLazadaClient(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get Lazada client: "+err.Error()))
		return
	}

	attributes, err := client.GetCategoryAttributes(categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorWithPlatform("lazada", "", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"categoryId": categoryID,
		"attributes": attributes,
	}))
}
