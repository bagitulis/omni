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

func (m *MockAPIClient) GetShippingDocumentParameter(orderSN, packageNumber string) (*shopeePkg.GetShippingDocumentParameterResponse, error) {
	args := m.Called(orderSN, packageNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.GetShippingDocumentParameterResponse), args.Error(1)
}

func (m *MockAPIClient) CreateShippingDocument(orderSN, packageNumber string) (*shopeePkg.CreateShippingDocumentResponse, error) {
	args := m.Called(orderSN, packageNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.CreateShippingDocumentResponse), args.Error(1)
}

func (m *MockAPIClient) CreateShippingDocumentWithOptions(orderSN, packageNumber string, options shopeePkg.ShippingDocumentRequestOptions) (*shopeePkg.CreateShippingDocumentResponse, error) {
	args := m.Called(orderSN, packageNumber, options)
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

func (m *MockAPIClient) GetShippingDocumentResultWithOptions(orderSN, packageNumber string, options shopeePkg.ShippingDocumentRequestOptions) (*shopeePkg.GetShippingDocumentResultResponse, error) {
	args := m.Called(orderSN, packageNumber, options)
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

func (m *MockAPIClient) SearchPackageList(packageStatus int, cursor string, pageSize int) (*shopeePkg.SearchPackageListResponse, error) {
	args := m.Called(packageStatus, cursor, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.SearchPackageListResponse), args.Error(1)
}

func (m *MockAPIClient) GetPackageDetail(packageNumbers []string) (*shopeePkg.GetPackageDetailResponse, error) {
	args := m.Called(packageNumbers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shopeePkg.GetPackageDetailResponse), args.Error(1)
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

	// Mock CreateShippingDocument FAILURE with batch_api_all_failed
	mockCreateResp := &shopeePkg.CreateShippingDocumentResponse{
		Error:   "common.batch_api_all_failed",
		Message: "All failed",
	}

	// Package lookup can be empty and should not block flow
	mockClient.On("GetShippingDocumentDataInfo", orderSN, "").Return(&shopeePkg.ShippingDocumentDataInfoResponse{}, nil)
	mockClient.On("GetShippingDocumentParameter", orderSN, "").Return(&shopeePkg.GetShippingDocumentParameterResponse{}, nil)
	mockClient.On("SearchPackageList", mock.AnythingOfType("int"), mock.AnythingOfType("string"), 100).Return(&shopeePkg.SearchPackageListResponse{}, nil)
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
	mockClient.On("CreateShippingDocumentWithOptions", orderSN, pkgNum, mock.Anything).Return(mockCreateResp, fmt.Errorf("shopee API error: %s - %s", mockCreateResp.Error, mockCreateResp.Message))

	// Mock tracking number exists (shipment arranged) so print flow can proceed
	mockClient.On("GetTrackingNumber", orderSN).Return(&shopeePkg.GetTrackingNumberResponse{
		Response: struct {
			TrackingNumber string `json:"tracking_number"`
			Hint           string `json:"hint,omitempty"`
		}{TrackingNumber: "SPX123456789"},
	}, nil)

	// Execute
	result, err := service.GetShippingLabel(ctx, orderSN, pkgNum, docType)

	// Assert
	assert.NoError(t, err) // Should NOT return error
	assert.NotNil(t, result)
	assert.Equal(t, "FAILED", result.Status)
	assert.Equal(t, "shopee API error: logistics.tracking_number_invalid - The tracking number is invalid", result.ErrorMessage)

	mockClient.AssertExpectations(t)
}

func TestGetShippingLabel_UsesRawDownloadErrorWhenTrackingUnavailable(t *testing.T) {
	mockClient := new(MockAPIClient)
	service := &ShippingService{pkgClient: mockClient}

	ctx := context.Background()
	orderSN := "260203Q98DEKHK"
	mockClient.On("GetShippingDocumentDataInfo", orderSN, "").Return(&shopeePkg.ShippingDocumentDataInfoResponse{}, nil)
	mockClient.On("GetShippingDocumentParameter", orderSN, "").Return(&shopeePkg.GetShippingDocumentParameterResponse{}, nil)
	mockClient.On("SearchPackageList", mock.AnythingOfType("int"), mock.AnythingOfType("string"), 100).Return(&shopeePkg.SearchPackageListResponse{}, nil)

	// ensureShipmentReady retries up to 3 times and returns empty tracking payload
	mockClient.On("GetTrackingNumber", orderSN).Return(&shopeePkg.GetTrackingNumberResponse{}, nil).Times(3)

	mockCreateResp := &shopeePkg.CreateShippingDocumentResponse{}
	mockClient.On("CreateShippingDocumentWithOptions", orderSN, "", mock.Anything).Return(mockCreateResp, nil)

	mockResultResp := &shopeePkg.GetShippingDocumentResultResponse{}
	mockResultResp.Response.ResultList = append(mockResultResp.Response.ResultList, struct {
		OrderSN       string `json:"order_sn"`
		PackageNumber string `json:"package_number"`
		Status        string `json:"status"`
		FailError     string `json:"fail_error,omitempty"`
		FailMessage   string `json:"fail_message,omitempty"`
	}{
		OrderSN: orderSN,
		Status:  "READY",
	})
	mockClient.On("GetShippingDocumentResultWithOptions", orderSN, "", mock.Anything).Return(mockResultResp, nil)

	mockDownloadResp := &shopeePkg.DownloadShippingDocumentResponse{
		Error:   "logistics.tracking_number_invalid",
		Message: "The tracking number is invalid",
	}
	mockClient.On("DownloadShippingDocument", orderSN, "", mock.AnythingOfType("string")).Return(mockDownloadResp, fmt.Errorf("logistics.tracking_number_invalid: The tracking number is invalid"))

	result, err := service.GetShippingLabel(ctx, orderSN, "", "THERMAL_AIR_WAYBILL")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "FAILED", result.Status)
	assert.Equal(t, "shopee API error: logistics.tracking_number_invalid - The tracking number is invalid", result.ErrorMessage)

	mockClient.AssertExpectations(t)
}

func TestGetShippingLabel_UsesResolvedPackageNumber(t *testing.T) {
	mockClient := new(MockAPIClient)
	service := &ShippingService{pkgClient: mockClient}

	ctx := context.Background()
	orderSN := "260203Q98DEKHK"
	resolvedPackage := "PKG-123"

	mockClient.On("GetShippingDocumentDataInfo", orderSN, "").Return(&shopeePkg.ShippingDocumentDataInfoResponse{
		Response: struct {
			OrderSN                 string  `json:"order_sn"`
			PackageNumber           string  `json:"package_number"`
			LogisticsChannelID      int     `json:"logistics_channel_id"`
			LogisticsChannelName    string  `json:"logistics_channel_name"`
			FirstMileTrackingNumber string  `json:"first_mile_tracking_number"`
			LastMileTrackingNumber  string  `json:"last_mile_tracking_number"`
			TrackingNumber          string  `json:"tracking_number"`
			ShippingCarrier         string  `json:"shipping_carrier"`
			SenderName              string  `json:"sender_name"`
			SenderPhone             string  `json:"sender_phone"`
			SenderAddress           string  `json:"sender_address"`
			SenderCity              string  `json:"sender_city"`
			SenderDistrict          string  `json:"sender_district"`
			SenderState             string  `json:"sender_state"`
			SenderZipcode           string  `json:"sender_zipcode"`
			SenderCountry           string  `json:"sender_country"`
			RecipientName           string  `json:"recipient_name"`
			RecipientPhone          string  `json:"recipient_phone"`
			RecipientAddress        string  `json:"recipient_address"`
			RecipientCity           string  `json:"recipient_city"`
			RecipientDistrict       string  `json:"recipient_district"`
			RecipientState          string  `json:"recipient_state"`
			RecipientZipcode        string  `json:"recipient_zipcode"`
			RecipientCountry        string  `json:"recipient_country"`
			RecipientSortCode       string  `json:"recipient_sort_code"`
			ServiceDescription      string  `json:"service_description"`
			BuyerCodAmount          float64 `json:"buyer_cod_amount"`
		}{PackageNumber: resolvedPackage},
	}, nil)
	mockClient.On("GetShippingDocumentParameter", orderSN, resolvedPackage).Return(&shopeePkg.GetShippingDocumentParameterResponse{}, nil)
	mockClient.On("GetShippingDocumentParameter", orderSN, "").Return(&shopeePkg.GetShippingDocumentParameterResponse{}, nil)

	mockClient.On("GetTrackingNumber", orderSN).Return(&shopeePkg.GetTrackingNumberResponse{}, nil).Times(3)
	mockClient.On("CreateShippingDocumentWithOptions", orderSN, resolvedPackage, mock.Anything).Return(&shopeePkg.CreateShippingDocumentResponse{}, nil)

	mockResultResp := &shopeePkg.GetShippingDocumentResultResponse{}
	mockResultResp.Response.ResultList = append(mockResultResp.Response.ResultList, struct {
		OrderSN       string `json:"order_sn"`
		PackageNumber string `json:"package_number"`
		Status        string `json:"status"`
		FailError     string `json:"fail_error,omitempty"`
		FailMessage   string `json:"fail_message,omitempty"`
	}{OrderSN: orderSN, PackageNumber: resolvedPackage, Status: "READY"})
	mockClient.On("GetShippingDocumentResultWithOptions", orderSN, resolvedPackage, mock.Anything).Return(mockResultResp, nil)

	mockDownloadResp := &shopeePkg.DownloadShippingDocumentResponse{}
	mockDownloadResp.RawPDF = []byte("pdf")
	mockClient.On("DownloadShippingDocument", orderSN, resolvedPackage, mock.AnythingOfType("string")).Return(mockDownloadResp, nil)

	result, err := service.GetShippingLabel(ctx, orderSN, "", "THERMAL_AIR_WAYBILL")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "SUCCESS", result.Status)
	assert.NotEmpty(t, result.FileData)

	mockClient.AssertExpectations(t)
}

