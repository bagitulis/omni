package shopee

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
)

// ShippingLabelResult represents the shipping label download result
type ShippingLabelResult struct {
	OrderSN      string `json:"order_sn"`
	Status       string `json:"status"`
	FileData     string `json:"file_data,omitempty"`     // Base64 encoded PDF
	ErrorMessage string `json:"error_message,omitempty"` // Error if failed
}

// GetShippingLabel downloads shipping document (waybill/label) for an order.
// This path only returns official Shopee documents from API responses.
func (s *ShippingService) GetShippingLabel(ctx context.Context, orderSN, packageNumber, documentType string) (*ShippingLabelResult, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}

	documentTypes := []string{
		documentType,
		"THERMAL_AIR_WAYBILL",
		"NORMAL_AIR_WAYBILL",
		"THERMAL_WAYBILL",
		"NORMAL_WAYBILL",
	}

	tried := make(map[string]struct{}, len(documentTypes))
	for _, docType := range documentTypes {
		if docType == "" {
			continue
		}
		if _, exists := tried[docType]; exists {
			continue
		}
		tried[docType] = struct{}{}

		result, downloadErr := s.tryStandardDownload(ctx, client, orderSN, packageNumber, docType)
		if downloadErr == nil && result.Status == "SUCCESS" && result.FileData != "" {
			return result, nil
		}
	}

	log.Warn().Str("order_sn", orderSN).Msg("Shopee did not return official shipping document")
	return &ShippingLabelResult{
		OrderSN:      orderSN,
		Status:       "FAILED",
		ErrorMessage: "official Shopee shipping document is not available for this order",
	}, nil
}

// tryStandardDownload attempts the standard Shopee document download flow
func (s *ShippingService) tryStandardDownload(ctx context.Context, client shippingClient, orderSN, packageNumber, documentType string) (*ShippingLabelResult, error) {
	// Try to create the shipping document first
	createResult, err := client.CreateShippingDocument(orderSN, packageNumber)

	var canProceedToDownload bool

	if err != nil {
		if createResult != nil && createResult.Error == "common.batch_api_all_failed" {
			if len(createResult.Response.ResultList) > 0 {
				failErr := createResult.Response.ResultList[0].FailError
				if failErr == "logistics.document_already_created" ||
					failErr == "logistics.tracking_number_invalid" {
					log.Info().Str("order_sn", orderSN).Str("fail_error", failErr).Msg("Create failed but will try download")
					canProceedToDownload = true
				}
			}
		}
		if !canProceedToDownload {
			canProceedToDownload = true // Try anyway
		}
	} else {
		canProceedToDownload = true
	}

	if !canProceedToDownload {
		return &ShippingLabelResult{
			OrderSN:      orderSN,
			Status:       "FAILED",
			ErrorMessage: "Cannot proceed to download",
		}, nil
	}

	// Try direct download
	downloadResult, err := client.DownloadShippingDocument(orderSN, packageNumber, documentType)
	if err != nil {
		// Check for should_print_first error - try poll flow
		if downloadResult != nil && downloadResult.Error == "logistics.shipping_document_should_print_first" {
			return s.pollAndDownload(ctx, client, orderSN, packageNumber, documentType)
		}
		return &ShippingLabelResult{
			OrderSN:      orderSN,
			Status:       "FAILED",
			ErrorMessage: err.Error(),
		}, nil
	}

	// Handle binary PDF response
	if len(downloadResult.RawPDF) > 0 {
		log.Info().Str("order_sn", orderSN).Int("pdf_size", len(downloadResult.RawPDF)).Msg("Received binary PDF")
		return &ShippingLabelResult{
			OrderSN:  orderSN,
			Status:   "SUCCESS",
			FileData: base64.StdEncoding.EncodeToString(downloadResult.RawPDF),
		}, nil
	}

	// Handle JSON response with file data
	if len(downloadResult.Response.ResultList) > 0 {
		docResult := downloadResult.Response.ResultList[0]
		if docResult.ShippingDocFile != "" {
			return &ShippingLabelResult{
				OrderSN:  orderSN,
				Status:   "SUCCESS",
				FileData: docResult.ShippingDocFile,
			}, nil
		}
	}

	return &ShippingLabelResult{
		OrderSN:      orderSN,
		Status:       "FAILED",
		ErrorMessage: "No document available from standard flow",
	}, nil
}

