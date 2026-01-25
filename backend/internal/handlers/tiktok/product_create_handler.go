package tiktok

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/repositories"
)

// ProductCreateHandler handles TikTok product creation endpoints
type ProductCreateHandler struct {
	basePath string
}

// NewProductCreateHandler creates a new product create handler
func NewProductCreateHandler(basePath string) *ProductCreateHandler {
	return &ProductCreateHandler{basePath: basePath}
}

// SaveDraft handles POST /api/tiktok/products/draft
func (h *ProductCreateHandler) SaveDraft(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var draft ProductDraft
	if err := c.ShouldBindJSON(&draft); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	if draft.ID == "" {
		draft.ID = generateID()
	}
	draft.TenantID = tenantID
	draft.Status = "draft"

	c.JSON(http.StatusOK, response.Success(gin.H{
		"draftId": draft.ID,
		"message": "Draft saved successfully",
	}))
}

// PublishDraft handles POST /api/tiktok/products/publish/:draftId
func (h *ProductCreateHandler) PublishDraft(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	draftID := c.Param("draftId")
	if draftID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing draftId"))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"draftId":   draftID,
		"productId": "published_" + draftID,
		"message":   "Draft published successfully",
	}))
}

// GetProductsFromDB handles GET /api/tiktok/products/db
func (h *ProductCreateHandler) GetProductsFromDB(c *gin.Context) {
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

	c.JSON(http.StatusOK, response.SuccessWithMeta(products, &response.Meta{
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (int(total) + pageSize - 1) / pageSize,
	}))
}

// GetProductFromDB handles GET /api/tiktok/products/db/:productId
func (h *ProductCreateHandler) GetProductFromDB(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	productID := c.Param("productId")
	if productID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing productId"))
		return
	}

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

// SearchProducts handles GET /api/tiktok/products/search
func (h *ProductCreateHandler) SearchProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	query := c.Query("q")
	page, pageSize := parsePagination(c)

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewTiktokProductRepository(db)
	products, total, err := repo.Search(c.Request.Context(), query, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to search products"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMeta(products, &response.Meta{
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (int(total) + pageSize - 1) / pageSize,
	}))
}

// GetCategories handles GET /api/tiktok/products/categories
func (h *ProductCreateHandler) GetCategories(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	parentID := c.Query("parentId")
	categories := getMockCategories(parentID)

	c.JSON(http.StatusOK, response.Success(gin.H{"categories": categories}))
}

// GetAttributes handles GET /api/tiktok/products/categories/:categoryId/attributes
func (h *ProductCreateHandler) GetAttributes(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	categoryID := c.Param("categoryId")
	if categoryID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing categoryId"))
		return
	}

	attributes := getMockAttributes()
	c.JSON(http.StatusOK, response.Success(gin.H{"attributes": attributes}))
}

// GetRules handles GET /api/tiktok/products/categories/:categoryId/rules
func (h *ProductCreateHandler) GetRules(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	categoryID := c.Param("categoryId")
	if categoryID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing categoryId"))
		return
	}

	rules := CategoryRule{
		MaxImages:      9,
		MaxSKUs:        50,
		MaxDescription: 10000,
		RequiredFields: []string{"title", "description", "price", "stock", "images"},
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"rules": rules}))
}

// GetBrands handles GET /api/tiktok/products/brands
func (h *ProductCreateHandler) GetBrands(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	brands := []Brand{
		{ID: "1", Name: "Brand A"},
		{ID: "2", Name: "Brand B"},
		{ID: "3", Name: "No Brand"},
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"brands": brands}))
}

// GetDeliveryOptions handles GET /api/tiktok/products/delivery-options
func (h *ProductCreateHandler) GetDeliveryOptions(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	options := []DeliveryOption{
		{ID: "standard", Name: "Standard Shipping", Type: "standard", MaxWeight: 30, IsAvailable: true},
		{ID: "express", Name: "Express Shipping", Type: "express", MaxWeight: 10, IsAvailable: true},
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"options": options}))
}

// GetWarehouses handles GET /api/tiktok/products/warehouses
func (h *ProductCreateHandler) GetWarehouses(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	warehouses := []Warehouse{
		{ID: "wh1", Name: "Main Warehouse", Address: "Jakarta", IsDefault: true, Status: "active"},
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"warehouses": warehouses}))
}

// Helper functions

func generateID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func getMockCategories(parentID string) []Category {
	if parentID != "" {
		return []Category{
			{ID: parentID + "_1", Name: "Subcategory 1", ParentID: parentID, Level: 2, IsLeaf: true},
			{ID: parentID + "_2", Name: "Subcategory 2", ParentID: parentID, Level: 2, IsLeaf: true},
		}
	}
	return []Category{
		{ID: "1", Name: "Electronics", Level: 1, IsLeaf: false},
		{ID: "2", Name: "Fashion", Level: 1, IsLeaf: false},
		{ID: "3", Name: "Home & Garden", Level: 1, IsLeaf: false},
	}
}

func getMockAttributes() []Attribute {
	return []Attribute{
		{ID: "brand", Name: "Brand", Type: "string", Required: true, InputType: "dropdown"},
		{ID: "color", Name: "Color", Type: "string", Required: false, Options: []string{"Red", "Blue", "Green"}, InputType: "dropdown"},
		{ID: "size", Name: "Size", Type: "string", Required: false, Options: []string{"S", "M", "L", "XL"}, InputType: "dropdown"},
	}
}
