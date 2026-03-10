package tiktok

import (
	"fmt"
)

// =============================================================================
// Category Recommendation Types
// =============================================================================

// RecommendCategoryRequest is the request body for POST /product/202309/categories/recommend
type RecommendCategoryRequest struct {
	ProductTitle    string                   `json:"product_title"`
	Description     string                   `json:"description,omitempty"`
	Images          []RecommendCategoryImage `json:"images,omitempty"`
	CategoryVersion string                   `json:"category_version,omitempty"` // "v1" or "v2"
}

// RecommendCategoryImage represents an image for category recommendation
type RecommendCategoryImage struct {
	URI string `json:"uri"`
}

// RecommendCategoryResponse is the API response
type RecommendCategoryResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    *RecommendCategoryData `json:"data"`
}

// RecommendCategoryData contains the recommendation result
type RecommendCategoryData struct {
	LeafCategoryID string                `json:"leaf_category_id"`
	Categories     []RecommendedCategory `json:"categories"`
}

// RecommendedCategory represents a recommended category in the hierarchy
type RecommendedCategory struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Level              int      `json:"level"`
	IsLeaf             bool     `json:"is_leaf"`
	PermissionStatuses []string `json:"permission_statuses,omitempty"`
}

// =============================================================================
// Category API Methods
// =============================================================================

// RecommendCategory calls TikTok API to get recommended category based on product info
// API: POST /product/202309/categories/recommend
// Reference: backend-node/tiktok_sdk/api/productV202309Api.ts - CategoriesRecommendPost
func (c *Client) RecommendCategory(title, description string, imageURIs []string) (*RecommendCategoryData, error) {
	apiPath := "/product/202309/categories/recommend"

	// Build request body
	reqBody := RecommendCategoryRequest{
		ProductTitle:    title,
		Description:     description,
		CategoryVersion: "v1", // Use v1 for non-US markets (ID is non-US)
	}

	// Add images if available (images should be TikTok CDN URIs from UploadImage)
	for _, uri := range imageURIs {
		reqBody.Images = append(reqBody.Images, RecommendCategoryImage{URI: uri})
	}

	params := make(map[string]string)

	var resp RecommendCategoryResponse
	err := c.doRequestWithBody("POST", apiPath, params, reqBody, &resp)
	if err != nil {
		return nil, fmt.Errorf("RecommendCategory failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("TikTok API error: code=%d, message=%s", resp.Code, resp.Message)
	}

	if resp.Data == nil || resp.Data.LeafCategoryID == "" {
		return nil, fmt.Errorf("no category recommendation returned")
	}

	return resp.Data, nil
}

// GetRecommendedCategoryID is a convenience method that returns just the leaf category ID
func (c *Client) GetRecommendedCategoryID(title, description string, imageURIs []string) (string, error) {
	data, err := c.RecommendCategory(title, description, imageURIs)
	if err != nil {
		return "", err
	}
	return data.LeafCategoryID, nil
}

// =============================================================================
// Get Category Attributes Types
// =============================================================================

// GetAttributesResponse is the response from Get Attributes API
type GetAttributesResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    *GetAttributesData `json:"data"`
}

// GetAttributesData contains the attributes list
type GetAttributesData struct {
	Attributes []CategoryAttribute `json:"attributes"`
}

// CategoryAttribute represents a product attribute for a category
type CategoryAttribute struct {
	ID                    string               `json:"id"`
	Name                  string               `json:"name"`
	Type                  string               `json:"type"`        // PRODUCT_PROPERTY or SALES_PROPERTY
	IsRequired            bool                 `json:"is_requried"` // Note: TikTok API has typo
	IsCustomizable        bool                 `json:"is_customizable"`
	IsMultipleSelection   bool                 `json:"is_multiple_selection"`
	Values                []AttributeValueInfo `json:"values"`
	RequirementConditions []interface{}        `json:"requirement_conditions,omitempty"`
}

// AttributeValueInfo represents an attribute value option
type AttributeValueInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetCategoryAttributes fetches required and optional attributes for a category
// API: GET /product/202309/categories/{category_id}/attributes
func (c *Client) GetCategoryAttributes(categoryID string) ([]CategoryAttribute, error) {
	apiPath := fmt.Sprintf("/product/202309/categories/%s/attributes", categoryID)
	params := map[string]string{
		"category_version": "v2", // Use v2 for SEA market
		"locale":           "id-ID",
	}

	var resp GetAttributesResponse
	err := c.doRequest("GET", apiPath, params, &resp)
	if err != nil {
		return nil, fmt.Errorf("GetCategoryAttributes failed: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("TikTok API error: code=%d, message=%s", resp.Code, resp.Message)
	}

	if resp.Data == nil {
		return nil, fmt.Errorf("no attributes data returned")
	}

	return resp.Data.Attributes, nil
}

// GetRequiredAttributes returns only required PRODUCT_PROPERTY attributes with their first valid value
func (c *Client) GetRequiredAttributes(categoryID string) ([]ProductAttribute, error) {
	attrs, err := c.GetCategoryAttributes(categoryID)
	if err != nil {
		return nil, err
	}

	var result []ProductAttribute
	for _, attr := range attrs {
		// Only include required PRODUCT_PROPERTY attributes
		if attr.Type != "PRODUCT_PROPERTY" || !attr.IsRequired {
			continue
		}

		// Skip if no values (customizable text input)
		if len(attr.Values) == 0 && attr.IsCustomizable {
			continue
		}

		// Use first available value
		values := []AttributeValue{}
		if len(attr.Values) > 0 {
			// Use value ID for built-in attributes
			values = append(values, AttributeValue{ID: attr.Values[0].ID})
		}

		result = append(result, ProductAttribute{
			ID:     attr.ID,
			Values: values,
		})
	}

	return result, nil
}
