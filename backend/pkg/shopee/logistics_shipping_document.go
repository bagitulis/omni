package shopee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

// CreateShippingDocumentRequest represents request to create shipping document
type CreateShippingDocumentRequest struct {
	OrderList []ShippingDocumentOrder `json:"order_list"`
}

// ShippingDocumentOrder represents an order for shipping document
type ShippingDocumentOrder struct {
	OrderSN       string `json:"order_sn"`
	PackageNumber string `json:"package_number,omitempty"`
}

// CreateShippingDocumentResponse represents create shipping document response
type CreateShippingDocumentResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ResultList []struct {
			OrderSN       string `json:"order_sn"`
			PackageNumber string `json:"package_number"`
			Status        string `json:"status"`
			FailError     string `json:"fail_error,omitempty"`
			FailMessage   string `json:"fail_message,omitempty"`
		} `json:"result_list"`
		Warning []string `json:"warning,omitempty"`
	} `json:"response"`
}

// GetShippingDocumentResultRequest represents request to get shipping document result
type GetShippingDocumentResultRequest struct {
	OrderList []ShippingDocumentOrder `json:"order_list"`
}

// GetShippingDocumentResultResponse represents get shipping document result response
type GetShippingDocumentResultResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ResultList []struct {
			OrderSN       string `json:"order_sn"`
			PackageNumber string `json:"package_number"`
			Status        string `json:"status"`
			FailError     string `json:"fail_error,omitempty"`
			FailMessage   string `json:"fail_message,omitempty"`
		} `json:"result_list"`
	} `json:"response"`
}

// DownloadShippingDocumentRequest represents request to download shipping document
type DownloadShippingDocumentRequest struct {
	OrderList    []ShippingDocumentOrder `json:"order_list"`
	DocumentType string                  `json:"document_type,omitempty"`
	DocumentSize string                  `json:"document_size,omitempty"`
}

// DownloadShippingDocumentResponse represents response for download (contains file data or status)
type DownloadShippingDocumentResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ResultList []struct {
			OrderSN         string `json:"order_sn"`
			PackageNumber   string `json:"package_number"`
			Status          string `json:"status"`
			FailError       string `json:"fail_error,omitempty"`
			FailMessage     string `json:"fail_message,omitempty"`
			ShippingDocFile string `json:"shipping_document_file,omitempty"`
		} `json:"result_list"`
	} `json:"response"`
	// RawPDF holds raw binary PDF data if response is not JSON
	RawPDF []byte `json:"-"`
}

// ShippingDocumentDataInfoRequest for get_shipping_document_data_info API
type ShippingDocumentDataInfoRequest struct {
	OrderSN       string `json:"order_sn"`
	PackageNumber string `json:"package_number,omitempty"`
}

// ShippingDocumentDataInfoResponse represents response from get_shipping_document_data_info
type ShippingDocumentDataInfoResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		OrderSN                 string  `json:"order_sn"`
		PackageNumber           string  `json:"package_number"`
		LogisticsChannelID      int     `json:"logistics_channel_id"`
		LogisticsChannelName    string  `json:"logistics_channel_name"`
		FirstMileTrackingNumber string  `json:"first_mile_tracking_number"`
		LastMileTrackingNumber  string  `json:"last_mile_tracking_number"`
		TrackingNumber          string  `json:"tracking_number"`
		ShippingCarrier         string  `json:"shipping_carrier"`
		SenderName              string  `json:"sender_name"`
		SenderPhone             string  `json:"sender_phone"`
		SenderAddress           string  `json:"sender_address"`
		SenderCity              string  `json:"sender_city"`
		SenderDistrict          string  `json:"sender_district"`
		SenderState             string  `json:"sender_state"`
		SenderZipcode           string  `json:"sender_zipcode"`
		SenderCountry           string  `json:"sender_country"`
		RecipientName           string  `json:"recipient_name"`
		RecipientPhone          string  `json:"recipient_phone"`
		RecipientAddress        string  `json:"recipient_address"`
		RecipientCity           string  `json:"recipient_city"`
		RecipientDistrict       string  `json:"recipient_district"`
		RecipientState          string  `json:"recipient_state"`
		RecipientZipcode        string  `json:"recipient_zipcode"`
		RecipientCountry        string  `json:"recipient_country"`
		RecipientSortCode       string  `json:"recipient_sort_code"`
		ServiceDescription      string  `json:"service_description"`
		BuyerCodAmount          float64 `json:"buyer_cod_amount"`
	} `json:"response"`
}

