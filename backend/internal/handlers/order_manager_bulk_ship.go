package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"

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
	ctx := c.Request.Context()
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}
	userID := middleware.GetUserID(c)

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
		results = h.bulkShipShopee(ctx, tenantID, userID, req.OrderSNs)
	case "tiktok":
		results = h.bulkShipTikTok(ctx, tenantID, userID, req.OrderSNs)
	case "lazada":
		results = h.bulkShipLazada(ctx, tenantID, userID, req.OrderSNs)
	default:
		c.JSON(http.StatusBadRequest, response.Error("Unsupported platform: "+platform))
		return
	}

	// Build summary
	summary := buildBulkSummary(results)

	data := map[string]any{
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

func (h *OrderManagerHandler) bulkShipShopee(ctx context.Context, tenantID, userID string, orderSNs []string) []BulkShipItemResult {
	client, err := h.getShopeeClient(tenantID)
	if err != nil {
		return allFailed(orderSNs, "Failed to get Shopee client: "+err.Error())
	}

	return parallelShipOrders(ctx, orderSNs, 3, func(orderSN string) BulkShipItemResult {
		result := BulkShipItemResult{OrderSN: orderSN}
		if cancelled, ok := cancelledBulkShipResult(ctx, orderSN); ok {
			return cancelled
		}
		if err := h.validateOrderStatus(tenantID, orderSN, "shopee"); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			return result
		}
		if cancelled, ok := cancelledBulkShipResult(ctx, orderSN); ok {
			return cancelled
		}
		shipReq := shopeePkg.ShipOrderRequest{OrderSN: orderSN}
		if _, err := client.ShipOrder(shipReq); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			return result
		}
		if err := h.updateOrderStatus(tenantID, orderSN, "SHIPPED", "shopee"); err != nil {
			log.Warn().Str("order_sn", orderSN).Err(err).Msg("[BulkShip/Shopee] Marketplace succeeded but local DB update failed")
			result.Status = "shipped_but_local_failed"
			result.MarketplaceOK = true
			result.Error = "Marketplace succeeded but local DB update failed: " + err.Error()
			return result
		}
		result.Status = "shipped"
		result.Message = "OK"
		log.Debug().Str("tenant_id", tenantID).Str("user_id", userID).Str("order_sn", orderSN).Msg("[BulkShip/Shopee] Order shipped")
		return result
	})
}

func (h *OrderManagerHandler) bulkShipTikTok(ctx context.Context, tenantID, userID string, orderSNs []string) []BulkShipItemResult {
	client, err := h.getTikTokClient(tenantID)
	if err != nil {
		return allFailed(orderSNs, "Failed to get TikTok client: "+err.Error())
	}

	return parallelShipOrders(ctx, orderSNs, 3, func(orderSN string) BulkShipItemResult {
		result := BulkShipItemResult{OrderSN: orderSN}
		if cancelled, ok := cancelledBulkShipResult(ctx, orderSN); ok {
			return cancelled
		}
		if err := h.validateOrderStatus(tenantID, orderSN, "tiktok"); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			return result
		}
		if cancelled, ok := cancelledBulkShipResult(ctx, orderSN); ok {
			return cancelled
		}
		pkgID, _, err := client.ResolveOrderToPackageID(ctx, orderSN)
		if err != nil {
			result.Status = "failed"
			result.Error = "Failed to resolve package: " + err.Error()
			return result
		}
		if cancelled, ok := cancelledBulkShipResult(ctx, orderSN); ok {
			return cancelled
		}
		shipReq := &tiktokPkg.ShipPackageRequest{HandoverMethod: "PICKUP"}
		if _, err = client.ArrangeShipment(ctx, pkgID, shipReq); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			return result
		}
		if err := h.updateOrderStatus(tenantID, orderSN, "AWAITING_COLLECTION", "tiktok"); err != nil {
			log.Warn().Str("order_sn", orderSN).Err(err).Msg("[BulkShip/TikTok] Marketplace succeeded but local DB update failed")
			result.Status = "shipped_but_local_failed"
			result.MarketplaceOK = true
			result.Error = "Marketplace succeeded but local DB update failed: " + err.Error()
			return result
		}
		result.Status = "shipped"
		result.Message = "OK"
		log.Debug().Str("tenant_id", tenantID).Str("user_id", userID).Str("order_sn", orderSN).Msg("[BulkShip/TikTok] Order shipped")
		return result
	})
}

