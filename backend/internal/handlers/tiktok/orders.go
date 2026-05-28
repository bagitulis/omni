package tiktok

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/utils"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// OrderHandler handles TikTok order HTTP requests
type OrderHandler struct {
	basePath string
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(basePath string) *OrderHandler {
	return &OrderHandler{basePath: basePath}
}

// GetOrders handles GET /api/tiktok/orders
func (h *OrderHandler) GetOrders(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	page, pageSize, err := utils.ValidatePagination(c.DefaultQuery("page", ""), c.DefaultQuery("pageSize", ""))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewTiktokOrderRepository(db)
	orders, total, err := repo.FindAll(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch orders"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMeta(orders, buildPaginationMeta(int(total), page, pageSize)))
}

// GetOrderByID handles GET /api/tiktok/orders/:orderId
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	orderID := c.Param("orderId")

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewTiktokOrderRepository(db)
	order, err := repo.FindByOrderID(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Order not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(order))
}

// ShipOrderRequest represents ship order request
type ShipOrderRequest struct {
	OrderID          string `json:"order_id" binding:"required"`
	PackageID        string `json:"package_id" binding:"required"`
	ShippingProvider string `json:"shipping_provider,omitempty"`
	TrackingNumber   string `json:"tracking_number,omitempty"`
	HandoverMethod   string `json:"handover_method,omitempty"` // PICKUP or DROP_OFF for TikTok Shipping
}

// ShipOrder handles POST /api/tiktok/orders/ship
func (h *OrderHandler) ShipOrder(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	var req ShipOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get TikTok client: "+err.Error()))
		return
	}

	// Build ship package request
	shipReq := &tiktokPkg.ShipPackageRequest{}

	// Seller Shipping: use tracking number and shipping provider
	if req.TrackingNumber != "" && req.ShippingProvider != "" {
		shipReq.SelfShipment = &tiktokPkg.SelfShipmentInfo{
			TrackingNumber:     req.TrackingNumber,
			ShippingProviderID: req.ShippingProvider,
		}
	} else {
		// TikTok Shipping: use handover method (PICKUP or DROP_OFF)
		if req.HandoverMethod == "" {
			req.HandoverMethod = "PICKUP" // Default to pickup
		}
		shipReq.HandoverMethod = req.HandoverMethod
	}

	resp, err := client.ShipPackage(c.Request.Context(), req.PackageID, shipReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to ship order: "+err.Error()))
		return
	}

	if resp.Code != 0 {
		c.JSON(http.StatusBadRequest, response.ErrorWithPlatform("tiktok", strconv.Itoa(resp.Code), resp.Message))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"package_id": resp.Data.PackageID,
	}))
}

// CancelOrderRequest represents cancel order request
type CancelOrderRequest struct {
	OrderID      string `json:"order_id" binding:"required"`
	CancelReason string `json:"cancel_reason" binding:"required"`
}

// CancelOrder handles POST /api/tiktok/orders/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	var req CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get TikTok client: "+err.Error()))
		return
	}

	cancelReq := tiktokPkg.CancelOrderRequest{
		OrderID:      req.OrderID,
		CancelReason: req.CancelReason,
	}

	resp, err := client.CancelOrder(c.Request.Context(), cancelReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to cancel order: "+err.Error()))
		return
	}

	if resp.Code != 0 {
		c.JSON(http.StatusBadRequest, response.ErrorWithPlatform("tiktok", strconv.Itoa(resp.Code), resp.Message))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"cancelled": true,
	}))
}

// getTiktokClient creates TikTok API client for tenant
func (h *OrderHandler) getTiktokClient(tenantID string) (*tiktokPkg.Client, error) {
	return NewTiktokClient(tenantID, h.basePath)
}

// parsePagination extracts and validates pagination params
func parsePagination(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return
}

// buildPaginationMeta creates pagination metadata
func buildPaginationMeta(total, page, pageSize int) *response.Meta {
	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}
	return &response.Meta{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
