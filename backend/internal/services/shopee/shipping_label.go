package shopee

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
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
	trackingErrorMessage := extractShopeeTrackingError(trackingResp, trackingErr)

	documentTypes := []string{
		documentType,
		"THERMAL_AIR_WAYBILL",
		"NORMAL_AIR_WAYBILL",
		"THERMAL_WAYBILL",
		"NORMAL_WAYBILL",
	}

	tried := make(map[string]struct{}, len(documentTypes))
	lastShopeeError := trackingErrorMessage
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
		ErrorMessage: "shopee API error: no error payload returned by create/result/download endpoints",
	}, nil
}

func (s *ShippingService) tryStandardDownload(ctx context.Context, client shippingClient, orderSN, packageNumber, documentType string) (*ShippingLabelResult, error) {
	_ = ctx
	createResult, err := client.CreateShippingDocument(orderSN, packageNumber)
	createErrorMessage := extractShopeeCreateError(createResult, err)

	if createResult == nil && err != nil {
		return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: err.Error()}, nil
	}

	if createResult != nil {
		if len(createResult.Response.ResultList) > 0 {
			item := createResult.Response.ResultList[0]
			if item.FailError != "" || item.FailMessage != "" {
				if item.FailError != "logistics.document_already_created" {
					return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: formatShopeeError(item.FailError, item.FailMessage)}, nil
				}
			}
		}

		if createResult.Error != "" || createResult.Message != "" {
			if createResult.Error != "common.batch_api_all_failed" {
				return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: formatShopeeError(createResult.Error, createResult.Message)}, nil
			}
		}
	}

	return s.pollAndDownload(ctx, client, orderSN, packageNumber, documentType, createErrorMessage)
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

func extractShopeeTrackingError(trackingResp *shopeePkg.GetTrackingNumberResponse, trackingErr error) string {
	if trackingErr != nil {
		return trackingErr.Error()
	}

	if trackingResp == nil {
		return ""
	}

	if trackingResp.Error != "" || trackingResp.Message != "" {
		return formatShopeeError(trackingResp.Error, trackingResp.Message)
	}

	if trackingResp.Response.Hint != "" {
		return formatShopeeError("", trackingResp.Response.Hint)
	}

	return ""
}

func (s *ShippingService) pollAndDownload(ctx context.Context, client shippingClient, orderSN, packageNumber, documentType, createErrorMessage string) (*ShippingLabelResult, error) {
	_ = ctx
	lastStatus := ""
	lastErrorMessage := ""
	for i := 0; i < 30; i++ {
		resultResp, err := client.GetShippingDocumentResult(orderSN, packageNumber)
		if err != nil {
			return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: err.Error()}, nil
		}

		if resultResp != nil {
			if resultResp.Error != "" || resultResp.Message != "" {
				return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: formatShopeeError(resultResp.Error, resultResp.Message)}, nil
			}

			if len(resultResp.Response.ResultList) > 0 {
				resultItem := resultResp.Response.ResultList[0]
				status := strings.ToUpper(resultItem.Status)
				if status != "" {
					lastStatus = status
				}
				if resultItem.FailError != "" || resultItem.FailMessage != "" {
					return &ShippingLabelResult{
						OrderSN:      orderSN,
						Status:       "FAILED",
						ErrorMessage: formatShopeeError(resultItem.FailError, resultItem.FailMessage),
					}, nil
				}
				if status == "FAILED" {
					return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: formatShopeeError("", "shipping_document_result status=FAILED")}, nil
				}
				if status == "READY" || status == "SUCCESS" {
					downloadResult, dlErr := client.DownloadShippingDocument(orderSN, packageNumber, documentType)
					if dlErr != nil {
						return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: extractShopeeDocumentError(downloadResult, dlErr)}, nil
					}
					if len(downloadResult.RawPDF) > 0 {
						return &ShippingLabelResult{OrderSN: orderSN, Status: "SUCCESS", FileData: base64.StdEncoding.EncodeToString(downloadResult.RawPDF)}, nil
					}
					if len(downloadResult.Response.ResultList) > 0 {
						docResult := downloadResult.Response.ResultList[0]
						if docResult.ShippingDocFile != "" {
							return &ShippingLabelResult{OrderSN: orderSN, Status: "SUCCESS", FileData: docResult.ShippingDocFile}, nil
						}
						if docResult.FailError != "" || docResult.FailMessage != "" {
							return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: formatShopeeError(docResult.FailError, docResult.FailMessage)}, nil
						}
					}
					if downloadResult.Error != "" || downloadResult.Message != "" {
						return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: formatShopeeError(downloadResult.Error, downloadResult.Message)}, nil
					}
					return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: formatShopeeError("", "download_shipping_document returned no file data")}, nil
				}
			}
		}

		if i < 29 {
			time.Sleep(2 * time.Second)
		}
	}

	if lastErrorMessage != "" {
		return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: lastErrorMessage}, nil
	}
	if lastStatus != "" {
		return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: formatShopeeError("", "shipping_document_result status="+lastStatus)}, nil
	}
	if createErrorMessage != "" {
		return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: createErrorMessage}, nil
	}

	return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: formatShopeeError("", "get_shipping_document_result returned no result status")}, nil
}
