package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/request"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	shopeeService "github.com/omni/backend/internal/services/shopee"
)

// BulkPrintLabels handles POST /api/orders/bulk-print-labels
// @Summary Bulk print shipping labels for multiple orders
// @Tags Orders
// @Accept json
// @Produce json
// @Param request body request.BulkPrintLabelsRequest true "Bulk print labels request"
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/bulk-print-labels [post]
func (h *OrderManagerHandler) BulkPrintLabels(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req request.BulkPrintLabelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	if len(req.OrderSNs) == 0 {
		c.JSON(http.StatusBadRequest, response.Error("order_sns cannot be empty"))
		return
	}

	// Use real ShippingService
	shippingService := shopeeService.NewShippingServiceWithCreds(tenantID, h.basePath)
	ctx := context.Background()

	labels := []map[string]interface{}{}
	failed := []map[string]interface{}{}

	for _, orderSN := range req.OrderSNs {
		result, err := shippingService.GetShippingLabel(ctx, orderSN, "", "THERMAL_AIR_WAYBILL")
		if err != nil {
			failed = append(failed, map[string]interface{}{
				"order_sn": orderSN,
				"error":    err.Error(),
			})
			continue
		}

		if result.Status == "FAILED" {
			failed = append(failed, map[string]interface{}{
				"order_sn": orderSN,
				"error":    result.ErrorMessage,
			})
			continue
		}

		labels = append(labels, map[string]interface{}{
			"order_sn":  result.OrderSN,
			"file_data": result.FileData, // Base64 encoded PDF
			"status":    result.Status,
		})
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"labels": labels,
		"failed": failed,
		"count":  len(labels),
	}))
}