func TestGetShippingLabel_ResolvesPackageFromPackageList(t *testing.T) {
	mockClient := new(MockAPIClient)
	service := &ShippingService{pkgClient: mockClient}

	ctx := context.Background()
	orderSN := "260211EHGAVETU"
	resolvedPackage := "PKG-PROCESSED-1"

	mockClient.On("GetTrackingNumber", orderSN).Return(&shopeePkg.GetTrackingNumberResponse{
		Response: struct {
			TrackingNumber string `json:"tracking_number"`
			Hint           string `json:"hint,omitempty"`
		}{TrackingNumber: "SPXID062814285902"},
	}, nil).Times(1)

	mockClient.On("GetShippingDocumentDataInfo", orderSN, "").Return(&shopeePkg.ShippingDocumentDataInfoResponse{}, nil)
	mockClient.On("GetShippingDocumentParameter", orderSN, resolvedPackage).Return(&shopeePkg.GetShippingDocumentParameterResponse{}, nil)
	mockClient.On("GetShippingDocumentParameter", orderSN, "").Return(&shopeePkg.GetShippingDocumentParameterResponse{}, nil)

	searchResp := &shopeePkg.SearchPackageListResponse{}
	searchResp.Response.PackagesList = append(searchResp.Response.PackagesList, shopeePkg.PackageBasic{
		OrderSN:       orderSN,
		PackageNumber: resolvedPackage,
	})
	mockClient.On("SearchPackageList", 3, "", 100).Return(searchResp, nil).Once()

	mockClient.On("CreateShippingDocumentWithOptions", orderSN, resolvedPackage, mock.Anything).Return(&shopeePkg.CreateShippingDocumentResponse{}, nil)

	resultResp := &shopeePkg.GetShippingDocumentResultResponse{}
	resultResp.Response.ResultList = append(resultResp.Response.ResultList, struct {
		OrderSN       string `json:"order_sn"`
		PackageNumber string `json:"package_number"`
		Status        string `json:"status"`
		FailError     string `json:"fail_error,omitempty"`
		FailMessage   string `json:"fail_message,omitempty"`
	}{OrderSN: orderSN, PackageNumber: resolvedPackage, Status: "READY"})
	mockClient.On("GetShippingDocumentResultWithOptions", orderSN, resolvedPackage, mock.Anything).Return(resultResp, nil).Once()

	downloadResp := &shopeePkg.DownloadShippingDocumentResponse{}
	downloadResp.RawPDF = []byte("pdf")
	mockClient.On("DownloadShippingDocument", orderSN, resolvedPackage, mock.AnythingOfType("string")).Return(downloadResp, nil).Once()

	result, err := service.GetShippingLabel(ctx, orderSN, "", "THERMAL_AIR_WAYBILL")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "SUCCESS", result.Status)
	assert.NotEmpty(t, result.FileData)

	mockClient.AssertExpectations(t)
}

