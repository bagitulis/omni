package shopee

import (
	"context"
	"encoding/json"
	"fmt"
)

// WholesaleTier represents a wholesale tier for Shopee API
type WholesaleTier struct {
	MinCount  int     `json:"min_count"`
	MaxCount  int     `json:"max_count"`
	UnitPrice float64 `json:"unit_price"`
}

// UpdateItemWholesaleRequest represents wholesale update request
type UpdateItemWholesaleRequest struct {
	ItemID    int64           `json:"item_id"`
	Wholesale []WholesaleTier `json:"wholesale"`
}

// UpdateItemWholesaleResponse represents wholesale update response
type UpdateItemWholesaleResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ItemID int64 `json:"item_id"`
	} `json:"response"`
}

// GetItemBaseInfoRequest represents get item info request
type GetItemBaseInfoRequest struct {
	ItemIDList string `json:"-"` // Comma-separated item IDs
}

// GetItemBaseInfoResponse represents get item info response
type GetItemBaseInfoResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ItemList []ItemBaseInfo `json:"item_list"`
	} `json:"response"`
}

// ItemBaseInfo represents item base information
type ItemBaseInfo struct {
	ItemID            int64           `json:"item_id"`
	ItemName          string          `json:"item_name"`
	ItemStatus        string          `json:"item_status"`
	WholesaleTierList []WholesaleTier `json:"wholesale_tier_list,omitempty"`
}

// ProductAPI provides product-related API methods
type ProductAPI struct {
	client *Client
}

// NewProductAPI creates a new ProductAPI instance
func NewProductAPI(client *Client) *ProductAPI {
	return &ProductAPI{client: client}
}

// UpdateItemWholesale updates wholesale tiers for an item
func (p *ProductAPI) UpdateItemWholesale(ctx context.Context, itemID int64, tiers []WholesaleTier) error {
	req := UpdateItemWholesaleRequest{
		ItemID:    itemID,
		Wholesale: tiers,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	var result UpdateItemWholesaleResponse
	if err := p.client.doPostRequest("/api/v2/product/update_item", nil, body, &result); err != nil {
		return err
	}

	if result.Error != "" {
		return fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return nil
}

// GetItemWholesale gets wholesale tiers for an item
func (p *ProductAPI) GetItemWholesale(ctx context.Context, itemID int64) ([]WholesaleTier, error) {
	params := map[string]string{
		"item_id_list": fmt.Sprintf("%d", itemID),
	}

	var result GetItemBaseInfoResponse
	if err := p.client.doRequest("GET", "/api/v2/product/get_item_base_info", params, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	if len(result.Response.ItemList) == 0 {
		return nil, fmt.Errorf("item not found: %d", itemID)
	}

	return result.Response.ItemList[0].WholesaleTierList, nil
}
