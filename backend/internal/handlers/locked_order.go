package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/orders"
	"gorm.io/gorm"
)

// LockedOrderHandler handles locked order endpoints
type LockedOrderHandler struct {
	fallbackDB *gorm.DB
}

// NewLockedOrderHandler creates a new locked order handler
func NewLockedOrderHandler(db *gorm.DB) *LockedOrderHandler {
	return &LockedOrderHandler{fallbackDB: db}
}

// getDB returns the appropriate database for the current request
func (h *LockedOrderHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// SaveLockedOrdersRequest represents the request body
type SaveLockedOrdersRequest struct {
	Items []orders.LockedOrderItem `json:"items"`
}

// SaveLockedOrders saves locked orders
// @Summary Save locked orders
// @Tags LockedOrders
// @Accept json
// @Param body body SaveLockedOrdersRequest true "Locked orders"
// @Success 200 {object} map[string]interface{}
// @Router /api/locked-orders [post]
func (h *LockedOrderHandler) SaveLockedOrders(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var req SaveLockedOrdersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	service := orders.NewLockedOrderService(db)
	count, err := service.SaveLockedOrders(c.Request.Context(), tenantID, req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   count,
		"message": "Locked orders saved successfully",
	})
}

// GetLockedOrders retrieves locked orders
// @Summary Get locked orders
// @Tags LockedOrders
// @Success 200 {object} map[string]interface{}
// @Router /api/locked-orders [get]
func (h *LockedOrderHandler) GetLockedOrders(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	service := orders.NewLockedOrderService(db)
	items, err := service.GetLockedOrders(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	totalQty, _ := service.GetTotalQty(c.Request.Context(), tenantID)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data":     items,
		"count":    len(items),
		"totalQty": totalQty,
	})
}

// ClearLockedOrders clears all locked orders
// @Summary Clear locked orders
// @Tags LockedOrders
// @Success 200 {object} map[string]interface{}
// @Router /api/locked-orders [delete]
func (h *LockedOrderHandler) ClearLockedOrders(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	service := orders.NewLockedOrderService(db)
	if err := service.ClearLockedOrders(c.Request.Context(), tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Locked orders cleared",
	})
}
