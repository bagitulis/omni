package shopee

import (
	"encoding/json"
	"fmt"
)

// CreateProductRequest represents request to create a product
type CreateProductRequest struct {
	OriginalPrice float64         `json:"original_price"`
	Description   string          `json:"description"`
	Weight        float64         `json:"weight"`
	ItemName      string          `json:"item_name"`
	ItemStatus    string          `json:"item_status"` // NORMAL, UNLIST
	CategoryID    int64           `json:"category_id"`
	Brand         *BrandInfo      `json:"brand,omitempty"`
	Image         ImageInfo       `json:"image"`
	Dimension     *DimensionInfo  `json:"dimension,omitempty"`
	LogisticInfo  []LogisticInfo  `json:"logistic_info"`
	SellerStock   []SellerStock   `json:"seller_stock,omitempty"`
	Attributes    []AttributeInfo `json:"attribute_list,omitempty"`
}

// BrandInfo represents brand information
type BrandInfo struct {
	BrandID       int64  `json:"brand_id"`
	OriginalBrand string `json:"original_brand_name,omitempty"`
}

// ImageInfo represents product images
type ImageInfo struct {
	ImageIDList  []string `json:"image_id_list,omitempty"`
	ImageURLList []string `json:"image_url_list,omitempty"`
}

// DimensionInfo represents product dimensions
type DimensionInfo struct {
	PackageLength int `json:"package_length"`
	PackageWidth  int `json:"package_width"`
	PackageHeight int `json:"package_height"`
}

// LogisticInfo represents logistics configuration
type LogisticInfo struct {
	LogisticID  int64   `json:"logistic_id"`
	Enabled     bool    `json:"enabled"`
	ShippingFee float64 `json:"shipping_fee,omitempty"`
	SizeID      int64   `json:"size_id,omitempty"`
}

// SellerStock represents seller stock info for Shopee API
type SellerStock struct {
	LocationID string `json:"location_id,omitempty"` // Optional - omit for default warehouse
	Stock      int    `json:"stock"`
}

// AttributeInfo represents product attribute
type AttributeInfo struct {
	AttributeID        int64            `json:"attribute_id"`
	AttributeValueList []AttributeValue `json:"attribute_value_list"`
}

// AttributeValue represents attribute value
type AttributeValue struct {
	ValueID       int64  `json:"value_id,omitempty"`
	OriginalValue string `json:"original_value_name,omitempty"`
	ValueUnit     string `json:"value_unit,omitempty"`
}

// CreateProductResponse represents create product API response
type CreateProductResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ItemID int64 `json:"item_id"`
	} `json:"response"`
}

// UpdateProductRequest represents request to update a product
type UpdateProductRequest struct {
	ItemID        int64           `json:"item_id"`
	ItemName      string          `json:"item_name,omitempty"`
	Description   string          `json:"description,omitempty"`
	OriginalPrice float64         `json:"original_price,omitempty"`
	Weight        float64         `json:"weight,omitempty"`
	ItemStatus    string          `json:"item_status,omitempty"`
	Dimension     *DimensionInfo  `json:"dimension,omitempty"`
	Image         *ImageInfo      `json:"image,omitempty"`
	Attributes    []AttributeInfo `json:"attribute_list,omitempty"`
}

// UpdateProductResponse represents update product API response
type UpdateProductResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ItemID int64 `json:"item_id"`
	} `json:"response"`
}

// DeleteProductRequest represents request to delete a product
type DeleteProductRequest struct {
	ItemID int64 `json:"item_id"`
}

// DeleteProductResponse represents delete product API response
type DeleteProductResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ItemID int64 `json:"item_id"`
	} `json:"response"`
}

// UpdateStockRequest represents request to update stock
type UpdateStockRequest struct {
	ItemID    int64           `json:"item_id"`
	StockList []StockListItem `json:"stock_list"`
}

// StockListItem represents stock update for a model/variant
type StockListItem struct {
	ModelID     int64         `json:"model_id"`
	SellerStock []SellerStock `json:"seller_stock"`
}

// UpdateStockResponse represents update stock API response
type UpdateStockResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		SuccessList []StockUpdateItem `json:"success_list"`
		FailureList []StockUpdateItem `json:"failure_list"`
	} `json:"response"`
}

// StockUpdateItem represents stock update result item
type StockUpdateItem struct {
	ItemID     int64  `json:"item_id"`
	ModelID    int64  `json:"model_id,omitempty"`
	LocationID string `json:"location_id,omitempty"`
}

// UpdatePriceRequest represents request to update price
type UpdatePriceRequest struct {
	ItemID    int64       `json:"item_id"`
	PriceList []PriceInfo `json:"price_list"`
}

// PriceInfo represents price information
type PriceInfo struct {
	ModelID       int64   `json:"model_id,omitempty"`
	OriginalPrice float64 `json:"original_price"`
}

// UpdatePriceResponse represents update price API response
type UpdatePriceResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		SuccessList []PriceUpdateItem `json:"success_list"`
		FailureList []PriceUpdateItem `json:"failure_list"`
	} `json:"response"`
}

// PriceUpdateItem represents price update result item
type PriceUpdateItem struct {
	ItemID  int64 `json:"item_id"`
	ModelID int64 `json:"model_id,omitempty"`
}

// CreateProduct creates a product via Shopee API
func (c *Client) CreateProduct(req CreateProductRequest) (*CreateProductResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var result CreateProductResponse
	if err := c.doPostRequest("/api/v2/product/add_item", nil, body, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// UpdateProduct updates a product via Shopee API
func (c *Client) UpdateProduct(req UpdateProductRequest) (*UpdateProductResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var result UpdateProductResponse
	if err := c.doPostRequest("/api/v2/product/update_item", nil, body, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// DeleteProduct deletes a product via Shopee API
func (c *Client) DeleteProduct(itemID int64) (*DeleteProductResponse, error) {
	req := DeleteProductRequest{ItemID: itemID}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var result DeleteProductResponse
	if err := c.doPostRequest("/api/v2/product/delete_item", nil, body, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// UpdateStock updates product stock via Shopee API
func (c *Client) UpdateStock(req UpdateStockRequest) (*UpdateStockResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var result UpdateStockResponse
	if err := c.doPostRequest("/api/v2/product/update_stock", nil, body, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// UpdatePrice updates product price via Shopee API
func (c *Client) UpdatePrice(req UpdatePriceRequest) (*UpdatePriceResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var result UpdatePriceResponse
	if err := c.doPostRequest("/api/v2/product/update_price", nil, body, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}
