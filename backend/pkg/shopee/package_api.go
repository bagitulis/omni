package shopee

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SearchPackageListRequest represents request body for search_package_list
type SearchPackageListRequest struct {
	Filter     PackageFilter     `json:"filter"`
	Pagination PackagePagination `json:"pagination"`
	Sort       PackageSort       `json:"sort"`
}

// PackageFilter for search_package_list
type PackageFilter struct {
	PackageStatus   int `json:"package_status"`   // 0=All, 1=Pending, 2=ToProcess, 3=Processed
	FulfillmentType int `json:"fulfillment_type"` // 2 = Seller fulfilled
}

// PackagePagination for search_package_list
type PackagePagination struct {
	PageSize int    `json:"page_size"`
	Cursor   string `json:"cursor,omitempty"`
}

// PackageSort for search_package_list
type PackageSort struct {
	SortType  int  `json:"sort_type"` // 1 = ShipByDate
	Ascending bool `json:"ascending"`
}

// SearchPackageListResponse represents response from search_package_list
type SearchPackageListResponse struct {
	Response struct {
		PackagesList []PackageBasic `json:"packages_list"`
		Pagination   struct {
			NextCursor string `json:"next_cursor"`
			More       bool   `json:"more"`
			TotalCount int    `json:"total_count"`
		} `json:"pagination"`
	} `json:"response"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// PackageBasic represents basic package info
type PackageBasic struct {
	OrderSN            string `json:"order_sn"`
	PackageNumber      string `json:"package_number"`
	LogisticsChannelID int64  `json:"logistics_channel_id"`
	IsShipmentArranged bool   `json:"is_shipment_arranged"`
}

// GetPackageDetailResponse represents response from get_package_detail
type GetPackageDetailResponse struct {
	Response struct {
		PackageList []PackageDetail `json:"package_list"`
	} `json:"response"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// PackageDetail represents detailed package info
type PackageDetail struct {
	OrderSN         string        `json:"order_sn"`
	PackageNumber   string        `json:"package_number"`
	ShippingCarrier string        `json:"shipping_carrier"`
	TrackingNumber  string        `json:"tracking_number"`
	ItemList        []PackageItem `json:"item_list"`
}

// PackageItem represents an item in a package
type PackageItem struct {
	ItemID        int64  `json:"item_id"`
	ModelID       int64  `json:"model_id"`
	ItemSKU       string `json:"item_sku"`
	ModelSKU      string `json:"model_sku"`
	ModelQuantity int    `json:"model_quantity"`
}

// SearchPackageList searches packages with filters (for processed orders)
// POST /api/v2/order/search_package_list
func (c *Client) SearchPackageList(packageStatus int, cursor string, pageSize int) (*SearchPackageListResponse, error) {
	path := "/api/v2/order/search_package_list"

	reqBody := SearchPackageListRequest{
		Filter: PackageFilter{
			PackageStatus:   packageStatus,
			FulfillmentType: 2, // Seller fulfilled
		},
		Pagination: PackagePagination{
			PageSize: pageSize,
			Cursor:   cursor,
		},
		Sort: PackageSort{
			SortType:  1, // ShipByDate
			Ascending: true,
		},
	}

	// Marshal request body to JSON
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	var result SearchPackageListResponse
	// Use existing doPostRequest signature: (path, params, body, result)
	err = c.doPostRequest(path, nil, bodyBytes, &result)
	if err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetPackageDetail gets package details including tracking and items
// GET /api/v2/order/get_package_detail
func (c *Client) GetPackageDetail(packageNumbers []string) (*GetPackageDetailResponse, error) {
	if len(packageNumbers) == 0 {
		return &GetPackageDetailResponse{}, nil
	}

	// Limit to 50 packages per request
	if len(packageNumbers) > 50 {
		packageNumbers = packageNumbers[:50]
	}

	path := "/api/v2/order/get_package_detail"
	params := map[string]string{
		"package_number_list": strings.Join(packageNumbers, ","),
	}

	var result GetPackageDetailResponse
	err := c.doRequest("GET", path, params, &result)
	if err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}
