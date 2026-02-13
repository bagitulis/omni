package shopee

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	shopeePkg "github.com/omni/backend/pkg/shopee"
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
	lastShopeeError := ""
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
		if result != nil && result.ErrorMessage != "" {
			lastShopeeError = result.ErrorMessage
		}
	}

	log.Warn().Str("order_sn", orderSN).Msg("Shopee did not return official shipping document")
	if lastShopeeError != "" {
		return &ShippingLabelResult{
			OrderSN:      orderSN,
			Status:       "FAILED",
			ErrorMessage: lastShopeeError,
		}, nil
	}

	return &ShippingLabelResult{
		OrderSN:      orderSN,
		Status:       "FAILED",
		ErrorMessage: "shopee API error: no shipping document details returned",
	}, nil
}

func (s *ShippingService) tryStandardDownload(ctx context.Context, client shippingClient, orderSN, packageNumber, documentType string) (*ShippingLabelResult, error) {
	createResult, err := client.CreateShippingDocument(orderSN, packageNumber)
	createErrorMessage := extractShopeeCreateError(createResult, err)

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
			ErrorMessage: createErrorMessage,
		}, nil
	}

	downloadResult, err := client.DownloadShippingDocument(orderSN, packageNumber, documentType)
	if err != nil {
		if downloadResult != nil && downloadResult.Error == "logistics.shipping_document_should_print_first" {
			return s.pollAndDownload(ctx, client, orderSN, packageNumber, documentType)
		}
		return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: extractShopeeDocumentError(downloadResult, err)}, nil
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

	if len(downloadResult.Response.ResultList) > 0 {
		docResult := downloadResult.Response.ResultList[0]
		if docResult.FailError != "" || docResult.FailMessage != "" {
			return &ShippingLabelResult{
				OrderSN:      orderSN,
				Status:       "FAILED",
				ErrorMessage: formatShopeeError(docResult.FailError, docResult.FailMessage),
			}, nil
		}
	}

	if createErrorMessage != "" {
		return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: createErrorMessage}, nil
	}

	return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: "shopee API error: empty shipping document response"}, nil
}

func extractShopeeCreateError(createResult *shopeePkg.CreateShippingDocumentResponse, err error) string {
	if createResult != nil {
		if createResult.Error != "" || createResult.Message != "" {
			return formatShopeeError(createResult.Error, createResult.Message)
		}
		if len(createResult.Response.ResultList) > 0 {
			item := createResult.Response.ResultList[0]
			if item.FailError != "" || item.FailMessage != "" {
				return formatShopeeError(item.FailError, item.FailMessage)
			}
		}
	}

	if err != nil {
		return err.Error()
	}

	return ""
}

func extractShopeeDocumentError(downloadResult *shopeePkg.DownloadShippingDocumentResponse, err error) string {
	if downloadResult != nil {
		if downloadResult.Error != "" || downloadResult.Message != "" {
			return formatShopeeError(downloadResult.Error, downloadResult.Message)
		}
		if len(downloadResult.Response.ResultList) > 0 {
			docResult := downloadResult.Response.ResultList[0]
			if docResult.FailError != "" || docResult.FailMessage != "" {
				return formatShopeeError(docResult.FailError, docResult.FailMessage)
			}
		}
	}

	if err != nil {
		return err.Error()
	}

	return "Shopee API request failed while downloading shipping document"
}

func formatShopeeError(errorCode, message string) string {
	switch {
	case errorCode != "" && message != "":
		return fmt.Sprintf("shopee API error: %s - %s", errorCode, message)
	case errorCode != "":
		return fmt.Sprintf("shopee API error: %s", errorCode)
	case message != "":
		return fmt.Sprintf("shopee API error: %s", message)
	default:
		return "shopee API error"
	}
}

func (s *ShippingService) pollAndDownload(ctx context.Context, client shippingClient, orderSN, packageNumber, documentType string) (*ShippingLabelResult, error) {
	for i := 0; i < 5; i++ {
		resultResp, err := client.GetShippingDocumentResult(orderSN, packageNumber)
		if err == nil && len(resultResp.Response.ResultList) > 0 {
			resultItem := resultResp.Response.ResultList[0]
			status := resultItem.Status
			if resultItem.FailError != "" || resultItem.FailMessage != "" {
				return &ShippingLabelResult{
					OrderSN:      orderSN,
					Status:       "FAILED",
					ErrorMessage: formatShopeeError(resultItem.FailError, resultItem.FailMessage),
				}, nil
			}
			if status == "READY" || status == "SUCCESS" {
				downloadResult, dlErr := client.DownloadShippingDocument(orderSN, packageNumber, documentType)
				if dlErr == nil {
					if len(downloadResult.RawPDF) > 0 {
						return &ShippingLabelResult{OrderSN: orderSN, Status: "SUCCESS", FileData: base64.StdEncoding.EncodeToString(downloadResult.RawPDF)}, nil
					}
					if len(downloadResult.Response.ResultList) > 0 {
						docResult := downloadResult.Response.ResultList[0]
						fileData := docResult.ShippingDocFile
						if fileData != "" {
							return &ShippingLabelResult{OrderSN: orderSN, Status: "SUCCESS", FileData: fileData}, nil
						}
						if docResult.FailError != "" || docResult.FailMessage != "" {
							return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: formatShopeeError(docResult.FailError, docResult.FailMessage)}, nil
						}
					}
				} else {
					return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: extractShopeeDocumentError(downloadResult, dlErr)}, nil
				}
			}
		} else if err != nil {
			return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: err.Error()}, nil
		}
		if i < 4 {
			time.Sleep(1 * time.Second)
		}
	}

	return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: "shopee API error: shipping document not ready after polling"}, nil
}
