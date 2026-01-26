// Package tiktok provides Product API types and methods for TikTok Shop
package tiktok

import "fmt"

// =============================================================================
// Product Create API
// =============================================================================

// CreateProductRequest represents TikTok product creation request
type CreateProductRequest struct {
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	CategoryID        string             `json:"category_id"`
	CategoryVersion   string             `json:"category_version,omitempty"` // "v2" for SEA shops
	MainImages        []ImageInfo        `json:"main_images"`
	Skus              []CreateProductSku `json:"skus"`
	PackageWeight     PackageWeight      `json:"package_weight"`
	ProductAttributes []ProductAttribute `json:"product_attributes,omitempty"`
	SaveMode          string             `json:"save_mode,omitempty"` // AS_DRAFT or LISTING
}

// ImageInfo represents image information
type ImageInfo struct {
	URI string `json:"uri"`
}

// ProductAttribute represents product attribute
type ProductAttribute struct {
	ID     string           `json:"id"`
	Values []AttributeValue `json:"values"`
}

// AttributeValue represents attribute value
// Use either ID (for built-in values) or Name (for custom values)
type AttributeValue struct {
	ID   string `json:"id,omitempty"`   // Built-in value ID
	Name string `json:"name,omitempty"` // Custom value name
}

// CreateProductSku represents SKU in product creation
type CreateProductSku struct {
	SellerSku       string         `json:"seller_sku"`
	Price           *SkuPrice      `json:"price,omitempty"`
	OriginalPrice   string         `json:"original_price,omitempty"` // Deprecated but kept for compatibility
	Inventory       []SkuInventory `json:"inventory,omitempty"`      // New inventory format
	StockInfos      []StockInfo    `json:"stock_infos,omitempty"`    // Legacy format
	SalesAttributes []SalesAttr    `json:"sales_attributes,omitempty"`
}

// SkuPrice represents SKU price information
type SkuPrice struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// SkuInventory represents SKU inventory with warehouse
type SkuInventory struct {
	WarehouseID string `json:"warehouse_id"`
	Quantity    int    `json:"quantity"`
}

// StockInfo represents stock information (legacy format)
type StockInfo struct {
	WarehouseID    string `json:"warehouse_id,omitempty"`
	AvailableStock int    `json:"available_stock"`
}

// SalesAttr represents sales attribute for SKU
// Based on TikTok SDK: CreateProductRequestBodySkusSalesAttributes
// For built-in attributes: use ID and ValueID
// For custom attributes: use Name and ValueName
type SalesAttr struct {
	ID        string `json:"id,omitempty"`         // Built-in attribute ID from Get Attributes API
	Name      string `json:"name,omitempty"`       // Custom attribute name (max 20 chars)
	ValueID   string `json:"value_id,omitempty"`   // Built-in value ID
	ValueName string `json:"value_name,omitempty"` // Custom value name (max 50 chars)
}

// PackageWeight represents package weight
type PackageWeight struct {
	Value string `json:"value"`
	Unit  string `json:"unit"`
}

// CreateProductResponse represents product creation response
type CreateProductResponse struct {
	BaseResponse
	Data struct {
		ProductID string `json:"product_id"`
		Skus      []struct {
			ID        string `json:"id"`
			SellerSku string `json:"seller_sku"`
		} `json:"skus"`
	} `json:"data"`
}

// CreateProduct creates a new product
func (c *Client) CreateProduct(req CreateProductRequest) (*CreateProductResponse, error) {
	// TikTok uses JSON body for product creation
	params := map[string]string{}

	var result CreateProductResponse
	err := c.doRequestWithBody("POST", "/product/202309/products", params, req, &result)
	return &result, err
}

// =============================================================================
// Product Update API
// =============================================================================

// UpdateProductRequest represents product update request
type UpdateProductRequest struct {
	ProductID   string             `json:"product_id"`
	Title       string             `json:"title,omitempty"`
	Description string             `json:"description,omitempty"`
	Skus        []UpdateProductSku `json:"skus,omitempty"`
}

// UpdateProductSku represents SKU in product update
type UpdateProductSku struct {
	ID            string      `json:"id"`
	SellerSku     string      `json:"seller_sku,omitempty"`
	OriginalPrice string      `json:"original_price,omitempty"`
	StockInfos    []StockInfo `json:"stock_infos,omitempty"`
}

