package shopee

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

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

// DownloadShippingLabel handles GET /api/shopee/shipping/download/:orderSn
// Downloads shipping label and saves to local Downloads folder
func (h *ShippingHandler) DownloadShippingLabel(c *gin.Context) {
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

	packageNumber := c.Query("package_number")
	documentType := c.DefaultQuery("document_type", "THERMAL_AIR_WAYBILL")

	apiClient := h.getAPIClient(tenantID)
	svc := shopeeService.NewShippingService(apiClient, tenantID)

	result, err := svc.GetShippingLabel(c.Request.Context(), orderSN, packageNumber, documentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	if result.Status == "FAILED" || result.FileData == "" {
		c.JSON(http.StatusInternalServerError, response.Error(result.ErrorMessage))
		return
	}

	// Decode base64 PDF data
	pdfData, err := base64.StdEncoding.DecodeString(result.FileData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to decode PDF: "+err.Error()))
		return
	}

	// Save to uploads folder (mounted volume accessible from host)
	uploadsDir := "/app/uploads/labels"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to create uploads directory: "+err.Error()))
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("shopee_label_%s_%s.pdf", orderSN, timestamp)
	filePath := filepath.Join(uploadsDir, filename)

	if err := os.WriteFile(filePath, pdfData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to save file: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"order_sn":  orderSN,
		"status":    "SUCCESS",
		"file_path": filePath,
		"file_size": len(pdfData),
		"file_data": result.FileData, // Include base64 data for direct browser download
		"message":   fmt.Sprintf("Shipping label saved to %s", filePath),
	}))
}

// GetShippingLabel handles GET /api/shopee/shipping/label/:orderSn
// Downloads shipping document (waybill/label) for an order
func (h *ShippingHandler) GetShippingLabel(c *gin.Context) {
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

	packageNumber := c.Query("package_number")
	documentType := c.DefaultQuery("document_type", "THERMAL_AIR_WAYBILL")

	apiClient := h.getAPIClient(tenantID)
	svc := shopeeService.NewShippingService(apiClient, tenantID)

	result, err := svc.GetShippingLabel(c.Request.Context(), orderSN, packageNumber, documentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	if result.Status == "FAILED" || result.FileData == "" {
		c.JSON(http.StatusInternalServerError, response.Error(result.ErrorMessage))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}
