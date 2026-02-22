package lazada

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// FlexibleString handles JSON that can be either string or number
// Lazada API returns item_id as number, but we need it as string
type FlexibleString string

// UnmarshalJSON implements json.Unmarshaler for FlexibleString
func (f *FlexibleString) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexibleString(s)
		return nil
	}

	// Try to unmarshal as number
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexibleString(n.String())
		return nil
	}

	// Try to unmarshal as int64 directly
	var i int64
	if err := json.Unmarshal(data, &i); err == nil {
		*f = FlexibleString(strconv.FormatInt(i, 10))
		return nil
	}

	return fmt.Errorf("FlexibleString: cannot unmarshal %s", string(data))
}

// String returns the string value
func (f FlexibleString) String() string {
	return string(f)
}

// OrderListResponse represents Lazada order list response
type OrderListResponse struct {
	Code string `json:"code"`
	Data struct {
		Count  int `json:"count"`
		Orders []struct {
			OrderID          string  `json:"order_id"`
			OrderNumber      string  `json:"order_number"`
			Status           string  `json:"status"`
			Price            float64 `json:"price"`
			CustomerName     string  `json:"customer_first_name"`
			PromisedShipDate string  `json:"promised_shipping_times"` // Shipping deadline (ISO date)
			ShippingType     string  `json:"delivery_info"`           // Shipping carrier/type
			CreatedAt        string  `json:"created_at"`              // Order creation time
			UpdatedAt        string  `json:"updated_at"`              // Order update time
		} `json:"orders"`
	} `json:"data"`
}

// GetOrders fetches orders from Lazada API
func (c *Client) GetOrders(status string, offset, limit int) (*OrderListResponse, error) {
	params := map[string]string{
		"status": status,
		"offset": fmt.Sprintf("%d", offset),
		"limit":  fmt.Sprintf("%d", limit),
	}

	var result OrderListResponse
	err := c.doRequest("GET", "/orders/get", params, &result)
	if err != nil {
		return nil, err
	}

	// Check for API error in response
	if result.Code != "0" && result.Code != "" {
		return nil, fmt.Errorf("lazada API error: %s", result.Code)
	}

	return &result, nil
}

// ProductListResponse represents Lazada product list response
type ProductListResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Data    struct {
		TotalProducts int       `json:"total_products"`
		Products      []Product `json:"products"`
	} `json:"data"`
}

