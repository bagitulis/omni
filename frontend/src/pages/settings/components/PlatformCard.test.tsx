import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { isPartnerKeyExpiring } from "./partnerKeyExpiry";

// Phase 8 — helper drives the "Rotate soon" banner. Frozen clock so results
// are deterministic. Mirrors backend pkg/shopee/expiry_test.go coverage.
const FIXED_NOW = new Date("2026-11-01T00:00:00Z").getTime();

describe("isPartnerKeyExpiring", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(FIXED_NOW);
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("returns not-expiring for undefined ISO", () => {
    const got = isPartnerKeyExpiring(undefined);
    expect(got).toEqual({ expired: false, expiring: false, days_left: 0 });
  });

  it("returns not-expiring for malformed ISO", () => {
    expect(isPartnerKeyExpiring("not-a-date")).toEqual({
      expired: false,
      expiring: false,
      days_left: 0,
    });
  });

  it("flags a key expiring in ≤ 7 days as expiring but not expired", () => {
    // 5 days ahead → within default 7-day window
    const iso = new Date(FIXED_NOW + 5 * 86_400_000).toISOString();
    const got = isPartnerKeyExpiring(iso);
    expect(got.expired).toBe(false);
    expect(got.expiring).toBe(true);
    expect(got.days_left).toBe(5);
  });

  it("flags an already-expired key as expired (days_left = 0)", () => {
    const iso = new Date(FIXED_NOW - 86_400_000).toISOString(); // yesterday
    const got = isPartnerKeyExpiring(iso);
    expect(got.expired).toBe(true);
    expect(got.expiring).toBe(true);
    expect(got.days_left).toBe(0);
  });

  it("returns not-expiring for a key 30 days in the future", () => {
    const iso = new Date(FIXED_NOW + 30 * 86_400_000).toISOString();
    const got = isPartnerKeyExpiring(iso);
    expect(got.expired).toBe(false);
    expect(got.expiring).toBe(false);
    expect(got.days_left).toBe(30);
  });

  it("honours a custom warning window", () => {
    const iso = new Date(FIXED_NOW + 10 * 86_400_000).toISOString();
    expect(isPartnerKeyExpiring(iso, 7).expiring).toBe(false); // outside window
    expect(isPartnerKeyExpiring(iso, 30).expiring).toBe(true); // inside window
  });

  it("clamps a negative window to zero (defensive)", () => {
    const iso = new Date(FIXED_NOW + 1 * 86_400_000).toISOString();
    // With window = -5 (garbage), only past dates should flag as expiring.
    expect(isPartnerKeyExpiring(iso, -5).expiring).toBe(false);
  });
});
