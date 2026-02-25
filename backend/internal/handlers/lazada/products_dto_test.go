package lazada

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateProductRequest_JSONFields(t *testing.T) {
	req := CreateProductRequest{
		Name:            "Test Product",
		Description:     "Test Description",
		Brand:           "TestBrand",
		PrimaryCategory: 123,
		SellerSku:       "SKU-001",
		Price:           10000.0,
		Quantity:        5,
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)

	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &out))

	assert.Equal(t, "Test Product", out["name"])
	assert.Equal(t, "Test Description", out["description"])
	assert.Equal(t, "TestBrand", out["brand"])
	assert.Equal(t, float64(123), out["primary_category"])
	assert.Equal(t, "SKU-001", out["seller_sku"])
	assert.Equal(t, float64(10000), out["price"])
	assert.Equal(t, float64(5), out["quantity"])
}

func TestCreateProductRequest_RequiredFields(t *testing.T) {
	// Minimal valid struct (all required fields present)
	req := CreateProductRequest{
		Name:            "Product",
		Description:     "Desc",
		PrimaryCategory: 1,
		SellerSku:       "SKU1",
		Price:           100,
		Quantity:        1,
	}
	assert.Equal(t, "Product", req.Name)
	assert.Equal(t, "Desc", req.Description)
	assert.Equal(t, int64(1), req.PrimaryCategory)
	assert.Equal(t, "SKU1", req.SellerSku)
	assert.Equal(t, float64(100), req.Price)
	assert.Equal(t, 1, req.Quantity)
}

func TestCreateProductRequest_OptionalBrand(t *testing.T) {
	// Brand is optional (omitempty)
	req := CreateProductRequest{
		Name:            "Product",
		Description:     "Desc",
		PrimaryCategory: 1,
		SellerSku:       "SKU1",
		Price:           100,
		Quantity:        1,
	}
	data, err := json.Marshal(req)
	assert.NoError(t, err)

	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &out))

	// brand field should be absent (omitempty + zero value)
	_, hasBrand := out["brand"]
	assert.False(t, hasBrand)
}

func TestUpdateProductRequest_JSONFields(t *testing.T) {
	req := UpdateProductRequest{
		Name:        "Updated Name",
		Description: "Updated Desc",
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)

	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &out))

	assert.Equal(t, "Updated Name", out["name"])
	assert.Equal(t, "Updated Desc", out["description"])
}

func TestUpdateProductRequest_EmptyIsValid(t *testing.T) {
	// Empty struct is valid (all fields omitempty)
	req := UpdateProductRequest{}
	data, err := json.Marshal(req)
	assert.NoError(t, err)

	var out map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &out))

	// Both fields are omitempty, should not appear in JSON
	_, hasName := out["name"]
	_, hasDesc := out["description"]
	assert.False(t, hasName)
	assert.False(t, hasDesc)
}
