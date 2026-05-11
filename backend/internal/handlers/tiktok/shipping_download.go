package tiktok

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/rs/zerolog/log"
)

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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
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

	documentType := "SHIPPING_LABEL"
	if req.IncludeProducts {
		documentType = "SHIPPING_LABEL_AND_PACKING_SLIP"
	}

	results := make([]BatchDownloadResult, 0, len(req.OrderIDs))
	successCount := 0

	for _, orderID := range req.OrderIDs {
		result := h.downloadSingleOrderLabel(c, tenantID, orderID, documentType)
		if result.Status == "SUCCESS" {
			successCount++
		}
		results = append(results, result)
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total":   len(req.OrderIDs),
		"success": successCount,
		"failed":  len(req.OrderIDs) - successCount,
		"results": results,
	}))
}

// DownloadShippingDocumentByOrder handles GET /api/tiktok/shipping/download/order/:orderId
// Downloads shipping document and saves to local Downloads folder
// Query params:
//   - document_type: SHIPPING_LABEL (default), PACKING_SLIP, SHIPPING_LABEL_AND_PACKING_SLIP
//   - include_products: true/false - shortcut for SHIPPING_LABEL_AND_PACKING_SLIP
func (h *ShippingHandler) DownloadShippingDocumentByOrder(c *gin.Context) {
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

	documentType := resolveDocumentType(c)

	docURL, orderDetail, err := h.Service.GetShippingLabelByOrder(c.Request.Context(), tenantID, orderID, documentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get shipping document: "+err.Error()))
		return
	}

	if docURL == "" {
		c.JSON(http.StatusInternalServerError, response.Error("No document URL returned"))
		return
	}

	pdfData, err := downloadPDF(docURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	filePath, err := saveLabelFile(orderID, pdfData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	result := gin.H{
		"order_id":  orderID,
		"status":    "SUCCESS",
		"file_path": filePath,
		"file_size": len(pdfData),
		"file_data": base64.StdEncoding.EncodeToString(pdfData),
		"message":   fmt.Sprintf("Shipping label saved to %s", filePath),
	}

	if orderDetail != nil {
		result["order_status"] = orderDetail.Status
		if len(orderDetail.Packages) > 0 {
			result["tracking_number"] = orderDetail.Packages[0].TrackingNumber
		}
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// downloadSingleOrderLabel downloads a shipping label for a single order
func (h *ShippingHandler) downloadSingleOrderLabel(c *gin.Context, tenantID, orderID, documentType string) BatchDownloadResult {
	result := BatchDownloadResult{OrderID: orderID}

	docURL, _, err := h.Service.GetShippingLabelByOrder(c.Request.Context(), tenantID, orderID, documentType)
	if err != nil {
		result.Status = "FAILED"
		result.Error = err.Error()
		return result
	}

	if docURL == "" {
		result.Status = "FAILED"
		result.Error = "No document URL returned"
		return result
	}

	pdfData, err := downloadPDF(docURL)
	if err != nil {
		result.Status = "FAILED"
		result.Error = err.Error()
		return result
	}

	filePath, err := saveLabelFile(orderID, pdfData)
	if err != nil {
		result.Status = "FAILED"
		result.Error = err.Error()
		return result
	}

	result.Status = "SUCCESS"
	result.FilePath = filePath
	result.FileSize = len(pdfData)
	result.FileData = base64.StdEncoding.EncodeToString(pdfData)
	return result
}

// resolveDocumentType determines the document type from query params
func resolveDocumentType(c *gin.Context) string {
	documentType := c.Query("document_type")
	if documentType == "" {
		if c.Query("include_products") == "true" {
			return "SHIPPING_LABEL_AND_PACKING_SLIP"
		}
		return "SHIPPING_LABEL"
	}
	return documentType
}

// allowedTikTokHosts contains permitted CDN hostnames for document downloads.
var allowedTikTokHosts = []string{
	".tiktokcdn.com",
	".tiktokcdn-us.com",
	".bytedance.com",
	".ibyteimg.com",
}

// maxDownloadSize is the maximum allowed response body size (50 MB).
const maxDownloadSize = 50 * 1024 * 1024

// httpDownloadClient is a shared HTTP client with a 30-second timeout for downloads.
var httpDownloadClient = &http.Client{
	Timeout: 30 * time.Second,
}

// validateDocURL checks that the URL uses HTTPS and points to an allowed TikTok CDN host.
func validateDocURL(docURL string) error {
	parsed, err := url.Parse(docURL)
	if err != nil {
		return fmt.Errorf("invalid document URL: %w", err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("document URL must use HTTPS, got %q", parsed.Scheme)
	}
	host := strings.ToLower(parsed.Hostname())
	for _, allowed := range allowedTikTokHosts {
		if host == strings.TrimPrefix(allowed, ".") || strings.HasSuffix(host, allowed) {
			return nil
		}
	}
	return fmt.Errorf("document URL host %q is not an allowed TikTok CDN domain", host)
}

// downloadPDF downloads PDF data from a URL
func downloadPDF(docURL string) ([]byte, error) {
	if err := validateDocURL(docURL); err != nil {
		log.Warn().Str("url", docURL).Err(err).Msg("Blocked download from untrusted URL")
		return nil, fmt.Errorf("URL validation failed: %w", err)
	}

	httpResp, err := httpDownloadClient.Get(docURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download document: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download document: HTTP %d", httpResp.StatusCode)
	}

	limitedReader := io.LimitReader(httpResp.Body, maxDownloadSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read document: %w", err)
	}
	if len(data) > maxDownloadSize {
		return nil, fmt.Errorf("document exceeds maximum allowed size of %d bytes", maxDownloadSize)
	}
	return data, nil
}

// saveLabelFile saves PDF data to the uploads directory
func saveLabelFile(orderID string, pdfData []byte) (string, error) {
	uploadsDir := "/app/uploads/labels"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create uploads directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("tiktok_label_%s_%s.pdf", orderID, timestamp)
	filePath := filepath.Join(uploadsDir, filename)

	if err := os.WriteFile(filePath, pdfData, 0644); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}
	return filePath, nil
}
