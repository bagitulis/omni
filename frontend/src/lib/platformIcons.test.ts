import { describe, it, expect } from "vitest";
import {
  getPlatformIcon,
  getPlatformIconEmoji,
  getPlatformLabel,
} from "./platformIcons";

describe("getPlatformIcon", () => {
  it("returns shopee emoji for shopee", () => {
    expect(getPlatformIcon("shopee")).toBe("🛍️");
  });

  it("returns lazada emoji for lazada", () => {
    expect(getPlatformIcon("lazada")).toBe("📦");
  });

  it("returns tiktok emoji for tiktok", () => {
    expect(getPlatformIcon("tiktok")).toBe("🎵");
  });

  it("is case-insensitive — handles SHOPEE uppercase", () => {
    expect(getPlatformIcon("SHOPEE")).toBe("🛍️");
  });

  it("is case-insensitive — handles Tiktok mixed-case", () => {
    expect(getPlatformIcon("TikTok")).toBe("🎵");
  });

  it("returns default cart emoji for unknown platform", () => {
    expect(getPlatformIcon("unknown")).toBe("🛒");
  });

  it("returns default cart emoji for empty string", () => {
    expect(getPlatformIcon("")).toBe("🛒");
  });

  it("returns default cart emoji for undefined", () => {
    expect(getPlatformIcon(undefined)).toBe("🛒");
  });
});

describe("getPlatformIconEmoji", () => {
  it("delegates to getPlatformIcon — returns shopee emoji", () => {
    expect(getPlatformIconEmoji("shopee")).toBe("🛍️");
  });

  it("delegates to getPlatformIcon — returns default for unknown", () => {
    expect(getPlatformIconEmoji("unknown")).toBe("🛒");
  });

  it("handles undefined", () => {
    expect(getPlatformIconEmoji(undefined)).toBe("🛒");
  });
});

describe("getPlatformLabel", () => {
  it("capitalizes shopee", () => {
    expect(getPlatformLabel("shopee")).toBe("Shopee");
  });

  it("capitalizes lazada", () => {
    expect(getPlatformLabel("lazada")).toBe("Lazada");
  });

  it("capitalizes tiktok", () => {
    expect(getPlatformLabel("tiktok")).toBe("Tiktok");
  });

  it("returns Unknown for undefined", () => {
    expect(getPlatformLabel(undefined)).toBe("Unknown");
  });

  it("returns Unknown for empty string", () => {
    expect(getPlatformLabel("")).toBe("Unknown");
  });

  it("preserves already-capitalized platform", () => {
    expect(getPlatformLabel("Shopee")).toBe("Shopee");
  });
});
