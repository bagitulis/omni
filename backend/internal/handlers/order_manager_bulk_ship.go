package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/request"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
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

	platform := req.Platform
	if platform == "" {
		c.JSON(http.StatusBadRequest, response.Error("platform is required"))
		return
	}

	var shipped []string
	var failed []map[string]interface{}

	switch platform {
	case "shopee":
		shipped, failed = h.bulkShipShopee(tenantID, req.OrderSNs)
	case "tiktok":
		shipped, failed = h.bulkShipTikTok(tenantID, req.OrderSNs)
	case "lazada":
		shipped, failed = h.bulkShipLazada(tenantID, req.OrderSNs)
	default:
		c.JSON(http.StatusBadRequest, response.Error("Unsupported platform: "+platform))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"shipped": shipped,
		"failed":  failed,
	}))
}

func (h *OrderManagerHandler) bulkShipShopee(tenantID string, orderSNs []string) ([]string, []map[string]interface{}) {
	shipped := []string{}
	failed := []map[string]interface{}{}

	client, err := h.getShopeeClient(tenantID)
	if err != nil {
		return nil, []map[string]interface{}{{"error": "Failed to get Shopee client: " + err.Error()}}
	}

	for _, orderSN := range orderSNs {
		if err := h.validateOrderStatus(tenantID, orderSN, "shopee"); err != nil {
			failed = append(failed, map[string]interface{}{"order_sn": orderSN, "error": err.Error()})
			continue
		}

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

		if err := h.updateOrderStatus(tenantID, orderSN, "SHIPPED", "shopee"); err != nil {
			log.Printf("[WARN] [BulkShip/Shopee] Local status update failed for %s: %v (marketplace action succeeded)", orderSN, err)
			failed = append(failed, map[string]interface{}{
				"order_sn": orderSN,
				"error":    "Shipped on marketplace but local DB update failed: " + err.Error(),
			})
			continue
		}

		shipped = append(shipped, orderSN)
	}

	return shipped, failed
}

func (h *OrderManagerHandler) bulkShipTikTok(tenantID string, orderSNs []string) ([]string, []map[string]interface{}) {
	shipped := []string{}
	failed := []map[string]interface{}{}

	client, err := h.getTikTokClient(tenantID)
	if err != nil {
		return nil, []map[string]interface{}{{"error": "Failed to get TikTok client: " + err.Error()}}
	}

	for _, orderSN := range orderSNs {
		if err := h.validateOrderStatus(tenantID, orderSN, "tiktok"); err != nil {
			failed = append(failed, map[string]interface{}{"order_sn": orderSN, "error": err.Error()})
			continue
		}

		pkgID, _, err := client.ResolveOrderToPackageID(orderSN)
		if err != nil {
			failed = append(failed, map[string]interface{}{"order_sn": orderSN, "error": "Failed to resolve package: " + err.Error()})
			continue
		}

		shipReq := &tiktokPkg.ShipPackageRequest{
			HandoverMethod: "PICKUP",
		}
		_, err = client.ArrangeShipment(pkgID, shipReq)
		if err != nil {
			failed = append(failed, map[string]interface{}{"order_sn": orderSN, "error": err.Error()})
			continue
		}

		if err := h.updateOrderStatus(tenantID, orderSN, "AWAITING_COLLECTION", "tiktok"); err != nil {
			log.Printf("[WARN] [BulkShip/TikTok] Local status update failed for %s: %v (marketplace action succeeded)", orderSN, err)
			failed = append(failed, map[string]interface{}{
				"order_sn": orderSN,
				"error":    "Shipped on marketplace but local DB update failed: " + err.Error(),
			})
			continue
		}
		shipped = append(shipped, orderSN)
	}
	return shipped, failed
}

