import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CredentialPlatformSummary } from "@/api/credentials";
import { useAuthStore } from "@/stores/authStore";
import PlatformsTab from "./PlatformsTab";

const getCredentialPlatforms = vi.fn();
const getCredentialAudit = vi.fn();
const saveManualToken = vi.fn();

vi.mock("@/api/credentials", async () => {
  const actual = await vi.importActual<typeof import("@/api/credentials")>("@/api/credentials");
  return {
    ...actual,
    getCredentialPlatforms: () => getCredentialPlatforms(),
    getCredentialAudit: (platform: string) => getCredentialAudit(platform),
    saveManualToken: (platform: string, payload: unknown) => saveManualToken(platform, payload),
  };
});

vi.mock("@/components/AntStaticApi", () => ({
  message: {
    error: vi.fn(),
    info: vi.fn(),
    success: vi.fn(),
  },
}));

const basePlatform = (overrides: Partial<CredentialPlatformSummary>): CredentialPlatformSummary => ({
  platform: "shopee",
  region: "id",
  status: "connected",
  app_configured: true,
  secret_mask: "configured",
  stores: [
    {
      store_identifier: "STORE-001",
      store_name: "[TEST DATA] Store",
      status: "connected",
      expires_at: "2026-06-01T00:00:00Z",
      last_refresh_at: "2026-05-26T00:00:00Z",
    },
  ],
  ...overrides,
});

const statusPlatforms: CredentialPlatformSummary[] = [
  basePlatform({ platform: "shopee", status: "disconnected", stores: [] }),
  basePlatform({ platform: "lazada", status: "incomplete", stores: [], secret_mask: "configured" }),
  basePlatform({ platform: "tiktok", status: "connected" }),
  basePlatform({ platform: "expired", status: "expired", stores: [] }),
  basePlatform({ platform: "refresh", status: "refresh_failed", stores: [] }),
  basePlatform({ platform: "action", status: "action_required", stores: [] }),
];

describe("PlatformsTab credential management", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    getCredentialAudit.mockResolvedValue([]);
    getCredentialPlatforms.mockResolvedValue(statusPlatforms);
    useAuthStore.setState({ user: { id: "1", username: "Dev", email: "dev@example.com", role: "developer" } });
  });

  it("separates store connections from app credentials and labels all credential states", async () => {
    render(<PlatformsTab />);

    expect(await screen.findByText("Store Connections")).toBeInTheDocument();
    expect(screen.getByText("App Credentials")).toBeInTheDocument();
    expect(screen.getAllByText("Disconnected").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Incomplete").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Connected").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Expired").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Refresh Failed").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Action Required").length).toBeGreaterThan(0);
  });

  it("shows Lazada as Indonesia while preserving masked secret display", async () => {
    render(<PlatformsTab />);

    await screen.findByText("Lazada App");
    expect(screen.getAllByText("Indonesia").length).toBeGreaterThan(0);
    expect(screen.getAllByText("configured").length).toBeGreaterThan(0);
    expect(screen.queryByText("raw-secret-value")).not.toBeInTheDocument();
  });

  it("hides manual token actions from tenant admins", async () => {
    useAuthStore.setState({ user: { id: "2", username: "Admin", email: "admin@example.com", role: "owner" } });

    render(<PlatformsTab />);

    await screen.findByText("App Credentials");
    expect(screen.queryByRole("button", { name: /manual token/i })).not.toBeInTheDocument();
  });

  it("opens developer manual token drawer with write-only token fields", async () => {
    const user = userEvent.setup();
    render(<PlatformsTab />);

    const manualButtons = await screen.findAllByRole("button", { name: /manual token/i });
    await user.click(manualButtons[0]);

    expect(screen.getByText(/Manual Token - Shopee/i)).toBeInTheDocument();
    expect(screen.getByLabelText("Access Token")).toHaveAttribute("type", "password");
    expect(screen.getByLabelText("Refresh Token")).toHaveAttribute("type", "password");
    expect(screen.getByLabelText("Shop Cipher")).toHaveAttribute("type", "password");
  });

  it("loads OAuth and refresh history in the history drawer", async () => {
    getCredentialAudit.mockResolvedValueOnce([
      {
        event_type: "refresh_failed",
        platform: "shopee",
        store_identifier_mask: "***001",
        code: "token_expired",
        created_at: "2026-05-26T00:00:00Z",
      },
    ]);
    const user = userEvent.setup();

    render(<PlatformsTab />);
    const historyButtons = await screen.findAllByRole("button", { name: /^history$/i });
    await user.click(historyButtons[0]);

    await waitFor(() => expect(getCredentialAudit).toHaveBeenCalledWith("tiktok"));
    expect(await screen.findByText("refresh_failed")).toBeInTheDocument();
    expect(screen.getByText("token_expired")).toBeInTheDocument();
  });

  it("renders responsive columns for narrow layouts", async () => {
    window.innerWidth = 320;
    render(<PlatformsTab />);

    expect(await screen.findByText("Credential Management")).toBeInTheDocument();
    expect(screen.getByText("Store Connections")).toBeInTheDocument();
  });
});