// UpdateProduct updates an existing product
func (c *Client) UpdateProduct(req UpdateProductRequest) (*BaseResponse, error) {
	params := map[string]string{}

	var result BaseResponse
	err := c.doRequestWithBody("PUT", "/product/202309/products/"+req.ProductID, params, req, &result)
	return &result, err
}

// DeleteProduct deactivates a product
func (c *Client) DeleteProduct(productID string) (*BaseResponse, error) {
	params := map[string]string{}
	body := map[string][]string{
		"product_ids": {productID},
	}

	var result BaseResponse
	err := c.doRequestWithBody("POST", "/product/202309/products/deactivate", params, body, &result)
	return &result, err
}

// =============================================================================
// Product Search v202502 (for sync)
// =============================================================================

// SearchProductsRequest represents search products request body
type SearchProductsRequest struct {
	Status string `json:"status,omitempty"` // ACTIVATE, DEACTIVATED, etc
}

// ProductSearchSku represents SKU in search response
type ProductSearchSku struct {
	ID        string `json:"id"`
	SellerSku string `json:"seller_sku"`
	Price     struct {
		Currency          string `json:"currency"`
		TaxExclusivePrice string `json:"tax_exclusive_price"`
		SalePrice         string `json:"sale_price"`
		OriginalPrice     string `json:"original_price"`
	} `json:"price"`
	Inventory []struct {
		WarehouseID string `json:"warehouse_id"`
		Quantity    int    `json:"quantity"`
	} `json:"inventory"`
}

// ProductSearchItem represents product in search response
type ProductSearchItem struct {
	ID          string             `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Status      string             `json:"status"`
	Skus        []ProductSearchSku `json:"skus"`
}

// SearchProductsResponse represents v202502 product search response
type SearchProductsResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Products      []ProductSearchItem `json:"products"`
		TotalCount    int                 `json:"total_count"`
		NextPageToken string              `json:"next_page_token"`
	} `json:"data"`
}

// SearchProductsV202502 searches products using v202502 API (POST with body)
func (c *Client) SearchProductsV202502(status string, pageSize int, pageToken string) (*SearchProductsResponse, error) {
	params := map[string]string{
		"page_size": fmt.Sprintf("%d", pageSize),
	}
	if pageToken != "" {
		params["page_token"] = pageToken
	}

	body := SearchProductsRequest{}
	if status != "" {
		body.Status = status
	}

	var result SearchProductsResponse
	err := c.doRequestWithBody("POST", "/product/202502/products/search", params, body, &result)
	return &result, err
}

// =============================================================================
// Product Detail v202309 (for variant info)
// =============================================================================

// ProductDetailSku represents SKU in detail response
type ProductDetailSku struct {
	ID              string             `json:"id"`
	SellerSku       string             `json:"seller_sku"`
	SalesAttributes []ProductSalesAttr `json:"sales_attributes"`
	Price           struct {
		Currency          string `json:"currency"`
		TaxExclusivePrice string `json:"tax_exclusive_price"`
		SalePrice         string `json:"sale_price"`
		OriginalPrice     string `json:"original_price"`
	} `json:"price"`
	Inventory []struct {
		WarehouseID string `json:"warehouse_id"`
		Quantity    int    `json:"quantity"`
	} `json:"inventory"`
}

// ProductSalesAttr represents sales attribute
type ProductSalesAttr struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ValueID   string `json:"value_id"`
	ValueName string `json:"value_name"`
}

// ProductImage represents product main image
type ProductImage struct {
	URI    string   `json:"uri"`
	URLs   []string `json:"urls"`
	Width  int      `json:"width"`
	Height int      `json:"height"`
}

// ProductDetailResponse represents product detail response
type ProductDetailResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ID          string             `json:"id"`
		Title       string             `json:"title"`
		Description string             `json:"description"`
		Status      string             `json:"status"`
		MainImages  []ProductImage     `json:"main_images"`
		Skus        []ProductDetailSku `json:"skus"`
	} `json:"data"`
}

// GetProductDetail fetches product detail with sales_attributes
func (c *Client) GetProductDetail(productID string) (*ProductDetailResponse, error) {
	params := map[string]string{}

	var result ProductDetailResponse
	err := c.doRequest("GET", "/product/202309/products/"+productID, params, &result)
	return &result, err
}
