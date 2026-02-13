package shopee

import (
	"context"
	"errors"
	"testing"

	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockShippingClient mocks shippingClient interface for testing
type MockShippingClient struct {
	mock.Mock
}

func (m *MockShippingClient) GetShippingParameter(orderSN string) (*shopeePkg.GetShippingParameterResponse, error) {
	args := m.Called(orderSN)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.GetShippingParameterResponse), args.Error(1)
}

func (m *MockShippingClient) ShipOrder(req shopeePkg.ShipOrderRequest) (*shopeePkg.ShipOrderResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.ShipOrderResponse), args.Error(1)
}

func (m *MockShippingClient) GetTrackingNumber(orderSN string) (*shopeePkg.GetTrackingNumberResponse, error) {
	args := m.Called(orderSN)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.GetTrackingNumberResponse), args.Error(1)
}

func (m *MockShippingClient) GetShippingDocumentParameter(orderSN, packageNumber string) (*shopeePkg.GetShippingDocumentParameterResponse, error) {
	args := m.Called(orderSN, packageNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.GetShippingDocumentParameterResponse), args.Error(1)
}

func (m *MockShippingClient) CreateShippingDocument(orderSN, packageNumber string) (*shopeePkg.CreateShippingDocumentResponse, error) {
	args := m.Called(orderSN, packageNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.CreateShippingDocumentResponse), args.Error(1)
}

func (m *MockShippingClient) CreateShippingDocumentWithOptions(orderSN, packageNumber string, options shopeePkg.ShippingDocumentRequestOptions) (*shopeePkg.CreateShippingDocumentResponse, error) {
	args := m.Called(orderSN, packageNumber, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.CreateShippingDocumentResponse), args.Error(1)
}

func (m *MockShippingClient) GetShippingDocumentResult(orderSN, packageNumber string) (*shopeePkg.GetShippingDocumentResultResponse, error) {
	args := m.Called(orderSN, packageNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.GetShippingDocumentResultResponse), args.Error(1)
}

func (m *MockShippingClient) GetShippingDocumentResultWithOptions(orderSN, packageNumber string, options shopeePkg.ShippingDocumentRequestOptions) (*shopeePkg.GetShippingDocumentResultResponse, error) {
	args := m.Called(orderSN, packageNumber, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.GetShippingDocumentResultResponse), args.Error(1)
}

func (m *MockShippingClient) DownloadShippingDocument(orderSN, packageNumber, documentType string) (*shopeePkg.DownloadShippingDocumentResponse, error) {
	args := m.Called(orderSN, packageNumber, documentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.DownloadShippingDocumentResponse), args.Error(1)
}

func (m *MockShippingClient) GetShippingDocumentDataInfo(orderSN, packageNumber string) (*shopeePkg.ShippingDocumentDataInfoResponse, error) {
	args := m.Called(orderSN, packageNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.ShippingDocumentDataInfoResponse), args.Error(1)
}

func (m *MockShippingClient) SearchPackageList(packageStatus int, cursor string, pageSize int) (*shopeePkg.SearchPackageListResponse, error) {
	args := m.Called(packageStatus, cursor, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.SearchPackageListResponse), args.Error(1)
}

func (m *MockShippingClient) GetPackageDetail(packageNumbers []string) (*shopeePkg.GetPackageDetailResponse, error) {
	args := m.Called(packageNumbers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.GetPackageDetailResponse), args.Error(1)
}

func TestNewShippingService_WithMock(t *testing.T) {
	mockClient := new(MockShippingClient)

	service := &ShippingService{
		pkgClient: mockClient,
		tenantID:  "tenant1",
	}

	assert.NotNil(t, service)
	assert.Equal(t, "tenant1", service.tenantID)
}

