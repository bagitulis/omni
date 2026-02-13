package shopee

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/rs/zerolog/log"
)

// ShippingLabelResult represents the shipping label download result.
type ShippingLabelResult struct {
	OrderSN      string `json:"order_sn"`
	Status       string `json:"status"`
	FileData     string `json:"file_data,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// GetShippingLabel downloads official Shopee shipping documents only.
func (s *ShippingService) GetShippingLabel(ctx context.Context, orderSN, packageNumber, documentType string) (*ShippingLabelResult, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}

	trackingResp, trackingErr := s.ensureShipmentReady(ctx, client, orderSN, packageNumber)
	if trackingErr != nil || trackingResp == nil || trackingResp.Response.TrackingNumber == "" {
		return &ShippingLabelResult{
			OrderSN:      orderSN,
			Status:       "FAILED",
			ErrorMessage: "shipment is not ready for printing: tracking number is unavailable",
		}, nil
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

func (s *ShippingService) tryStandardDownload(ctx context.Context, client shippingClient, orderSN, packageNumber, documentType string) (*ShippingLabelResult, error) {
	createResult, err := client.CreateShippingDocument(orderSN, packageNumber)

	canProceedToDownload := err == nil
	if err != nil {
		if createResult != nil && createResult.Error == "common.batch_api_all_failed" && len(createResult.Response.ResultList) > 0 {
			failErr := createResult.Response.ResultList[0].FailError
			if failErr == "logistics.document_already_created" || failErr == "logistics.tracking_number_invalid" {
				canProceedToDownload = true
			}
		}
		if !canProceedToDownload {
			canProceedToDownload = true
		}
	}

	if !canProceedToDownload {
		return &ShippingLabelResult{
			OrderSN:      orderSN,
			Status:       "FAILED",
			ErrorMessage: "cannot proceed to download",
		}, nil
	}

	downloadResult, err := client.DownloadShippingDocument(orderSN, packageNumber, documentType)
	if err != nil {
		if downloadResult != nil && downloadResult.Error == "logistics.shipping_document_should_print_first" {
			return s.pollAndDownload(ctx, client, orderSN, packageNumber, documentType)
		}
		return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: err.Error()}, nil
	}

	if len(downloadResult.RawPDF) > 0 {
		return &ShippingLabelResult{
			OrderSN:  orderSN,
			Status:   "SUCCESS",
			FileData: base64.StdEncoding.EncodeToString(downloadResult.RawPDF),
		}, nil
	}

	if len(downloadResult.Response.ResultList) > 0 {
		docResult := downloadResult.Response.ResultList[0]
		if docResult.ShippingDocFile != "" {
			return &ShippingLabelResult{OrderSN: orderSN, Status: "SUCCESS", FileData: docResult.ShippingDocFile}, nil
		}
	}

	return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: "no document available from Shopee API"}, nil
}

func (s *ShippingService) pollAndDownload(ctx context.Context, client shippingClient, orderSN, packageNumber, documentType string) (*ShippingLabelResult, error) {
	for i := 0; i < 5; i++ {
		resultResp, err := client.GetShippingDocumentResult(orderSN, packageNumber)
		if err == nil && len(resultResp.Response.ResultList) > 0 {
			status := resultResp.Response.ResultList[0].Status
			if status == "READY" || status == "SUCCESS" {
				downloadResult, dlErr := client.DownloadShippingDocument(orderSN, packageNumber, documentType)
				if dlErr == nil {
					if len(downloadResult.RawPDF) > 0 {
						return &ShippingLabelResult{OrderSN: orderSN, Status: "SUCCESS", FileData: base64.StdEncoding.EncodeToString(downloadResult.RawPDF)}, nil
					}
					if len(downloadResult.Response.ResultList) > 0 {
						fileData := downloadResult.Response.ResultList[0].ShippingDocFile
						if fileData != "" {
							return &ShippingLabelResult{OrderSN: orderSN, Status: "SUCCESS", FileData: fileData}, nil
						}
					}
				}
			}
		}
		if i < 4 {
			time.Sleep(1 * time.Second)
		}
	}

	return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: "document not ready after polling"}, nil
}
