package tiktok

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	serviceTiktok "github.com/omni/backend/internal/services/tiktok"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTikTokClient defines the mock for TikTok API Client
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

func TestTikTok_Shipping_RealRoute_Simulation(t *testing.T) {
	// 1. Setup Gin
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Mock Middleware
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "123") // Simulate Auth middleware
		c.Next()
	})

	// 2. Setup Mock Client
	mockClient := new(MockTikTokClient)

	// Real Package ID from User
	realPackageID := "582445147764327664"

	// --- Scenario 1: Arrange Shipment ---

	// Mock ArrangeShipment
	mockArrangeResp := &tiktokPkg.ShipPackageResponse{
		BaseResponse: tiktokPkg.BaseResponse{Code: 0, Message: "Success"},
	}
	mockArrangeResp.Data.PackageID = realPackageID

	// Expectation: Matches the body sent by verify_tiktok_shipping.sh
	mockClient.On("ArrangeShipment", realPackageID, mock.AnythingOfType("*tiktok.ShipPackageRequest")).Return(mockArrangeResp, nil)

	// Setup Service with Mock Factory
	factory := func(ctx context.Context, tenantID string) (serviceTiktok.TikTokClient, error) {
		return mockClient, nil
	}

	service := serviceTiktok.NewShippingServiceWithFactory("base_path", factory)

	// Create Handler and inject service (now possible via public Service field)
	handler := &ShippingHandler{
		Service: service,
	}

	// Register Route
	r.POST("/api/tiktok/shipping/arrange", handler.ArrangeShipment)
	r.GET("/api/tiktok/shipping/document/:packageId", handler.GetShippingDocument)

	// Perform Request 1: Arrange Shipment
	reqBody := `{"package_id": "582445147764327664", "handover_method": "PICKUP"}`
	req1, _ := http.NewRequest("POST", "/api/tiktok/shipping/arrange", strings.NewReader(reqBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	// Assertions 1
	assert.Equal(t, http.StatusOK, w1.Code)

	var resp1 map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &resp1)
	assert.NoError(t, err)
	assert.Equal(t, true, resp1["success"])

	data1 := resp1["data"].(map[string]interface{})
	data1Inner := data1["data"].(map[string]interface{})
	assert.Equal(t, realPackageID, data1Inner["package_id"])

	// --- Scenario 2: Get Shipping Document ---

	// Mock GetShippingDocument
	mockClient.On("GetShippingDocument", realPackageID, "SHIPPING_LABEL").Return("https://tiktok.com/label.pdf", nil)

	// Perform Request 2: Get Document
	req2, _ := http.NewRequest("GET", "/api/tiktok/shipping/document/"+realPackageID+"?document_type=SHIPPING_LABEL", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	// Assertions 2
	assert.Equal(t, http.StatusOK, w2.Code)

	var resp2 map[string]interface{}
	err2 := json.Unmarshal(w2.Body.Bytes(), &resp2)
	assert.NoError(t, err2)
	assert.Equal(t, true, resp2["success"])

	data2 := resp2["data"].(map[string]interface{})
	assert.Equal(t, "https://tiktok.com/label.pdf", data2["doc_url"])

	// Verify All Mock Expectations
	mockClient.AssertExpectations(t)
}