// pollAndDownload polls for document ready status then downloads
func (s *ShippingService) pollAndDownload(ctx context.Context, client shippingClient, orderSN, packageNumber, documentType string) (*ShippingLabelResult, error) {
	// Poll for document ready status (max 5 attempts)
	for i := 0; i < 5; i++ {
		resultResp, err := client.GetShippingDocumentResult(orderSN, packageNumber)
		if err == nil && len(resultResp.Response.ResultList) > 0 {
			status := resultResp.Response.ResultList[0].Status
			if status == "READY" || status == "SUCCESS" {
				// Document is ready, download it
				downloadResult, err := client.DownloadShippingDocument(orderSN, packageNumber, documentType)
				if err == nil {
					if len(downloadResult.RawPDF) > 0 {
						return &ShippingLabelResult{
							OrderSN:  orderSN,
							Status:   "SUCCESS",
							FileData: base64.StdEncoding.EncodeToString(downloadResult.RawPDF),
						}, nil
					}
					if len(downloadResult.Response.ResultList) > 0 && downloadResult.Response.ResultList[0].ShippingDocFile != "" {
						return &ShippingLabelResult{
							OrderSN:  orderSN,
							Status:   "SUCCESS",
							FileData: downloadResult.Response.ResultList[0].ShippingDocFile,
						}, nil
					}
				}
			}
		}
		if i < 4 {
			time.Sleep(1 * time.Second)
		}
	}

	return &ShippingLabelResult{
		OrderSN:      orderSN,
		Status:       "FAILED",
		ErrorMessage: "Document not ready after polling",
	}, nil
}

// generateLabelFromData gets label data via API and generates PDF locally
// This is the FALLBACK that MUST SUCCEED for SPX orders
func (s *ShippingService) generateLabelFromData(ctx context.Context, client shippingClient, orderSN, packageNumber string) (*ShippingLabelResult, error) {
	// Get shipping document data info
	dataInfo, err := client.GetShippingDocumentDataInfo(orderSN, packageNumber)
	if err != nil {
		log.Error().Err(err).Str("order_sn", orderSN).Msg("Failed to get shipping document data info")
		// Last resort: try to get tracking number at least
		trackingResp, trackErr := client.GetTrackingNumber(orderSN)
		if trackErr != nil {
			return &ShippingLabelResult{
				OrderSN:      orderSN,
				Status:       "FAILED",
				ErrorMessage: fmt.Sprintf("Cannot get label data: %s", err.Error()),
			}, nil
		}
		// Generate minimal label with tracking number only
		return s.generateMinimalLabel(orderSN, trackingResp.Response.TrackingNumber)
	}

	// Generate PDF from the data
	pdfData, err := s.createLabelPDF(orderSN, dataInfo)
	if err != nil {
		log.Error().Err(err).Str("order_sn", orderSN).Msg("Failed to generate PDF")
		return &ShippingLabelResult{
			OrderSN:      orderSN,
			Status:       "FAILED",
			ErrorMessage: fmt.Sprintf("Failed to generate PDF: %s", err.Error()),
		}, nil
	}

	log.Info().Str("order_sn", orderSN).Int("pdf_size", len(pdfData)).Msg("Generated shipping label PDF from data")

	return &ShippingLabelResult{
		OrderSN:  orderSN,
		Status:   "SUCCESS",
		FileData: base64.StdEncoding.EncodeToString(pdfData),
	}, nil
}

