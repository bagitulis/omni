package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/request"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	labelSvc "github.com/omni/backend/internal/services/label"
)

// BulkPrintLabels handles POST /api/orders/bulk-print-labels
// Routes to the correct platform's shipping label service per order.
// @Summary Bulk print shipping labels for multiple orders (multi-platform)
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

	labelService := labelSvc.NewLabelService(h.basePath)
	ctx := context.Background()

	labels := []map[string]interface{}{}
	failed := []map[string]interface{}{}

	for _, orderSN := range req.OrderSNs {
		result := labelService.GetLabelWithOptions(
			ctx,
			tenantID,
			orderSN,
			req.Platform,
			labelSvc.LabelOptions{
				IncludeProducts:    req.IncludeProducts,
				TikTokDocumentType: req.TikTokDocumentType,
			},
		)

		if result.Status == "FAILED" {
			failed = append(failed, map[string]interface{}{
				"order_sn": orderSN,
				"platform": result.Platform,
				"error":    result.ErrorMessage,
			})
			continue
		}

		labels = append(labels, map[string]interface{}{
			"order_sn":  result.OrderSN,
			"platform":  result.Platform,
			"file_data": result.FileData,
			"status":    result.Status,
		})
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"labels": labels,
		"failed": failed,
		"count":  len(labels),
	}))
}
