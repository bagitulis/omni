package shopee

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	shopeeService "github.com/omni/backend/internal/services/shopee"
)

// WalletHandler handles Shopee wallet endpoints
type WalletHandler struct {
	getAPIClient func(tenantID string) shopeeService.APIClient
}

// NewWalletHandler creates a new wallet handler
func NewWalletHandler(getAPIClient func(tenantID string) shopeeService.APIClient) *WalletHandler {
	return &WalletHandler{getAPIClient: getAPIClient}
}

// GetBalance handles GET /api/shopee/wallet/balance
func (h *WalletHandler) GetBalance(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	apiClient := h.getAPIClient(tenantID)
	svc := shopeeService.NewWalletService(apiClient, tenantID)

	balance, err := svc.GetBalance(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(balance))
}

// GetTransactions handles GET /api/shopee/wallet/transactions
func (h *WalletHandler) GetTransactions(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	// Parse date range from query
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startStr != "" {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.Error("Invalid start_date format"))
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0) // Default: 1 month ago
	}

	if endStr != "" {
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.Error("Invalid end_date format"))
			return
		}
	} else {
		endDate = time.Now()
	}

	apiClient := h.getAPIClient(tenantID)
	svc := shopeeService.NewWalletService(apiClient, tenantID)

	filter := shopeeService.TransactionFilter{
		StartDate: startDate,
		EndDate:   endDate,
		Type:      c.Query("type"),
		PageSize:  50,
		PageToken: c.Query("page_token"),
	}

	result, err := svc.GetTransactions(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// GetNetIncome handles GET /api/shopee/wallet/income
func (h *WalletHandler) GetNetIncome(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startStr != "" {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.Error("Invalid start_date format"))
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0)
	}

	if endStr != "" {
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.Error("Invalid end_date format"))
			return
		}
	} else {
		endDate = time.Now()
	}

	apiClient := h.getAPIClient(tenantID)
	svc := shopeeService.NewWalletService(apiClient, tenantID)

	income, err := svc.CalculateNetIncome(c.Request.Context(), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(income))
}
