package lazada

import (
	"testing"
)

// TestIsMaskedPII_HappyMaskedCases confirms every canonical Lazada masking
// pattern is detected. Live examples from Lazada Open Platform announcement
// docId=145548 (2026-07-01 rollout) plus the industry-standard variants we
// have seen from Shopee/Shopify.
func TestIsMaskedPII_HappyMaskedCases(t *testing.T) {
	cases := map[string]string{
		"asterisk_head_tail":       "J***s",           // first + last char preserved
		"asterisk_only_leading":    "***s",            // trailing char preserved
		"asterisk_only_trailing":   "J***",            // leading char preserved
		"asterisk_phone":           "+62812***56",     // country code + last 2
		"asterisk_phone_local":     "0812***56",       // no country code
		"asterisk_address_line":    "Jl. K***, RT 0*", // multi-token line
		"asterisk_all_middle_dots": "A••••B",          // bullet used by Lazada UI
		"asterisk_full_mask":       "*****",           // fully masked field
	}
	for name, val := range cases {
		t.Run(name, func(t *testing.T) {
			if !IsMaskedPII(val) {
				t.Errorf("IsMaskedPII(%q) = false, want true", val)
			}
		})
	}
}

// TestIsMaskedPII_UnmaskedCases guards against false positives on real
// buyer data that happens to contain a `*` (rare but legal in names/addresses).
func TestIsMaskedPII_UnmaskedCases(t *testing.T) {
	cases := map[string]string{
		"plain_name":       "Jane Smith",
		"plain_phone":      "+6281234567890",
		"plain_address":    "Jl. Kemang Raya No. 12, Jakarta Selatan",
		"single_asterisk":  "5* rating in review name", // one lone asterisk in prose
		"empty":            "",
		"whitespace_only":  "   ",
		"unicode_name":     "Zoë O'Brien",
		"comma_addr":       "Level 3, Block B, Unit 42",
	}
	for name, val := range cases {
		t.Run(name, func(t *testing.T) {
			if IsMaskedPII(val) {
				t.Errorf("IsMaskedPII(%q) = true, want false", val)
			}
		})
	}
}

// TestDetectMaskedFields_OrderListResponse asserts that a Lazada
// GetOrders response with masked customer_first_name is flagged.
// Uses the real payload shape from pkg/lazada.OrderListResponse.
func TestDetectMaskedFields_OrderListResponse(t *testing.T) {
	resp := &OrderListResponse{}
	resp.Data.Count = 2
	resp.Data.Orders = []struct {
		OrderID          string  `json:"order_id"`
		OrderNumber      string  `json:"order_number"`
		Status           string  `json:"status"`
		Price            float64 `json:"price"`
		CustomerName     string  `json:"customer_first_name"`
		PromisedShipDate string  `json:"promised_shipping_times"`
		ShippingType     string  `json:"delivery_info"`
		CreatedAt        string  `json:"created_at"`
		UpdatedAt        string  `json:"updated_at"`
	}{
		{OrderID: "1001", CustomerName: "J***s"},          // masked
		{OrderID: "1002", CustomerName: "Jane Smith"},     // unmasked (DBS)
	}
	report := DetectMaskedFieldsInOrderList(resp)
	if len(report.PerOrder) != 2 {
		t.Fatalf("PerOrder length = %d, want 2", len(report.PerOrder))
	}
	if !report.PerOrder["1001"].AnyMasked {
		t.Errorf("Order 1001 (J***s) not flagged as masked")
	}
	if report.PerOrder["1002"].AnyMasked {
		t.Errorf("Order 1002 (Jane Smith) wrongly flagged as masked")
	}
	if !report.HasAnyMasked() {
		t.Errorf("HasAnyMasked() = false, want true (order 1001 is masked)")
	}
	// Assert the masked-field list contains the exact field name that Lazada
	// returned masked. Consumers use this to decide UI treatment per field.
	if got := report.PerOrder["1001"].MaskedFields; len(got) != 1 || got[0] != "customer_first_name" {
		t.Errorf("Order 1001 MaskedFields = %v, want [customer_first_name]", got)
	}
	// Fallback: an entirely empty response is not "masked" — it's empty.
	empty := &OrderListResponse{}
	emptyReport := DetectMaskedFieldsInOrderList(empty)
	if emptyReport.HasAnyMasked() {
		t.Errorf("Empty response reported HasAnyMasked() = true, want false")
	}
}

// TestDetectMaskedFields_NilInput is a defensive negative-path test:
// passing nil must not panic and must report no masking.
func TestDetectMaskedFields_NilInput(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DetectMaskedFieldsInOrderList(nil) panicked: %v", r)
		}
	}()
	report := DetectMaskedFieldsInOrderList(nil)
	if report.HasAnyMasked() {
		t.Errorf("nil input: HasAnyMasked() = true, want false")
	}
	if len(report.PerOrder) != 0 {
		t.Errorf("nil input: PerOrder length = %d, want 0", len(report.PerOrder))
	}
}
