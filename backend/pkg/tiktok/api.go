package tiktok

import "fmt"

// OrderListResponse represents TikTok order list response
type OrderListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		OrderList []struct {
			OrderID     string  `json:"order_id"`
			OrderStatus string  `json:"order_status"`
			TotalAmount float64 `json:"payment_info.total_amount"`
			CreateTime  int64   `json:"create_time"`
			UpdateTime  int64   `json:"update_time"`
		} `json:"order_list"`
		TotalCount int `json:"total_count"`
	} `json:"data"`
}

// GetOrders fetches orders from TikTok API
func (c *Client) GetOrders(status string, pageSize int) (*OrderListResponse, error) {
	params := map[string]string{
		"page_size": fmt.Sprintf("%d", pageSize),
	}
	if status != "" {
		params["order_status"] = status
	}

	var result OrderListResponse
	err := c.doRequest("GET", "/order/202309/orders/search", params, &result)
	return &result, err
}

// ProductListResponse represents TikTok product list response
type ProductListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Products []struct {
			ProductID   string  `json:"id"`
			Name        string  `json:"title"`
			Description string  `json:"description"`
			Price       float64 `json:"sale_price"`
			Status      string  `json:"status"`
		} `json:"products"`
		TotalCount int `json:"total_count"`
	} `json:"data"`
}

// GetProducts fetches products from TikTok API
func (c *Client) GetProducts(pageSize int) (*ProductListResponse, error) {
	params := map[string]string{
		"page_size": fmt.Sprintf("%d", pageSize),
	}

	var result ProductListResponse
	err := c.doRequest("GET", "/product/202309/products/search", params, &result)
	return &result, err
}

// ===== Order Operations =====

// BaseResponse is common response structure
type BaseResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// ShipPackageRequest represents ship package request
type ShipPackageRequest struct {
	OrderID          string `json:"order_id"`
	PackageID        string `json:"package_id"`
	ShippingProvider string `json:"shipping_provider_id"`
	TrackingNumber   string `json:"tracking_number"`
}

// ShipPackageResponse represents ship package response
type ShipPackageResponse struct {
	BaseResponse
	Data struct {
		PackageID string `json:"package_id"`
	} `json:"data"`
}

// ShipPackage marks a package as shipped
func (c *Client) ShipPackage(req ShipPackageRequest) (*ShipPackageResponse, error) {
	params := map[string]string{
		"order_id":             req.OrderID,
		"package_id":           req.PackageID,
		"shipping_provider_id": req.ShippingProvider,
		"tracking_number":      req.TrackingNumber,
	}

	var result ShipPackageResponse
	err := c.doRequest("POST", "/fulfillment/202309/packages/ship", params, &result)
	return &result, err
}

// CancelOrderRequest represents cancel order request
type CancelOrderRequest struct {
	OrderID      string `json:"order_id"`
	CancelReason string `json:"cancel_reason"`
}

// CancelOrder cancels an order
func (c *Client) CancelOrder(req CancelOrderRequest) (*BaseResponse, error) {
	params := map[string]string{
		"order_id":      req.OrderID,
		"cancel_reason": req.CancelReason,
	}

	var result BaseResponse
	err := c.doRequest("POST", "/order/202309/orders/cancel", params, &result)
	return &result, err
}

// ===== Product Operations =====

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

// ===== Inventory Update API =====

// UpdateInventoryRequest represents TikTok inventory update request
// API: POST /product/202309/products/{product_id}/inventory/update
type UpdateInventoryRequest struct {
	ProductID string             `json:"-"` // Used in URL path, not body
	Skus      []InventorySkuInfo `json:"skus"`
}

// InventorySkuInfo represents SKU inventory info
type InventorySkuInfo struct {
	ID        string              `json:"id"`
	Inventory []InventoryQuantity `json:"inventory"`
}

// InventoryQuantity represents inventory quantity (warehouse optional per TikTok API)
type InventoryQuantity struct {
	WarehouseID string `json:"warehouse_id,omitempty"`
	Quantity    int    `json:"quantity"`
}

// UpdateInventoryResponse represents inventory update response
type UpdateInventoryResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Skus []struct {
			ID      string `json:"id"`
			Success bool   `json:"success"`
		} `json:"skus"`
	} `json:"data"`
}

// UpdateInventory updates product inventory
// TikTok v202309 API: POST /product/202309/products/{product_id}/inventory/update
func (c *Client) UpdateInventory(req UpdateInventoryRequest) (*UpdateInventoryResponse, error) {
	params := map[string]string{}

	body := struct {
		Skus []InventorySkuInfo `json:"skus"`
	}{
		Skus: req.Skus,
	}

	var result UpdateInventoryResponse
	endpoint := fmt.Sprintf("/product/202309/products/%s/inventory/update", req.ProductID)
	err := c.doRequestWithBody("POST", endpoint, params, body, &result)

	if err != nil {
		return nil, err
	}

	// TikTok returns code 0 for success
	if result.Code != 0 {
		return &result, fmt.Errorf("tiktok API error: code=%d, message=%s", result.Code, result.Message)
	}

	return &result, nil
}

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

// ===== Product Search v202502 (for sync) =====

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

// ===== Product Detail v202309 (for variant info) =====

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

// =============================================================================
// Order Search API
// =============================================================================

// OrderSearchRequest represents order search request
type OrderSearchRequest struct {
	OrderStatus  string `json:"order_status,omitempty"`
	CreateTimeGe int64  `json:"create_time_ge,omitempty"` // Unix timestamp
	CreateTimeLt int64  `json:"create_time_lt,omitempty"` // Unix timestamp
}

