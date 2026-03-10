package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/request"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
)

// ===== Bulk Ship Result Types =====

// BulkShipItemResult represents the result for a single order in bulk ship.
type BulkShipItemResult struct {
	OrderSN       string `json:"order_sn"`
	Status        string `json:"status"`                   // "shipped", "failed", "shipped_but_local_failed"
	Error         string `json:"error,omitempty"`          // Error message if failed
	Message       string `json:"message,omitempty"`        // Success message
	MarketplaceOK bool   `json:"marketplace_ok,omitempty"` // True if marketplace action succeeded
}

// BulkShipSummary provides a summary of the bulk ship operation.
type BulkShipSummary struct {
	Total   int `json:"total"`
	Shipped int `json:"shipped"`
	Failed  int `json:"failed"`
}

// ===== Main Handler =====

// BulkShipOrders handles POST /api/orders/bulk-ship
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

	var results []BulkShipItemResult

	switch platform {
	case "shopee":
		results = h.bulkShipShopee(tenantID, req.OrderSNs)
	case "tiktok":
		results = h.bulkShipTikTok(tenantID, req.OrderSNs)
	case "lazada":
		results = h.bulkShipLazada(tenantID, req.OrderSNs)
	default:
		c.JSON(http.StatusBadRequest, response.Error("Unsupported platform: "+platform))
		return
	}

	// Build summary
	summary := buildBulkSummary(results)

	data := map[string]interface{}{
		"summary": summary,
		"results": results,
	}

	// Determine HTTP status code
	switch {
	case summary.Failed == 0:
		c.JSON(http.StatusOK, response.Success(data))
	case summary.Shipped == 0:
		c.JSON(http.StatusUnprocessableEntity, response.ErrorWithData("All orders failed to ship", data))
	default:
		c.JSON(207, response.PartialSuccess(data)) // 207 Multi-Status
	}
}

// ===== Platform Bulk Shippers =====

func (h *OrderManagerHandler) bulkShipShopee(tenantID string, orderSNs []string) []BulkShipItemResult {
	results := make([]BulkShipItemResult, 0, len(orderSNs))

	client, err := h.getShopeeClient(tenantID)
	if err != nil {
		return allFailed(orderSNs, "Failed to get Shopee client: "+err.Error())
	}

	for _, orderSN := range orderSNs {
		result := BulkShipItemResult{OrderSN: orderSN}

		if err := h.validateOrderStatus(tenantID, orderSN, "shopee"); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			results = append(results, result)
			continue
		}

		shipReq := shopeePkg.ShipOrderRequest{OrderSN: orderSN}
		if _, err := client.ShipOrder(shipReq); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			results = append(results, result)
			continue
		}

		if err := h.updateOrderStatus(tenantID, orderSN, "SHIPPED", "shopee"); err != nil {
			log.Warn().Str("order_sn", orderSN).Err(err).Msg("[BulkShip/Shopee] Marketplace succeeded but local DB update failed")
			result.Status = "shipped_but_local_failed"
			result.MarketplaceOK = true
			result.Error = "Marketplace succeeded but local DB update failed: " + err.Error()
			results = append(results, result)
			continue
		}

		result.Status = "shipped"
		result.Message = "OK"
		results = append(results, result)
	}

	return results
}

