package shopee

import (
	"net/http"

	"github.com/gin-gonic/gin"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// EscrowHandler handles escrow detail endpoints
type EscrowHandler struct {
	getAPIClient func(tenantID string) shopeeService.APIClient
}

// NewEscrowHandler creates a new escrow handler
func NewEscrowHandler(getAPIClient func(tenantID string) shopeeService.APIClient) *EscrowHandler {
	return &EscrowHandler{getAPIClient: getAPIClient}
}

// GetEscrowDetail handles POST /api/shopee/wallet/escrow-detail
// Gets single escrow detail for an order
func (h *EscrowHandler) GetEscrowDetail(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	var req struct {
		OrderSN string `json:"order_sn" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	apiClient := h.getAPIClient(tenantID)
	if apiClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Shopee API client not initialized",
		})
		return
	}

	// Type assert to concrete client
	client, ok := apiClient.(*shopeePkg.Client)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invalid Shopee API client type",
		})
		return
	}

	escrowSvc := shopeeService.NewEscrowService(client, tenantID)
	detail, err := escrowSvc.GetEscrowDetail(c.Request.Context(), req.OrderSN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if detail == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "No escrow detail found",
			"data": gin.H{
				"response": nil,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Escrow detail fetched successfully",
		"data": gin.H{
			"response": detail,
		},
	})
}

// GetEscrowDetailBatch handles POST /api/shopee/wallet/escrow-detail-batch
// Gets escrow details for multiple orders
func (h *EscrowHandler) GetEscrowDetailBatch(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	var req struct {
		OrderSNList []string `json:"order_sn_list" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if len(req.OrderSNList) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "order_sn_list is required and must not be empty",
		})
		return
	}

	// Limit batch size
	const maxBatchSize = 50
	if len(req.OrderSNList) > maxBatchSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Maximum batch size is 50 orders",
		})
		return
	}

	apiClient := h.getAPIClient(tenantID)
	if apiClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Shopee API client not initialized",
		})
		return
	}

	// Type assert to concrete client
	client, ok := apiClient.(*shopeePkg.Client)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Invalid Shopee API client type",
		})
		return
	}

	escrowSvc := shopeeService.NewEscrowService(client, tenantID)
	details, err := escrowSvc.GetEscrowDetailsBatch(c.Request.Context(), req.OrderSNList)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Escrow details fetched successfully",
		"data": gin.H{
			"responses":     details,
			"total":         len(details),
			"requested":     len(req.OrderSNList),
			"success_count": len(details),
		},
	})
}
