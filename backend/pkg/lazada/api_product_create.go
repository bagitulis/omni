package lazada

import "fmt"

// ===== Product CRUD Operations =====

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

func buildProductPayload(req CreateProductRequest) string {
	// Simplified XML builder - in production use encoding/xml
	return fmt.Sprintf(`<Request><Product><PrimaryCategory>%d</PrimaryCategory><Attributes><name>%s</name><description>%s</description><brand>%s</brand></Attributes></Product></Request>`,
		req.PrimaryCategory, req.Name, req.Description, req.Brand)
}

func buildUpdatePayload(req UpdateProductRequest) string {
	return fmt.Sprintf(`<Request><Product><ItemId>%s</ItemId><Attributes><name>%s</name><description>%s</description></Attributes></Product></Request>`,
		req.ItemID, req.Name, req.Description)
}
