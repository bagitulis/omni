package shopee

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/services"
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// shippingClient interface for pkg/shopee.Client methods used by shipping service
type shippingClient interface {
	GetShippingParameter(orderSN string) (*shopeePkg.GetShippingParameterResponse, error)
	ShipOrder(req shopeePkg.ShipOrderRequest) (*shopeePkg.ShipOrderResponse, error)
	GetTrackingNumber(orderSN string) (*shopeePkg.GetTrackingNumberResponse, error)
	CreateShippingDocument(orderSN, packageNumber string) (*shopeePkg.CreateShippingDocumentResponse, error)
	DownloadShippingDocument(orderSN, packageNumber, documentType string) (*shopeePkg.DownloadShippingDocumentResponse, error)
}

// ShippingService handles Shopee shipping operations
type ShippingService struct {
	apiClient   APIClient
	pkgClient   shippingClient // Direct reference to pkg/shopee.Client for shipping operations
	tenantID    string
	credService *services.CredentialService
	dbPath      string
}

// NewShippingService creates a new shipping service with API client
func NewShippingService(apiClient APIClient, tenantID string) *ShippingService {
	// Type assert to get the underlying pkg/shopee.Client for shipping operations
	pkgClient, _ := apiClient.(shippingClient)
	return &ShippingService{
		apiClient: apiClient,
		pkgClient: pkgClient,
		tenantID:  tenantID,
	}
}

// NewShippingServiceWithCreds creates a shipping service with credential support
func NewShippingServiceWithCreds(tenantID, dbPath string) *ShippingService {
	return &ShippingService{
		tenantID:    tenantID,
		dbPath:      dbPath,
		credService: services.NewCredentialService(dbPath),
	}
}

// ShippingOption represents a shipping option
type ShippingOption struct {
	LogisticID        int64   `json:"logistic_id"`
	LogisticName      string  `json:"logistic_name"`
	Enabled           bool    `json:"enabled"`
	ShippingFeeType   string  `json:"shipping_fee_type"`
	EstimatedCost     float64 `json:"estimated_cost"`
	EstimatedDays     int     `json:"estimated_days"`
	IsFreeShipping    bool    `json:"is_free_shipping"`
	HasCOD            bool    `json:"has_cod"`
	TrackingAvailable bool    `json:"tracking_available"`
}

// ShipmentInfo represents shipment information
type ShipmentInfo struct {
	OrderSN          string `json:"order_sn"`
	PackageNumber    string `json:"package_number"`
	LogisticID       int64  `json:"logistic_id"`
	LogisticName     string `json:"logistic_name"`
	TrackingNumber   string `json:"tracking_number"`
	ShippingStatus   string `json:"shipping_status"`
	PickupDoneTime   int64  `json:"pickup_done_time,omitempty"`
	DeliveryDoneTime int64  `json:"delivery_done_time,omitempty"`
}

// TrackingInfo represents tracking information
type TrackingInfo struct {
	TrackingNumber string         `json:"tracking_number"`
	LogisticName   string         `json:"logistic_name"`
	Status         string         `json:"status"`
	History        []TrackingStep `json:"history"`
}

// TrackingStep represents a tracking step
type TrackingStep struct {
	Time        int64  `json:"time"`
	Description string `json:"description"`
	Location    string `json:"location,omitempty"`
}

// ArrangeShipmentRequest represents request to arrange shipment
type ArrangeShipmentRequest struct {
	OrderSN    string       `json:"order_sn"`
	PackageNum string       `json:"package_number,omitempty"`
	Pickup     *PickupInfo  `json:"pickup,omitempty"`
	DropOff    *DropOffInfo `json:"dropoff,omitempty"`
	TrackingNo string       `json:"tracking_number,omitempty"`
}

// PickupInfo represents pickup information
type PickupInfo struct {
	AddressID    int64  `json:"address_id"`
	PickupTimeID string `json:"pickup_time_id,omitempty"`
}

// DropOffInfo represents drop-off information
type DropOffInfo struct {
	BranchID int64 `json:"branch_id"`
}

// getClient returns a Shopee client - either the passed one or creates from credentials
func (s *ShippingService) getClient() (shippingClient, error) {
	// If we already have a pkgClient (from NewShippingService), use it
	if s.pkgClient != nil {
		return s.pkgClient, nil
	}

	// Otherwise, create from credentials
	if s.credService == nil {
		return nil, fmt.Errorf("credential service not initialized")
	}

	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "shopee")
	if err != nil {
		return nil, err
	}

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)
	return client, nil
}

type ShippingOptions struct {
	Pickup  []shopeePkg.PickupAddressInfo `json:"pickup"`
	Dropoff []shopeePkg.BranchInfo        `json:"dropoff"`
}

