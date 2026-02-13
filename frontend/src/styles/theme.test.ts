import { describe, expect, it } from "vitest";
import { antdDarkTheme, antdTheme } from "./theme";

describe("theme tooltip contrast", () => {
  it("uses readable tooltip colors in light theme", () => {
    expect(antdTheme.components?.Tooltip).toMatchObject({
      colorBgSpotlight: "#f1f5f9",
      colorTextLightSolid: "#334155",
    });
  });

  it("uses readable tooltip colors in dark theme", () => {
    expect(antdDarkTheme.components?.Tooltip).toMatchObject({
      colorBgSpotlight: "#1E293B",
      colorTextLightSolid: "#E2E8F0",
    });
  });
});
