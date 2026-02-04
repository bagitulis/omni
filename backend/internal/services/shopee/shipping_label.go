package shopee

import (
	"context"
	"fmt"
)

// ShippingLabelResult represents the shipping label download result
type ShippingLabelResult struct {
	OrderSN      string `json:"order_sn"`
	Status       string `json:"status"`
	FileData     string `json:"file_data,omitempty"`     // Base64 encoded PDF
	ErrorMessage string `json:"error_message,omitempty"` // Error if failed
}

// GetShippingLabel downloads shipping document (waybill/label) for an order
func (s *ShippingService) GetShippingLabel(ctx context.Context, orderSN, packageNumber, documentType string) (*ShippingLabelResult, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}

	// First, create the shipping document
	if _, err := s.ensureShipmentReady(ctx, client, orderSN, packageNumber); err != nil {
		return &ShippingLabelResult{
			OrderSN:      orderSN,
			Status:       "FAILED",
			ErrorMessage: err.Error(),
		}, nil
	}

	createResult, err := client.CreateShippingDocument(orderSN, packageNumber)
	if err != nil {
		// Handle "common.batch_api_all_failed" which is a business error, not a system error
		if createResult != nil && createResult.Error == "common.batch_api_all_failed" {
			// Fall through to process ResultList below
		} else {
			return nil, fmt.Errorf("create shipping document: %w", err)
		}
	}

	// Check create result status
	if len(createResult.Response.ResultList) > 0 {
		status := createResult.Response.ResultList[0].Status
		if status == "FAILED" {
			return &ShippingLabelResult{
				OrderSN:      orderSN,
				Status:       "FAILED",
				ErrorMessage: fmt.Sprintf("%s: %s", createResult.Response.ResultList[0].FailError, createResult.Response.ResultList[0].FailMessage),
			}, nil
		}
	}

	// Download the shipping document
	downloadResult, err := client.DownloadShippingDocument(orderSN, packageNumber, documentType)
	if err != nil {
		return nil, fmt.Errorf("download shipping document: %w", err)
	}

	if len(downloadResult.Response.ResultList) == 0 {
		return &ShippingLabelResult{
			OrderSN:      orderSN,
			Status:       "FAILED",
			ErrorMessage: "No shipping document available",
		}, nil
	}

	docResult := downloadResult.Response.ResultList[0]
	return &ShippingLabelResult{
		OrderSN:      orderSN,
		Status:       docResult.Status,
		FileData:     docResult.ShippingDocFile,
		ErrorMessage: docResult.FailMessage,
	}, nil
}