// Product represents a Lazada product with SKUs
type Product struct {
	ItemID      FlexibleString    `json:"item_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Brand       string            `json:"brand"`
	Price       float64           `json:"price"`
	Status      string            `json:"status"`
	Images      []string          `json:"images"`
	Attributes  ProductAttributes `json:"attributes"`
	Skus        []ProductSku      `json:"skus"`
}

// ProductAttributes represents product attributes from Lazada API
type ProductAttributes struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Brand       string `json:"brand"`
}

// ProductSku represents a SKU in Lazada product
type ProductSku struct {
	SkuID        FlexibleString `json:"SkuId"`
	ShopSku      string         `json:"ShopSku"`
	SellerSku    string         `json:"SellerSku"`
	Price        float64        `json:"price"`
	SpecialPrice float64        `json:"special_price"`
	Quantity     int            `json:"quantity"`
	Available    int            `json:"Available"`
	SaleProp     interface{}    `json:"saleProp"`
	Pilihan      string         `json:"Pilihan"`
	Variation    string         `json:"Variation"`
}

// GetProducts fetches products from Lazada API
func (c *Client) GetProducts(offset, limit int) (*ProductListResponse, error) {
	// Backward-compatible wrapper. Prefer GetProductsWithContext.
	return c.GetProductsWithContext(context.Background(), offset, limit)
}

func shouldRetryProductGet(code, message string) bool {
	// Lazada returns numeric string codes (e.g. "1002") while embedding the
	// prefixed error code in message (e.g. "E1002: ...").
	if code == "1002" {
		return true
	}
	if code == "506" {
		// Lazada frequently returns transient ISP errors as 506.
		// We retry a few times before giving up.
		return true
	}

	msg := strings.ToLower(message)
	if strings.Contains(msg, "e1002") || strings.Contains(msg, "sentinel") {
		return true
	}
	if strings.Contains(msg, "system busy") || strings.Contains(msg, "throttle") {
		return true
	}
	return false
}

// GetProductsWithContext fetches products from Lazada API with retry/backoff.
func (c *Client) GetProductsWithContext(ctx context.Context, offset, limit int) (*ProductListResponse, error) {
	params := map[string]string{
		"filter": "live", // Filter live products only, same as Node.js
		"offset": fmt.Sprintf("%d", offset),
		"limit":  fmt.Sprintf("%d", limit),
	}

	// Retry strategy: exponential backoff with small jitter.
	// Keep this conservative to avoid amplifying traffic during throttling.
	const maxRetries = 5
	baseDelay := 1 * time.Second

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		var raw json.RawMessage
		err := c.RawGet(ctx, "/products/get", params, &raw)
		if err == nil {
			var result ProductListResponse
			if unmarshalErr := json.Unmarshal(raw, &result); unmarshalErr != nil {
				return nil, unmarshalErr
			}

			if result.Code == "0" || result.Code == "" {
				return &result, nil
			}

			// Retryable Lazada-side throttling/system errors.
			if shouldRetryProductGet(result.Code, result.Message) && attempt < maxRetries {
				jitterMs := rand.Intn(250) // 0-249ms
				delay := baseDelay * time.Duration(1<<attempt)
				if delay > 30*time.Second {
					delay = 30 * time.Second
				}
				delay += time.Duration(jitterMs) * time.Millisecond

				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(delay):
					continue
				}
			}

			if result.Message != "" {
				return nil, fmt.Errorf("lazada API error: code=%s, message=%s", result.Code, result.Message)
			}
			return nil, fmt.Errorf("lazada API error: code=%s", result.Code)
		}

		lastErr = err
		if attempt >= maxRetries {
			break
		}

		delay := baseDelay * time.Duration(1<<attempt)
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			continue
		}
	}

	return nil, lastErr
}

// GetProductItem fetches a single product by item ID from Lazada API.
// Uses /products/get with item_id filter.
func (c *Client) GetProductItem(ctx context.Context, itemID int64) (*Product, error) {
	params := map[string]string{
		"filter":  "live",
		"item_id": strconv.FormatInt(itemID, 10),
	}

	var raw json.RawMessage
	if err := c.RawGet(ctx, "/products/get", params, &raw); err != nil {
		return nil, err
	}

	var result ProductListResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "" {
		return nil, fmt.Errorf("lazada GetProductItem error: %s - %s", result.Code, result.Message)
	}
	if len(result.Data.Products) == 0 {
		return nil, fmt.Errorf("lazada product not found: %d", itemID)
	}
	return &result.Data.Products[0], nil
}

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

// ===== Product Operations =====

// CreateProductRequest represents product creation request
type CreateProductRequest struct {
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	Brand           string             `json:"brand"`
	PrimaryCategory int64              `json:"primary_category"`
	Skus            []CreateProductSku `json:"skus"`
	Images          []string           `json:"images"`
}

// CreateProductSku represents SKU in product creation
type CreateProductSku struct {
	SellerSku    string  `json:"seller_sku"`
	Price        float64 `json:"price"`
	Quantity     int     `json:"quantity"`
	SpecialPrice float64 `json:"special_price,omitempty"`
}

// CreateProductResponse represents product creation response
type CreateProductResponse struct {
	BaseResponse
	Data struct {
		ItemID  FlexibleString `json:"item_id"`
		SkuList []struct {
			SkuID     FlexibleString `json:"sku_id"`
			SellerSku string         `json:"seller_sku"`
		} `json:"sku_list"`
		ItemStatus string `json:"item_status,omitempty"`
	} `json:"data"`
}

// CreateProduct creates a new product
func (c *Client) CreateProduct(req CreateProductRequest) (*CreateProductResponse, error) {
	// Build XML payload (Lazada uses XML for product creation)
	params := map[string]string{
		"payload": buildProductPayload(req),
	}

	var result CreateProductResponse
	err := c.doRequest("POST", "/product/create", params, &result)
	return &result, err
}

// CreateProductWithPayload creates a product using raw XML payload
// Per Lazada docs: POST /product/create with XML payload in query string
// This gives full control over the XML structure
func (c *Client) CreateProductWithPayload(xmlPayload string) (*CreateProductResponse, error) {
	params := map[string]string{
		"payload": xmlPayload,
	}

	var result CreateProductResponse
	err := c.doRequest("POST", "/product/create", params, &result)
	return &result, err
}

// UpdateProductRequest represents product update request
type UpdateProductRequest struct {
	ItemID      string             `json:"item_id"`
	Name        string             `json:"name,omitempty"`
	Description string             `json:"description,omitempty"`
	Skus        []UpdateProductSku `json:"skus,omitempty"`
}

// UpdateProductSku represents SKU in product update
type UpdateProductSku struct {
	SkuID        string  `json:"sku_id"`
	SellerSku    string  `json:"seller_sku,omitempty"`
	Price        float64 `json:"price,omitempty"`
	Quantity     int     `json:"quantity,omitempty"`
	SpecialPrice float64 `json:"special_price,omitempty"`
}

// UpdateProduct updates an existing product
func (c *Client) UpdateProduct(req UpdateProductRequest) (*BaseResponse, error) {
	params := map[string]string{
		"payload": buildUpdatePayload(req),
	}

	var result BaseResponse
	err := c.doRequest("POST", "/product/update", params, &result)
	return &result, err
}

// DeleteProduct removes a product
func (c *Client) DeleteProduct(itemID string) (*BaseResponse, error) {
	params := map[string]string{
		"item_id": itemID,
	}

	var result BaseResponse
	err := c.doRequest("POST", "/product/remove", params, &result)
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

func buildProductPayload(req CreateProductRequest) string {
	// Simplified XML builder - in production use encoding/xml
	return fmt.Sprintf(`<Request><Product><PrimaryCategory>%d</PrimaryCategory><Attributes><name>%s</name><description>%s</description><brand>%s</brand></Attributes></Product></Request>`,
		req.PrimaryCategory, req.Name, req.Description, req.Brand)
}

func buildUpdatePayload(req UpdateProductRequest) string {
	return fmt.Sprintf(`<Request><Product><ItemId>%s</ItemId><Attributes><name>%s</name><description>%s</description></Attributes></Product></Request>`,
		req.ItemID, req.Name, req.Description)
}