// GetShippingOptions gets available shipping options for an order
func (s *ShippingService) GetShippingOptions(ctx context.Context, orderSN string) (*ShippingOptions, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}

	result, err := client.GetShippingParameter(orderSN)
	if err != nil {
		return nil, fmt.Errorf("get shipping parameter: %w", err)
	}

	return &ShippingOptions{
		Pickup:  result.Response.InfoNeeded.Pickup,
		Dropoff: result.Response.InfoNeeded.Dropoff,
	}, nil
}

// ArrangeShipment arranges shipment for an order
func (s *ShippingService) ArrangeShipment(ctx context.Context, req ArrangeShipmentRequest) (*ShipmentInfo, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}

	shipReq := shopeePkg.ShipOrderRequest{
		OrderSN:       req.OrderSN,
		PackageNumber: req.PackageNum,
	}

	if req.TrackingNo != "" {
		shipReq.NonIntegrated = &shopeePkg.NonIntegratedInfo{
			TrackingNumber: req.TrackingNo,
		}
	} else {
		if req.Pickup != nil {
			shipReq.Pickup = &shopeePkg.PickupInfo{
				AddressID:    req.Pickup.AddressID,
				PickupTimeID: req.Pickup.PickupTimeID,
			}
		}

		if req.DropOff != nil {
			shipReq.Dropoff = &shopeePkg.DropoffInfo{
				BranchID: req.DropOff.BranchID,
			}
		}
	}

	_, err = client.ShipOrder(shipReq)
	if err != nil {
		return nil, fmt.Errorf("ship order: %w", err)
	}

	return &ShipmentInfo{
		OrderSN:        req.OrderSN,
		ShippingStatus: "SHIPPED",
	}, nil
}

// GetTrackingInfo gets tracking information for an order
func (s *ShippingService) GetTrackingInfo(ctx context.Context, orderSN string) (*TrackingInfo, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}

	result, err := client.GetTrackingNumber(orderSN)
	if err != nil {
		return nil, fmt.Errorf("get tracking number: %w", err)
	}

	return &TrackingInfo{
		TrackingNumber: result.Response.TrackingNumber,
		Status:         "SHIPPED",
		History:        []TrackingStep{},
	}, nil
}

// GetShipmentInfo gets shipment info for an order
func (s *ShippingService) GetShipmentInfo(ctx context.Context, orderSN string) (*ShipmentInfo, error) {
	trackingInfo, err := s.GetTrackingInfo(ctx, orderSN)
	if err != nil {
		return nil, err
	}

	return &ShipmentInfo{
		OrderSN:        orderSN,
		TrackingNumber: trackingInfo.TrackingNumber,
		ShippingStatus: trackingInfo.Status,
	}, nil
}

// CalculateShippingFee calculates shipping fee (not directly available in Shopee API v2)
func (s *ShippingService) CalculateShippingFee(ctx context.Context, orderSN string, logisticID int64) (float64, error) {
	// Shopee doesn't expose shipping fee calculation directly
	// This would need to be retrieved from order details
	return 0, nil
}

// ShippingLabelResult represents the shipping label download result
type ShippingLabelResult struct {
	OrderSN      string `json:"order_sn"`
	Status       string `json:"status"`
	FileData     string `json:"file_data,omitempty"`     // Base64 encoded PDF
	ErrorMessage string `json:"error_message,omitempty"` // Error if failed
}

// GetShippingLabel downloads shipping document (waybill/label) for an order
func (s *ShippingService) GetShippingLabel(ctx context.Context, orderSN, packageNumber, documentType string) (*ShippingLabelResult, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}

	// First, create the shipping document
	createResult, err := client.CreateShippingDocument(orderSN, packageNumber)
	if err != nil {
		return nil, fmt.Errorf("create shipping document: %w", err)
	}

	// Check create result status
	if len(createResult.Response.ResultList) > 0 {
		status := createResult.Response.ResultList[0].Status
		if status == "FAILED" {
			return &ShippingLabelResult{
				OrderSN:      orderSN,
				Status:       "FAILED",
				ErrorMessage: createResult.Response.ResultList[0].FailMessage,
			}, nil
		}
	}

	// Download the shipping document
	downloadResult, err := client.DownloadShippingDocument(orderSN, packageNumber, documentType)
	if err != nil {
		return nil, fmt.Errorf("download shipping document: %w", err)
	}

	if len(downloadResult.Response.ResultList) == 0 {
		return &ShippingLabelResult{
			OrderSN:      orderSN,
			Status:       "FAILED",
			ErrorMessage: "No shipping document available",
		}, nil
	}

	docResult := downloadResult.Response.ResultList[0]
	return &ShippingLabelResult{
		OrderSN:      orderSN,
		Status:       docResult.Status,
		FileData:     docResult.ShippingDocFile,
		ErrorMessage: docResult.FailMessage,
	}, nil
}
