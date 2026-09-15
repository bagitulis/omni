package lazada

import "strings"

// Package lazada — PII masking detection.
//
// Background: Lazada Open Platform began masking buyer PII in order-list /
// order-detail responses on 2026-07-01 (announcement docId=145548). Field
// values arrive as "J***s", "+62812***56", etc. Sellers can apply for a
// Delivery-by-Seller (DBS) exception; DBS orders remain unmasked.
//
// Consumers of the Lazada SDK must be able to (a) detect that a payload has
// been masked, (b) surface which fields, per order, are masked, so the UI can
// render an "Apply for DBS unmasking" hint instead of trying to parse the
// masked string.
//
// This file does NOT unmask anything and does NOT transform the raw payload —
// downstream code stores masked strings verbatim.

// IsMaskedPII reports whether a single field value looks like a Lazada masking
// pattern. Currently: any value containing two or more consecutive `*` or `•`
// characters. Single stray asterisks (common in prose, e.g. "5* rating") are
// NOT flagged, to avoid false positives on real buyer data.
func IsMaskedPII(v string) bool {
	s := strings.TrimSpace(v)
	if s == "" {
		return false
	}
	return strings.Contains(s, "**") || strings.Contains(s, "••")
}

// FieldMaskReport is the per-order verdict emitted by
// DetectMaskedFieldsInOrderList.
//
// AnyMasked is true when at least one PII-carrying field on this order is
// masked. MaskedFields lists the JSON field names (snake_case, as returned by
// Lazada) that were flagged. Consumers surface these to the FE so it can pick
// per-field UI treatment.
type FieldMaskReport struct {
	AnyMasked    bool     `json:"any_masked"`
	MaskedFields []string `json:"masked_fields"`
}

// OrderListMaskReport aggregates per-order masking verdicts for a
// GetOrders response.
type OrderListMaskReport struct {
	PerOrder map[string]FieldMaskReport `json:"per_order"`
}

// HasAnyMasked returns true if any order in the response has at least one
// masked PII field.
func (r OrderListMaskReport) HasAnyMasked() bool {
	for _, v := range r.PerOrder {
		if v.AnyMasked {
			return true
		}
	}
	return false
}

// DetectMaskedFieldsInOrderList walks a Lazada GetOrders response and reports,
// per order, which PII fields were returned masked. Safe on nil input.
//
// Only fields that Lazada explicitly masks per docId=145548 are checked:
// buyer / customer name. Address and phone are not present on the list
// response (they live on the detail response); when that struct is added, this
// function should be mirrored by DetectMaskedFieldsInOrderDetail.
func DetectMaskedFieldsInOrderList(resp *OrderListResponse) OrderListMaskReport {
	report := OrderListMaskReport{PerOrder: map[string]FieldMaskReport{}}
	if resp == nil {
		return report
	}
	for _, o := range resp.Data.Orders {
		fields := []string{}
		if IsMaskedPII(o.CustomerName) {
			fields = append(fields, "customer_first_name")
		}
		// If additional PII fields are added to OrderListResponse in future
		// (buyer_last_name, phone, address_line, ...), append their JSON names
		// here. Keep this function surgical — no other side effects.
		report.PerOrder[o.OrderID] = FieldMaskReport{
			AnyMasked:    len(fields) > 0,
			MaskedFields: fields,
		}
	}
	return report
}
