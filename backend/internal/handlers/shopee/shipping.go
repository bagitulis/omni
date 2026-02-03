package shopee

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	shopeeService "github.com/omni/backend/internal/services/shopee"
)

// ShippingHandler handles Shopee shipping endpoints
type ShippingHandler struct {
	getAPIClient func(tenantID string) shopeeService.APIClient
}

// NewShippingHandler creates a new shipping handler
func NewShippingHandler(getAPIClient func(tenantID string) shopeeService.APIClient) *ShippingHandler {
	return &ShippingHandler{getAPIClient: getAPIClient}
}

// GetOptions handles GET /api/shopee/shipping/options
func (h *ShippingHandler) GetOptions(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	orderSN := c.Query("orderSn")
	if orderSN == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing orderSn parameter"))
		return
	}

	apiClient := h.getAPIClient(tenantID)
	svc := shopeeService.NewShippingService(apiClient, tenantID)

	options, err := svc.GetShippingOptions(c.Request.Context(), orderSN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(options))
}

// ArrangeShipment handles POST /api/shopee/shipping/arrange
func (h *ShippingHandler) ArrangeShipment(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req shopeeService.ArrangeShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	apiClient := h.getAPIClient(tenantID)
	svc := shopeeService.NewShippingService(apiClient, tenantID)

	result, err := svc.ArrangeShipment(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// GetTracking handles GET /api/shopee/shipping/tracking/:orderSn
func (h *ShippingHandler) GetTracking(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	orderSN := c.Param("orderSn")
	if orderSN == "" {
		c.JSON(http.StatusBadRequest, response.Error("orderSn required"))
		return
	}

	apiClient := h.getAPIClient(tenantID)
	svc := shopeeService.NewShippingService(apiClient, tenantID)

	tracking, err := svc.GetTrackingInfo(c.Request.Context(), orderSN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(tracking))
}

// GetShipment handles GET /api/shopee/shipping/info/:orderSn
func (h *ShippingHandler) GetShipment(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	orderSN := c.Param("orderSn")
	if orderSN == "" {
		c.JSON(http.StatusBadRequest, response.Error("orderSn required"))
		return
	}

	apiClient := h.getAPIClient(tenantID)
	svc := shopeeService.NewShippingService(apiClient, tenantID)

	info, err := svc.GetShipmentInfo(c.Request.Context(), orderSN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(info))
}