func (h *OrderManagerHandler) bulkShipLazada(tenantID string, orderSNs []string) ([]string, []map[string]interface{}) {
	shipped := []string{}
	failed := []map[string]interface{}{}

	client, err := h.getLazadaClient(tenantID)
	if err != nil {
		return nil, []map[string]interface{}{{"error": "Failed to get Lazada client: " + err.Error()}}
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return nil, []map[string]interface{}{{"error": "DB error: " + err.Error()}}
	}

	for _, orderSN := range orderSNs {
		if err := h.validateOrderStatus(tenantID, orderSN, "lazada"); err != nil {
			failed = append(failed, map[string]interface{}{"order_sn": orderSN, "error": err.Error()})
			continue
		}

		var items []struct {
			ItemID string `gorm:"column:item_id"`
		}
		if err := db.Table("LazadaOrderItem").Select("item_id").Where("order_sn = ?", orderSN).Scan(&items).Error; err != nil {
			failed = append(failed, map[string]interface{}{"order_sn": orderSN, "error": "Failed to get items: " + err.Error()})
			continue
		}

		var order struct {
			ShippingCarrier string `gorm:"column:shipping_carrier"`
			OrderStatus     string `gorm:"column:order_status"`
		}
		if err := db.Table("LazadaOrder").Select("shipping_carrier, order_status").Where("order_sn = ?", orderSN).First(&order).Error; err != nil {
			failed = append(failed, map[string]interface{}{"order_sn": orderSN, "error": "Failed to get order info: " + err.Error()})
			continue
		}

		itemIDs := make([]string, len(items))
		for i, item := range items {
			itemIDs[i] = item.ItemID
		}

		provider := order.ShippingCarrier
		if provider == "" {
			provider = "dropship"
		}

		if order.OrderStatus == "pending" {
			_, err := client.SetStatusToPackedByMarketplace(itemIDs, provider, "dropship")
			if err != nil {
				failed = append(failed, map[string]interface{}{"order_sn": orderSN, "error": "Pack failed: " + err.Error()})
				continue
			}
		}

		_, err = client.SetStatusToReadyToShip(itemIDs, provider, "", "dropship")
		if err != nil {
			failed = append(failed, map[string]interface{}{"order_sn": orderSN, "error": "RTS failed: " + err.Error()})
			continue
		}

		if err := h.updateOrderStatus(tenantID, orderSN, "ready_to_ship", "lazada"); err != nil {
			log.Printf("[WARN] [BulkShip/Lazada] Local status update failed for %s: %v (marketplace action succeeded)", orderSN, err)
			failed = append(failed, map[string]interface{}{
				"order_sn": orderSN,
				"error":    "Shipped on marketplace but local DB update failed: " + err.Error(),
			})
			continue
		}
		shipped = append(shipped, orderSN)
	}
	return shipped, failed
}

// --- Platform Client Helpers (DRY: single credential service creation) ---

// platformOrderTable maps platform names to their order table names.
var platformOrderTable = map[string]string{
	"shopee": "ShopeeOrder",
	"tiktok": "TiktokOrder",
	"lazada": "LazadaOrder",
}

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

func (h *OrderManagerHandler) getTikTokClient(tenantID string) (*tiktokPkg.Client, error) {
	credService := services.NewCredentialService(h.basePath)
	creds, err := credService.GetPlatformCredentials(tenantID, "tiktok")
	if err != nil {
		return nil, err
	}
	client := tiktokPkg.NewClient(creds.AppKey, creds.AppSecret)
	client.SetCredentials(creds.AccessToken, creds.ShopCipher)
	return client, nil
}

func (h *OrderManagerHandler) getLazadaClient(tenantID string) (*lazadaPkg.Client, error) {
	credService := services.NewCredentialService(h.basePath)
	creds, err := credService.GetPlatformCredentials(tenantID, "lazada")
	if err != nil {
		return nil, err
	}
	region := creds.Region
	if region == "" {
		region = "id"
	}
	client := lazadaPkg.NewClient(creds.AppKey, creds.AppSecret, region)
	client.SetAccessToken(creds.AccessToken)
	return client, nil
}

// --- Order Status Helpers ---

func (h *OrderManagerHandler) validateOrderStatus(tenantID, orderSN, platform string) error {
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return err
	}

	table, ok := platformOrderTable[platform]
	if !ok {
		return fmt.Errorf("unsupported platform: %s", platform)
	}

	var status string
	if err := db.Table(table).Select("order_status").Where("order_sn = ?", orderSN).Scan(&status).Error; err != nil {
		return err
	}

	// Platform-specific expected statuses
	expectedStatuses := map[string][]string{
		"shopee": {"READY_TO_SHIP"},
		"tiktok": {"AWAITING_SHIPMENT"},
		"lazada": {"pending", "packed"},
	}

	allowed := expectedStatuses[platform]
	for _, s := range allowed {
		if status == s {
			return nil
		}
	}
	return fmt.Errorf("order status is %s, expected one of %v", status, allowed)
}

func (h *OrderManagerHandler) updateOrderStatus(tenantID, orderSN, status, platform string) error {
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return err
	}

	table, ok := platformOrderTable[platform]
	if !ok {
		return fmt.Errorf("unsupported platform: %s", platform)
	}

	return db.Table(table).Where("order_sn = ?", orderSN).Update("order_status", status).Error
}