func TestShippingService_GetShippingOptions_Success(t *testing.T) {
	mockClient := new(MockShippingClient)
	ctx := context.Background()

	mockResp := &shopeePkg.GetShippingParameterResponse{}
	mockResp.Response.Pickup.AddressList = []shopeePkg.PickupAddressInfo{
		{AddressID: 1, Address: "Test Address 1"},
		{AddressID: 2, Address: "Test Address 2"},
	}
	mockResp.Response.Dropoff.BranchList = []shopeePkg.BranchInfo{
		{BranchID: 100, Address: "Branch 1"},
	}

	mockClient.On("GetShippingParameter", "ORDER123").Return(mockResp, nil)

	service := &ShippingService{pkgClient: mockClient, tenantID: "tenant1"}

	options, err := service.GetShippingOptions(ctx, "ORDER123")

	assert.NoError(t, err)
	assert.NotNil(t, options)
	assert.Len(t, options.Pickup, 2)
	assert.Len(t, options.Dropoff, 1)
	mockClient.AssertExpectations(t)
}

func TestShippingService_GetShippingOptions_APIError(t *testing.T) {
	mockClient := new(MockShippingClient)
	ctx := context.Background()

	mockClient.On("GetShippingParameter", "ORDER123").Return(nil, errors.New("API error"))

	service := &ShippingService{pkgClient: mockClient, tenantID: "tenant1"}

	options, err := service.GetShippingOptions(ctx, "ORDER123")

	assert.Error(t, err)
	assert.Nil(t, options)
	assert.Contains(t, err.Error(), "get shipping parameter")
	mockClient.AssertExpectations(t)
}

func TestShippingService_GetTrackingInfo_Success(t *testing.T) {
	mockClient := new(MockShippingClient)
	ctx := context.Background()

	mockResp := &shopeePkg.GetTrackingNumberResponse{}
	mockResp.Response.TrackingNumber = "TRACK123456"

	mockClient.On("GetTrackingNumber", "ORDER123").Return(mockResp, nil)

	service := &ShippingService{pkgClient: mockClient, tenantID: "tenant1"}

	info, err := service.GetTrackingInfo(ctx, "ORDER123")

	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "TRACK123456", info.TrackingNumber)
	assert.Equal(t, "SHIPPED", info.Status)
	mockClient.AssertExpectations(t)
}

func TestShippingService_GetTrackingInfo_APIError(t *testing.T) {
	mockClient := new(MockShippingClient)
	ctx := context.Background()

	mockClient.On("GetTrackingNumber", "ORDER123").Return(nil, errors.New("tracking error"))

	service := &ShippingService{pkgClient: mockClient, tenantID: "tenant1"}

	info, err := service.GetTrackingInfo(ctx, "ORDER123")

	assert.Error(t, err)
	assert.Nil(t, info)
	mockClient.AssertExpectations(t)
}

func TestShippingService_GetShipmentInfo_Success(t *testing.T) {
	mockClient := new(MockShippingClient)
	ctx := context.Background()

	mockResp := &shopeePkg.GetTrackingNumberResponse{}
	mockResp.Response.TrackingNumber = "TRACK789"

	mockClient.On("GetTrackingNumber", "ORDER456").Return(mockResp, nil)

	service := &ShippingService{pkgClient: mockClient, tenantID: "tenant1"}

	info, err := service.GetShipmentInfo(ctx, "ORDER456")

	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "ORDER456", info.OrderSN)
	assert.Equal(t, "TRACK789", info.TrackingNumber)
	mockClient.AssertExpectations(t)
}

func TestShippingService_ArrangeShipment_WithDropoff(t *testing.T) {
	mockClient := new(MockShippingClient)
	ctx := context.Background()

	// Mock ShipOrder success
	mockShipResp := &shopeePkg.ShipOrderResponse{}
	mockClient.On("ShipOrder", mock.MatchedBy(func(req shopeePkg.ShipOrderRequest) bool {
		return req.OrderSN == "ORDER123" && req.Dropoff != nil && req.Dropoff.BranchID == 100
	})).Return(mockShipResp, nil)

	// Mock GetTrackingNumber for ensureShipmentReady
	mockTrackResp := &shopeePkg.GetTrackingNumberResponse{}
	mockTrackResp.Response.TrackingNumber = "TRACK_DROP"
	mockClient.On("GetTrackingNumber", "ORDER123").Return(mockTrackResp, nil)

	service := &ShippingService{pkgClient: mockClient, tenantID: "tenant1"}

	req := ArrangeShipmentRequest{
		OrderSN: "ORDER123",
		DropOff: &DropOffInfo{BranchID: 100},
	}

	info, err := service.ArrangeShipment(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "TRACK_DROP", info.TrackingNumber)
	mockClient.AssertExpectations(t)
}

