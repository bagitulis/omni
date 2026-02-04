package tiktok

import (
	"context"
	"testing"

	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTikTokClient is a mock of the TikTokClient interface
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

func TestTikTok_ArrangeShipment_RealCase(t *testing.T) {
	// Setup
	mockClient := new(MockTikTokClient)

	// Create service with mock factory
	service := NewShippingServiceWithFactory("test_path", func(ctx context.Context, tenantID string) (TikTokClient, error) {
		return mockClient, nil
	})

	ctx := context.Background()
	tenantID := "1"
	packageID := "582445147764327664" // Real Package ID from user

	// Mock Request
	req := &tiktokPkg.ShipPackageRequest{
		HandoverMethod: "PICKUP",
	}

	// Mock Response
	mockResp := &tiktokPkg.ShipPackageResponse{
		BaseResponse: tiktokPkg.BaseResponse{
			Code:    0,
			Message: "Success",
		},
	}
	mockResp.Data.PackageID = packageID

	// Expectation
	mockClient.On("ArrangeShipment", packageID, req).Return(mockResp, nil)

	// Execute
	resp, err := service.ArrangeShipment(ctx, tenantID, packageID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, packageID, resp.Data.PackageID)

	mockClient.AssertExpectations(t)
}

func TestTikTok_GetShippingLabel_RealCase(t *testing.T) {
	// Setup
	mockClient := new(MockTikTokClient)

	// Create service with mock factory
	service := NewShippingServiceWithFactory("test_path", func(ctx context.Context, tenantID string) (TikTokClient, error) {
		return mockClient, nil
	})

	ctx := context.Background()
	tenantID := "1"
	packageID := "582445147764327664" // Real Package ID from user
	docType := "SHIPPING_LABEL"

	// Mock Response
	expectedURL := "https://tiktok.com/shipping/label.pdf"

	// Expectation
	mockClient.On("GetShippingDocument", packageID, docType).Return(expectedURL, nil)

	// Execute
	url, err := service.GetShippingLabel(ctx, tenantID, packageID, docType)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, url)

	mockClient.AssertExpectations(t)
}
