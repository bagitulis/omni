package lazadasdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

func (c *Client) GetOrders(ctx context.Context, p GetOrdersParams) (*OrderListResponse, error) {
	params := map[string]string{}

	// Lazada uses ISO8601 strings; RFC3339 is accepted by Lazada in practice.
	if p.CreatedAfter != nil {
		params["created_after"] = p.CreatedAfter.Format(time.RFC3339)
	}
	if p.CreatedBefore != nil {
		params["created_before"] = p.CreatedBefore.Format(time.RFC3339)
	}
	if p.UpdatedAfter != nil {
		params["update_after"] = p.UpdatedAfter.Format(time.RFC3339)
	}
	if p.UpdatedBefore != nil {
		params["update_before"] = p.UpdatedBefore.Format(time.RFC3339)
	}
	if p.Status != "" {
		params["status"] = p.Status
	}
	if p.SortDirection != "" {
		params["sort_direction"] = p.SortDirection
	}
	if p.SortBy != "" {
		params["sort_by"] = p.SortBy
	}
	if p.Offset > 0 {
		params["offset"] = strconv.Itoa(p.Offset)
	}
	if p.Limit > 0 {
		params["limit"] = strconv.Itoa(p.Limit)
	}

	var resp OrderListResponse
	if err := c.callRaw(ctx, "GET", "/orders/get", params, &resp); err != nil {
		return nil, err
	}
	if resp.Code != "" && resp.Code != "0" {
		return &resp, fmt.Errorf("lazada API error: code=%s", resp.Code)
	}
	return &resp, nil
}

func (c *Client) GetOrder(ctx context.Context, orderID string) (*RawResponse, error) {
	if orderID == "" {
		return nil, ErrBadRequest("order_id is required")
	}
	var resp RawResponse
	if err := c.callRaw(ctx, "GET", "/order/get", map[string]string{"order_id": orderID}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetOrderItems(ctx context.Context, orderID string) (*RawResponse, error) {
	if orderID == "" {
		return nil, ErrBadRequest("order_id is required")
	}
	var resp RawResponse
	if err := c.callRaw(ctx, "GET", "/order/items/get", map[string]string{"order_id": orderID}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetMultipleOrderItems(ctx context.Context, orderIDs []int64) (*RawResponse, error) {
	if len(orderIDs) == 0 {
		return nil, ErrBadRequest("order_ids is required")
	}
	encoded, err := marshalJSONArrayInts(orderIDs)
	if err != nil {
		return nil, err
	}
	var resp RawResponse
	if err := c.callRaw(ctx, "GET", "/orders/items/get", map[string]string{"order_ids": encoded}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetDocument(ctx context.Context, docType string, orderItemIDs []int64) (*RawResponse, error) {
	if docType == "" {
		return nil, ErrBadRequest("doc_type is required")
	}
	if len(orderItemIDs) == 0 {
		return nil, ErrBadRequest("order_item_ids is required")
	}
	encoded, err := marshalJSONArrayInts(orderItemIDs)
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"doc_type":       docType,
		"order_item_ids": encoded,
	}
	var resp RawResponse
	if err := c.callRaw(ctx, "GET", "/order/document/get", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetFailureReasons(ctx context.Context) (*RawResponse, error) {
	var resp RawResponse
	if err := c.callRaw(ctx, "GET", "/order/failure_reason/get", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) SetInvoiceNumber(ctx context.Context, orderItemID int64, invoiceNumber string) (*RawResponse, error) {
	if orderItemID == 0 {
		return nil, ErrBadRequest("order_item_id is required")
	}
	if invoiceNumber == "" {
		return nil, ErrBadRequest("invoice_number is required")
	}
	params := map[string]string{
		"order_item_id":  strconv.FormatInt(orderItemID, 10),
		"invoice_number": invoiceNumber,
	}
	var resp RawResponse
	if err := c.callRaw(ctx, "POST", "/order/invoice_number/set", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) SetStatusToCanceled(ctx context.Context, orderItemID int64, reasonID int, reasonDetail string) (*RawResponse, error) {
	if orderItemID == 0 {
		return nil, ErrBadRequest("order_item_id is required")
	}
	if reasonID == 0 {
		return nil, ErrBadRequest("reason_id is required")
	}
	params := map[string]string{
		"order_item_id": strconv.FormatInt(orderItemID, 10),
		"reason_id":     strconv.Itoa(reasonID),
	}
	if reasonDetail != "" {
		params["reason_detail"] = reasonDetail
	}
	var resp RawResponse
	if err := c.callRaw(ctx, "POST", "/order/cancel", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetStatusToPacked wraps POST /order/pack.
// deliveryType is typically "dropship"; shippingProvider is required for dropship.
func (c *Client) SetStatusToPacked(ctx context.Context, deliveryType string, orderItemIDs []int64, shippingProvider string) (*RawResponse, error) {
	if deliveryType == "" {
		return nil, ErrBadRequest("delivery_type is required")
	}
	if len(orderItemIDs) == 0 {
		return nil, ErrBadRequest("order_item_ids is required")
	}
	if deliveryType == "dropship" && shippingProvider == "" {
		return nil, ErrBadRequest("shipping_provider is required for dropship")
	}
	encoded, err := marshalJSONArrayStrings(orderItemIDs)
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"delivery_type":  deliveryType,
		"order_item_ids": encoded,
	}
	if shippingProvider != "" {
		params["shipping_provider"] = shippingProvider
	}
	var resp RawResponse
	if err := c.callRaw(ctx, "POST", "/order/pack", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetStatusToRTS wraps POST /order/rts.
func (c *Client) SetStatusToRTS(ctx context.Context, deliveryType string, orderItemIDs []int64, shipmentProvider string, trackingNumber string) (*RawResponse, error) {
	if len(orderItemIDs) == 0 {
		return nil, ErrBadRequest("order_item_ids is required")
	}
	if shipmentProvider == "" {
		return nil, ErrBadRequest("shipment_provider is required")
	}
	if trackingNumber == "" {
		return nil, ErrBadRequest("tracking_number is required")
	}
	if deliveryType == "" {
		deliveryType = "dropship"
	}
	encoded, err := marshalJSONArrayStrings(orderItemIDs)
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"delivery_type":     deliveryType,
		"order_item_ids":    encoded,
		"shipment_provider": shipmentProvider,
		"tracking_number":   trackingNumber,
	}
	var resp RawResponse
	if err := c.callRaw(ctx, "POST", "/order/rts", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func marshalJSONArrayInts(ids []int64) (string, error) {
	data, err := json.Marshal(ids)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func marshalJSONArrayStrings(ids []int64) (string, error) {
	strs := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			return "", errors.New("ids must be non-zero")
		}
		strs = append(strs, strconv.FormatInt(id, 10))
	}
	data, err := json.Marshal(strs)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
