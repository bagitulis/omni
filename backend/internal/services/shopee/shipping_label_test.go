package shopee

import (
	"context"
	"fmt"
	"testing"

	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAPIClient is a mock of the APIClient interface
type MockAPIClient struct {
	mock.Mock
}

func (m *MockAPIClient) GetShippingParameter(orderSN string) (*shopeePkg.GetShippingParameterResponse, error) {
	args := m.Called(orderSN)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.GetShippingParameterResponse), args.Error(1)
}

func (m *MockAPIClient) ShipOrder(req shopeePkg.ShipOrderRequest) (*shopeePkg.ShipOrderResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*shopeePkg.ShipOrderResponse), args.Error(1)
}

func (m *MockAPIClient) GetTrackingNumber(orderSN string) (*shopeePkg.GetTrackingNumberResponse, error) {
	args := m.Called(orderSN)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.GetTrackingNumberResponse), args.Error(1)
}

func (m *MockAPIClient) CreateShippingDocument(orderSN, packageNumber string) (*shopeePkg.CreateShippingDocumentResponse, error) {
	args := m.Called(orderSN, packageNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.CreateShippingDocumentResponse), args.Error(1)
}

func (m *MockAPIClient) DownloadShippingDocument(orderSN, packageNumber, documentType string) (*shopeePkg.DownloadShippingDocumentResponse, error) {
	args := m.Called(orderSN, packageNumber, documentType)
	return args.Get(0).(*shopeePkg.DownloadShippingDocumentResponse), args.Error(1)
}

func (m *MockAPIClient) GetShippingDocumentResult(orderSN, packageNumber string) (*shopeePkg.GetShippingDocumentResultResponse, error) {
	args := m.Called(orderSN, packageNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.GetShippingDocumentResultResponse), args.Error(1)
}

func (m *MockAPIClient) GetShippingDocumentDataInfo(orderSN, packageNumber string) (*shopeePkg.ShippingDocumentDataInfoResponse, error) {
	args := m.Called(orderSN, packageNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.ShippingDocumentDataInfoResponse), args.Error(1)
}

func TestGetShippingLabel_HandlesBatchApiAllFailed(t *testing.T) {
	// Setup
	mockClient := new(MockAPIClient)
	service := &ShippingService{
		pkgClient: mockClient, // Inject mock directly
	}
	ctx := context.Background()
	orderSN := "260203Q98DEKHK"
	pkgNum := ""
	docType := "THERMAL_AIR_WAYBILL"

	// Mock EnsureShipmentReady success (Tracking Exists)
	mockTrackingResp := &shopeePkg.GetTrackingNumberResponse{}
	mockTrackingResp.Response.TrackingNumber = "JP123456789"
	mockClient.On("GetTrackingNumber", orderSN).Return(mockTrackingResp, nil)

	// Mock CreateShippingDocument FAILURE with batch_api_all_failed
	mockCreateResp := &shopeePkg.CreateShippingDocumentResponse{
		Error:   "common.batch_api_all_failed",
		Message: "All failed",
	}
	// Add the failure details
	failDetail := struct {
		OrderSN       string `json:"order_sn"`
		PackageNumber string `json:"package_number"`
		Status        string `json:"status"`
		FailError     string `json:"fail_error,omitempty"`
		FailMessage   string `json:"fail_message,omitempty"`
	}{
		OrderSN:     orderSN,
		Status:      "FAILED",
		FailError:   "logistics.tracking_number_invalid",
		FailMessage: "The tracking number is invalid",
	}
	mockCreateResp.Response.ResultList = append(mockCreateResp.Response.ResultList, failDetail)

	// The client.CreateShippingDocument will return the response AND an error because Error string is set
	mockClient.On("CreateShippingDocument", orderSN, pkgNum).Return(mockCreateResp, fmt.Errorf("shopee API error: %s - %s", mockCreateResp.Error, mockCreateResp.Message))

	// Mock DownloadShippingDocument to also fail with the tracking number error
	// The code proceeds to download even after batch_api_all_failed with tracking_number_invalid
	mockDownloadResp := &shopeePkg.DownloadShippingDocumentResponse{
		Error:   "logistics.tracking_number_invalid",
		Message: "The tracking number is invalid",
	}
	mockClient.On("DownloadShippingDocument", orderSN, pkgNum, docType).Return(mockDownloadResp, fmt.Errorf("logistics.tracking_number_invalid: The tracking number is invalid"))

	// Mock GetShippingDocumentDataInfo to fail - this is called in the fallback path
	mockClient.On("GetShippingDocumentDataInfo", orderSN, pkgNum).Return(nil, fmt.Errorf("logistics.tracking_number_invalid: Cannot get document data"))

	// Mock GetTrackingNumber to also fail so fallback chain fails completely
	// Override the earlier mock by removing it and setting a new one
	// Note: We need to set this AFTER the fallback-path specific mock
	mockClient.On("GetTrackingNumber", orderSN).Unset()
	mockClient.On("GetTrackingNumber", orderSN).Return(nil, fmt.Errorf("logistics.tracking_number_invalid: Tracking number is invalid"))

	// Execute
	result, err := service.GetShippingLabel(ctx, orderSN, pkgNum, docType)

	// Assert
	assert.NoError(t, err) // Should NOT return error
	assert.NotNil(t, result)
	assert.Equal(t, "FAILED", result.Status)
	assert.Contains(t, result.ErrorMessage, "logistics.tracking_number_invalid")

	mockClient.AssertExpectations(t)
}
