import { describe, expect, it } from "vitest";
import { neutral, platform, primary, status } from "./colors";

describe("colors", () => {
  describe("primary", () => {
    it("has DEFAULT value", () => {
      expect(primary.DEFAULT).toBe("#0369a1");
    });
    it("has hover value", () => {
      expect(primary.hover).toBe("#0284c7");
    });
    it("has active value", () => {
      expect(primary.active).toBe("#075985");
    });
    it("has light value", () => {
      expect(primary.light).toBe("#e0f2fe");
    });
    it("has lighter value", () => {
      expect(primary.lighter).toBe("#f0f9ff");
    });
  });

  describe("neutral", () => {
    it("has white", () => {
      expect(neutral.white).toBe("#ffffff");
    });
    it("has background", () => {
      expect(neutral.background).toBe("#f8fafc");
    });
    it("has surface", () => {
      expect(neutral.surface).toBe("#ffffff");
    });
    it("has border", () => {
      expect(neutral.border).toBe("#e2e8f0");
    });
    it("has textMuted", () => {
      expect(neutral.textMuted).toBe("#64748b");
    });
    it("has text", () => {
      expect(neutral.text).toBe("#334155");
    });
    it("has textStrong", () => {
      expect(neutral.textStrong).toBe("#0f172a");
    });
  });

  describe("status", () => {
    it("success has DEFAULT, light, dark", () => {
      expect(status.success.DEFAULT).toBe("#16a34a");
      expect(status.success.light).toBe("#dcfce7");
      expect(status.success.dark).toBe("#15803d");
    });
    it("warning has DEFAULT, light, dark", () => {
      expect(status.warning.DEFAULT).toBe("#d97706");
      expect(status.warning.light).toBe("#fef3c7");
      expect(status.warning.dark).toBe("#b45309");
    });
    it("error has DEFAULT, light, dark", () => {
      expect(status.error.DEFAULT).toBe("#dc2626");
      expect(status.error.light).toBe("#fee2e2");
      expect(status.error.dark).toBe("#b91c1c");
    });
    it("info has DEFAULT, light, dark", () => {
      expect(status.info.DEFAULT).toBe("#0369a1");
      expect(status.info.light).toBe("#e0f2fe");
      expect(status.info.dark).toBe("#075985");
    });
  });

  describe("platform", () => {
    it("shopee color", () => {
      expect(platform.shopee).toBe("#ee4d2d");
    });
    it("tiktok color", () => {
      expect(platform.tiktok).toBe("#000000");
    });
    it("lazada color", () => {
      expect(platform.lazada).toBe("#0f146d");
    });
  });
});
