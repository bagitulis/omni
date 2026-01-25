package shopee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
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
// Shopee API: POST /api/v2/product/update_stock
// Format: { item_id, stock_list: [{ model_id, seller_stock: [{ stock }] }] }
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

// UploadImageRequest represents image upload request
type UploadImageRequest struct {
	Image []byte `json:"-"` // Raw image bytes
}

// UploadImageResponse represents image upload response
type UploadImageResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ImageInfo struct {
			ImageID      string   `json:"image_id"`
			ImageURLList []string `json:"image_url_list,omitempty"`
		} `json:"image_info"`
	} `json:"response"`
}

// UploadImage uploads image to Shopee using multipart/form-data
func (c *Client) UploadImage(imageBytes []byte) (*UploadImageResponse, error) {
	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add image field
	part, err := writer.CreateFormFile("image", "product_image.jpg")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(imageBytes); err != nil {
		return nil, fmt.Errorf("write image data: %w", err)
	}
	writer.Close()

	// Build signed URL
	reqURL := c.buildURL("/api/v2/media_space/upload_image", nil)

	// Create request with multipart body
	req, err := http.NewRequest("POST", reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read raw response for debugging
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Parse response using flexible structure
	var rawResp map[string]interface{}
	if err := json.Unmarshal(respBody, &rawResp); err != nil {
		return nil, fmt.Errorf("decode response: %w (body: %s)", err, string(respBody[:min(200, len(respBody))]))
	}

	// Build result from raw response
	result := &UploadImageResponse{}
	if errStr, ok := rawResp["error"].(string); ok {
		result.Error = errStr
	}
	if msgStr, ok := rawResp["message"].(string); ok {
		result.Message = msgStr
	}

	// Extract image_id from response.image_info
	if respData, ok := rawResp["response"].(map[string]interface{}); ok {
		if imageInfo, ok := respData["image_info"].(map[string]interface{}); ok {
			if imageID, ok := imageInfo["image_id"].(string); ok {
				result.Response.ImageInfo.ImageID = imageID
			}
		}
	}

	if result.Error != "" {
		return result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return result, nil
}

// UploadImageFromURL downloads image from URL and uploads to Shopee CDN
func (c *Client) UploadImageFromURL(imageURL string) (string, error) {
	// Download image
	resp, err := c.httpClient.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("download image failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download image failed with status %d", resp.StatusCode)
	}

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read image bytes failed: %w", err)
	}

	// Upload to Shopee
	uploadResp, err := c.UploadImage(imageBytes)
	if err != nil {
		return "", err
	}

	return uploadResp.Response.ImageInfo.ImageID, nil
}

// GetWalletTransactionRequest represents wallet transaction request
type GetWalletTransactionRequest struct {
	TransactionTypes   []string `json:"-"` // Not used in query
	TransactionTabType string   `json:"-"` // Used in query
	MoneyFlow          string   `json:"-"` // MONEY_IN or MONEY_OUT
	PageNo             int      `json:"-"`
	PageSize           int      `json:"-"`
	StartDate          int64    `json:"-"`
	EndDate            int64    `json:"-"`
}

// WalletTransactionResponse represents wallet transaction response
type WalletTransactionResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		TransactionList []WalletTransaction `json:"transaction_list"`
		More            bool                `json:"more"`
	} `json:"response"`
}

// WalletTransaction represents a wallet transaction
type WalletTransaction struct {
	TransactionID   int64   `json:"transaction_id"`
	Status          string  `json:"status"`
	TransactionType string  `json:"transaction_type"`
	Amount          float64 `json:"amount"`
	CreateTime      int64   `json:"create_time"`
	OrderSN         string  `json:"order_sn,omitempty"`
	RefundSN        string  `json:"refund_sn,omitempty"`
	Description     string  `json:"description,omitempty"`
	BuyerUsername   string  `json:"buyer_user_name,omitempty"`
}

