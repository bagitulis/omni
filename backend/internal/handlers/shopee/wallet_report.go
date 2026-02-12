package shopee

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/google"
	shopeeService "github.com/omni/backend/internal/services/shopee"
)

// WalletReportHandler handles wallet reporting and export endpoints
type WalletReportHandler struct {
	getAPIClient      func(tenantID string) shopeeService.APIClient
	googleAuthService *google.AuthService
}

// NewWalletReportHandler creates a new wallet report handler
func NewWalletReportHandler(getAPIClient func(tenantID string) shopeeService.APIClient, googleAuth *google.AuthService) *WalletReportHandler {
	return &WalletReportHandler{
		getAPIClient:      getAPIClient,
		googleAuthService: googleAuth,
	}
}

// GetWalletReport handles POST /api/shopee/wallet/report
// Generates wallet report for specified month/year
func (h *WalletReportHandler) GetWalletReport(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId - authentication required",
		})
		return
	}

	var req struct {
		Month           int    `json:"month" binding:"required,min=1,max=12"`
		Year            int    `json:"year" binding:"required,min=2020"`
		TransactionType string `json:"transaction_type"`
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

	walletSvc := shopeeService.NewWalletService(apiClient, tenantID)
	transactions, err := walletSvc.GetMonthlyTransactions(c.Request.Context(), req.Month, req.Year, req.TransactionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if len(transactions) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "No transactions found for the selected period",
			"data":    []interface{}{},
		})
		return
	}

	// Process transactions
	processed := walletSvc.ProcessTransactions(transactions)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Wallet transactions fetched successfully",
		"data": gin.H{
			"transactions": processed.Transactions,
			"totalAmount":  processed.TotalAmount,
			"count":        processed.Count,
			"orderNumbers": processed.OrderNumbers,
		},
	})
}

// ExportWallet handles POST /api/shopee/wallet/export
// Exports wallet transactions to various formats
func (h *WalletReportHandler) ExportWallet(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId - authentication required",
		})
		return
	}

	var req struct {
		Month           int    `json:"month" binding:"required,min=1,max=12"`
		Year            int    `json:"year" binding:"required,min=2020"`
		TransactionType string `json:"transaction_type"`
		Format          string `json:"format"` // csv, xlsx, json
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

	walletSvc := shopeeService.NewWalletService(apiClient, tenantID)
	transactions, err := walletSvc.GetMonthlyTransactions(c.Request.Context(), req.Month, req.Year, req.TransactionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if len(transactions) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "No transactions found for the selected period",
			"data": gin.H{
				"transactions": []interface{}{},
				"count":        0,
				"totalAmount":  0,
				"period":       strconv.Itoa(req.Month) + "/" + strconv.Itoa(req.Year),
			},
		})
		return
	}

	processed := walletSvc.ProcessTransactions(transactions)

	// Return data for export
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Export data prepared",
		"data": gin.H{
			"transactions": processed.Transactions,
			"count":        processed.Count,
			"totalAmount":  processed.TotalAmount,
			"period":       strconv.Itoa(req.Month) + "/" + strconv.Itoa(req.Year),
		},
	})
}

// ExportToSheets handles POST /api/shopee/wallet/export-to-sheets
// Exports wallet transactions to Google Sheets
func (h *WalletReportHandler) ExportToSheets(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId - authentication required",
		})
		return
	}

	var req struct {
		Month           int    `json:"month" binding:"required,min=1,max=12"`
		Year            int    `json:"year" binding:"required,min=2020"`
		TransactionType string `json:"transaction_type"`
		SpreadsheetID   string `json:"spreadsheet_id" binding:"required"`
		SheetName       string `json:"sheet_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing required fields: spreadsheet_id, sheet_name, month, and year are required. " + err.Error(),
		})
		return
	}

	// Check if Google Auth is configured
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

	// Get transactions
	walletSvc := shopeeService.NewWalletService(apiClient, tenantID)
	transactions, err := walletSvc.GetMonthlyTransactions(c.Request.Context(), req.Month, req.Year, req.TransactionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if len(transactions) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "No transactions to export",
			"data": gin.H{
				"exported": 0,
			},
		})
		return
	}

	processed := walletSvc.ProcessTransactions(transactions)

	// Export to Google Sheets
	sheetsService := google.NewSheetsService(h.googleAuthService, tenantID)

	// Prepare data for sheets
	headers := []interface{}{"Date", "Order SN", "Description", "Amount", "Status", "Transaction Type", "Tab Type", "Buyer Name"}
	values := [][]interface{}{headers}

	for _, tx := range processed.Transactions {
		row := []interface{}{
			tx["date"],
			tx["order_sn"],
			tx["description"],
			tx["amount"],
			tx["status"],
			tx["transaction_type"],
			tx["tab_type"],
			tx["buyer_name"],
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
		"message": "Wallet transactions exported to Google Sheets",
		"data": gin.H{
			"count":       processed.Count,
			"totalAmount": processed.TotalAmount,
		},
	})
}
