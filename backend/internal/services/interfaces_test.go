package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time checks that mock types satisfy the interfaces defined in interfaces.go.
// If any method is missing this file will not compile.
var _ OrderService = (*mockOrderSvc)(nil)
var _ ProductService = (*mockProductSvc)(nil)
var _ TokenService = (*mockTokenSvc)(nil)

// mockOrderSvc implements OrderService for compile-time verification.
type mockOrderSvc struct{}

func (m *mockOrderSvc) GetOrders(_ context.Context, _ string, _ OrderFilter) ([]Order, int, error) {
	return nil, 0, nil
}
func (m *mockOrderSvc) GetOrderByID(_ context.Context, _ string, _ string) (*Order, error) {
	return nil, nil
}
func (m *mockOrderSvc) SyncOrders(_ context.Context, _ string) error { return nil }

// mockProductSvc implements ProductService for compile-time verification.
type mockProductSvc struct{}

func (m *mockProductSvc) GetProducts(_ context.Context, _ string, _, _ int) ([]Product, int, error) {
	return nil, 0, nil
}
func (m *mockProductSvc) GetProductByID(_ context.Context, _ string, _ string) (*Product, error) {
	return nil, nil
}
func (m *mockProductSvc) SyncProducts(_ context.Context, _ string) error { return nil }

// mockTokenSvc implements TokenService for compile-time verification.
type mockTokenSvc struct{}

func (m *mockTokenSvc) GetAccessToken(_ context.Context, _, _ string) (string, error) {
	return "", nil
}
func (m *mockTokenSvc) RefreshToken(_ context.Context, _, _ string) error { return nil }
func (m *mockTokenSvc) IsTokenValid(_ context.Context, _, _ string) bool  { return true }

// TestOrderFilter_Fields verifies that OrderFilter has expected fields.
func TestOrderFilter_Fields(t *testing.T) {
	f := OrderFilter{
		Platform:  "shopee",
		Status:    "completed",
		StartDate: "2024-01-01",
		EndDate:   "2024-12-31",
		Page:      1,
		PageSize:  20,
	}

	assert.Equal(t, "shopee", f.Platform)
	assert.Equal(t, "completed", f.Status)
	assert.Equal(t, "2024-01-01", f.StartDate)
	assert.Equal(t, "2024-12-31", f.EndDate)
	assert.Equal(t, 1, f.Page)
	assert.Equal(t, 20, f.PageSize)
}

// TestOrder_JSONTags verifies Order struct serializes to snake_case JSON.
func TestOrder_JSONTags(t *testing.T) {
	o := Order{
		ID:          "ord-1",
		Platform:    "shopee",
		OrderSN:     "SN123",
		Status:      "shipped",
		TotalAmount: 99.99,
		Currency:    "IDR",
		BuyerName:   "John Doe",
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-02T00:00:00Z",
	}

	data, err := json.Marshal(o)
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &m))

	assert.Equal(t, "ord-1", m["id"])
	assert.Equal(t, "shopee", m["platform"])
	assert.Equal(t, "SN123", m["order_sn"])
	assert.Equal(t, "shipped", m["status"])
	assert.InDelta(t, 99.99, m["total_amount"], 0.001)
	assert.Equal(t, "IDR", m["currency"])
	assert.Equal(t, "John Doe", m["buyer_name"])
	assert.Equal(t, "2024-01-01T00:00:00Z", m["created_at"])
	assert.Equal(t, "2024-01-02T00:00:00Z", m["updated_at"])
}

// TestProduct_JSONTags verifies Product struct serializes to snake_case JSON.
func TestProduct_JSONTags(t *testing.T) {
	p := Product{
		ID:        "prod-1",
		Platform:  "lazada",
		Name:      "Widget",
		SKU:       "SKU-001",
		Price:     19.50,
		Stock:     100,
		Status:    "active",
		Images:    []string{"img1.jpg", "img2.jpg"},
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-02T00:00:00Z",
	}

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &m))

	assert.Equal(t, "prod-1", m["id"])
	assert.Equal(t, "lazada", m["platform"])
	assert.Equal(t, "Widget", m["name"])
	assert.Equal(t, "SKU-001", m["sku"])
	assert.InDelta(t, 19.50, m["price"], 0.001)
	assert.EqualValues(t, 100, m["stock"])
	assert.Equal(t, "active", m["status"])
	assert.Equal(t, "2024-01-01T00:00:00Z", m["created_at"])
	assert.Equal(t, "2024-01-02T00:00:00Z", m["updated_at"])

	images, ok := m["images"].([]interface{})
	require.True(t, ok)
	assert.Len(t, images, 2)
}

// TestOrderService_MockImplementation verifies mock satisfies interface at runtime.
func TestOrderService_MockImplementation(t *testing.T) {
	var svc OrderService = &mockOrderSvc{}

	orders, count, err := svc.GetOrders(context.Background(), "t1", OrderFilter{})
	assert.NoError(t, err)
	assert.Nil(t, orders)
	assert.Equal(t, 0, count)

	order, err := svc.GetOrderByID(context.Background(), "t1", "order-1")
	assert.NoError(t, err)
	assert.Nil(t, order)

	err = svc.SyncOrders(context.Background(), "t1")
	assert.NoError(t, err)
}

// TestProductService_MockImplementation verifies mock satisfies interface at runtime.
func TestProductService_MockImplementation(t *testing.T) {
	var svc ProductService = &mockProductSvc{}

	products, count, err := svc.GetProducts(context.Background(), "t1", 1, 10)
	assert.NoError(t, err)
	assert.Nil(t, products)
	assert.Equal(t, 0, count)

	product, err := svc.GetProductByID(context.Background(), "t1", "p-1")
	assert.NoError(t, err)
	assert.Nil(t, product)

	err = svc.SyncProducts(context.Background(), "t1")
	assert.NoError(t, err)
}

// TestTokenService_MockImplementation verifies mock satisfies interface at runtime.
func TestTokenService_MockImplementation(t *testing.T) {
	var svc TokenService = &mockTokenSvc{}

	token, err := svc.GetAccessToken(context.Background(), "t1", "shopee")
	assert.NoError(t, err)
	assert.Empty(t, token)

	err = svc.RefreshToken(context.Background(), "t1", "shopee")
	assert.NoError(t, err)

	valid := svc.IsTokenValid(context.Background(), "t1", "shopee")
	assert.True(t, valid)
}
