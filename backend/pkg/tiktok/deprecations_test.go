package tiktok

import (
	"testing"
)

// Package tiktok — deprecation helpers (Phase 7).
//
// Phase 6 audit flagged two soft-deprecations on TikTok Shop Partner Center:
//   1. `/product/202309/products/search` superseded by
//      `/product/202502/products/search` (POST). We still keep GetProducts
//      alive to avoid breaking any external caller of pkg/tiktok, but the
//      canonical documented path is now the newer version.
//   2. `handover_method` in Ship Package + Batch Ship Package is a no-op for
//      SEA cross-border sellers (per changelog page 4260w8mg / d4fd... Jul-Aug
//      2025). Passing it for those sellers wastes bytes; passing it for
//      local-fulfilment sellers is still meaningful.
//
// This file houses the helpers that make the deprecation policy testable.

// TestShouldSendHandoverMethod_LocalSEA — local (non-cross-border) SEA
// sellers still need handover_method. Return true.
func TestShouldSendHandoverMethod_LocalSEA(t *testing.T) {
	cases := []struct {
		region string
		xborder bool
	}{
		{"ID", false},
		{"MY", false},
		{"TH", false},
		{"VN", false},
		{"PH", false},
		{"SG", false},
	}
	for _, tc := range cases {
		if !ShouldSendHandoverMethod(tc.region, tc.xborder) {
			t.Errorf("region=%q xborder=%v: want true, got false", tc.region, tc.xborder)
		}
	}
}

// TestShouldSendHandoverMethod_CrossBorderSEA_IsNoop — the exact deprecation
// case. Cross-border sellers targeting SEA markets should NOT send the field.
func TestShouldSendHandoverMethod_CrossBorderSEA_IsNoop(t *testing.T) {
	cases := []string{"ID", "MY", "TH", "VN", "PH", "SG"}
	for _, region := range cases {
		if ShouldSendHandoverMethod(region, true) {
			t.Errorf("region=%q xborder=true: want false (SEA cross-border no-op), got true", region)
		}
	}
}

// TestShouldSendHandoverMethod_NonSEA — regions outside SEA (US, UK, EU, MX,
// BR, JP) are unaffected by the deprecation; keep sending.
func TestShouldSendHandoverMethod_NonSEA(t *testing.T) {
	cases := []struct {
		region string
		xborder bool
	}{
		{"US", true},
		{"US", false},
		{"UK", true},
		{"GB", true},
		{"DE", true},
		{"FR", true},
		{"BR", false},
		{"MX", false},
		{"JP", false},
	}
	for _, tc := range cases {
		if !ShouldSendHandoverMethod(tc.region, tc.xborder) {
			t.Errorf("region=%q xborder=%v: want true (non-SEA), got false", tc.region, tc.xborder)
		}
	}
}

// TestShouldSendHandoverMethod_EmptyRegion_DefaultsTrue — negative path:
// when the caller does not know the region, err on the side of sending
// (previous behaviour). A missing-region log/metric can be added later; the
// helper itself must not silently drop the field.
func TestShouldSendHandoverMethod_EmptyRegion_DefaultsTrue(t *testing.T) {
	if !ShouldSendHandoverMethod("", false) {
		t.Errorf(`region="" xborder=false: want true (default keep sending), got false`)
	}
	if !ShouldSendHandoverMethod("", true) {
		t.Errorf(`region="" xborder=true: want true (default keep sending), got false`)
	}
}

// TestShouldSendHandoverMethod_CaseInsensitive — region codes arrive from
// several sources (config, JWT claim, order payload). Accept mixed case.
func TestShouldSendHandoverMethod_CaseInsensitive(t *testing.T) {
	if ShouldSendHandoverMethod("id", true) {
		t.Errorf(`lower-case "id" xborder=true: want false, got true`)
	}
	if ShouldSendHandoverMethod("Id", true) {
		t.Errorf(`mixed-case "Id" xborder=true: want false, got true`)
	}
}

// TestBuildShipPackageRequest_DropsHandoverMethodForCrossBorderSEA — end-to-end
// test through the request builder. The builder is what production callers
// use; the underlying helper is an implementation detail.
func TestBuildShipPackageRequest_DropsHandoverMethodForCrossBorderSEA(t *testing.T) {
	got := BuildShipPackageRequest("ID", true, "PICKUP", nil, nil)
	if got.HandoverMethod != "" {
		t.Errorf("SEA cross-border: HandoverMethod = %q, want empty", got.HandoverMethod)
	}
}

// TestBuildShipPackageRequest_KeepsHandoverMethodForLocalSeller — the happy
// path stays intact.
func TestBuildShipPackageRequest_KeepsHandoverMethodForLocalSeller(t *testing.T) {
	got := BuildShipPackageRequest("ID", false, "PICKUP", nil, nil)
	if got.HandoverMethod != "PICKUP" {
		t.Errorf("SEA local: HandoverMethod = %q, want PICKUP", got.HandoverMethod)
	}
}

// TestBuildShipPackageRequest_PassesThroughOtherFields — negative-path guard
// that the region check does not accidentally drop pickup slot / self
// shipment info.
func TestBuildShipPackageRequest_PassesThroughOtherFields(t *testing.T) {
	slot := &PickupSlotInfo{StartTime: 100, EndTime: 200}
	self := &SelfShipmentInfo{TrackingNumber: "T1", ShippingProviderID: "SP1"}
	got := BuildShipPackageRequest("ID", true, "PICKUP", slot, self)
	if got.PickupSlot == nil || got.PickupSlot.StartTime != 100 {
		t.Errorf("PickupSlot dropped: %+v", got.PickupSlot)
	}
	if got.SelfShipment == nil || got.SelfShipment.TrackingNumber != "T1" {
		t.Errorf("SelfShipment dropped: %+v", got.SelfShipment)
	}
}
