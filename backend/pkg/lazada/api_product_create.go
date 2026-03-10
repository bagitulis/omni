package lazada

import (
	"encoding/xml"
	"fmt"
)

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
	params := map[string]string{
		"payload": buildProductPayload(req),
	}

	var result CreateProductResponse
	if err := c.doRequest("POST", "/product/create", params, &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "" {
		return &result, fmt.Errorf("lazada API error (code %s): %s", result.Code, result.Message)
	}
	return &result, nil
}

// CreateProductWithPayload creates a product using raw XML payload
func (c *Client) CreateProductWithPayload(xmlPayload string) (*CreateProductResponse, error) {
	params := map[string]string{
		"payload": xmlPayload,
	}

	var result CreateProductResponse
	if err := c.doRequest("POST", "/product/create", params, &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "" {
		return &result, fmt.Errorf("lazada API error (code %s): %s", result.Code, result.Message)
	}
	return &result, nil
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
	if err := c.doRequest("POST", "/product/update", params, &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "" {
		return &result, fmt.Errorf("lazada API error (code %s): %s", result.Code, result.Message)
	}
	return &result, nil
}

// DeleteProduct removes a product
func (c *Client) DeleteProduct(itemID string) (*BaseResponse, error) {
	params := map[string]string{
		"item_id": itemID,
	}

	var result BaseResponse
	if err := c.doRequest("POST", "/product/remove", params, &result); err != nil {
		return nil, err
	}
	if result.Code != "0" && result.Code != "" {
		return &result, fmt.Errorf("lazada API error (code %s): %s", result.Code, result.Message)
	}
	return &result, nil
}

// ===== XML Payload Builders (encoding/xml — injection-safe) =====

// lazadaProductXML is the top-level XML wrapper for Lazada product API.
type lazadaProductXML struct {
	XMLName xml.Name           `xml:"Request"`
	Product lazadaProductInner `xml:"Product"`
}

type lazadaProductInner struct {
	PrimaryCategory int64               `xml:"PrimaryCategory"`
	Attributes      lazadaAttributesXML `xml:"Attributes"`
	Skus            *lazadaSkusXML      `xml:"Skus,omitempty"`
	Images          *lazadaImagesXML    `xml:"Images,omitempty"`
}

type lazadaAttributesXML struct {
	Name        string `xml:"name"`
	Description string `xml:"description"`
	Brand       string `xml:"brand,omitempty"`
}

type lazadaSkusXML struct {
	Sku []lazadaSkuXML `xml:"Sku"`
}

type lazadaSkuXML struct {
	SellerSku    string  `xml:"SellerSku"`
	Price        float64 `xml:"price"`
	Quantity     int     `xml:"quantity"`
	SpecialPrice float64 `xml:"special_price,omitempty"`
}

type lazadaImagesXML struct {
	Image []lazadaImageXML `xml:"Image"`
}

type lazadaImageXML struct {
	URL string `xml:"Url"`
}

// lazadaUpdateXML is the top-level XML wrapper for update API.
type lazadaUpdateXML struct {
	XMLName xml.Name          `xml:"Request"`
	Product lazadaUpdateInner `xml:"Product"`
}

type lazadaUpdateInner struct {
	ItemID     string              `xml:"ItemId"`
	Attributes lazadaAttributesXML `xml:"Attributes"`
}

func buildProductPayload(req CreateProductRequest) string {
	payload := lazadaProductXML{
		Product: lazadaProductInner{
			PrimaryCategory: req.PrimaryCategory,
			Attributes: lazadaAttributesXML{
				Name:        req.Name,
				Description: req.Description,
				Brand:       req.Brand,
			},
		},
	}

	// Add SKUs
	if len(req.Skus) > 0 {
		skus := &lazadaSkusXML{Sku: make([]lazadaSkuXML, len(req.Skus))}
		for i, s := range req.Skus {
			skus.Sku[i] = lazadaSkuXML{
				SellerSku:    s.SellerSku,
				Price:        s.Price,
				Quantity:     s.Quantity,
				SpecialPrice: s.SpecialPrice,
			}
		}
		payload.Product.Skus = skus
	}

	// Add Images
	if len(req.Images) > 0 {
		imgs := &lazadaImagesXML{Image: make([]lazadaImageXML, len(req.Images))}
		for i, url := range req.Images {
			imgs.Image[i] = lazadaImageXML{URL: url}
		}
		payload.Product.Images = imgs
	}

	data, err := xml.Marshal(payload)
	if err != nil {
		// Fallback: should not happen with well-formed structs
		return fmt.Sprintf(`<Request><Product><PrimaryCategory>%d</PrimaryCategory></Product></Request>`, req.PrimaryCategory)
	}
	return string(data)
}

func buildUpdatePayload(req UpdateProductRequest) string {
	payload := lazadaUpdateXML{
		Product: lazadaUpdateInner{
			ItemID: req.ItemID,
			Attributes: lazadaAttributesXML{
				Name:        req.Name,
				Description: req.Description,
			},
		},
	}

	data, err := xml.Marshal(payload)
	if err != nil {
		return fmt.Sprintf(`<Request><Product><ItemId>%s</ItemId></Product></Request>`, req.ItemID)
	}
	return string(data)
}
