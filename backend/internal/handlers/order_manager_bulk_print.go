package handlers

import (
	"context"
	"net/http"
	"sync"

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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
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

	type labelResult struct {
		index   int
		label   map[string]interface{}
		failed  map[string]interface{}
		success bool
	}

	resultsCh := make([]labelResult, len(req.OrderSNs))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3) // Limit concurrency for external API calls

	for i, orderSN := range req.OrderSNs {
		wg.Add(1)
		go func(idx int, sn string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := labelService.GetLabelWithOptions(
				ctx, tenantID, sn, req.Platform,
				labelSvc.LabelOptions{
					IncludeProducts:    req.IncludeProducts,
					TikTokDocumentType: req.TikTokDocumentType,
				},
			)

			if result.Status == "FAILED" {
				resultsCh[idx] = labelResult{index: idx, failed: map[string]interface{}{
					"order_sn": sn, "platform": result.Platform, "error": result.ErrorMessage,
				}}
			} else {
				resultsCh[idx] = labelResult{index: idx, success: true, label: map[string]interface{}{
					"order_sn": result.OrderSN, "platform": result.Platform,
					"file_data": result.FileData, "status": result.Status,
				}}
			}
		}(i, orderSN)
	}
	wg.Wait()

	labels := []map[string]interface{}{}
	failed := []map[string]interface{}{}
	for _, r := range resultsCh {
		if r.success {
			labels = append(labels, r.label)
		} else if r.failed != nil {
			failed = append(failed, r.failed)
		}
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"labels": labels,
		"failed": failed,
		"count":  len(labels),
	}))
}