// CreateShippingDocument creates shipping document for orders
func (c *Client) CreateShippingDocument(orderSN string, packageNumber string) (*CreateShippingDocumentResponse, error) {
	req := CreateShippingDocumentRequest{
		OrderList: []ShippingDocumentOrder{
			{OrderSN: orderSN, PackageNumber: packageNumber},
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

// GetShippingDocumentResult gets shipping document creation result
func (c *Client) GetShippingDocumentResult(orderSN string, packageNumber string) (*GetShippingDocumentResultResponse, error) {
	req := GetShippingDocumentResultRequest{
		OrderList: []ShippingDocumentOrder{
			{OrderSN: orderSN, PackageNumber: packageNumber},
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

// DownloadShippingDocument downloads shipping document (waybill/label) for an order
func (c *Client) DownloadShippingDocument(orderSN string, packageNumber string, documentType string) (*DownloadShippingDocumentResponse, error) {
	if documentType == "" {
		documentType = "THERMAL_AIR_WAYBILL"
	}

	req := DownloadShippingDocumentRequest{
		OrderList: []ShippingDocumentOrder{
			{OrderSN: orderSN, PackageNumber: packageNumber},
		},
		DocumentType: documentType,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	return c.doDownloadRequest("/api/v2/logistics/download_shipping_document", body)
}

// GetShippingDocumentDataInfo gets shipping document data without downloading PDF
func (c *Client) GetShippingDocumentDataInfo(orderSN string, packageNumber string) (*ShippingDocumentDataInfoResponse, error) {
	req := ShippingDocumentDataInfoRequest{
		OrderSN:       orderSN,
		PackageNumber: packageNumber,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var result ShippingDocumentDataInfoResponse
	if err := c.doPostRequest("/api/v2/logistics/get_shipping_document_data_info", nil, body, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// doDownloadRequest handles shipping document download which can return JSON or binary PDF
func (c *Client) doDownloadRequest(path string, body []byte) (*DownloadShippingDocumentResponse, error) {
	reqURL, urlErr := c.buildURL(path, nil)
	if urlErr != nil {
		return nil, urlErr
	}

	log.Info().Str("path", path).Msg("[Shopee API] POST download request")

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(body))
	if err != nil {
		log.Error().Err(err).Msg("[Shopee API] Failed to create request")
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("[Shopee API] HTTP request failed")
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("[Shopee API] Failed to read response")
		return nil, err
	}

	log.Info().Str("status", resp.Status).Msg("[Shopee API] Download response")
	contentType := resp.Header.Get("Content-Type")

	var result DownloadShippingDocumentResponse

	if strings.Contains(contentType, "application/pdf") || strings.Contains(contentType, "application/octet-stream") {
		log.Info().Int("size", len(respBody)).Msg("[Shopee API] Received binary PDF")
		result.RawPDF = respBody
		return &result, nil
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Error().Err(err).Msg("[Shopee API] Failed to parse response")
		return nil, err
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// doPostRequest executes POST request with body
func (c *Client) doPostRequest(path string, params map[string]string, body []byte, result interface{}) error {
	reqURL, urlErr := c.buildURL(path, params)
	if urlErr != nil {
		return urlErr
	}

	log.Info().Str("path", path).Msg("[Shopee API] POST request")

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(body))
	if err != nil {
		log.Error().Err(err).Msg("[Shopee API] Failed to create request")
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("[Shopee API] HTTP request failed")
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("[Shopee API] Failed to read response")
		return err
	}

	log.Info().Str("status", resp.Status).Str("path", path).Msg("[Shopee API] Response")

	if err := json.Unmarshal(respBody, result); err != nil {
		log.Error().Err(err).Msg("[Shopee API] Failed to parse response")
		return err
	}

	return nil
}
