import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, type Mock } from "vitest";
import SettingsPage from "./SettingsPage";

// Mock dependencies
vi.mock("react-router-dom", () => ({
  useSearchParams: vi.fn(() => [new URLSearchParams("tab=general"), vi.fn()]),
}));

vi.mock("./tabs/GeneralTab", () => ({
  default: () => <div data-testid="general-tab">General Tab</div>,
}));

vi.mock("./tabs/PlatformsTab", () => ({
  default: () => <div data-testid="platforms-tab">Platforms Tab</div>,
}));

vi.mock("./tabs/WebhooksTab", () => ({
  default: () => <div data-testid="webhooks-tab">Webhooks Tab</div>,
}));

vi.mock("./tabs/AccountTab", () => ({
  default: () => <div data-testid="account-tab">Account Tab</div>,
}));

vi.mock("./tabs/GoogleSheetsTab", () => ({
  default: () => <div data-testid="google-sheets-tab">Google Sheets Tab</div>,
}));

// Mock Ant Design
vi.mock("antd", async () => {
  const actual = await vi.importActual("antd");
  return {
    ...actual,
    theme: {
      useToken: () => ({
        token: { colorBgContainer: "#fff", borderRadius: 4 },
      }),
    },
  };
});

describe("SettingsPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Mock matchMedia
    Object.defineProperty(window, "matchMedia", {
      writable: true,
      value: vi.fn().mockImplementation((query) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
    });
  });

  it("renders page title", () => {
    render(<SettingsPage />);
    expect(screen.getByText("Settings")).toBeInTheDocument();
  });

  it("renders default tab content", () => {
    render(<SettingsPage />);
    expect(screen.getByTestId("general-tab")).toBeInTheDocument();
  });

  it("renders other tabs on navigation", async () => {
    // Since we mocked useSearchParams, we can't easily simulate navigation in the same render
    // without a more complex mock. But we can verify all tabs are present in the DOM (Antd Tabs usually render tabs in DOM, just hidden)
    // or just check if tab headers are there.

    render(<SettingsPage />);
    expect(screen.getByText("General")).toBeInTheDocument();
    expect(screen.getByText("Platforms")).toBeInTheDocument();
    expect(screen.getByText("Webhooks")).toBeInTheDocument();
    expect(screen.getByText("Google Sheets")).toBeInTheDocument();
    expect(screen.getByText("Account")).toBeInTheDocument();
  });

  it("changes URL param on tab click", async () => {
    const { useSearchParams } = await import("react-router-dom");
    const setSearchParams = vi.fn();
    (useSearchParams as unknown as Mock).mockReturnValue([
      new URLSearchParams("tab=general"),
      setSearchParams,
    ]);

    render(<SettingsPage />);
    // Click on Platforms tab
    // Note: Finding the exact click target in Antd Tabs can be tricky in JSDOM.
    // Usually finding by text works.
    fireEvent.click(screen.getByText("Platforms"));

    expect(setSearchParams).toHaveBeenCalledWith({ tab: "platforms" });
  });

  it("renders correct tab based on URL param", async () => {
    const { useSearchParams } = await import("react-router-dom");
    (useSearchParams as unknown as Mock).mockReturnValue([
      new URLSearchParams("tab=platforms"),
      vi.fn(),
    ]);

    render(<SettingsPage />);
    expect(screen.getByTestId("platforms-tab")).toBeInTheDocument();
  });
});