// OrderSearchResponse represents order search response
type OrderSearchResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Orders        []TiktokOrder `json:"orders"`
		NextPageToken string        `json:"next_page_token"`
		TotalCount    int           `json:"total_count"`
	} `json:"data"`
}

// TiktokOrder represents a TikTok order
type TiktokOrder struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	CreateTime   int64  `json:"create_time"`
	UpdateTime   int64  `json:"update_time"`
	BuyerEmail   string `json:"buyer_email,omitempty"`
	BuyerMessage string `json:"buyer_message,omitempty"`
	PaymentInfo  struct {
		Currency            string `json:"currency"`
		OriginalTotalAmount string `json:"original_total_product_price"`
		TotalAmount         string `json:"total_amount"`
		SubTotal            string `json:"sub_total"`
		ShippingFee         string `json:"shipping_fee"`
		PlatformDiscount    string `json:"platform_discount"`
		SellerDiscount      string `json:"seller_discount"`
		ShippingFeeDiscount string `json:"shipping_fee_seller_discount"`
		ShippingFeePlatform string `json:"shipping_fee_platform_discount"`
	} `json:"payment_info"`
	LineItems []TiktokOrderItem `json:"line_items"`
}

// TiktokOrderItem represents an order item
type TiktokOrderItem struct {
	ID               string `json:"id"`
	SkuID            string `json:"sku_id"`
	SkuName          string `json:"sku_name"`
	ProductID        string `json:"product_id"`
	ProductName      string `json:"product_name"`
	SellerSku        string `json:"seller_sku"`
	Quantity         int    `json:"quantity"`
	OriginalPrice    string `json:"original_price"`
	SalePrice        string `json:"sale_price"`
	PlatformDiscount string `json:"platform_discount"`
	SellerDiscount   string `json:"seller_discount"`
}

// SearchOrders searches orders from TikTok API
func (c *Client) SearchOrders(req OrderSearchRequest, pageSize int, pageToken string) (*OrderSearchResponse, error) {
	params := map[string]string{
		"page_size": fmt.Sprintf("%d", pageSize),
	}
	if pageToken != "" {
		params["page_token"] = pageToken
	}

	// Build request body
	body := make(map[string]interface{})
	if req.OrderStatus != "" {
		body["order_status"] = req.OrderStatus
	}
	if req.CreateTimeGe > 0 {
		body["create_time_ge"] = req.CreateTimeGe
	}
	if req.CreateTimeLt > 0 {
		body["create_time_lt"] = req.CreateTimeLt
	}

	var result OrderSearchResponse
	err := c.doRequestWithBody("POST", "/order/202309/orders/search", params, body, &result)
	return &result, err
}

// =============================================================================
// Finance API - Get Transactions by Order
// =============================================================================

// OrderTransactionResponse represents order transaction response
type OrderTransactionResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		OrderID               string                 `json:"order_id"`
		Currency              string                 `json:"currency"`
		SettlementAmount      string                 `json:"settlement_amount"` // Total settlement amount for the order
		SkuTransactions       []SkuTransaction       `json:"sku_transactions"`
		OrderLevelCharges     []OrderLevelCharge     `json:"order_level_charges"`
		StatementTransactions []StatementTransaction `json:"statement_transactions,omitempty"` // v202309
	} `json:"data"`
}

// SkuTransaction represents SKU level transaction
type SkuTransaction struct {
	SkuID                 string `json:"sku_id"`
	ProductName           string `json:"product_name"`
	SkuName               string `json:"sku_name"`
	Quantity              int    `json:"quantity"`
	SkuSubtotalBeforeDisc string `json:"sku_subtotal_before_discount"`
	SkuPlatformDiscount   string `json:"sku_platform_discount"`
	SkuSellerDiscount     string `json:"sku_seller_discount"`
	SkuExtPlatformDisc    string `json:"sku_ext_platform_discount"`
	SkuExtSellerDisc      string `json:"sku_ext_seller_discount"`
	SkuSubtotalAfterDisc  string `json:"sku_subtotal_after_discount"`
	RetailDeliveryFee     string `json:"sku_retail_delivery_fee"`
	SkuEstimatedPkg       string `json:"sku_estimated_package_on_buyer"`
	TransactionFee        string `json:"transaction_fee"`
	ReferralFee           string `json:"referral_fee"`
	AffiliateCommission   string `json:"affiliate_commission"`
	AffiliatePartnerComm  string `json:"affiliate_partner_commission"`
	SkuNetSales           string `json:"sku_net_sales"`
	SkuNetPayout          string `json:"sku_net_payout"`
}

// OrderLevelCharge represents order level charges
type OrderLevelCharge struct {
	ChargeType   string `json:"charge_type"`
	ChargeAmount string `json:"charge_amount"`
}

// StatementTransaction represents statement transaction (v202309)
type StatementTransaction struct {
	StatementID     string `json:"statement_id"`
	StatementTime   int64  `json:"statement_time"`
	TransactionType string `json:"transaction_type"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
}

// GetOrderTransactions fetches transaction details for an order (v202501 API)
func (c *Client) GetOrderTransactions(orderID string) (*OrderTransactionResponse, error) {
	params := map[string]string{}

	var result OrderTransactionResponse
	// Use v202501 API for full SKU transaction details
	err := c.doRequest("GET", "/finance/202501/orders/"+orderID+"/statement_transactions", params, &result)
	return &result, err
}

// GetOrderTransactionsV202309 fetches transactions using older API version
func (c *Client) GetOrderTransactionsV202309(orderID string) (*OrderTransactionResponse, error) {
	params := map[string]string{}

	var result OrderTransactionResponse
	err := c.doRequest("GET", "/finance/202309/orders/"+orderID+"/statement_transactions", params, &result)
	return &result, err
}