func (h *OrderManagerHandler) bulkShipTikTok(tenantID string, orderSNs []string) []BulkShipItemResult {
	results := make([]BulkShipItemResult, 0, len(orderSNs))

	client, err := h.getTikTokClient(tenantID)
	if err != nil {
		return allFailed(orderSNs, "Failed to get TikTok client: "+err.Error())
	}

	for _, orderSN := range orderSNs {
		result := BulkShipItemResult{OrderSN: orderSN}

		if err := h.validateOrderStatus(tenantID, orderSN, "tiktok"); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			results = append(results, result)
			continue
		}

		pkgID, _, err := client.ResolveOrderToPackageID(orderSN)
		if err != nil {
			result.Status = "failed"
			result.Error = "Failed to resolve package: " + err.Error()
			results = append(results, result)
			continue
		}

		shipReq := &tiktokPkg.ShipPackageRequest{HandoverMethod: "PICKUP"}
		if _, err = client.ArrangeShipment(pkgID, shipReq); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			results = append(results, result)
			continue
		}

		if err := h.updateOrderStatus(tenantID, orderSN, "AWAITING_COLLECTION", "tiktok"); err != nil {
			log.Warn().Str("order_sn", orderSN).Err(err).Msg("[BulkShip/TikTok] Marketplace succeeded but local DB update failed")
			result.Status = "shipped_but_local_failed"
			result.MarketplaceOK = true
			result.Error = "Marketplace succeeded but local DB update failed: " + err.Error()
			results = append(results, result)
			continue
		}

		result.Status = "shipped"
		result.Message = "OK"
		results = append(results, result)
	}
	return results
}

func (h *OrderManagerHandler) bulkShipLazada(tenantID string, orderSNs []string) []BulkShipItemResult {
	results := make([]BulkShipItemResult, 0, len(orderSNs))

	client, err := h.getLazadaClient(tenantID)
	if err != nil {
		return allFailed(orderSNs, "Failed to get Lazada client: "+err.Error())
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return allFailed(orderSNs, "DB error: "+err.Error())
	}

	for _, orderSN := range orderSNs {
		result := BulkShipItemResult{OrderSN: orderSN}

		if err := h.validateOrderStatus(tenantID, orderSN, "lazada"); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			results = append(results, result)
			continue
		}

		// Fetch order items
		itemIDs, err := fetchLazadaItemIDsFromDB(db, orderSN)
		if err != nil {
			result.Status = "failed"
			result.Error = "Failed to get items: " + err.Error()
			results = append(results, result)
			continue
		}

		// Fetch order info for shipping carrier and status
		orderInfo, err := fetchLazadaOrderInfoFromDB(db, orderSN)
		if err != nil {
			result.Status = "failed"
			result.Error = "Failed to get order info: " + err.Error()
			results = append(results, result)
			continue
		}

		provider := orderInfo.ShippingCarrier
		if provider == "" {
			provider = "JNE" // Default carrier for Indonesia
		}

		// Pack first if order is pending
		if orderInfo.OrderStatus == "pending" {
			packResp, err := client.SetStatusToPackedByMarketplace(itemIDs, provider, "dropship")
			if err != nil {
				result.Status = "failed"
				result.Error = "Pack failed: " + err.Error()
				results = append(results, result)
				continue
			}
			if packResp.Code != "0" && packResp.Code != "" {
				result.Status = "failed"
				result.Error = fmt.Sprintf("Pack failed (code %s): %s", packResp.Code, packResp.Message)
				results = append(results, result)
				continue
			}
		}

		// Set ready to ship
		rtsResp, err := client.SetStatusToReadyToShip(itemIDs, provider, "", "dropship")
		if err != nil {
			result.Status = "failed"
			result.Error = "RTS failed: " + err.Error()
			results = append(results, result)
			continue
		}
		if rtsResp.Code != "0" && rtsResp.Code != "" {
			result.Status = "failed"
			result.Error = fmt.Sprintf("RTS failed (code %s): %s", rtsResp.Code, rtsResp.Message)
			results = append(results, result)
			continue
		}

		if err := h.updateOrderStatus(tenantID, orderSN, "ready_to_ship", "lazada"); err != nil {
			log.Warn().Str("order_sn", orderSN).Err(err).Msg("[BulkShip/Lazada] Marketplace succeeded but local DB update failed")
			result.Status = "shipped_but_local_failed"
			result.MarketplaceOK = true
			result.Error = "Marketplace succeeded but local DB update failed: " + err.Error()
			results = append(results, result)
			continue
		}

		result.Status = "shipped"
		result.Message = "OK"
		results = append(results, result)
	}
	return results
}
