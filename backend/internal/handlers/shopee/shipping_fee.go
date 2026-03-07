package shopee

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/google"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// ShippingFeeHandler handles shipping fee processing and export
type ShippingFeeHandler struct {
	getAPIClient      func(tenantID string) shopeeService.APIClient
	googleAuthService *google.AuthService
}

// NewShippingFeeHandler creates a new shipping fee handler
func NewShippingFeeHandler(getAPIClient func(tenantID string) shopeeService.APIClient, googleAuth *google.AuthService) *ShippingFeeHandler {
	return &ShippingFeeHandler{
		getAPIClient:      getAPIClient,
		googleAuthService: googleAuth,
	}
}

// ProcessShippingFee handles POST /api/shopee/shipping/process-fee
// Processes shipping fees for orders
func (h *ShippingFeeHandler) ProcessShippingFee(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	var req struct {
		OrderSNs []string `json:"order_sn_list" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if len(req.OrderSNs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "order_sn_list is required and must not be empty",
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

	shippingSvc := shopeeService.NewShippingFeeService(client, tenantID)
	results, err := shippingSvc.ProcessShippingFees(c.Request.Context(), req.OrderSNs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Shipping fees processed",
		"data":    results,
	})
}

// ExportShippingFee handles POST /api/shopee/shipping/export-fee
// Exports shipping fee data
func (h *ShippingFeeHandler) ExportShippingFee(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	var req struct {
		Month  int    `json:"month" binding:"required,min=1,max=12"`
		Year   int    `json:"year" binding:"required,min=2020"`
		Format string `json:"format"` // csv, xlsx, json
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing required fields: month and year are required. " + err.Error(),
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

	shippingSvc := shopeeService.NewShippingFeeService(client, tenantID)
	feeData, err := shippingSvc.GetMonthlyShippingFees(c.Request.Context(), req.Month, req.Year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if len(feeData.Fees) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "No shipping fee data found for the selected period",
			"data": gin.H{
				"fees":   []interface{}{},
				"total":  0,
				"count":  0,
				"period": strconv.Itoa(req.Month) + "/" + strconv.Itoa(req.Year),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Shipping fee data ready for export",
		"data": gin.H{
			"fees":   feeData.Fees,
			"total":  feeData.Total,
			"count":  feeData.Count,
			"period": strconv.Itoa(req.Month) + "/" + strconv.Itoa(req.Year),
		},
	})
}

// ExportToSheets handles POST /api/shopee/shipping/export-to-sheets
// Exports shipping fees to Google Sheets
func (h *ShippingFeeHandler) ExportToSheets(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	var req struct {
		Month         int    `json:"month" binding:"required,min=1,max=12"`
		Year          int    `json:"year" binding:"required,min=2020"`
		SpreadsheetID string `json:"spreadsheet_id" binding:"required"`
		SheetName     string `json:"sheet_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing required fields: spreadsheet_id, sheet_name, month, and year are required. " + err.Error(),
		})
		return
	}

	if h.googleAuthService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Google Sheets integration not configured. Service account credentials required.",
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

	// Get shipping fee data
	shippingSvc := shopeeService.NewShippingFeeService(client, tenantID)
	feeData, err := shippingSvc.GetMonthlyShippingFees(c.Request.Context(), req.Month, req.Year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if len(feeData.Fees) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "No shipping fee data to export",
			"data":    gin.H{"exported": 0},
		})
		return
	}

	// Export to Google Sheets
	sheetsService := google.NewSheetsService(h.googleAuthService, tenantID)

	// Prepare data
	headers := []interface{}{"Order SN", "Shipping Fee", "Actual Shipping Fee", "Discount", "Subsidy", "Created Date"}
	values := [][]interface{}{headers}

	for _, fee := range feeData.Fees {
		row := []interface{}{
			fee["order_sn"],
			fee["shipping_fee"],
			fee["actual_shipping_fee"],
			fee["discount"],
			fee["subsidy"],
			fee["created_date"],
		}
		values = append(values, row)
	}

	// Write to sheet
	sheetRange := req.SheetName + "!A1"
	if err := sheetsService.WriteRange(c.Request.Context(), req.SpreadsheetID, sheetRange, values); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to export to Google Sheets: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Shipping fees exported to Google Sheets",
		"data": gin.H{
			"count": feeData.Count,
			"total": feeData.Total,
		},
	})
}
