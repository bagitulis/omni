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
	service *serviceTiktok.ShippingService
}

// NewShippingHandler creates a new shipping handler
func NewShippingHandler(basePath string) *ShippingHandler {
	return &ShippingHandler{
		service: serviceTiktok.NewShippingService(basePath),
	}
}

// arrangeShipmentRequest represents the request body for arranging shipment
type arrangeShipmentRequest struct {
	PackageID      string                      `json:"package_id" binding:"required"`
	HandoverMethod string                      `json:"handover_method,omitempty"`
	PickupSlot     *tiktokPkg.PickupSlotInfo   `json:"pickup_slot,omitempty"`
	SelfShipment   *tiktokPkg.SelfShipmentInfo `json:"self_shipment,omitempty"`
}

// ArrangeShipment handles POST /api/tiktok/shipping/arrange
func (h *ShippingHandler) ArrangeShipment(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req arrangeShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request body: "+err.Error()))
		return
	}

	if req.SelfShipment != nil && req.SelfShipment.TrackingNumber == "" && req.SelfShipment.ShippingProviderID != "" {
		c.JSON(http.StatusBadRequest, response.Error("tracking_number is required for self-shipment"))
		return
	}

	// Map to SDK request
	sdkReq := &tiktokPkg.ShipPackageRequest{
		HandoverMethod: req.HandoverMethod,
		PickupSlot:     req.PickupSlot,
		SelfShipment:   req.SelfShipment,
	}

	result, err := h.service.ArrangeShipment(c.Request.Context(), tenantID, req.PackageID, sdkReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to arrange shipment: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// ShippingDocumentResponse represents the shipping document response
type ShippingDocumentResponse struct {
	DocURL string `json:"doc_url"`
}

// GetShippingDocument handles GET /api/tiktok/shipping/document/:packageId
func (h *ShippingHandler) GetShippingDocument(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	packageID := c.Param("packageId")
	if packageID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing packageId"))
		return
	}

	documentType := c.DefaultQuery("document_type", "SHIPPING_LABEL")

	docURL, err := h.service.GetShippingLabel(c.Request.Context(), tenantID, packageID, documentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get shipping document: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(ShippingDocumentResponse{
		DocURL: docURL,
	}))
}
