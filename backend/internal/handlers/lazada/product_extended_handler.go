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

// Category represents a Lazada category
type Category struct {
	ID       int64      `json:"id"`
	Name     string     `json:"name"`
	ParentID int64      `json:"parentId,omitempty"`
	Level    int        `json:"level"`
	IsLeaf   bool       `json:"isLeaf"`
	Children []Category `json:"children,omitempty"`
}

// GetCategories handles GET /api/lazada/products/categories
func (h *ProductExtendedHandler) GetCategories(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	// In production, call Lazada API
	categories := []Category{
		{ID: 1, Name: "Electronics", Level: 1, IsLeaf: false},
		{ID: 2, Name: "Fashion", Level: 1, IsLeaf: false},
		{ID: 3, Name: "Home & Living", Level: 1, IsLeaf: false},
		{ID: 4, Name: "Health & Beauty", Level: 1, IsLeaf: false},
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"categories": categories}))
}

// Attribute represents a Lazada category attribute
type Attribute struct {
	Name           string   `json:"name"`
	Label          string   `json:"label"`
	InputType      string   `json:"inputType"`
	IsMandatory    bool     `json:"isMandatory"`
	IsSaleProp     bool     `json:"isSaleProp"`
	Options        []Option `json:"options,omitempty"`
	AttributeType  string   `json:"attributeType"`
}

// Option represents an attribute option
type Option struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// GetAttributes handles GET /api/lazada/products/attributes/:categoryId
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

	// In production, call Lazada API
	attributes := []Attribute{
		{
			Name:          "brand",
			Label:         "Brand",
			InputType:     "singleSelect",
			IsMandatory:   true,
			IsSaleProp:    false,
			AttributeType: "normal",
		},
		{
			Name:        "color_family",
			Label:       "Color Family",
			InputType:   "singleSelect",
			IsMandatory: false,
			IsSaleProp:  true,
			Options: []Option{
				{Name: "Black", Value: "Black"},
				{Name: "White", Value: "White"},
				{Name: "Red", Value: "Red"},
			},
			AttributeType: "sku",
		},
		{
			Name:          "warranty_type",
			Label:         "Warranty Type",
			InputType:     "singleSelect",
			IsMandatory:   true,
			IsSaleProp:    false,
			AttributeType: "normal",
		},
	}

	_ = categoryID // Would be used to fetch category-specific attributes

	c.JSON(http.StatusOK, response.Success(gin.H{
		"categoryId": categoryID,
		"attributes": attributes,
	}))
}
