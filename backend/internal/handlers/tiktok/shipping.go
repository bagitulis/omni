package tiktok

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req arrangeShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request body: "+err.Error()))
		return
	}

	// Validate: either package_id or order_id must be provided
	if req.PackageID == "" && req.OrderID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Either package_id or order_id is required"))
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

	var result *tiktokPkg.ShipPackageResponse
	var err error

	if req.OrderID != "" {
		// Use order ID flow (resolves to package ID internally)
		result, err = h.Service.ArrangeShipmentByOrder(c.Request.Context(), tenantID, req.OrderID, sdkReq)
	} else {
		// Use package ID directly
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

// BatchDownloadRequest represents the request body for batch download
type BatchDownloadRequest struct {
	OrderIDs        []string `json:"order_ids"`
	IncludeProducts bool     `json:"include_products"`
}

// BatchDownloadResult represents the result for a single order in batch
type BatchDownloadResult struct {
	OrderID  string `json:"order_id"`
	Status   string `json:"status"`
	FilePath string `json:"file_path,omitempty"`
	FileSize int    `json:"file_size,omitempty"`
	FileData string `json:"file_data,omitempty"`
	Error    string `json:"error,omitempty"`
}

// BatchDownloadShippingDocuments handles POST /api/tiktok/shipping/download/batch
// Downloads shipping documents for multiple orders
func (h *ShippingHandler) BatchDownloadShippingDocuments(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req BatchDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request body: "+err.Error()))
		return
	}

	if len(req.OrderIDs) == 0 {
		c.JSON(http.StatusBadRequest, response.Error("order_ids is required"))
		return
	}

	if len(req.OrderIDs) > 20 {
		c.JSON(http.StatusBadRequest, response.Error("Maximum 20 orders per batch"))
		return
	}

	// Determine document type
	documentType := "SHIPPING_LABEL"
	if req.IncludeProducts {
		documentType = "SHIPPING_LABEL_AND_PACKING_SLIP"
	}

	results := make([]BatchDownloadResult, 0, len(req.OrderIDs))
	successCount := 0

	for _, orderID := range req.OrderIDs {
		result := BatchDownloadResult{OrderID: orderID}

		// Get document URL
		docURL, _, err := h.Service.GetShippingLabelByOrder(c.Request.Context(), tenantID, orderID, documentType)
		if err != nil {
			result.Status = "FAILED"
			result.Error = err.Error()
			results = append(results, result)
			continue
		}

		if docURL == "" {
			result.Status = "FAILED"
			result.Error = "No document URL returned"
			results = append(results, result)
			continue
		}

		// Download the PDF
		httpResp, err := http.Get(docURL)
		if err != nil {
			result.Status = "FAILED"
			result.Error = "Failed to download: " + err.Error()
			results = append(results, result)
			continue
		}

		pdfData, err := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		if err != nil {
			result.Status = "FAILED"
			result.Error = "Failed to read document"
			results = append(results, result)
			continue
		}

		// Save to uploads folder
		uploadsDir := "/app/uploads/labels"
		os.MkdirAll(uploadsDir, 0755)

		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("tiktok_label_%s_%s.pdf", orderID, timestamp)
		filePath := filepath.Join(uploadsDir, filename)

		if err := os.WriteFile(filePath, pdfData, 0644); err != nil {
			result.Status = "FAILED"
			result.Error = "Failed to save file"
			results = append(results, result)
			continue
		}

		result.Status = "SUCCESS"
		result.FilePath = filePath
		result.FileSize = len(pdfData)
		result.FileData = base64.StdEncoding.EncodeToString(pdfData)
		results = append(results, result)
		successCount++
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total":   len(req.OrderIDs),
		"success": successCount,
		"failed":  len(req.OrderIDs) - successCount,
		"results": results,
	}))
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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
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

	// Add order details if available
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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
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

// DownloadShippingDocumentByOrder handles GET /api/tiktok/shipping/download/order/:orderId
// Downloads shipping document and saves to local Downloads folder
// Query params:
//   - document_type: SHIPPING_LABEL (default), PACKING_SLIP, SHIPPING_LABEL_AND_PACKING_SLIP
//   - include_products: true/false - shortcut for SHIPPING_LABEL_AND_PACKING_SLIP
func (h *ShippingHandler) DownloadShippingDocumentByOrder(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	orderID := c.Param("orderId")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing orderId"))
		return
	}

	// Determine document type
	documentType := c.Query("document_type")
	if documentType == "" {
		// Check include_products shortcut
		if c.Query("include_products") == "true" {
			documentType = "SHIPPING_LABEL_AND_PACKING_SLIP"
		} else {
			documentType = "SHIPPING_LABEL"
		}
	}

	docURL, orderDetail, err := h.Service.GetShippingLabelByOrder(c.Request.Context(), tenantID, orderID, documentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get shipping document: "+err.Error()))
		return
	}

	if docURL == "" {
		c.JSON(http.StatusInternalServerError, response.Error("No document URL returned"))
		return
	}

	// Download the PDF from URL
	httpResp, err := http.Get(docURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to download document: "+err.Error()))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, response.Error(fmt.Sprintf("Failed to download document: HTTP %d", httpResp.StatusCode)))
		return
	}

	pdfData, err := io.ReadAll(httpResp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to read document: "+err.Error()))
		return
	}

	// Save to uploads folder (mounted volume accessible from host)
	uploadsDir := "/app/uploads/labels"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to create uploads directory: "+err.Error()))
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("tiktok_label_%s_%s.pdf", orderID, timestamp)
	filePath := filepath.Join(uploadsDir, filename)

	if err := os.WriteFile(filePath, pdfData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to save file: "+err.Error()))
		return
	}

	// Encode PDF as base64 for frontend download
	base64Data := base64.StdEncoding.EncodeToString(pdfData)

	result := gin.H{
		"order_id":  orderID,
		"status":    "SUCCESS",
		"file_path": filePath,
		"file_size": len(pdfData),
		"file_data": base64Data, // Include base64 data for direct browser download
		"message":   fmt.Sprintf("Shipping label saved to %s", filePath),
	}

	// Add order details if available
	if orderDetail != nil {
		result["order_status"] = orderDetail.Status
		if len(orderDetail.Packages) > 0 {
			result["tracking_number"] = orderDetail.Packages[0].TrackingNumber
		}
	}

	c.JSON(http.StatusOK, response.Success(result))
}
