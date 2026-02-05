package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPlatformFactory(t *testing.T) {
	factory := NewPlatformFactory()
	assert.NotNil(t, factory)
}

func TestPlatformFactory_SupportedPlatforms(t *testing.T) {
	factory := NewPlatformFactory()
	platforms := factory.SupportedPlatforms()

	assert.Contains(t, platforms, "shopee")
	assert.Contains(t, platforms, "lazada")
	assert.Contains(t, platforms, "tiktok")
	assert.Len(t, platforms, 3)
}

func TestPlatformFactory_GetOrderService_NotInitialized(t *testing.T) {
	factory := NewPlatformFactory()

	tests := []struct {
		name     string
		platform string
		wantErr  string
	}{
		{name: "shopee not initialized", platform: "shopee", wantErr: "shopee order service not initialized"},
		{name: "lazada not initialized", platform: "lazada", wantErr: "lazada order service not initialized"},
		{name: "tiktok not initialized", platform: "tiktok", wantErr: "tiktok order service not initialized"},
		{name: "unsupported platform", platform: "amazon", wantErr: "unsupported platform: amazon"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := factory.GetOrderService(tt.platform)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// mockOrderService implements OrderService for testing
type mockOrderService struct{}

func (m *mockOrderService) GetOrders(ctx context.Context, tenantID string, filter OrderFilter) ([]Order, int, error) {
	return []Order{}, 0, nil
}

func (m *mockOrderService) GetOrderByID(ctx context.Context, tenantID string, orderID string) (*Order, error) {
	return nil, nil
}

func (m *mockOrderService) SyncOrders(ctx context.Context, tenantID string) error {
	return nil
}

func TestPlatformFactory_SetAndGetShopeeOrderService(t *testing.T) {
	factory := NewPlatformFactory()
	mockService := &mockOrderService{}

	factory.SetShopeeOrderService(mockService)

	svc, err := factory.GetOrderService("shopee")
	require.NoError(t, err)
	assert.Equal(t, mockService, svc)
}

func TestPlatformFactory_SetAndGetLazadaOrderService(t *testing.T) {
	factory := NewPlatformFactory()
	mockService := &mockOrderService{}

	factory.SetLazadaOrderService(mockService)

	svc, err := factory.GetOrderService("lazada")
	require.NoError(t, err)
	assert.Equal(t, mockService, svc)
}

func TestPlatformFactory_SetAndGetTiktokOrderService(t *testing.T) {
	factory := NewPlatformFactory()
	mockService := &mockOrderService{}

	factory.SetTiktokOrderService(mockService)

	svc, err := factory.GetOrderService("tiktok")
	require.NoError(t, err)
	assert.Equal(t, mockService, svc)
}

func TestPlatformFactory_GetProductService(t *testing.T) {
	factory := NewPlatformFactory()

	tests := []struct {
		name     string
		platform string
		wantNil  bool
		wantErr  bool
	}{
		{name: "shopee - nil until set", platform: "shopee", wantNil: true, wantErr: false},
		{name: "lazada - nil until set", platform: "lazada", wantNil: true, wantErr: false},
		{name: "tiktok - nil until set", platform: "tiktok", wantNil: true, wantErr: false},
		{name: "unsupported platform", platform: "amazon", wantNil: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := factory.GetProductService(tt.platform)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if tt.wantNil {
				assert.Nil(t, svc)
			}
		})
	}
}

func TestPlatformFactory_MultipleServices(t *testing.T) {
	factory := NewPlatformFactory()

	shopeeService := &mockOrderService{}
	lazadaService := &mockOrderService{}
	tiktokService := &mockOrderService{}

	factory.SetShopeeOrderService(shopeeService)
	factory.SetLazadaOrderService(lazadaService)
	factory.SetTiktokOrderService(tiktokService)

	// Verify each service returns the correct mock
	s1, err := factory.GetOrderService("shopee")
	require.NoError(t, err)
	assert.Equal(t, shopeeService, s1)

	s2, err := factory.GetOrderService("lazada")
	require.NoError(t, err)
	assert.Equal(t, lazadaService, s2)

	s3, err := factory.GetOrderService("tiktok")
	require.NoError(t, err)
	assert.Equal(t, tiktokService, s3)
}

func TestOrderFilter(t *testing.T) {
	filter := OrderFilter{
		Platform:  "shopee",
		Status:    "pending",
		StartDate: "2024-01-01",
		EndDate:   "2024-01-31",
		Page:      1,
		PageSize:  20,
	}

	assert.Equal(t, "shopee", filter.Platform)
	assert.Equal(t, "pending", filter.Status)
	assert.Equal(t, "2024-01-01", filter.StartDate)
	assert.Equal(t, "2024-01-31", filter.EndDate)
	assert.Equal(t, 1, filter.Page)
	assert.Equal(t, 20, filter.PageSize)
}

func TestOrder(t *testing.T) {
	order := Order{
		ID:          "ord-123",
		Platform:    "shopee",
		OrderSN:     "SH123456789",
		Status:      "shipped",
		TotalAmount: 150000.0,
		Currency:    "IDR",
		BuyerName:   "John Doe",
		CreatedAt:   "2024-01-15T10:00:00Z",
		UpdatedAt:   "2024-01-16T08:00:00Z",
	}

	assert.Equal(t, "ord-123", order.ID)
	assert.Equal(t, "shopee", order.Platform)
	assert.Equal(t, "SH123456789", order.OrderSN)
	assert.Equal(t, "shipped", order.Status)
	assert.Equal(t, 150000.0, order.TotalAmount)
	assert.Equal(t, "IDR", order.Currency)
	assert.Equal(t, "John Doe", order.BuyerName)
}

func TestProduct(t *testing.T) {
	product := Product{
		ID:        "prod-123",
		Platform:  "lazada",
		Name:      "Test Product",
		SKU:       "SKU-001",
		Price:     99000.0,
		Stock:     50,
		Status:    "active",
		Images:    []string{"https://example.com/img1.jpg", "https://example.com/img2.jpg"},
		CreatedAt: "2024-01-10T12:00:00Z",
		UpdatedAt: "2024-01-12T14:00:00Z",
	}

	assert.Equal(t, "prod-123", product.ID)
	assert.Equal(t, "lazada", product.Platform)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, "SKU-001", product.SKU)
	assert.Equal(t, 99000.0, product.Price)
	assert.Equal(t, 50, product.Stock)
	assert.Equal(t, "active", product.Status)
	assert.Len(t, product.Images, 2)
}
