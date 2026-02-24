package request_test

import (
	"encoding/json"
	"testing"

	"github.com/omni/backend/internal/dto/request"
)

// ---------------------------------------------------------------------------
// Shopee request structs
// ---------------------------------------------------------------------------

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
	if r.TrackingNumber != "TRK001" {
		t.Errorf("TrackingNumber = %q, want TRK001", r.TrackingNumber)
	}
	if r.ShippingCarrier != "JNE" {
		t.Errorf("ShippingCarrier = %q, want JNE", r.ShippingCarrier)
	}
	if r.AddressID != 42 {
		t.Errorf("AddressID = %d, want 42", r.AddressID)
	}
	if r.PickupTimeID != "PT-001" {
		t.Errorf("PickupTimeID = %q, want PT-001", r.PickupTimeID)
	}
	if r.BranchID != 7 {
		t.Errorf("BranchID = %d, want 7", r.BranchID)
	}
}

func TestShipOrderRequest_JSONTags(t *testing.T) {
	r := request.ShipOrderRequest{
		OrderSN:         "SN123",
		TrackingNumber:  "TRK001",
		ShippingCarrier: "JNE",
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	for _, key := range []string{"order_sn", "shipping_carrier"} {
		if _, ok := m[key]; !ok {
			t.Errorf("expected JSON key %q", key)
		}
	}
}

func TestCancelOrderRequest_Fields(t *testing.T) {
	r := request.CancelOrderRequest{
		OrderSN:      "SN456",
		CancelReason: "customer request",
	}
	if r.OrderSN != "SN456" {
		t.Errorf("OrderSN = %q, want SN456", r.OrderSN)
	}
	if r.CancelReason != "customer request" {
		t.Errorf("CancelReason = %q, want customer request", r.CancelReason)
	}
}

func TestCancelOrderRequest_JSONTags(t *testing.T) {
	r := request.CancelOrderRequest{OrderSN: "SN1", CancelReason: "reason"}
	b, _ := json.Marshal(r)
	var m map[string]interface{}
	json.Unmarshal(b, &m) //nolint:errcheck
	for _, key := range []string{"order_sn", "cancel_reason"} {
		if _, ok := m[key]; !ok {
			t.Errorf("expected JSON key %q", key)
		}
	}
}

func TestCreateProductRequest_Fields(t *testing.T) {
	r := request.CreateProductRequest{
		Name:          "Widget",
		Description:   "A fine widget",
		CategoryID:    100,
		OriginalPrice: 29.99,
		Stock:         50,
		Images:        []string{"img1.jpg"},
		SKU:           "SKU-001",
		Weight:        0.5,
	}
	if r.Name != "Widget" {
		t.Errorf("Name = %q, want Widget", r.Name)
	}
	if r.CategoryID != 100 {
		t.Errorf("CategoryID = %d, want 100", r.CategoryID)
	}
	if r.OriginalPrice != 29.99 {
		t.Errorf("OriginalPrice = %f, want 29.99", r.OriginalPrice)
	}
	if r.Stock != 50 {
		t.Errorf("Stock = %d, want 50", r.Stock)
	}
	if len(r.Images) != 1 {
		t.Errorf("Images len = %d, want 1", len(r.Images))
	}
}

func TestUpdateProductRequest_Fields(t *testing.T) {
	r := request.UpdateProductRequest{
		Name:          "Updated Widget",
		OriginalPrice: 34.99,
		Stock:         60,
		Status:        "active",
	}
	if r.Name != "Updated Widget" {
		t.Errorf("Name = %q, want Updated Widget", r.Name)
	}
	if r.Status != "active" {
		t.Errorf("Status = %q, want active", r.Status)
	}
}

func TestUpdateStockRequest_Fields(t *testing.T) {
	r := request.UpdateStockRequest{Stock: 100}
	if r.Stock != 100 {
		t.Errorf("Stock = %d, want 100", r.Stock)
	}
}

func TestUpdateStockRequest_JSONTag(t *testing.T) {
	r := request.UpdateStockRequest{Stock: 5}
	b, _ := json.Marshal(r)
	var m map[string]interface{}
	json.Unmarshal(b, &m) //nolint:errcheck
	if _, ok := m["stock"]; !ok {
		t.Error("expected JSON key stock")
	}
}

func TestUpdatePriceRequest_Fields(t *testing.T) {
	r := request.UpdatePriceRequest{OriginalPrice: 99.99}
	if r.OriginalPrice != 99.99 {
		t.Errorf("OriginalPrice = %f, want 99.99", r.OriginalPrice)
	}
}

func TestUpdatePriceRequest_JSONTag(t *testing.T) {
	r := request.UpdatePriceRequest{OriginalPrice: 10.0}
	b, _ := json.Marshal(r)
	var m map[string]interface{}
	json.Unmarshal(b, &m) //nolint:errcheck
	if _, ok := m["original_price"]; !ok {
		t.Error("expected JSON key original_price")
	}
}

func TestBulkShipRequest_Fields(t *testing.T) {
	r := request.BulkShipRequest{
		OrderSNs: []string{"SN1", "SN2", "SN3"},
		Platform: "shopee",
	}
	if len(r.OrderSNs) != 3 {
		t.Errorf("OrderSNs len = %d, want 3", len(r.OrderSNs))
	}
	if r.Platform != "shopee" {
		t.Errorf("Platform = %q, want shopee", r.Platform)
	}
}

func TestBulkShipRequest_JSONTags(t *testing.T) {
	r := request.BulkShipRequest{OrderSNs: []string{"SN1"}, Platform: "shopee"}
	b, _ := json.Marshal(r)
	var m map[string]interface{}
	json.Unmarshal(b, &m) //nolint:errcheck
	for _, key := range []string{"order_sns", "platform"} {
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
	if r.TikTokDocumentType != "SHIPPING_LABEL" {
		t.Errorf("TikTokDocumentType = %q, want SHIPPING_LABEL", r.TikTokDocumentType)
	}
}

// ---------------------------------------------------------------------------
// Lazada request structs
// ---------------------------------------------------------------------------

func TestLazadaShipOrderRequest_Fields(t *testing.T) {
	r := request.LazadaShipOrderRequest{
		OrderID:        "LZ123",
		TrackingNumber: "TRK-LZ",
		ShippingType:   "dropship",
	}
	if r.OrderID != "LZ123" {
		t.Errorf("OrderID = %q, want LZ123", r.OrderID)
	}
	if r.TrackingNumber != "TRK-LZ" {
		t.Errorf("TrackingNumber = %q, want TRK-LZ", r.TrackingNumber)
	}
}

func TestLazadaShipOrderRequest_JSONTags(t *testing.T) {
	r := request.LazadaShipOrderRequest{OrderID: "O1", TrackingNumber: "T1"}
	b, _ := json.Marshal(r)
	var m map[string]interface{}
	json.Unmarshal(b, &m) //nolint:errcheck
	for _, key := range []string{"order_id", "tracking_number"} {
		if _, ok := m[key]; !ok {
			t.Errorf("expected JSON key %q", key)
		}
	}
}

func TestLazadaCancelOrderRequest_Fields(t *testing.T) {
	r := request.LazadaCancelOrderRequest{
		OrderID:      "LZ456",
		CancelReason: "out of stock",
	}
	if r.OrderID != "LZ456" {
		t.Errorf("OrderID = %q, want LZ456", r.OrderID)
	}
}

func TestLazadaCreateProductRequest_Fields(t *testing.T) {
	r := request.LazadaCreateProductRequest{
		Name:  "Test Product",
		Price: 49.99,
		Stock: 20,
		SKU:   "LZ-SKU",
	}
	if r.Price != 49.99 {
		t.Errorf("Price = %f, want 49.99", r.Price)
	}
	if r.Stock != 20 {
		t.Errorf("Stock = %d, want 20", r.Stock)
	}
}

func TestLazadaUpdateProductRequest_Fields(t *testing.T) {
	r := request.LazadaUpdateProductRequest{
		Name:   "Updated",
		Price:  55.0,
		Stock:  30,
		Status: "active",
	}
	if r.Status != "active" {
		t.Errorf("Status = %q, want active", r.Status)
	}
}

// ---------------------------------------------------------------------------
// TikTok request structs
// ---------------------------------------------------------------------------

func TestTiktokShipOrderRequest_Fields(t *testing.T) {
	r := request.TiktokShipOrderRequest{
		OrderID:          "TT789",
		TrackingNumber:   "TRK-TT",
		ShippingProvider: "J&T",
	}
	if r.OrderID != "TT789" {
		t.Errorf("OrderID = %q, want TT789", r.OrderID)
	}
	if r.ShippingProvider != "J&T" {
		t.Errorf("ShippingProvider = %q, want J&T", r.ShippingProvider)
	}
}

func TestTiktokShipOrderRequest_JSONTags(t *testing.T) {
	r := request.TiktokShipOrderRequest{OrderID: "O1", TrackingNumber: "T1"}
	b, _ := json.Marshal(r)
	var m map[string]interface{}
	json.Unmarshal(b, &m) //nolint:errcheck
	for _, key := range []string{"order_id", "tracking_number"} {
		if _, ok := m[key]; !ok {
			t.Errorf("expected JSON key %q", key)
		}
	}
}

func TestTiktokCancelOrderRequest_Fields(t *testing.T) {
	r := request.TiktokCancelOrderRequest{
		OrderID:      "TT999",
		CancelReason: "changed mind",
	}
	if r.CancelReason != "changed mind" {
		t.Errorf("CancelReason = %q, want changed mind", r.CancelReason)
	}
}

func TestTiktokCreateProductRequest_Fields(t *testing.T) {
	r := request.TiktokCreateProductRequest{
		Name:        "TT Product",
		CategoryID:  "CAT-001",
		Price:       15.99,
		Stock:       100,
		Description: "desc",
		SKU:         "TT-SKU",
	}
	if r.CategoryID != "CAT-001" {
		t.Errorf("CategoryID = %q, want CAT-001", r.CategoryID)
	}
	if r.Price != 15.99 {
		t.Errorf("Price = %f, want 15.99", r.Price)
	}
}

func TestTiktokUpdateProductRequest_Fields(t *testing.T) {
	r := request.TiktokUpdateProductRequest{
		Name:   "Updated TT",
		Price:  20.0,
		Stock:  200,
		Status: "inactive",
	}
	if r.Status != "inactive" {
		t.Errorf("Status = %q, want inactive", r.Status)
	}
}

// ---------------------------------------------------------------------------
// JSON round-trip: verify snake_case keys survive marshal/unmarshal
// ---------------------------------------------------------------------------

func TestCreateProductRequest_JSONRoundTrip(t *testing.T) {
	original := request.CreateProductRequest{
		Name:          "Round Trip",
		Description:   "test",
		CategoryID:    99,
		OriginalPrice: 12.34,
		Stock:         10,
		SKU:           "RT-SKU",
		Weight:        1.5,
	}
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded request.CreateProductRequest
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.OriginalPrice != original.OriginalPrice {
		t.Errorf("OriginalPrice = %f, want %f", decoded.OriginalPrice, original.OriginalPrice)
	}
	if decoded.CategoryID != original.CategoryID {
		t.Errorf("CategoryID = %d, want %d", decoded.CategoryID, original.CategoryID)
	}
}
