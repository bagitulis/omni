package tiktok

import "strings"

// Package tiktok — soft-deprecation policy helpers (Phase 7).
//
// TikTok Shop Partner Center announced two soft-deprecations that touch this
// codebase without breaking today. This file centralises the policy so
// callers do not scatter region checks and version bumps.
//
// Reference:
//   docs/integrations/tiktok-api-drift-audit.md (Phase 6)
//   partner.tiktokshop.com changelog pages 4260w8mg, m5p1u1fo (handover_method)
//   partner.tiktokshop.com POST /product/202502/products/search

// seaRegions is the set of SEA markets where TikTok Shop has flagged
// handover_method as no-op for cross-border sellers. Keep as a private set
// so callers use ShouldSendHandoverMethod instead of open-coding the list.
var seaRegions = map[string]struct{}{
	"ID": {}, "MY": {}, "TH": {}, "VN": {}, "PH": {}, "SG": {},
}

// ShouldSendHandoverMethod returns true when the caller should include the
// handover_method field in a Ship Package / Batch Ship Package request.
//
// Rules:
//   - SEA cross-border (region ∈ SEA AND isCrossBorder == true) → false.
//     Passing the field is a no-op per TikTok changelog; drop it.
//   - Everything else → true. Includes SEA local, all non-SEA markets, and
//     the "unknown region" default (err on the side of sending — the field
//     is optional and TikTok ignores it silently rather than 4xx-ing).
//
// Region matching is case-insensitive.
func ShouldSendHandoverMethod(region string, isCrossBorder bool) bool {
	if !isCrossBorder {
		return true
	}
	r := strings.ToUpper(strings.TrimSpace(region))
	if r == "" {
		return true
	}
	_, isSEA := seaRegions[r]
	return !isSEA
}

// BuildShipPackageRequest constructs a ShipPackageRequest with the
// handover_method deprecation policy applied.
//
// Callers pass their seller region + a cross-border flag; the returned
// request drops handover_method automatically when it would be a no-op.
// PickupSlot and SelfShipment are always passed through — the deprecation
// only affects the handover_method key.
func BuildShipPackageRequest(
	region string,
	isCrossBorder bool,
	handoverMethod string,
	pickupSlot *PickupSlotInfo,
	selfShipment *SelfShipmentInfo,
) *ShipPackageRequest {
	req := &ShipPackageRequest{
		PickupSlot:   pickupSlot,
		SelfShipment: selfShipment,
	}
	if ShouldSendHandoverMethod(region, isCrossBorder) {
		req.HandoverMethod = handoverMethod
	}
	return req
}