func TestShippingService_ArrangeShipment_WithPickupAndTimeSlot(t *testing.T) {
	mockClient := new(MockShippingClient)
	ctx := context.Background()

	// Mock ShipOrder success - should include pickup_time_id
	mockShipResp := &shopeePkg.ShipOrderResponse{}
	mockClient.On("ShipOrder", mock.MatchedBy(func(req shopeePkg.ShipOrderRequest) bool {
		return req.OrderSN == "ORDER123" && req.Pickup != nil && req.Pickup.AddressID == 1 && req.Pickup.PickupTimeID == "slot_1"
	})).Return(mockShipResp, nil)

	// Mock GetTrackingNumber for ensureShipmentReady
	mockTrackResp := &shopeePkg.GetTrackingNumberResponse{}
	mockTrackResp.Response.TrackingNumber = "TRACK_PICKUP"
	mockClient.On("GetTrackingNumber", "ORDER123").Return(mockTrackResp, nil)

	service := &ShippingService{pkgClient: mockClient, tenantID: "tenant1"}

	req := ArrangeShipmentRequest{
		OrderSN: "ORDER123",
		Pickup:  &PickupInfo{AddressID: 1, PickupTimeID: "slot_1"}, // Time slot provided
	}

	info, err := service.ArrangeShipment(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "TRACK_PICKUP", info.TrackingNumber)
	mockClient.AssertExpectations(t)
}

func TestShippingService_ArrangeShipment_MissingInfo(t *testing.T) {
	mockClient := new(MockShippingClient)
	ctx := context.Background()

	service := &ShippingService{pkgClient: mockClient, tenantID: "tenant1"}

	// No pickup, dropoff, or tracking number
	req := ArrangeShipmentRequest{
		OrderSN: "ORDER123",
	}

	info, err := service.ArrangeShipment(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "pickup, dropoff, or tracking_number is required")
}

func TestShippingService_ArrangeShipment_WithTrackingNumber(t *testing.T) {
	mockClient := new(MockShippingClient)
	ctx := context.Background()

	// Mock ShipOrder with non-integrated tracking
	mockShipResp := &shopeePkg.ShipOrderResponse{}
	mockClient.On("ShipOrder", mock.MatchedBy(func(req shopeePkg.ShipOrderRequest) bool {
		return req.OrderSN == "ORDER123" && req.NonIntegrated != nil && req.NonIntegrated.TrackingNumber == "MANUAL_TRACK"
	})).Return(mockShipResp, nil)

	// Mock GetTrackingNumber
	mockTrackResp := &shopeePkg.GetTrackingNumberResponse{}
	mockTrackResp.Response.TrackingNumber = "MANUAL_TRACK"
	mockClient.On("GetTrackingNumber", "ORDER123").Return(mockTrackResp, nil)

	service := &ShippingService{pkgClient: mockClient, tenantID: "tenant1"}

	req := ArrangeShipmentRequest{
		OrderSN:    "ORDER123",
		TrackingNo: "MANUAL_TRACK",
	}

	info, err := service.ArrangeShipment(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, info)
	mockClient.AssertExpectations(t)
}

func TestShippingService_CalculateShippingFee(t *testing.T) {
	mockClient := new(MockShippingClient)
	ctx := context.Background()

	service := &ShippingService{pkgClient: mockClient, tenantID: "tenant1"}

	// This method returns 0 as Shopee doesn't expose shipping fee calculation directly
	fee, err := service.CalculateShippingFee(ctx, "ORDER123", 1)

	assert.NoError(t, err)
	assert.Equal(t, float64(0), fee)
}
