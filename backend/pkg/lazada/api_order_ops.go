package lazada

import "fmt"

// ===== Order Operations =====

// BaseResponse is common response structure
type BaseResponse struct {
	Code      string `json:"code"`
	RequestID string `json:"request_id"`
}

// SetStatusToPackedByMarketplaceResponse represents packed response
type SetStatusToPackedByMarketplaceResponse struct {
	BaseResponse
	Data struct {
		OrderItems []struct {
			OrderItemID  string `json:"order_item_id"`
			ShipmentType string `json:"shipment_type"`
			ShipmentCode string `json:"shipment_code"`
		} `json:"order_items"`
	} `json:"data"`
}

// SetStatusToPackedByMarketplace marks order items as packed.
// deliveryType: "dropship" (seller ships via 3PL) or "pickup" (carrier picks up from seller).
// Per Lazada Open Platform API, "dropship" is the default for most seller-fulfilled orders.
func (c *Client) SetStatusToPackedByMarketplace(orderItemIDs []string, shipmentProvider, deliveryType string) (*SetStatusToPackedByMarketplaceResponse, error) {
	if deliveryType == "" {
		deliveryType = "dropship"
	}
	params := map[string]string{
		"order_item_ids":    fmt.Sprintf("[%s]", joinQuoted(orderItemIDs)),
		"delivery_type":     deliveryType,
		"shipping_provider": shipmentProvider,
	}

	var result SetStatusToPackedByMarketplaceResponse
	err := c.doRequest("POST", "/order/pack", params, &result)
	return &result, err
}

// SetStatusToReadyToShipResponse represents ready to ship response
type SetStatusToReadyToShipResponse struct {
	BaseResponse
	Data struct {
		OrderItems []struct {
			OrderItemID  string `json:"order_item_id"`
			TrackingCode string `json:"tracking_code"`
		} `json:"order_items"`
	} `json:"data"`
}

// SetStatusToReadyToShip marks order items as ready to ship.
// deliveryType: "dropship" (seller ships via 3PL) or "pickup" (carrier picks up from seller).
// Per Lazada Open Platform API, "dropship" is the default for most seller-fulfilled orders.
func (c *Client) SetStatusToReadyToShip(orderItemIDs []string, shipmentProvider, trackingNumber, deliveryType string) (*SetStatusToReadyToShipResponse, error) {
	if deliveryType == "" {
		deliveryType = "dropship"
	}
	params := map[string]string{
		"order_item_ids":    fmt.Sprintf("[%s]", joinQuoted(orderItemIDs)),
		"delivery_type":     deliveryType,
		"shipping_provider": shipmentProvider,
	}
	if trackingNumber != "" {
		params["tracking_number"] = trackingNumber
	}

	var result SetStatusToReadyToShipResponse
	err := c.doRequest("POST", "/order/rts", params, &result)
	return &result, err
}

// CancelOrder cancels order items
func (c *Client) CancelOrder(orderItemID, reasonDetail, reasonID string) (*BaseResponse, error) {
	params := map[string]string{
		"order_item_id": orderItemID,
		"reason_detail": reasonDetail,
		"reason_id":     reasonID,
	}

	var result BaseResponse
	err := c.doRequest("POST", "/order/cancel", params, &result)
	return &result, err
}

// GetDocumentRequest represents get document request
type GetDocumentRequest struct {
	OrderItemIDs []string `json:"order_item_ids"`
	DocType      string   `json:"doc_type"` // "shippingLabel", "invoice", "carrierManifest"
}

// GetDocumentResponse represents get document response
type GetDocumentResponse struct {
	BaseResponse
	Data struct {
		Document struct {
			File     string `json:"file"`      // Base64 encoded PDF
			URL      string `json:"url"`       // PDF URL
			MimeType string `json:"mime_type"` // "application/pdf"
		} `json:"document"`
	} `json:"data"`
}

// GetDocument retrieves shipping documents (shipping label, invoice, etc.)
func (c *Client) GetDocument(req GetDocumentRequest) (*GetDocumentResponse, error) {
	params := map[string]string{
		"doc_type":       req.DocType,
		"order_item_ids": fmt.Sprintf("[%s]", joinQuoted(req.OrderItemIDs)),
	}

	var result GetDocumentResponse
	err := c.doRequest("POST", "/order/document/get", params, &result)
	return &result, err
}

// ===== Helper Functions =====

func joinQuoted(items []string) string {
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = fmt.Sprintf(`"%s"`, item)
	}
	return joinStrings(quoted, ",")
}

func joinStrings(items []string, sep string) string {
	if len(items) == 0 {
		return ""
	}
	result := items[0]
	for i := 1; i < len(items); i++ {
		result += sep + items[i]
	}
	return result
}
