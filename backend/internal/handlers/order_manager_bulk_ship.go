package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/request"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services"
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// BulkShipOrders handles POST /api/orders/bulk-ship
// @Summary Bulk ship multiple orders
// @Tags Orders
// @Accept json
// @Produce json
// @Param request body request.BulkShipRequest true "Bulk ship request"
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/bulk-ship [post]
func (h *OrderManagerHandler) BulkShipOrders(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req request.BulkShipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	if len(req.OrderSNs) == 0 {
		c.JSON(http.StatusBadRequest, response.Error("order_sns cannot be empty"))
		return
	}

	// Get Shopee credentials from GlobalConfig
	client, err := h.getShopeeClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get Shopee client: "+err.Error()))
		return
	}

	// Process each order
	shipped := []string{}
	failed := []map[string]interface{}{}

	for _, orderSN := range req.OrderSNs {
		shipReq := shopeePkg.ShipOrderRequest{
			OrderSN: orderSN,
		}

		_, err := client.ShipOrder(shipReq)
		if err != nil {
			failed = append(failed, map[string]interface{}{
				"order_sn": orderSN,
				"error":    err.Error(),
			})
			continue
		}

		// Update local database
		if err := h.updateOrderStatus(tenantID, orderSN, "SHIPPED"); err != nil {
			orderManagerLogger.WithTenantID(tenantID).Error("Failed to update order status: " + err.Error())
			// Log but don't fail - order is already shipped on Shopee
		}

		shipped = append(shipped, orderSN)
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"shipped": shipped,
		"failed":  failed,
	}))
}

// getShopeeClient creates Shopee client with tenant credentials
func (h *OrderManagerHandler) getShopeeClient(tenantID string) (*shopeePkg.Client, error) {
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
func (h *OrderManagerHandler) updateOrderStatus(tenantID, orderSN, status string) error {
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return err
	}
	return db.Exec("UPDATE ShopeeOrder SET order_status = ? WHERE order_sn = ?", status, orderSN).Error
}
