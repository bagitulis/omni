package request_test

import (
	"encoding/json"
	"testing"

	"github.com/omni/backend/internal/dto/request"
)

func TestShipOrderRequest_Fields(t *testing.T) {
	r := request.ShipOrderRequest{
		OrderSN:         "SN123",
		TrackingNumber:  "TRK001",
		ShippingCarrier: "JNE",
		AddressID:       42,
		PickupTimeID:    "PT-001",
		BranchID:        7,
	}
	if r.OrderSN != "SN123" {
		t.Errorf("OrderSN = %q, want SN123", r.OrderSN)
	}
	if r.AddressID != 42 {
		t.Errorf("AddressID = %d, want 42", r.AddressID)
	}
}

func TestShipOrderRequest_JSONTags(t *testing.T) {
	r := request.ShipOrderRequest{OrderSN: "SN123", ShippingCarrier: "JNE"}
	b, _ := json.Marshal(r)
	var m map[string]interface{}
	json.Unmarshal(b, &m) //nolint:errcheck
	for _, key := range []string{"order_sn", "shipping_carrier"} {
		if _, ok := m[key]; !ok {
			t.Errorf("expected JSON key %q", key)
		}
	}
}

func TestBulkPrintLabelsRequest_Fields(t *testing.T) {
	r := request.BulkPrintLabelsRequest{
		OrderSNs:           []string{"SN1"},
		Platform:           "tiktok",
		IncludeProducts:    true,
		TikTokDocumentType: "SHIPPING_LABEL",
	}
	if r.Platform != "tiktok" {
		t.Errorf("Platform = %q, want tiktok", r.Platform)
	}
	if !r.IncludeProducts {
		t.Error("IncludeProducts should be true")
	}
}

func TestTiktokShipOrderRequest_Fields(t *testing.T) {
	r := request.TiktokShipOrderRequest{
		OrderID:          "TT789",
		TrackingNumber:   "TRK-TT",
		ShippingProvider: "J&T",
	}
	if r.OrderID != "TT789" {
		t.Errorf("OrderID = %q, want TT789", r.OrderID)
	}
}

func TestLazadaShipOrderRequest_Fields(t *testing.T) {
	r := request.LazadaShipOrderRequest{
		OrderID:        "LZ123",
		TrackingNumber: "TRK-LZ",
		ShippingType:   "dropship",
	}
	if r.OrderID != "LZ123" {
		t.Errorf("OrderID = %q, want LZ123", r.OrderID)
	}
}
