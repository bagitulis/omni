package shopee

import (
	"fmt"
	"strconv"
)

// ProductListResponse represents Shopee product list API response
type ProductListResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		Item []struct {
			ItemID int64 `json:"item_id"`
		} `json:"item"`
		TotalCount  int64 `json:"total_count"`
		HasNextPage bool  `json:"has_next_page"`
		NextOffset  int   `json:"next_offset"`
	} `json:"response"`
}

// ProductDetailResponse represents Shopee product detail API response
type ProductDetailResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ItemList []ProductDetail `json:"item_list"`
	} `json:"response"`
}

// ProductDetail represents a Shopee product
type ProductDetail struct {
	ItemID        int64    `json:"item_id"`
	ItemName      string   `json:"item_name"`
	Description   string   `json:"description"`
	CategoryID    int64    `json:"category_id"`
	OriginalPrice float64  `json:"original_price"`
	CurrentPrice  float64  `json:"current_price"`
	Stock         int      `json:"stock"`
	ItemStatus    string   `json:"item_status"`
	Images        []string `json:"images"`
}

// GetProductList fetches product list from Shopee API
func (c *Client) GetProductList(offset, limit int) (*ProductListResponse, error) {
	params := map[string]string{
		"offset":      strconv.Itoa(offset),
		"page_size":   strconv.Itoa(limit),
		"item_status": "NORMAL",
	}

	var result ProductListResponse
	err := c.doRequest("GET", "/api/v2/product/get_item_list", params, &result)
	if err != nil {
		return nil, err
	}

	// Check for API error in response
	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetProductDetail fetches product details from Shopee API
// API: /api/v2/product/get_item_base_info
// Note: need_tax_info and need_complaint_policy are optional
func (c *Client) GetProductDetail(itemIDs []int64) (*ProductDetailResponse, error) {
	// Join item IDs with comma
	ids := ""
	for i, id := range itemIDs {
		if i > 0 {
			ids += ","
		}
		ids += strconv.FormatInt(id, 10)
	}

	params := map[string]string{
		"item_id_list": ids,
	}

	var result ProductDetailResponse
	err := c.doRequest("GET", "/api/v2/product/get_item_base_info", params, &result)
	if err != nil {
		return nil, err
	}

	// Check for API error in response
	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetProductDetailWithImages fetches product details including images from Shopee API
// Uses /api/v2/product/get_item_base_info with need_media field to get image data
// Note: "image" is included in default response, but may be a struct instead of array
func (c *Client) GetProductDetailWithImages(itemIDs []int64) (*ProductDetailWithImagesResponse, error) {
	// Join item IDs with comma
	ids := ""
	for i, id := range itemIDs {
		if i > 0 {
			ids += ","
		}
		ids += strconv.FormatInt(id, 10)
	}

	params := map[string]string{
		"item_id_list": ids,
	}

	var result ProductDetailWithImagesResponse
	err := c.doRequest("GET", "/api/v2/product/get_item_base_info", params, &result)
	if err != nil {
		return nil, err
	}

	// Check for API error in response
	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// ProductDetailWithImagesResponse represents Shopee product with image structure
type ProductDetailWithImagesResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ItemList []ProductDetailWithImages `json:"item_list"`
	} `json:"response"`
}

// ProductDetailWithImages contains product with image structure (not array)
// Also includes price_info for products without variants
type ProductDetailWithImages struct {
	ItemID      int64  `json:"item_id"`
	ItemName    string `json:"item_name"`
	Description string `json:"description"`
	CategoryID  int64  `json:"category_id"`
	Image       struct {
		ImageURLList []string `json:"image_url_list"`
		ImageIDList  []string `json:"image_id_list"`
	} `json:"image"`
	// Price info for products without variants (single-SKU products)
	PriceInfo []struct {
		OriginalPrice float64 `json:"original_price"`
		CurrentPrice  float64 `json:"current_price"`
	} `json:"price_info"`
	// Stock info for products without variants
	StockInfoV2 struct {
		SummaryInfo struct {
			TotalAvailableStock int `json:"total_available_stock"`
		} `json:"summary_info"`
	} `json:"stock_info_v2"`
}

// ModelListResponse represents Shopee model list API response
type ModelListResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		TierVariation []TierVariation `json:"tier_variation"`
		Model         []Model         `json:"model"`
	} `json:"response"`
}

// TierVariation represents variant tier in Shopee
type TierVariation struct {
	Name       string   `json:"name"`
	OptionList []Option `json:"option_list"`
}

// Option represents variant option
type Option struct {
	Option string `json:"option"`
}

// Model represents a product SKU/model in Shopee
type Model struct {
	ModelID   int64  `json:"model_id"`
	ModelSKU  string `json:"model_sku"`
	TierIndex []int  `json:"tier_index"`
	PriceInfo []struct {
		OriginalPrice float64 `json:"original_price"`
		CurrentPrice  float64 `json:"current_price"`
	} `json:"price_info"`
	StockInfoV2 struct {
		SummaryInfo struct {
			TotalAvailableStock int `json:"total_available_stock"`
		} `json:"summary_info"`
	} `json:"stock_info_v2"`
}

// GetModelList fetches model list for an item from Shopee API
func (c *Client) GetModelList(itemID int64) (*ModelListResponse, error) {
	params := map[string]string{
		"item_id": strconv.FormatInt(itemID, 10),
	}

	var result ModelListResponse
	err := c.doRequest("GET", "/api/v2/product/get_model_list", params, &result)
	if err != nil {
		return nil, err
	}

	// Check for API error in response
	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}