// createLabelPDF generates a thermal shipping label PDF from data
func (s *ShippingService) createLabelPDF(orderSN string, data interface{}) ([]byte, error) {
	// Extract fields from the known ShippingDocumentDataInfoResponse type
	var trackingNumber, carrier, recipientName, recipientPhone, recipientAddress string
	var recipientCity, recipientState, recipientZip string
	var senderName, senderPhone, senderAddress, senderCity, senderState string
	var sortCode string

	if infoResp, ok := data.(*shopeePkg.ShippingDocumentDataInfoResponse); ok {
		r := infoResp.Response
		trackingNumber = r.TrackingNumber
		if trackingNumber == "" {
			trackingNumber = r.LastMileTrackingNumber
		}
		if trackingNumber == "" {
			trackingNumber = r.FirstMileTrackingNumber
		}
		carrier = r.ShippingCarrier
		recipientName = r.RecipientName
		recipientPhone = r.RecipientPhone
		recipientAddress = r.RecipientAddress
		recipientCity = r.RecipientCity
		recipientState = r.RecipientState
		recipientZip = r.RecipientZipcode
		senderName = r.SenderName
		senderPhone = r.SenderPhone
		senderAddress = r.SenderAddress
		senderCity = r.SenderCity
		senderState = r.SenderState
		sortCode = r.RecipientSortCode
	}

	if trackingNumber == "" {
		trackingNumber = "TRACKING_PENDING"
	}

	// Create PDF (10cm x 15cm thermal label size)
	pdf := gofpdf.New("P", "mm", "A6", "")
	pdf.SetMargins(5, 5, 5)
	pdf.AddPage()

	// Header - Shopee logo area
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(238, 77, 45) // Shopee orange
	pdf.Rect(0, 0, 105, 15, "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(105, 15, "SHOPEE EXPRESS", "", 1, "C", false, 0, "")

	// Reset colors
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFillColor(255, 255, 255)

	// Tracking number (large, prominent)
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(95, 10, fmt.Sprintf("Tracking: %s", trackingNumber), "1", 1, "C", false, 0, "")

	// Barcode placeholder (we'll draw a text representation)
	pdf.Ln(3)
	pdf.SetFont("Courier", "B", 10)
	pdf.CellFormat(95, 8, trackingNumber, "1", 1, "C", false, 0, "")

	// Carrier info
	pdf.Ln(3)
	pdf.SetFont("Arial", "", 10)
	if carrier == "" {
		carrier = "SPX Express"
	}
	pdf.CellFormat(95, 6, fmt.Sprintf("Carrier: %s", carrier), "", 1, "L", false, 0, "")

	// Separator
	pdf.Ln(2)
	pdf.Line(5, pdf.GetY(), 100, pdf.GetY())
	pdf.Ln(3)

	// Recipient section
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(95, 6, "PENERIMA / RECIPIENT:", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	if recipientName == "" {
		recipientName = "(Data dari Shopee Seller Center)"
	}
	pdf.CellFormat(95, 5, recipientName, "", 1, "L", false, 0, "")

	if recipientPhone != "" {
		pdf.CellFormat(95, 5, fmt.Sprintf("Tel: %s", recipientPhone), "", 1, "L", false, 0, "")
	}

	if recipientAddress != "" {
		pdf.MultiCell(95, 5, recipientAddress, "", "L", false)
	}

	if recipientCity != "" || recipientState != "" || recipientZip != "" {
		pdf.CellFormat(95, 5, fmt.Sprintf("%s, %s %s", recipientCity, recipientState, recipientZip), "", 1, "L", false, 0, "")
	}

	// Sort code if available
	if sortCode != "" {
		pdf.Ln(2)
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(95, 6, fmt.Sprintf("Sort: %s", sortCode), "1", 1, "C", false, 0, "")
	}

	// Separator
	pdf.Ln(3)
	pdf.Line(5, pdf.GetY(), 100, pdf.GetY())
	pdf.Ln(3)

	// Sender section
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(95, 5, "PENGIRIM / SENDER:", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	if senderName == "" {
		senderName = "(Lihat detail di Shopee)"
	}
	pdf.CellFormat(95, 4, senderName, "", 1, "L", false, 0, "")

	if senderPhone != "" {
		pdf.CellFormat(95, 4, fmt.Sprintf("Tel: %s", senderPhone), "", 1, "L", false, 0, "")
	}

	if senderAddress != "" {
		pdf.MultiCell(95, 4, senderAddress, "", "L", false)
	}

	if senderCity != "" || senderState != "" {
		pdf.CellFormat(95, 4, fmt.Sprintf("%s, %s", senderCity, senderState), "", 1, "L", false, 0, "")
	}

	// Order number at bottom
	pdf.Ln(3)
	pdf.Line(5, pdf.GetY(), 100, pdf.GetY())
	pdf.Ln(2)
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(95, 4, fmt.Sprintf("Order: %s", orderSN), "", 1, "C", false, 0, "")
	pdf.CellFormat(95, 4, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04")), "", 1, "C", false, 0, "")

	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// generateMinimalLabel creates a minimal label with just tracking number
func (s *ShippingService) generateMinimalLabel(orderSN, trackingNumber string) (*ShippingLabelResult, error) {
	pdf := gofpdf.New("P", "mm", "A6", "")
	pdf.SetMargins(5, 5, 5)
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(238, 77, 45)
	pdf.Rect(0, 0, 105, 15, "F")
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(105, 15, "SHOPEE EXPRESS", "", 1, "C", false, 0, "")

	pdf.SetTextColor(0, 0, 0)

	// Tracking number
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 18)
	pdf.CellFormat(95, 12, "TRACKING NUMBER", "", 1, "C", false, 0, "")

	pdf.Ln(5)
	pdf.SetFont("Courier", "B", 14)
	pdf.CellFormat(95, 10, trackingNumber, "1", 1, "C", false, 0, "")

	// Order info
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(95, 6, fmt.Sprintf("Order: %s", orderSN), "", 1, "C", false, 0, "")

	// Note
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 9)
	pdf.MultiCell(95, 5, "Untuk detail lengkap label pengiriman, silakan cetak dari Shopee Seller Center.", "", "C", false)

	pdf.Ln(5)
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(95, 4, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")), "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return &ShippingLabelResult{
			OrderSN:      orderSN,
			Status:       "FAILED",
			ErrorMessage: fmt.Sprintf("Failed to generate PDF: %s", err.Error()),
		}, nil
	}

	log.Info().Str("order_sn", orderSN).Str("tracking", trackingNumber).Msg("Generated minimal shipping label")

	return &ShippingLabelResult{
		OrderSN:  orderSN,
		Status:   "SUCCESS",
		FileData: base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}
