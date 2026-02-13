import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { downloadOrderLabel } from "./labelDownload";

describe("downloadOrderLabel", () => {
  const originalOpen = window.open;
  const originalCreateObjectURL = URL.createObjectURL;
  const originalRevokeObjectURL = URL.revokeObjectURL;

  beforeEach(() => {
    window.open = vi.fn();
    URL.createObjectURL = vi.fn(() => "blob:test");
    URL.revokeObjectURL = vi.fn();
  });

  afterEach(() => {
    window.open = originalOpen;
    URL.createObjectURL = originalCreateObjectURL;
    URL.revokeObjectURL = originalRevokeObjectURL;
    vi.restoreAllMocks();
  });

  it("opens http/https labels in new tab", () => {
    downloadOrderLabel("https://example.com/label.pdf", "ORDER-1");

    expect(window.open).toHaveBeenCalledWith(
      "https://example.com/label.pdf",
      "_blank",
      "noopener,noreferrer",
    );
  });

  it("downloads base64 labels as pdf", () => {
    const createElementSpy = vi.spyOn(document, "createElement");
    const appendSpy = vi
      .spyOn(document.body, "appendChild")
      .mockImplementation((node) => node);
    const removeSpy = vi
      .spyOn(document.body, "removeChild")
      .mockImplementation((node) => node);
    const clickSpy = vi.fn();

    createElementSpy.mockReturnValue({
      set href(_value: string) {},
      set download(_value: string) {},
      click: clickSpy,
    } as unknown as HTMLAnchorElement);

    downloadOrderLabel(btoa("pdf-data"), "ORDER-2");

    expect(URL.createObjectURL).toHaveBeenCalled();
    expect(appendSpy).toHaveBeenCalled();
    expect(clickSpy).toHaveBeenCalled();
    expect(removeSpy).toHaveBeenCalled();
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:test");
  });

  it("throws when file data is empty", () => {
    expect(() => downloadOrderLabel("", "ORDER-3")).toThrow(
      "Empty label data for order ORDER-3",
    );
  });
});
