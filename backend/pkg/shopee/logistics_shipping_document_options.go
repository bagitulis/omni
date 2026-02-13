package shopee

import (
	"encoding/json"
	"fmt"
)

// ShippingDocumentRequestOptions contains optional fields for shipping document APIs.
type ShippingDocumentRequestOptions struct {
	TrackingNumber       string
	ShippingDocumentType string
}

type shippingDocumentOrderWithOptions struct {
	OrderSN              string `json:"order_sn"`
	PackageNumber        string `json:"package_number,omitempty"`
	TrackingNumber       string `json:"tracking_number,omitempty"`
	ShippingDocumentType string `json:"shipping_document_type,omitempty"`
}

type createShippingDocumentWithOptionsRequest struct {
	OrderList []shippingDocumentOrderWithOptions `json:"order_list"`
}

type getShippingDocumentResultWithOptionsRequest struct {
	OrderList []shippingDocumentOrderWithOptions `json:"order_list"`
}

// GetShippingDocumentParameterResponse represents selectable/suggested document types.
type GetShippingDocumentParameterResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ResultList []struct {
			OrderSN                        string   `json:"order_sn"`
			PackageNumber                  string   `json:"package_number"`
			SuggestShippingDocumentType    string   `json:"suggest_shipping_document_type"`
			SelectableShippingDocumentType []string `json:"selectable_shipping_document_type"`
			FailError                      string   `json:"fail_error,omitempty"`
			FailMessage                    string   `json:"fail_message,omitempty"`
		} `json:"result_list"`
	} `json:"response"`
}

type getShippingDocumentParameterRequest struct {
	OrderList []struct {
		OrderSN       string `json:"order_sn"`
		PackageNumber string `json:"package_number,omitempty"`
	} `json:"order_list"`
}

// CreateShippingDocumentWithOptions creates shipping document with tracking/document type options.
func (c *Client) CreateShippingDocumentWithOptions(orderSN, packageNumber string, options ShippingDocumentRequestOptions) (*CreateShippingDocumentResponse, error) {
	req := createShippingDocumentWithOptionsRequest{
		OrderList: []shippingDocumentOrderWithOptions{
			{
				OrderSN:              orderSN,
				PackageNumber:        packageNumber,
				TrackingNumber:       options.TrackingNumber,
				ShippingDocumentType: options.ShippingDocumentType,
			},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var result CreateShippingDocumentResponse
	if err := c.doPostRequest("/api/v2/logistics/create_shipping_document", nil, body, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetShippingDocumentResultWithOptions gets shipping document result with document type option.
func (c *Client) GetShippingDocumentResultWithOptions(orderSN, packageNumber string, options ShippingDocumentRequestOptions) (*GetShippingDocumentResultResponse, error) {
	req := getShippingDocumentResultWithOptionsRequest{
		OrderList: []shippingDocumentOrderWithOptions{
			{
				OrderSN:              orderSN,
				PackageNumber:        packageNumber,
				ShippingDocumentType: options.ShippingDocumentType,
			},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var result GetShippingDocumentResultResponse
	if err := c.doPostRequest("/api/v2/logistics/get_shipping_document_result", nil, body, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetShippingDocumentParameter gets selectable/suggested shipping document types.
func (c *Client) GetShippingDocumentParameter(orderSN, packageNumber string) (*GetShippingDocumentParameterResponse, error) {
	req := getShippingDocumentParameterRequest{}
	req.OrderList = append(req.OrderList, struct {
		OrderSN       string `json:"order_sn"`
		PackageNumber string `json:"package_number,omitempty"`
	}{
		OrderSN:       orderSN,
		PackageNumber: packageNumber,
	})

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var result GetShippingDocumentParameterResponse
	if err := c.doPostRequest("/api/v2/logistics/get_shipping_document_parameter", nil, body, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}
