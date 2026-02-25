package tiktok

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductDraft_JSONFields(t *testing.T) {
	draft := ProductDraft{
		ID:          "draft-001",
		TenantID:    "tenant-abc",
		Title:       "Test Product",
		Description: "A test product description",
		CategoryID:  "cat-001",
		Images:      []string{"https://example.com/img1.jpg"},
		SKUs: []ProductSKU{
			{SKU: "SKU001", Price: 99.99, Stock: 10},
		},
		Status:    "draft",
		CreatedAt: 1700000000,
		UpdatedAt: 1700000001,
	}

	data, err := json.Marshal(draft)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	// Verify snake_case JSON fields
	assert.Equal(t, "draft-001", decoded["id"])
	assert.Equal(t, "tenant-abc", decoded["tenant_id"])
	assert.Equal(t, "Test Product", decoded["title"])
	assert.Equal(t, "A test product description", decoded["description"])
	assert.Equal(t, "cat-001", decoded["category_id"])
	assert.Equal(t, "draft", decoded["status"])
	assert.Equal(t, float64(1700000000), decoded["created_at"])
	assert.Equal(t, float64(1700000001), decoded["updated_at"])
}

func TestProductDraft_Images(t *testing.T) {
	draft := ProductDraft{
		Images: []string{"img1.jpg", "img2.jpg"},
	}

	data, err := json.Marshal(draft)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	images, ok := decoded["images"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, images, 2)
	assert.Equal(t, "img1.jpg", images[0])
}

func TestProductSKU_JSONFields(t *testing.T) {
	sku := ProductSKU{
		SKU:      "SKU001",
		Price:    49.99,
		Stock:    100,
		ImageURL: "https://example.com/sku.jpg",
	}

	data, err := json.Marshal(sku)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, "SKU001", decoded["sku"])
	assert.Equal(t, float64(49.99), decoded["price"])
	assert.Equal(t, float64(100), decoded["stock"])
	assert.Equal(t, "https://example.com/sku.jpg", decoded["image_url"])
}

func TestProductSKU_ImageURL_Omitempty(t *testing.T) {
	sku := ProductSKU{
		SKU:      "SKU001",
		Price:    49.99,
		Stock:    5,
		ImageURL: "", // empty - should be omitted
	}

	data, err := json.Marshal(sku)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	_, hasImageURL := decoded["image_url"]
	assert.False(t, hasImageURL, "image_url should be omitted when empty")
}

func TestCategory_JSONFields(t *testing.T) {
	cat := Category{
		ID:       "cat-001",
		Name:     "Electronics",
		ParentID: "root",
		Level:    1,
		IsLeaf:   false,
		Children: []Category{
			{ID: "cat-002", Name: "Phones", Level: 2, IsLeaf: true},
		},
	}

	data, err := json.Marshal(cat)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, "cat-001", decoded["id"])
	assert.Equal(t, "Electronics", decoded["name"])
	assert.Equal(t, "root", decoded["parent_id"])
	assert.Equal(t, float64(1), decoded["level"])
	assert.Equal(t, false, decoded["is_leaf"])
}

func TestCategory_ParentID_Omitempty(t *testing.T) {
	cat := Category{
		ID:       "cat-001",
		Name:     "Root Category",
		ParentID: "", // empty - should be omitted
		Level:    1,
	}

	data, err := json.Marshal(cat)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	_, hasParentID := decoded["parent_id"]
	assert.False(t, hasParentID, "parent_id should be omitted when empty")
}

func TestAttribute_JSONFields(t *testing.T) {
	attr := Attribute{
		ID:          "attr-001",
		Name:        "Color",
		Type:        "string",
		Required:    false,
		Options:     []string{"Red", "Blue", "Green"},
		InputType:   "dropdown",
		CustomValue: false,
	}

	data, err := json.Marshal(attr)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, "attr-001", decoded["id"])
	assert.Equal(t, "Color", decoded["name"])
	assert.Equal(t, "string", decoded["type"])
	assert.Equal(t, false, decoded["required"])
	assert.Equal(t, "dropdown", decoded["input_type"])
	assert.Equal(t, false, decoded["custom_value"])

	options, ok := decoded["options"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, options, 3)
}

func TestAttribute_Options_Omitempty(t *testing.T) {
	attr := Attribute{
		ID:        "attr-001",
		Name:      "Brand",
		Type:      "string",
		Required:  true,
		InputType: "text",
		Options:   nil, // nil - should be omitted
	}

	data, err := json.Marshal(attr)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	_, hasOptions := decoded["options"]
	assert.False(t, hasOptions, "options should be omitted when nil")
}

func TestCategoryRule_JSONFields(t *testing.T) {
	rule := CategoryRule{
		MaxImages:      9,
		MaxSKUs:        50,
		MaxDescription: 10000,
		RequiredFields: []string{"title", "description", "price"},
	}

	data, err := json.Marshal(rule)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, float64(9), decoded["max_images"])
	assert.Equal(t, float64(50), decoded["max_skus"])
	assert.Equal(t, float64(10000), decoded["max_description"])

	fields, ok := decoded["required_fields"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, fields, 3)
}

func TestBrand_JSONFields(t *testing.T) {
	brand := Brand{
		ID:   "brand-001",
		Name: "Nike",
	}

	data, err := json.Marshal(brand)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, "brand-001", decoded["id"])
	assert.Equal(t, "Nike", decoded["name"])
}

func TestDeliveryOption_JSONFields(t *testing.T) {
	opt := DeliveryOption{
		ID:          "standard",
		Name:        "Standard Shipping",
		Type:        "standard",
		MaxWeight:   30.0,
		MaxSize:     "60x60x60",
		IsAvailable: true,
	}

	data, err := json.Marshal(opt)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, "standard", decoded["id"])
	assert.Equal(t, "Standard Shipping", decoded["name"])
	assert.Equal(t, "standard", decoded["type"])
	assert.Equal(t, float64(30), decoded["max_weight"])
	assert.Equal(t, "60x60x60", decoded["max_size"])
	assert.Equal(t, true, decoded["is_available"])
}

func TestWarehouse_JSONFields(t *testing.T) {
	wh := Warehouse{
		ID:        "wh-001",
		Name:      "Main Warehouse",
		Address:   "Jakarta, Indonesia",
		IsDefault: true,
		Status:    "active",
	}

	data, err := json.Marshal(wh)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, "wh-001", decoded["id"])
	assert.Equal(t, "Main Warehouse", decoded["name"])
	assert.Equal(t, "Jakarta, Indonesia", decoded["address"])
	assert.Equal(t, true, decoded["is_default"])
	assert.Equal(t, "active", decoded["status"])
}

func TestProductDraft_Unmarshal(t *testing.T) {
	raw := `{
		"id": "draft-xyz",
		"tenant_id": "tenant-1",
		"title": "My Product",
		"description": "Description here",
		"category_id": "cat-100",
		"images": ["img1.jpg"],
		"skus": [{"sku": "S001", "price": 15.5, "stock": 20}],
		"status": "draft",
		"created_at": 1700000000,
		"updated_at": 1700000005
	}`

	var draft ProductDraft
	err := json.Unmarshal([]byte(raw), &draft)
	assert.NoError(t, err)
	assert.Equal(t, "draft-xyz", draft.ID)
	assert.Equal(t, "tenant-1", draft.TenantID)
	assert.Equal(t, "My Product", draft.Title)
	assert.Len(t, draft.Images, 1)
	assert.Len(t, draft.SKUs, 1)
	assert.Equal(t, "S001", draft.SKUs[0].SKU)
	assert.Equal(t, float64(15.5), draft.SKUs[0].Price)
}
