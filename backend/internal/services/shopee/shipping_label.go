package shopee

import (
	"context"
	"slices"
	"strings"

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

	effectivePackageNumber := packageNumber
	trackingResp, trackingErr := s.ensureShipmentReady(ctx, client, orderSN, effectivePackageNumber)
	trackingNumber := ""
	if trackingResp != nil {
		trackingNumber = strings.TrimSpace(trackingResp.Response.TrackingNumber)
	}

	resolvedPackageNumber, packageResolveError := resolvePackageNumber(ctx, client, orderSN, packageNumber, trackingResp)
	if resolvedPackageNumber != "" {
		effectivePackageNumber = resolvedPackageNumber
	}

	trackingErrorMessage := firstNonEmpty(
		extractShopeeTrackingError(trackingResp, trackingErr),
		packageResolveError,
	)

	documentTypes, documentTypeError := resolveDocumentTypes(client, orderSN, effectivePackageNumber, documentType)
	if documentTypeError != "" {
		lastShopeeError := firstNonEmpty(documentTypeError, trackingErrorMessage)
		return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: lastShopeeError}, nil
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

		result, downloadErr := s.tryStandardDownload(ctx, client, orderSN, effectivePackageNumber, trackingNumber, docType)
		if downloadErr == nil && result.Status == "SUCCESS" && result.FileData != "" {
			return result, nil
		}
		if result != nil && result.ErrorMessage != "" {
			lastShopeeError = result.ErrorMessage
		}
	}

	log.Warn().Str("order_sn", orderSN).Msg("Shopee did not return official shipping document")
	if lastShopeeError != "" {
		return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: lastShopeeError}, nil
	}
	return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED"}, nil
}

func (s *ShippingService) tryStandardDownload(ctx context.Context, client shippingClient, orderSN, packageNumber, trackingNumber, documentType string) (*ShippingLabelResult, error) {
	_ = ctx
	createResult, err := client.CreateShippingDocumentWithOptions(orderSN, packageNumber, shopeePkg.ShippingDocumentRequestOptions{
		TrackingNumber: trackingNumber, ShippingDocumentType: documentType,
	})
	createErrorMessage := extractShopeeCreateError(createResult, err)
	effectivePackageNumber := packageNumber

	if createResult == nil && err != nil {
		return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: err.Error()}, nil
	}

	if createResult != nil {
		if len(createResult.Response.ResultList) > 0 {
			item := createResult.Response.ResultList[0]
			if effectivePackageNumber == "" && item.PackageNumber != "" {
				effectivePackageNumber = item.PackageNumber
			}
			if item.FailError != "" || item.FailMessage != "" {
				if item.FailError == "logistics.tracking_number_invalid" && packageNumber != "" {
					return s.tryStandardDownload(ctx, client, orderSN, "", trackingNumber, documentType)
				}
				if item.FailError != "logistics.document_already_created" {
					return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: firstNonEmpty(formatShopeeError(item.FailError, item.FailMessage), createErrorMessage)}, nil
				}
			}
		}
		if createResult.Error != "" || createResult.Message != "" {
			if createResult.Error != "common.batch_api_all_failed" {
				return &ShippingLabelResult{OrderSN: orderSN, Status: "FAILED", ErrorMessage: firstNonEmpty(formatShopeeError(createResult.Error, createResult.Message), createErrorMessage)}, nil
			}
		}
	}

	return s.pollAndDownload(ctx, client, orderSN, effectivePackageNumber, documentType, createErrorMessage)
}

func resolveDocumentTypes(client shippingClient, orderSN, packageNumber, requestedType string) ([]string, string) {
	result := make([]string, 0, 8)
	seen := make(map[string]struct{}, 8)
	parameterTypes := make([]string, 0, 4)
	appendType := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		if !isKnownShopeeDocumentType(trimmed) {
			return
		}
		if _, exists := seen[trimmed]; exists {
			return
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	appendParameterType := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || !isKnownShopeeDocumentType(trimmed) {
			return
		}
		if slices.Contains(parameterTypes, trimmed) {
			return
		}
		parameterTypes = append(parameterTypes, trimmed)
		appendType(trimmed)
	}

	var lastError string
	apiResult, err := client.GetShippingDocumentParameter(orderSN, packageNumber)
	if err != nil {
		lastError = err.Error()
	} else {
		lastError = appendDocumentTypesFromParameter(apiResult, appendParameterType)
	}

	if packageNumber != "" && len(parameterTypes) == 0 {
		fallbackResult, fallbackErr := client.GetShippingDocumentParameter(orderSN, "")
		if fallbackErr != nil {
			if lastError == "" {
				lastError = fallbackErr.Error()
			}
		} else {
			fallbackParamErr := appendDocumentTypesFromParameter(fallbackResult, appendParameterType)
			if lastError == "" {
				lastError = fallbackParamErr
			}
		}
	}

	trimmedRequestedType := strings.TrimSpace(requestedType)
	if len(parameterTypes) == 0 || slices.Contains(parameterTypes, trimmedRequestedType) {
		appendType(trimmedRequestedType)
	}

	appendKnownShopeeDocumentTypes(appendType)

	return result, lastError
}

func isKnownShopeeDocumentType(value string) bool {
	return slices.Contains(knownShopeeDocumentTypes(), value)
}

func appendKnownShopeeDocumentTypes(appendType func(string)) {
	for _, documentType := range knownShopeeDocumentTypes() {
		appendType(documentType)
	}
}

func knownShopeeDocumentTypes() []string {
	return []string{
		"THERMAL_AIR_WAYBILL",
		"NORMAL_AIR_WAYBILL",
		"THERMAL_JOB_AIR_WAYBILL",
		"NORMAL_JOB_AIR_WAYBILL",
		"THERMAL_UNPACKAGED_LABEL",
	}
}
