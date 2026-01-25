package tiktok

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/repositories"
)

// ProductComplianceHandler handles TikTok product compliance endpoints
type ProductComplianceHandler struct {
	basePath string
}

// NewProductComplianceHandler creates a new product compliance handler
func NewProductComplianceHandler(basePath string) *ProductComplianceHandler {
	return &ProductComplianceHandler{basePath: basePath}
}

// ComplianceStatus represents product compliance status
type ComplianceStatus struct {
	ProductID         string            `json:"productId"`
	Status            string            `json:"status"`
	Violations        []ComplianceViolation `json:"violations,omitempty"`
	LastCheckedAt     int64             `json:"lastCheckedAt"`
	RecommendedAction string            `json:"recommendedAction,omitempty"`
}

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Field       string `json:"field,omitempty"`
	Severity    string `json:"severity"`
}

// GetCompliance handles GET /api/tiktok/products/:productId/compliance
func (h *ProductComplianceHandler) GetCompliance(c *gin.Context) {
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

	// In production, call TikTok API
	status := ComplianceStatus{
		ProductID:     productID,
		Status:        "compliant",
		Violations:    []ComplianceViolation{},
		LastCheckedAt: currentTimestamp(),
	}

	c.JSON(http.StatusOK, response.Success(status))
}

// UpdateComplianceRequest represents compliance update request
type UpdateComplianceRequest struct {
	Action string `json:"action" binding:"required"`
	Fields map[string]interface{} `json:"fields,omitempty"`
}

// UpdateCompliance handles POST /api/tiktok/products/:productId/compliance
func (h *ProductComplianceHandler) UpdateCompliance(c *gin.Context) {
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

	var req UpdateComplianceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	// In production, call TikTok API to update compliance
	c.JSON(http.StatusOK, response.Success(gin.H{
		"productId": productID,
		"action":    req.Action,
		"status":    "updated",
	}))
}

// GlobalProduct represents a global product
type GlobalProduct struct {
	GlobalProductID string   `json:"globalProductId"`
	Title           string   `json:"title"`
	Status          string   `json:"status"`
	Markets         []string `json:"markets"`
	LocalProducts   []string `json:"localProducts,omitempty"`
}

// GetGlobalProducts handles GET /api/tiktok/products/global-products
func (h *ProductComplianceHandler) GetGlobalProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewTiktokProductRepository(db)
	products, _, err := repo.FindAll(c.Request.Context(), 1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch products"))
		return
	}

	// Convert to global products format
	globalProducts := make([]GlobalProduct, 0)
	for _, p := range products {
		globalProducts = append(globalProducts, GlobalProduct{
			GlobalProductID: p.ProductID,
			Title:           p.Name,
			Status:          p.Status,
			Markets:         []string{"ID"},
		})
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"products": globalProducts}))
}

// PublishGlobalRequest represents global publish request
type PublishGlobalRequest struct {
	ProductID string   `json:"productId" binding:"required"`
	Markets   []string `json:"markets" binding:"required"`
}

// PublishGlobal handles POST /api/tiktok/products/publish-global
func (h *ProductComplianceHandler) PublishGlobal(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req PublishGlobalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	// In production, call TikTok API
	c.JSON(http.StatusOK, response.Success(gin.H{
		"productId": req.ProductID,
		"markets":   req.Markets,
		"status":    "publishing",
		"message":   "Product is being published to selected markets",
	}))
}

// currentTimestamp returns current Unix timestamp
func currentTimestamp() int64 {
	return time.Now().Unix()
}
