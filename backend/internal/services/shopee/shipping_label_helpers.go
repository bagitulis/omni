package shopee

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	shopeePkg "github.com/omni/backend/pkg/shopee"
)

func (s *ShippingService) pollAndDownload(ctx context.Context, client shippingClient, orderSN, packageNumber, documentType, createErrorMessage string) (*ShippingLabelResult, error) {
	lastStatus := ""
	effectivePackageNumber := packageNumber
	for i := 0; i < 30; i++ {
		resultResp, err := client.GetShippingDocumentResultWithOptions(orderSN, effectivePackageNumber, shopeePkg.ShippingDocumentRequestOptions{
			ShippingDocumentType: documentType,
		})
		if err != nil {
			return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: err.Error()}, nil
		}

		if resultResp != nil {
			if resultResp.Error != "" || resultResp.Message != "" {
				return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: formatShopeeError(resultResp.Error, resultResp.Message)}, nil
			}

			if len(resultResp.Response.ResultList) > 0 {
				resultItem := resultResp.Response.ResultList[0]
				if effectivePackageNumber == "" && resultItem.PackageNumber != "" {
					effectivePackageNumber = resultItem.PackageNumber
				}
				status := strings.ToUpper(resultItem.Status)
				if status != "" {
					lastStatus = status
				}
				if resultItem.FailError != "" || resultItem.FailMessage != "" {
					return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: formatShopeeError(resultItem.FailError, resultItem.FailMessage)}, nil
				}
				if status == "FAILED" {
					return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: createErrorMessage}, nil
				}
				if status == "READY" || status == "SUCCESS" {
					return s.downloadDocument(client, orderSN, effectivePackageNumber, documentType, createErrorMessage)
				}
			}
		}

		if i < 29 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(2 * time.Second):
			}
		}
	}

	if lastStatus != "" || createErrorMessage != "" {
		return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: createErrorMessage}, nil
	}
	return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED"}, nil
}

func (s *ShippingService) downloadDocument(client shippingClient, orderSN, packageNumber, documentType, createErrorMessage string) (*ShippingLabelResult, error) {
	downloadResult, dlErr := client.DownloadShippingDocument(orderSN, packageNumber, documentType)
	if dlErr != nil {
		return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: firstNonEmpty(extractShopeeDocumentError(downloadResult, dlErr), dlErr.Error())}, nil
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
	return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: createErrorMessage}, nil
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
	return ""
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
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
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

func appendDocumentTypesFromParameter(resp *shopeePkg.GetShippingDocumentParameterResponse, appendType func(string)) string {
	if resp == nil {
		return ""
	}
	if resp.Error != "" || resp.Message != "" {
		return formatShopeeError(resp.Error, resp.Message)
	}
	if len(resp.Response.ResultList) == 0 {
		return ""
	}
	item := resp.Response.ResultList[0]
	if item.FailError != "" || item.FailMessage != "" {
		return formatShopeeError(item.FailError, item.FailMessage)
	}
	appendType(item.SuggestShippingDocumentType)
	for _, selectable := range item.SelectableShippingDocumentType {
		appendType(selectable)
	}
	return ""
}
