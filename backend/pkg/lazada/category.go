// Package lazada provides category API functions for Lazada platform
package lazada

import (
	"encoding/json"
	"fmt"
)

// =============================================================================
// Category Suggestion/Recommendation Types
// =============================================================================

// CategorySuggestionResponse represents the response from GET /category/suggestion/get
type CategorySuggestionResponse struct {
	Code    string               `json:"code"`
	Message string               `json:"message,omitempty"`
	Data    *CategorySuggestions `json:"data,omitempty"`
}

// CategorySuggestions contains category suggestion results
// Per Lazada API docs: /product/category/suggestion/get
type CategorySuggestions struct {
	Suggestions []CategorySuggestion `json:"categorySuggestions"`
}

// CategorySuggestion represents a single category suggestion
// API response example: { categoryPath: "...", categoryName: "T-Shirt", categoryId: "2342" }
type CategorySuggestion struct {
	CategoryID   FlexibleInt `json:"categoryId"` // Can be string or number
	CategoryName string      `json:"categoryName"`
	CategoryPath string      `json:"categoryPath"`
}

// FlexibleInt handles JSON that can be either string or number for categoryId
type FlexibleInt int64

// UnmarshalJSON implements json.Unmarshaler for FlexibleInt
func (f *FlexibleInt) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as int first
	var i int64
	if err := json.Unmarshal(data, &i); err == nil {
		*f = FlexibleInt(i)
		return nil
	}

	// Try to unmarshal as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		var n int64
		_, err := fmt.Sscanf(s, "%d", &n)
		if err == nil {
			*f = FlexibleInt(n)
			return nil
		}
	}

	return fmt.Errorf("FlexibleInt: cannot unmarshal %s", string(data))
}

// Int64 returns the int64 value
func (f FlexibleInt) Int64() int64 {
	return int64(f)
}

// =============================================================================
// Category Tree Types
// =============================================================================

// CategoryTreeResponse represents the response from GET /category/tree/get
type CategoryTreeResponse struct {
	Code    string     `json:"code"`
	Message string     `json:"message,omitempty"`
	Data    []Category `json:"data,omitempty"`
}

// Category represents a Lazada category
type Category struct {
	CategoryID int64      `json:"category_id"`
	Name       string     `json:"name"`
	Var        bool       `json:"var"`  // has variations
	Leaf       bool       `json:"leaf"` // is leaf category
	Children   []Category `json:"children,omitempty"`
}

// =============================================================================
// Category Attributes Types
// =============================================================================

// CategoryAttributesResponse represents the response from GET /category/attributes/get
type CategoryAttributesResponse struct {
	Code    string              `json:"code"`
	Message string              `json:"message,omitempty"`
	Data    []CategoryAttribute `json:"data,omitempty"`
}

// CategoryAttribute represents a Lazada category attribute
type CategoryAttribute struct {
	Name          string                    `json:"name"`
	InputType     string                    `json:"input_type"` // dropDownList, freeText, etc.
	IsMandatory   bool                      `json:"is_mandatory"`
	AttributeType string                    `json:"attribute_type"` // sku, normal
	Label         string                    `json:"label"`
	Options       []CategoryAttributeOption `json:"options,omitempty"`
}

// CategoryAttributeOption represents an attribute option
type CategoryAttributeOption struct {
	Name string `json:"name"`
}

// =============================================================================
// Category API Methods
// =============================================================================

// GetCategorySuggestion gets recommended category based on product name
// API: GET /product/category/suggestion/get
// Per Lazada docs: https://open.lazada.com/apps/doc/api?path=/product/category/suggestion/get
// Request: product_name (required)
// Response: { code: "0", data: { categorySuggestions: [{ categoryPath, categoryName, categoryId }] } }
func (c *Client) GetCategorySuggestion(productName string) (*CategorySuggestion, error) {
	if productName == "" {
		return nil, fmt.Errorf("product name is required for category suggestion")
	}

	params := map[string]string{
		"product_name": productName,
	}

	var resp CategorySuggestionResponse
	err := c.doRequest("GET", "/product/category/suggestion/get", params, &resp)
	if err != nil {
		return nil, fmt.Errorf("category suggestion API failed: %w", err)
	}

	if resp.Code != "0" && resp.Code != "" {
		return nil, fmt.Errorf("lazada API error: code=%s, message=%s", resp.Code, resp.Message)
	}

	if resp.Data == nil || len(resp.Data.Suggestions) == 0 {
		return nil, fmt.Errorf("no category suggestions found for: %s", productName)
	}

	return &resp.Data.Suggestions[0], nil
}

// GetRecommendedCategoryID is a convenience method that returns just the category ID
func (c *Client) GetRecommendedCategoryID(productName string) (int64, error) {
	suggestion, err := c.GetCategorySuggestion(productName)
	if err != nil {
		return 0, err
	}
	return suggestion.CategoryID.Int64(), nil
}

// GetCategoryTree fetches the full category tree
// API: GET /category/tree/get
func (c *Client) GetCategoryTree() ([]Category, error) {
	params := map[string]string{}

	var resp CategoryTreeResponse
	err := c.doRequest("GET", "/category/tree/get", params, &resp)
	if err != nil {
		return nil, fmt.Errorf("category tree API failed: %w", err)
	}

	if resp.Code != "0" && resp.Code != "" {
		return nil, fmt.Errorf("lazada API error: code=%s, message=%s", resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// GetCategoryAttributes fetches attributes for a specific category
// API: GET /category/attributes/get
func (c *Client) GetCategoryAttributes(categoryID int64) ([]CategoryAttribute, error) {
	params := map[string]string{
		"primary_category_id": fmt.Sprintf("%d", categoryID),
	}

	var resp CategoryAttributesResponse
	err := c.doRequest("GET", "/category/attributes/get", params, &resp)
	if err != nil {
		return nil, fmt.Errorf("category attributes API failed: %w", err)
	}

	if resp.Code != "0" && resp.Code != "" {
		return nil, fmt.Errorf("lazada API error: code=%s, message=%s", resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// GetRequiredAttributes returns only required/mandatory attributes for a category
func (c *Client) GetRequiredAttributes(categoryID int64) ([]CategoryAttribute, error) {
	attrs, err := c.GetCategoryAttributes(categoryID)
	if err != nil {
		return nil, err
	}

	required := make([]CategoryAttribute, 0)
	for _, attr := range attrs {
		if attr.IsMandatory {
			required = append(required, attr)
		}
	}

	return required, nil
}
