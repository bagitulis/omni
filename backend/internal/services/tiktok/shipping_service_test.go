package tiktok

import (
	"context"
	"errors"
	"testing"

	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTikTokClient for testing
type MockTikTokClient struct {
	mock.Mock
}

func (m *MockTikTokClient) ArrangeShipment(packageID string, req *tiktokPkg.ShipPackageRequest) (*tiktokPkg.ShipPackageResponse, error) {
	args := m.Called(packageID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tiktokPkg.ShipPackageResponse), args.Error(1)
}

func (m *MockTikTokClient) GetShippingDocument(packageID, documentType string) (string, error) {
	args := m.Called(packageID, documentType)
	return args.String(0), args.Error(1)
}

func (m *MockTikTokClient) GetOrderPackages(orderID string) (*tiktokPkg.GetPackageDetailResponse, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tiktokPkg.GetPackageDetailResponse), args.Error(1)
}

func (m *MockTikTokClient) GetOrderDetail(orderIDs []string) (*tiktokPkg.OrderDetailResponse, error) {
	args := m.Called(orderIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tiktokPkg.OrderDetailResponse), args.Error(1)
}

func (m *MockTikTokClient) GetHandoverTimeSlots(packageID string) (*tiktokPkg.HandoverTimeSlotsResponse, error) {
	args := m.Called(packageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tiktokPkg.HandoverTimeSlotsResponse), args.Error(1)
}

func (m *MockTikTokClient) GetHandoverTimeSlotsForOrder(orderID string, lineItemIDs []string) (*tiktokPkg.HandoverTimeSlotsResponse, error) {
	args := m.Called(orderID, lineItemIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tiktokPkg.HandoverTimeSlotsResponse), args.Error(1)
}

func (m *MockTikTokClient) ResolveOrderToPackageID(orderID string) (string, string, error) {
	args := m.Called(orderID)
	return args.String(0), args.String(1), args.Error(2)
}

func TestGetShippingLabel_DirectSuccess(t *testing.T) {
	mockClient := new(MockTikTokClient)
	// Expect direct success with packageID
	mockClient.On("GetShippingDocument", "pkg_123", "SHIPPING_LABEL").Return("http://pdf.url", nil)

	factory := func(ctx context.Context, tenantID string) (TikTokClient, error) {
		return mockClient, nil
	}
	svc := NewShippingServiceWithFactory("base", factory)

	url, err := svc.GetShippingLabel(context.Background(), "tenant1", "pkg_123", "SHIPPING_LABEL")

	assert.NoError(t, err)
	assert.Equal(t, "http://pdf.url", url)
	mockClient.AssertExpectations(t)
}

func TestGetShippingLabel_FailoverToOrderLookup(t *testing.T) {
	mockClient := new(MockTikTokClient)

	// 1. First attempt fails (invalid package id, actually order sn)
	mockClient.On("GetShippingDocument", "order_sn_123", "SHIPPING_LABEL").Return("", errors.New("not found"))

	// 2. Lookup order packages
	packagesResp := &tiktokPkg.GetPackageDetailResponse{
		Data: struct {
			Packages []tiktokPkg.PackageInfo `json:"packages"`
		}{
			Packages: []tiktokPkg.PackageInfo{
				{ID: "real_pkg_456"},
			},
		},
	}
	mockClient.On("GetOrderPackages", "order_sn_123").Return(packagesResp, nil)

	// 3. Retry with real package ID
	mockClient.On("GetShippingDocument", "real_pkg_456", "SHIPPING_LABEL").Return("http://pdf.url/real", nil)

	factory := func(ctx context.Context, tenantID string) (TikTokClient, error) {
		return mockClient, nil
	}
	svc := NewShippingServiceWithFactory("base", factory)

	url, err := svc.GetShippingLabel(context.Background(), "tenant1", "order_sn_123", "SHIPPING_LABEL")

	assert.NoError(t, err)
	assert.Equal(t, "http://pdf.url/real", url)
	mockClient.AssertExpectations(t)
}

func TestGetShippingLabel_FailoverFails(t *testing.T) {
	mockClient := new(MockTikTokClient)

	// 1. First attempt fails
	mockClient.On("GetShippingDocument", "order_sn_123", "SHIPPING_LABEL").Return("", errors.New("original error"))

	// 2. Lookup fails too
	mockClient.On("GetOrderPackages", "order_sn_123").Return(nil, errors.New("lookup error"))

	factory := func(ctx context.Context, tenantID string) (TikTokClient, error) {
		return mockClient, nil
	}
	svc := NewShippingServiceWithFactory("base", factory)

	url, err := svc.GetShippingLabel(context.Background(), "tenant1", "order_sn_123", "SHIPPING_LABEL")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "original error") // Should return original error
	assert.Equal(t, "", url)
	mockClient.AssertExpectations(t)
}
