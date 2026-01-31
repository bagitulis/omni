package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/products"
	"gorm.io/gorm"
)

// ProductCloneHandler handles product clone endpoints
type ProductCloneHandler struct {
	fallbackDB *gorm.DB
	dbPath     string
}

// NewProductCloneHandler creates a new product clone handler
func NewProductCloneHandler(db *gorm.DB) *ProductCloneHandler {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data" // Default path
	}
	return &ProductCloneHandler{fallbackDB: db, dbPath: dbPath}
}

// getDB returns the tenant database with proper schema context
func (h *ProductCloneHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// Clone handles POST /api/products/clone
func (h *ProductCloneHandler) Clone(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	var req products.CloneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate platforms
	validPlatforms := map[string]bool{"shopee": true, "lazada": true, "tiktok": true}
	if !validPlatforms[req.SourcePlatform] || !validPlatforms[req.TargetPlatform] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid platform"})
		return
	}

	if req.SourcePlatform == req.TargetPlatform {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Source and target platform must be different"})
		return
	}

	// Get tenant DB with proper schema context
	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant database"})
		return
	}

	// Use service with credentials support for API calls
	svc := products.NewCloneServiceWithCreds(db, tenantID, h.dbPath)
	result, err := svc.Clone(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetStatus handles GET /api/products/clone/status/:id
func (h *ProductCloneHandler) GetStatus(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	cloneID := c.Param("id")
	if cloneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Clone ID required"})
		return
	}

	// Get tenant DB with proper schema context
	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant database"})
		return
	}

	svc := products.NewCloneService(db, tenantID)
	result, err := svc.GetCloneStatus(c.Request.Context(), cloneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.Status == "not_found" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Clone job not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// BatchClone handles POST /api/products/clone/batch
func (h *ProductCloneHandler) BatchClone(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	var req products.BatchCloneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate platforms
	validPlatforms := map[string]bool{"shopee": true, "lazada": true, "tiktok": true}
	if !validPlatforms[req.SourcePlatform] || !validPlatforms[req.TargetPlatform] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid platform"})
		return
	}

	if req.SourcePlatform == req.TargetPlatform {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Source and target platform must be different"})
		return
	}

	if len(req.SourceItemIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one source item ID required"})
		return
	}

	if len(req.SourceItemIDs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 100 items per batch"})
		return
	}

	// Get tenant DB with proper schema context
	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant database"})
		return
	}

	svc := products.NewCloneService(db, tenantID)
	result, err := svc.BatchClone(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetProductData handles GET /api/clone/product-data
func (h *ProductCloneHandler) GetProductData(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	platform := c.Query("platform")
	sku := c.Query("sku")

	if platform == "" || sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "platform and sku are required"})
		return
	}

	validPlatforms := map[string]bool{"shopee": true, "lazada": true, "tiktok": true}
	if !validPlatforms[platform] {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid platform"})
		return
	}

	// Get tenant DB with proper schema context
	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get tenant database"})
		return
	}

	// Use WithCreds version to support API calls for fetching images
	svc := products.NewCloneServiceWithCreds(db, tenantID, h.dbPath)
	productData, err := svc.GetProductData(c.Request.Context(), platform, sku)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "product": productData})
}

// GetAvailableTargets handles GET /api/clone/available-targets
func (h *ProductCloneHandler) GetAvailableTargets(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	sku := c.Query("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "sku is required"})
		return
	}

	// Get tenant DB with proper schema context
	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get tenant database"})
		return
	}

	svc := products.NewCloneServiceWithCreds(db, tenantID, h.dbPath)
	result, err := svc.GetAvailableTargets(c.Request.Context(), sku)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"sku":     result.SKU,
		"status":  result.Status,
		"sources": result.Sources,
		"targets": result.Targets,
	})
}

// Preview handles GET /api/clone/preview
// Returns preview of clone operation including conflict detection and adjustments
func (h *ProductCloneHandler) Preview(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	// Parse query params
	sourcePlatform := c.Query("source_platform")
	targetPlatform := c.Query("target_platform")
	sourceItemID := c.Query("source_item_id")
	sku := c.Query("sku")

	if sourcePlatform == "" || targetPlatform == "" || sourceItemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "source_platform, target_platform, and source_item_id are required",
		})
		return
	}

	// Validate platforms
	validPlatforms := map[string]bool{"shopee": true, "lazada": true, "tiktok": true}
	if !validPlatforms[sourcePlatform] || !validPlatforms[targetPlatform] {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid platform"})
		return
	}

	if sourcePlatform == targetPlatform {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Source and target platform must be different"})
		return
	}

	// Get tenant DB with proper schema context
	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get tenant database"})
		return
	}

	// Build clone request for conflict check
	req := products.CloneRequest{
		SourcePlatform: sourcePlatform,
		TargetPlatform: targetPlatform,
		SourceItemID:   sourceItemID,
		SKU:            sku,
	}

	svc := products.NewCloneServiceWithCreds(db, tenantID, h.dbPath)
	result, err := svc.CheckConflict(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