func (h *OrderManagerHandler) bulkShipLazada(ctx context.Context, tenantID, userID string, orderSNs []string) []BulkShipItemResult {
	client, err := h.getLazadaClient(tenantID)
	if err != nil {
		return allFailed(orderSNs, "Failed to get Lazada client: "+err.Error())
	}
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return allFailed(orderSNs, "DB error: "+err.Error())
	}

	return parallelShipOrders(ctx, orderSNs, 3, func(orderSN string) BulkShipItemResult {
		result := BulkShipItemResult{OrderSN: orderSN}
		if cancelled, ok := cancelledBulkShipResult(ctx, orderSN); ok {
			return cancelled
		}
		if err := h.validateOrderStatus(tenantID, orderSN, "lazada"); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			return result
		}
		if cancelled, ok := cancelledBulkShipResult(ctx, orderSN); ok {
			return cancelled
		}
		itemIDs, err := fetchLazadaItemIDsFromDB(db, orderSN)
		if err != nil {
			result.Status = "failed"
			result.Error = "Failed to get items: " + err.Error()
			return result
		}
		orderInfo, err := fetchLazadaOrderInfoFromDB(db, orderSN)
		if err != nil {
			result.Status = "failed"
			result.Error = "Failed to get order info: " + err.Error()
			return result
		}
		provider := orderInfo.ShippingCarrier
		if provider == "" {
			provider = "JNE"
		}
		if orderInfo.OrderStatus == "pending" {
			if cancelled, ok := cancelledBulkShipResult(ctx, orderSN); ok {
				return cancelled
			}
			packResp, err := client.SetStatusToPackedByMarketplace(itemIDs, provider, "dropship")
			if err != nil {
				result.Status = "failed"
				result.Error = "Pack failed: " + err.Error()
				return result
			}
			if packResp.Code != "0" && packResp.Code != "" {
				result.Status = "failed"
				result.Error = fmt.Sprintf("Pack failed (code %s): %s", packResp.Code, packResp.Message)
				return result
			}
		}
		if cancelled, ok := cancelledBulkShipResult(ctx, orderSN); ok {
			return cancelled
		}
		rtsResp, err := client.SetStatusToReadyToShip(itemIDs, provider, "", "dropship")
		if err != nil {
			result.Status = "failed"
			result.Error = "RTS failed: " + err.Error()
			return result
		}
		if rtsResp.Code != "0" && rtsResp.Code != "" {
			result.Status = "failed"
			result.Error = fmt.Sprintf("RTS failed (code %s): %s", rtsResp.Code, rtsResp.Message)
			return result
		}
		if err := h.updateOrderStatus(tenantID, orderSN, "ready_to_ship", "lazada"); err != nil {
			log.Warn().Str("order_sn", orderSN).Err(err).Msg("[BulkShip/Lazada] Marketplace succeeded but local DB update failed")
			result.Status = "shipped_but_local_failed"
			result.MarketplaceOK = true
			result.Error = "Marketplace succeeded but local DB update failed: " + err.Error()
			return result
		}
		result.Status = "shipped"
		result.Message = "OK"
		log.Debug().Str("tenant_id", tenantID).Str("user_id", userID).Str("order_sn", orderSN).Msg("[BulkShip/Lazada] Order shipped")
		return result
	})
}

func cancelledBulkShipResult(ctx context.Context, orderSN string) (BulkShipItemResult, bool) {
	select {
	case <-ctx.Done():
		return BulkShipItemResult{OrderSN: orderSN, Status: "cancelled", Error: ctx.Err().Error()}, true
	default:
		return BulkShipItemResult{}, false
	}
}

// parallelShipOrders executes ship operations in parallel with concurrency limit and request cancellation.
func parallelShipOrders(ctx context.Context, orderSNs []string, maxConcurrency int, shipFn func(string) BulkShipItemResult) []BulkShipItemResult {
	results := make([]BulkShipItemResult, len(orderSNs))
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency)

	for i, orderSN := range orderSNs {
		wg.Add(1)
		go func(idx int, sn string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				results[idx] = BulkShipItemResult{OrderSN: sn, Status: "cancelled", Error: ctx.Err().Error()}
				return
			case sem <- struct{}{}:
			}
			defer func() { <-sem }()
			select {
			case <-ctx.Done():
				results[idx] = BulkShipItemResult{OrderSN: sn, Status: "cancelled", Error: ctx.Err().Error()}
				return
			default:
			}
			results[idx] = shipFn(sn)
		}(i, orderSN)
	}

	wg.Wait()
	return results
}
