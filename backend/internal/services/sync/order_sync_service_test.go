package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type getOrdersByStatusCall struct {
	platform PlatformType
	status   string
	limit    int
}

type orderQueryRepositoryMock struct {
	ordersByPlatform map[PlatformType][]Order
	calls            []getOrdersByStatusCall
}

func (m *orderQueryRepositoryMock) SaveOrders(_ context.Context, _ PlatformType, _ []Order) error {
	return nil
}

func (m *orderQueryRepositoryMock) GetOrdersByStatus(
	_ context.Context,
	platform PlatformType,
	status string,
	limit int,
) ([]Order, error) {
	m.calls = append(m.calls, getOrdersByStatusCall{
		platform: platform,
		status:   status,
		limit:    limit,
	})

	if m.ordersByPlatform == nil {
		return nil, nil
	}

	orders := m.ordersByPlatform[platform]
	if len(orders) == 0 {
		return nil, nil
	}

	result := make([]Order, len(orders))
	copy(result, orders)
	return result, nil
}

func (m *orderQueryRepositoryMock) ClearOrdersByStatus(_ context.Context, _ PlatformType, _ string) error {
	return nil
}

func (m *orderQueryRepositoryMock) GetOrdersCount(_ context.Context, _ PlatformType, _ string) (int64, error) {
	return 0, nil
}

func TestOrderQueryServiceGetOrdersByCategorySkipsUnsupportedPlatformStatus(t *testing.T) {
	repo := &orderQueryRepositoryMock{
		ordersByPlatform: map[PlatformType][]Order{
			PlatformShopee: {{OrderSN: "SP-1"}},
			PlatformLazada: {{OrderSN: "LZ-1"}},
			PlatformTiktok: {{OrderSN: "TT-1"}},
		},
	}
	service := NewOrderQueryService(repo)

	orders, err := service.GetOrdersByCategory(context.Background(), StatusUnpaid, nil)

	assert.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.Len(t, repo.calls, 2)
	assert.ElementsMatch(t,
		[]getOrdersByStatusCall{
			{platform: PlatformShopee, status: "UNPAID", limit: 0},
			{platform: PlatformTiktok, status: "UNPAID", limit: 0},
		},
		repo.calls,
	)
}

func TestOrderQueryServiceGetOrdersByCategorySpecificUnsupportedPlatformReturnsEmpty(t *testing.T) {
	repo := &orderQueryRepositoryMock{
		ordersByPlatform: map[PlatformType][]Order{
			PlatformLazada: {{OrderSN: "LZ-1"}},
		},
	}
	service := NewOrderQueryService(repo)
	lazada := PlatformLazada

	orders, err := service.GetOrdersByCategory(context.Background(), StatusUnpaid, &lazada)

	assert.NoError(t, err)
	assert.Empty(t, orders)
	assert.Empty(t, repo.calls)
}

func TestOrderQueryServiceGetOrdersByCategorySpecificSupportedPlatform(t *testing.T) {
	repo := &orderQueryRepositoryMock{
		ordersByPlatform: map[PlatformType][]Order{
			PlatformLazada: {{OrderSN: "LZ-TO-PACK"}},
		},
	}
	service := NewOrderQueryService(repo)
	lazada := PlatformLazada

	orders, err := service.GetOrdersByCategory(context.Background(), StatusUnprocess, &lazada)

	assert.NoError(t, err)
	assert.Len(t, orders, 1)
	assert.Len(t, repo.calls, 1)
	assert.Equal(t, getOrdersByStatusCall{platform: PlatformLazada, status: "topack", limit: 0}, repo.calls[0])
}
