import { describe, it, expect } from "vitest";
import { PLATFORM_BRAND_COLORS } from "./platformColors";

describe("PLATFORM_BRAND_COLORS", () => {
  it("contains shopee brand color", () => {
    expect(PLATFORM_BRAND_COLORS.shopee).toBe("#ee4d2d");
  });

  it("contains lazada brand color", () => {
    expect(PLATFORM_BRAND_COLORS.lazada).toBe("#0f146d");
  });

  it("contains tiktok brand color", () => {
    expect(PLATFORM_BRAND_COLORS.tiktok).toBe("#000000");
  });

  it("has exactly 3 entries", () => {
    expect(Object.keys(PLATFORM_BRAND_COLORS)).toHaveLength(3);
  });

  it("all values are valid hex color strings", () => {
    const hexColorPattern = /^#[0-9a-fA-F]{6}$/;
    for (const [platform, color] of Object.entries(PLATFORM_BRAND_COLORS)) {
      expect(color, `${platform} color should be a valid 6-digit hex`).toMatch(
        hexColorPattern,
      );
    }
  });

  it("is typed as Record<string, string>", () => {
    // Verify runtime shape — all values are strings
    for (const value of Object.values(PLATFORM_BRAND_COLORS)) {
      expect(typeof value).toBe("string");
    }
  });
});
