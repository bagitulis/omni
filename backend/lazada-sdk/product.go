package lazadasdk

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type GetProductsParams struct {
	Filter        string // live|inactive|deleted|image-missing|all
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
	Search        string
	Offset        int
	Limit         int
	Options       string
	SkuSellerList []string
}

func (c *Client) GetProducts(ctx context.Context, p GetProductsParams) (*ProductListResponse, error) {
	if p.Filter == "" {
		return nil, ErrBadRequest("filter is required")
	}

	params := map[string]string{
		"filter": p.Filter,
	}
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
	if p.Search != "" {
		params["search"] = p.Search
	}
	if p.Offset > 0 {
		params["offset"] = strconv.Itoa(p.Offset)
	}
	if p.Limit > 0 {
		params["limit"] = strconv.Itoa(p.Limit)
	}
	if p.Options != "" {
		params["options"] = p.Options
	}
	if len(p.SkuSellerList) > 0 {
		data, err := json.Marshal(p.SkuSellerList)
		if err != nil {
			return nil, fmt.Errorf("marshal sku_seller_list: %w", err)
		}
		params["sku_seller_list"] = string(data)
	}

	var resp ProductListResponse
	if err := c.callRaw(ctx, "GET", "/products/get", params, &resp); err != nil {
		return nil, err
	}
	if resp.Code != "" && resp.Code != "0" {
		if resp.Message != "" {
			return &resp, fmt.Errorf("lazada API error: code=%s, message=%s", resp.Code, resp.Message)
		}
		return &resp, fmt.Errorf("lazada API error: code=%s", resp.Code)
	}
	return &resp, nil
}

// GetProductItem wraps GET /product/item/get.
// Lazada allows querying by item_id OR seller_sku.
func (c *Client) GetProductItem(ctx context.Context, itemID string, sellerSKU string) (*RawResponse, error) {
	params := map[string]string{}
	if itemID != "" {
		params["item_id"] = itemID
	}
	if sellerSKU != "" {
		params["seller_sku"] = sellerSKU
	}
	if len(params) == 0 {
		return nil, ErrBadRequest("item_id or seller_sku is required")
	}

	var resp RawResponse
	if err := c.callRaw(ctx, "GET", "/product/item/get", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdatePriceQuantity sends an XML payload to POST /product/price_quantity/update.
func (c *Client) UpdatePriceQuantity(ctx context.Context, xmlPayload string) (*RawResponse, error) {
	if xmlPayload == "" {
		return nil, ErrBadRequest("payload is required")
	}
	var resp RawResponse
	if err := c.callRaw(ctx, "POST", "/product/price_quantity/update", map[string]string{"payload": xmlPayload}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateProduct sends an XML payload to POST /product/update.
func (c *Client) UpdateProduct(ctx context.Context, xmlPayload string) (*RawResponse, error) {
	if xmlPayload == "" {
		return nil, ErrBadRequest("payload is required")
	}
	var resp RawResponse
	if err := c.callRaw(ctx, "POST", "/product/update", map[string]string{"payload": xmlPayload}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