// GetWalletTransactions gets wallet transactions using GET request with query params
func (c *Client) GetWalletTransactions(req GetWalletTransactionRequest) (*WalletTransactionResponse, error) {
	// Build query params
	params := map[string]string{
		"page_no":   strconv.Itoa(req.PageNo),
		"page_size": strconv.Itoa(req.PageSize),
	}

	if req.StartDate > 0 {
		params["create_time_from"] = strconv.FormatInt(req.StartDate, 10)
	}
	if req.EndDate > 0 {
		params["create_time_to"] = strconv.FormatInt(req.EndDate, 10)
	}
	if req.TransactionTabType != "" {
		params["transaction_tab_type"] = req.TransactionTabType
	}
	if req.MoneyFlow != "" {
		params["money_flow"] = req.MoneyFlow
	}

	reqURL := c.buildURL("/api/v2/payment/get_wallet_transaction_list", params)
	httpReq, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	var result WalletTransactionResponse
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func createPostRequest(url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// =============================================================================
// Category Recommendation API
// =============================================================================

// CategoryRecommendResponse represents category recommendation response
type CategoryRecommendResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Warning  string `json:"warning"`
	Response struct {
		CategoryID []int64 `json:"category_id"`
	} `json:"response"`
}

// GetCategoryRecommend gets category recommendation by product name
// API: GET /api/v2/product/category_recommend
func (c *Client) GetCategoryRecommend(itemName string, imageID string) (*CategoryRecommendResponse, error) {
	params := map[string]string{
		"item_name": itemName,
	}
	if imageID != "" {
		params["product_cover_image"] = imageID
	}

	var result CategoryRecommendResponse
	if err := c.doRequest("GET", "/api/v2/product/category_recommend", params, &result); err != nil {
		return nil, fmt.Errorf("category recommend request failed: %w", err)
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// =============================================================================
// Recommend Attribute API
// =============================================================================

// RecommendAttributeResponse represents recommend attribute response
type RecommendAttributeResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Warning  string `json:"warning"`
	Response struct {
		AttributeList []RecommendedAttribute `json:"attribute_list"`
	} `json:"response"`
}

// RecommendedAttribute represents a recommended attribute
type RecommendedAttribute struct {
	AttributeID      int64                  `json:"attribute_id"`
	OriginalAttrName string                 `json:"original_attribute_name"`
	DisplayAttrName  string                 `json:"display_attribute_name"`
	IsMandatory      bool                   `json:"is_mandatory"`
	InputType        string                 `json:"input_type"`
	FormatType       string                 `json:"format_type"`
	AttributeUnit    []string               `json:"attribute_unit"`
	AttributeValues  []RecommendedAttrValue `json:"attribute_value_list"`
}

// RecommendedAttrValue represents attribute value
type RecommendedAttrValue struct {
	ValueID         int64  `json:"value_id"`
	OriginalValName string `json:"original_value_name"`
	DisplayValName  string `json:"display_value_name"`
	ValueUnit       string `json:"value_unit,omitempty"`
}

// GetRecommendAttribute gets recommended attributes for a category
// API: GET /api/v2/product/get_recommend_attribute
func (c *Client) GetRecommendAttribute(categoryID int64, itemName string) (*RecommendAttributeResponse, error) {
	params := map[string]string{
		"category_id": fmt.Sprintf("%d", categoryID),
		"item_name":   itemName,
	}

	var result RecommendAttributeResponse
	if err := c.doRequest("GET", "/api/v2/product/get_recommend_attribute", params, &result); err != nil {
		return nil, fmt.Errorf("get recommend attribute request failed: %w", err)
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// =============================================================================
// Logistics API
// =============================================================================

// GetLogisticChannelsResponse represents logistics channels response
type GetLogisticChannelsResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		LogisticsList []LogisticChannel `json:"logistics_channel_list"`
	} `json:"response"`
}

// LogisticChannel represents a logistics channel
type LogisticChannel struct {
	LogisticID   int64  `json:"logistics_channel_id"`
	LogisticName string `json:"logistics_channel_name"`
	Enabled      bool   `json:"enabled"`
	Preferred    bool   `json:"preferred"`
}

// GetLogisticChannels gets available logistics channels for shop
// API: GET /api/v2/logistics/get_channel_list
func (c *Client) GetLogisticChannels() (*GetLogisticChannelsResponse, error) {
	var result GetLogisticChannelsResponse
	if err := c.doRequest("GET", "/api/v2/logistics/get_channel_list", nil, &result); err != nil {
		return nil, fmt.Errorf("get logistics channels failed: %w", err)
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}
