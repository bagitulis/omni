package lazada

import "fmt"

// ===== Stock/Quantity Update API =====

// UpdatePriceQuantityRequest represents Lazada price/quantity update request
type UpdatePriceQuantityRequest struct {
	ItemID    string `json:"item_id"`
	SkuID     string `json:"sku_id"`
	SellerSku string `json:"seller_sku"`
	Quantity  int    `json:"quantity"`
}

// UpdatePriceQuantityResponse represents update response
type UpdatePriceQuantityResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// UpdatePriceQuantity updates product stock quantity via Lazada API
// API: POST /product/price_quantity/update with XML payload
func (c *Client) UpdatePriceQuantity(req UpdatePriceQuantityRequest) (*UpdatePriceQuantityResponse, error) {
	// Lazada expects XML payload as string parameter
	xmlPayload := fmt.Sprintf(`<Request>
  <Product>
    <Skus>
      <Sku>
        <ItemId>%s</ItemId>
        <SkuId>%s</SkuId>
        <SellerSku>%s</SellerSku>
        <Quantity>%d</Quantity>
      </Sku>
    </Skus>
  </Product>
</Request>`, req.ItemID, req.SkuID, req.SellerSku, req.Quantity)

	params := map[string]string{
		"payload": xmlPayload,
	}

	var result UpdatePriceQuantityResponse
	err := c.doRequest("POST", "/product/price_quantity/update", params, &result)
	if err != nil {
		return nil, err
	}

	// Lazada returns code "0" for success
	if result.Code != "0" {
		return &result, fmt.Errorf("lazada API error: code=%s, message=%s", result.Code, result.Message)
	}

	return &result, nil
}

// ===== Price Update API =====

// UpdatePriceRequest represents Lazada price update request
type UpdatePriceRequest struct {
	ItemID    string  `json:"item_id"`
	SkuID     string  `json:"sku_id"`
	SellerSku string  `json:"seller_sku"`
	Price     float64 `json:"price"`
}

// UpdatePriceResponse represents price update response
type UpdatePriceResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// UpdatePrice updates product price via Lazada API
// API: POST /product/price_quantity/update with XML payload (uses Price instead of Quantity)
func (c *Client) UpdatePrice(req UpdatePriceRequest) (*UpdatePriceResponse, error) {
	// Lazada expects XML payload as string parameter
	xmlPayload := fmt.Sprintf(`<Request>
  <Product>
    <Skus>
      <Sku>
        <ItemId>%s</ItemId>
        <SkuId>%s</SkuId>
        <SellerSku>%s</SellerSku>
        <Price>%.2f</Price>
      </Sku>
    </Skus>
  </Product>
</Request>`, req.ItemID, req.SkuID, req.SellerSku, req.Price)

	params := map[string]string{
		"payload": xmlPayload,
	}

	var result UpdatePriceResponse
	err := c.doRequest("POST", "/product/price_quantity/update", params, &result)
	if err != nil {
		return nil, err
	}

	// Lazada returns code "0" for success
	if result.Code != "0" {
		return &result, fmt.Errorf("lazada API error: code=%s, message=%s", result.Code, result.Message)
	}

	return &result, nil
}
