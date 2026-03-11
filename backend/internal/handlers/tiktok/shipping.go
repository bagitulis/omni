package tiktok

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	serviceTiktok "github.com/omni/backend/internal/services/tiktok"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// ShippingHandler handles TikTok shipping requests
type ShippingHandler struct {
	Service *serviceTiktok.ShippingService
}

// NewShippingHandler creates a new shipping handler
func NewShippingHandler(basePath string) *ShippingHandler {
	return &ShippingHandler{
		Service: serviceTiktok.NewShippingService(basePath),
	}
}

// arrangeShipmentRequest represents the request body for arranging shipment
type arrangeShipmentRequest struct {
	PackageID      string                      `json:"package_id,omitempty"`
	OrderID        string                      `json:"order_id,omitempty"` // Alternative to package_id
	HandoverMethod string                      `json:"handover_method,omitempty"`
	PickupSlot     *tiktokPkg.PickupSlotInfo   `json:"pickup_slot,omitempty"`
	SelfShipment   *tiktokPkg.SelfShipmentInfo `json:"self_shipment,omitempty"`
}

// ArrangeShipment handles POST /api/tiktok/shipping/arrange
func (h *ShippingHandler) ArrangeShipment(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	var req arrangeShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request body: "+err.Error()))
		return
	}

	if req.PackageID == "" && req.OrderID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Either package_id or order_id is required"))
		return
	}

	if req.SelfShipment != nil && req.SelfShipment.TrackingNumber == "" && req.SelfShipment.ShippingProviderID != "" {
		c.JSON(http.StatusBadRequest, response.Error("tracking_number is required for self-shipment"))
		return
	}

	sdkReq := &tiktokPkg.ShipPackageRequest{
		HandoverMethod: req.HandoverMethod,
		PickupSlot:     req.PickupSlot,
		SelfShipment:   req.SelfShipment,
	}

	var result *tiktokPkg.ShipPackageResponse
	var err error

	if req.OrderID != "" {
		result, err = h.Service.ArrangeShipmentByOrder(c.Request.Context(), tenantID, req.OrderID, sdkReq)
	} else {
		result, err = h.Service.ArrangeShipment(c.Request.Context(), tenantID, req.PackageID, sdkReq)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to arrange shipment: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// ShippingDocumentResponse represents the shipping document response
type ShippingDocumentResponse struct {
	DocURL         string `json:"doc_url"`
	TrackingNumber string `json:"tracking_number,omitempty"`
	OrderID        string `json:"order_id,omitempty"`
	OrderStatus    string `json:"order_status,omitempty"`
}

// GetShippingDocument handles GET /api/tiktok/shipping/document/:packageId
func (h *ShippingHandler) GetShippingDocument(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	packageID := c.Param("packageId")
	if packageID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing packageId"))
		return
	}

	documentType := c.DefaultQuery("document_type", "SHIPPING_LABEL")

	docURL, err := h.Service.GetShippingLabel(c.Request.Context(), tenantID, packageID, documentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get shipping document: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(ShippingDocumentResponse{
		DocURL: docURL,
	}))
}

// GetShippingDocumentByOrder handles GET /api/tiktok/shipping/document/order/:orderId
func (h *ShippingHandler) GetShippingDocumentByOrder(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	orderID := c.Param("orderId")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing orderId"))
		return
	}

	documentType := c.DefaultQuery("document_type", "SHIPPING_LABEL")

	docURL, orderDetail, err := h.Service.GetShippingLabelByOrder(c.Request.Context(), tenantID, orderID, documentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get shipping document: "+err.Error()))
		return
	}

	resp := ShippingDocumentResponse{
		DocURL:  docURL,
		OrderID: orderID,
	}

	if orderDetail != nil {
		resp.OrderStatus = orderDetail.Status
		if len(orderDetail.Packages) > 0 {
			resp.TrackingNumber = orderDetail.Packages[0].TrackingNumber
		}
	}

	c.JSON(http.StatusOK, response.Success(resp))
}

// HandoverTimeSlotsResponse represents available time slots response
type HandoverTimeSlotsResponse struct {
	TimeSlots []tiktokPkg.HandoverTimeSlot `json:"time_slots"`
}

// GetHandoverTimeSlots handles GET /api/tiktok/shipping/timeslots/:orderOrPackageId
func (h *ShippingHandler) GetHandoverTimeSlots(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	orderOrPackageID := c.Param("orderOrPackageId")
	if orderOrPackageID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing orderOrPackageId"))
		return
	}

	resp, err := h.Service.GetHandoverTimeSlots(c.Request.Context(), tenantID, orderOrPackageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get handover time slots: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(HandoverTimeSlotsResponse{
		TimeSlots: resp.Data.TimeSlots,
	}))
}

// GetOrderDetail handles GET /api/tiktok/shipping/order/:orderId
func (h *ShippingHandler) GetOrderDetail(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	orderID := c.Param("orderId")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing orderId"))
		return
	}

	resp, err := h.Service.GetOrderDetail(c.Request.Context(), tenantID, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get order detail: "+err.Error()))
		return
	}

	if len(resp.Data.Orders) == 0 {
		c.JSON(http.StatusNotFound, response.Error("Order not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(resp.Data.Orders[0]))
}