func TestGetShippingLabel_RetriesCreateWithoutPackageOnTrackingInvalid(t *testing.T) {
	mockClient := new(MockAPIClient)
	service := &ShippingService{pkgClient: mockClient}

	ctx := context.Background()
	orderSN := "260211EHGAVETU"
	resolvedPackage := "OFG224499838282608"

	mockClient.On("GetTrackingNumber", orderSN).Return(&shopeePkg.GetTrackingNumberResponse{
		Response: struct {
			TrackingNumber string `json:"tracking_number"`
			Hint           string `json:"hint,omitempty"`
		}{TrackingNumber: "SPXID062814285902"},
	}, nil).Once()

	mockClient.On("GetShippingDocumentDataInfo", orderSN, "").Return(&shopeePkg.ShippingDocumentDataInfoResponse{}, nil).Once()
	mockClient.On("GetShippingDocumentParameter", orderSN, resolvedPackage).Return(&shopeePkg.GetShippingDocumentParameterResponse{}, nil).Once()
	mockClient.On("GetShippingDocumentParameter", orderSN, "").Return(&shopeePkg.GetShippingDocumentParameterResponse{}, nil).Once()

	searchResp := &shopeePkg.SearchPackageListResponse{}
	searchResp.Response.PackagesList = append(searchResp.Response.PackagesList, shopeePkg.PackageBasic{
		OrderSN:       orderSN,
		PackageNumber: resolvedPackage,
	})
	mockClient.On("SearchPackageList", 3, "", 100).Return(searchResp, nil).Once()

	failCreate := &shopeePkg.CreateShippingDocumentResponse{}
	failCreate.Response.ResultList = append(failCreate.Response.ResultList, struct {
		OrderSN       string `json:"order_sn"`
		PackageNumber string `json:"package_number"`
		Status        string `json:"status"`
		FailError     string `json:"fail_error,omitempty"`
		FailMessage   string `json:"fail_message,omitempty"`
	}{
		OrderSN:       orderSN,
		PackageNumber: resolvedPackage,
		Status:        "FAILED",
		FailError:     "logistics.tracking_number_invalid",
		FailMessage:   "The tracking number is invalid. Please check the tracking number.",
	})
	mockClient.On("CreateShippingDocumentWithOptions", orderSN, resolvedPackage, mock.Anything).Return(failCreate, nil).Once()
	mockClient.On("CreateShippingDocumentWithOptions", orderSN, "", mock.Anything).Return(&shopeePkg.CreateShippingDocumentResponse{}, nil).Once()

	resultResp := &shopeePkg.GetShippingDocumentResultResponse{}
	resultResp.Response.ResultList = append(resultResp.Response.ResultList, struct {
		OrderSN       string `json:"order_sn"`
		PackageNumber string `json:"package_number"`
		Status        string `json:"status"`
		FailError     string `json:"fail_error,omitempty"`
		FailMessage   string `json:"fail_message,omitempty"`
	}{OrderSN: orderSN, Status: "READY"})
	mockClient.On("GetShippingDocumentResultWithOptions", orderSN, "", mock.Anything).Return(resultResp, nil).Once()

	downloadResp := &shopeePkg.DownloadShippingDocumentResponse{}
	downloadResp.RawPDF = []byte("pdf")
	mockClient.On("DownloadShippingDocument", orderSN, "", mock.AnythingOfType("string")).Return(downloadResp, nil).Once()

	result, err := service.GetShippingLabel(ctx, orderSN, "", "THERMAL_AIR_WAYBILL")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "SUCCESS", result.Status)
	assert.NotEmpty(t, result.FileData)

	mockClient.AssertExpectations(t)
}
