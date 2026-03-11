package shopee

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/request"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services"
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// ShipOrder handles POST /api/shopee/orders/ship
func (h *OrderHandler) ShipOrder(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	var req request.ShipOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	// Get Shopee credentials from GlobalConfig
	client, err := h.getShopeeClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get Shopee client: "+err.Error()))
		return
	}

	// Call Shopee API to ship order
	shipReq := shopeePkg.ShipOrderRequest{
		OrderSN: req.OrderSN,
	}

	if req.AddressID != 0 {
		shipReq.Pickup = &shopeePkg.PickupInfo{
			AddressID:    req.AddressID,
			PickupTimeID: req.PickupTimeID,
		}
	}

	if req.BranchID != 0 {
		shipReq.Dropoff = &shopeePkg.DropoffInfo{
			BranchID: req.BranchID,
		}
	}

	// Check if tracking number provided (non-integrated logistics)
	if req.TrackingNumber != "" {
		shipReq.NonIntegrated = &shopeePkg.NonIntegratedInfo{
			TrackingNumber: req.TrackingNumber,
		}
	}

	result, err := client.ShipOrder(shipReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorWithPlatform("shopee", "", err.Error()))
		return
	}

	// Update local database
	if err := h.updateOrderStatus(tenantID, req.OrderSN, "SHIPPED"); err != nil {
		log.Info().Msgf("[WARN] [Shopee/ShipOrder] Local status update failed for %s: %v", req.OrderSN, err)
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"message":         "Order shipped successfully",
		"order_sn":        result.Response.OrderSN,
		"tracking_number": req.TrackingNumber,
	}))
}

// CancelOrder handles POST /api/shopee/orders/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	var req request.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	// Get Shopee credentials from GlobalConfig
	client, err := h.getShopeeClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get Shopee client: "+err.Error()))
		return
	}

	// Call Shopee API to cancel order
	cancelReq := shopeePkg.CancelOrderRequest{
		OrderSN:      req.OrderSN,
		CancelReason: req.CancelReason,
	}

	result, err := client.CancelOrder(cancelReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorWithPlatform("shopee", "", err.Error()))
		return
	}

	// Update local database
	if err := h.updateOrderStatus(tenantID, req.OrderSN, "CANCELLED"); err != nil {
		log.Info().Msgf("[WARN] [Shopee/CancelOrder] Local status update failed for %s: %v", req.OrderSN, err)
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"message":  "Order cancelled successfully",
		"order_sn": result.Response.OrderSN,
	}))
}

// getShopeeClient creates Shopee client with tenant credentials
func (h *OrderHandler) getShopeeClient(tenantID string) (*shopeePkg.Client, error) {
	credService := services.NewCredentialService(h.basePath)
	creds, err := credService.GetPlatformCredentials(tenantID, "shopee")
	if err != nil {
		return nil, err
	}

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)
	return client, nil
}

// updateOrderStatus updates order status in local database
func (h *OrderHandler) updateOrderStatus(tenantID, orderSN, status string) error {
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return err
	}
	return db.Exec("UPDATE ShopeeOrder SET order_status = ? WHERE order_sn = ?", status, orderSN).Error
}
