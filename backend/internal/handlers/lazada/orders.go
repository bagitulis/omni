package lazada

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/repositories"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
)

// OrderHandler handles Lazada order HTTP requests
type OrderHandler struct {
	basePath string
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(basePath string) *OrderHandler {
	return &OrderHandler{basePath: basePath}
}

// GetOrders handles GET /api/lazada/orders
func (h *OrderHandler) GetOrders(c *gin.Context) {
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

	repo := repositories.NewLazadaOrderRepository(db)
	orders, total, err := repo.FindAll(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch orders"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMeta(orders, buildPaginationMeta(int(total), page, pageSize)))
}

// GetOrderByID handles GET /api/lazada/orders/:orderId
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	orderID := c.Param("orderId")

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewLazadaOrderRepository(db)
	order, err := repo.FindByOrderID(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Order not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(order))
}

// ShipOrderRequest represents ship order request
type ShipOrderRequest struct {
	OrderItemIDs     []string `json:"orderItemIds" binding:"required"`
	ShippingProvider string   `json:"shippingProvider" binding:"required"`
	TrackingNumber   string   `json:"trackingNumber,omitempty"`
}

// ShipOrder handles POST /api/lazada/orders/ship
func (h *OrderHandler) ShipOrder(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req ShipOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	// Get Lazada client from platform config
	client, err := h.getLazadaClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get Lazada client: "+err.Error()))
		return
	}

	// First pack the order
	packResp, err := client.SetStatusToPackedByMarketplace(req.OrderItemIDs, req.ShippingProvider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to pack order: "+err.Error()))
		return
	}
	if packResp.Code != "0" {
		c.JSON(http.StatusBadRequest, response.Error("Pack failed: code "+packResp.Code))
		return
	}

	// Then set ready to ship
	rtsResp, err := client.SetStatusToReadyToShip(req.OrderItemIDs, req.ShippingProvider, req.TrackingNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to set ready to ship: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"success": rtsResp.Code == "0",
		"data":    rtsResp.Data,
	}))
}

// CancelOrderRequest represents cancel order request
type CancelOrderRequest struct {
	OrderItemID  string `json:"orderItemId" binding:"required"`
	ReasonID     string `json:"reasonId" binding:"required"`
	ReasonDetail string `json:"reasonDetail,omitempty"`
}

// CancelOrder handles POST /api/lazada/orders/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	client, err := h.getLazadaClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get Lazada client: "+err.Error()))
		return
	}

	resp, err := client.CancelOrder(req.OrderItemID, req.ReasonDetail, req.ReasonID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to cancel order: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"success": resp.Code == "0",
	}))
}

// getLazadaClient creates Lazada API client for tenant
func (h *OrderHandler) getLazadaClient(tenantID string) (*lazadaPkg.Client, error) {
	ctx := context.Background()
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return nil, err
	}

	platformRepo := repositories.NewPlatformConfigRepository(db)
	platformConfig, err := platformRepo.FindByTenantAndPlatform(ctx, tenantID, "lazada")
	if err != nil {
		return nil, err
	}

	// Get global config for app credentials
	systemDB, err := config.GetSystemDB(h.basePath)
	if err != nil {
		return nil, err
	}
	globalRepo := repositories.NewGlobalConfigRepository(systemDB)
	creds, err := globalRepo.GetLazadaCredentials(ctx)
	if err != nil {
		return nil, err
	}

	client := lazadaPkg.NewClient(creds.AppKey, creds.AppSecret, platformConfig.Region)
	client.SetAccessToken(platformConfig.AccessToken)
	return client, nil
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
