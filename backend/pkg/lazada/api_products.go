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
// Strategy: try /product/item/get first, fallback to /products/get if that fails.
// /product/item/get is the dedicated single-product endpoint but sometimes returns
// E506 for certain products. /products/get is proven reliable in full-sync flows.
func (c *Client) GetProductItem(ctx context.Context, itemID int64) (*Product, error) {
	itemIDStr := strconv.FormatInt(itemID, 10)

	// Attempt 1: /product/item/get (dedicated single-product endpoint)
	product, err := c.tryGetProductItem(ctx, itemIDStr)
	if err == nil {
		return product, nil
	}

	// Attempt 2: /products/get with item_id filter (proven working endpoint)
	product, fallbackErr := c.tryGetProductViaList(ctx, itemIDStr)
	if fallbackErr == nil {
		return product, nil
	}

	return nil, fmt.Errorf(
		"lazada GetProductItem failed: primary (%v), fallback (%v)",
		err, fallbackErr,
	)
}

// tryGetProductItem uses /product/item/get with retry.
func (c *Client) tryGetProductItem(ctx context.Context, itemID string) (*Product, error) {
	params := map[string]string{"item_id": itemID}

	const maxRetries = 2
	baseDelay := 1 * time.Second
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			if delay > 5*time.Second {
				delay = 5 * time.Second
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		var raw json.RawMessage
		if err := c.RawGet(ctx, "/product/item/get", params, &raw); err != nil {
			lastErr = err
			continue
		}

		product, err := c.parseProductResponse(raw)
		if err != nil {
			if shouldRetryProductGet(err.Error(), "") && attempt < maxRetries {
				lastErr = err
				continue
			}
			return nil, err
		}
		return product, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("lazada /product/item/get failed for item_id=%s", itemID)
}

// tryGetProductViaList uses /products/get (the list endpoint) to find a product.
// This endpoint is proven reliable in the full-sync flow.
func (c *Client) tryGetProductViaList(ctx context.Context, itemID string) (*Product, error) {
	// Use /products/get with a small result window — Lazada will return
	// all products but we filter by item_id client-side.
	// We request a larger batch since the API doesn't support filtering by single item_id.
	params := map[string]string{
		"filter": "all",
		"offset": "0",
		"limit":  "50",
	}

	var raw json.RawMessage
	if err := c.RawGet(ctx, "/products/get", params, &raw); err != nil {
		return nil, fmt.Errorf("lazada /products/get fallback error: %w", err)
	}

	var result ProductListResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("lazada /products/get parse error: %w", err)
	}
	if result.Code != "0" && result.Code != "" {
		return nil, fmt.Errorf("lazada /products/get error: %s - %s", result.Code, result.Message)
	}

	// Find product by item_id in list
	for i := range result.Data.Products {
		if string(result.Data.Products[i].ItemID) == itemID {
			return &result.Data.Products[i], nil
		}
	}

	return nil, fmt.Errorf("lazada: product item_id=%s not found in /products/get response (%d products)",
		itemID, len(result.Data.Products))
}

// parseProductResponse handles both single and list response formats from Lazada.
func (c *Client) parseProductResponse(raw json.RawMessage) (*Product, error) {
	// Try list format: {"code":"0","data":{"products":[...]}}
	var listResult ProductListResponse
	if err := json.Unmarshal(raw, &listResult); err == nil {
		if listResult.Code == "0" || listResult.Code == "" {
			if len(listResult.Data.Products) > 0 {
				return &listResult.Data.Products[0], nil
			}
		}
		if listResult.Code != "0" && listResult.Code != "" {
			return nil, fmt.Errorf("%s - %s", listResult.Code, listResult.Message)
		}
	}

	// Try single format: {"code":"0","data":{...product fields...}}
	var singleResult struct {
		Code    string  `json:"code"`
		Message string  `json:"message,omitempty"`
		Data    Product `json:"data"`
	}
	if err := json.Unmarshal(raw, &singleResult); err != nil {
		return nil, fmt.Errorf("parse error: %w (raw: %.200s)", err, string(raw))
	}
	if singleResult.Code != "0" && singleResult.Code != "" {
		return nil, fmt.Errorf("%s - %s", singleResult.Code, singleResult.Message)
	}

	return &singleResult.Data, nil
}


